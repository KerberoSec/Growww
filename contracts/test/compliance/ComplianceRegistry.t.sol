// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {IdentityRegistry} from "../../src/compliance/IdentityRegistry.sol";
import {ComplianceRegistry} from "../../src/compliance/ComplianceRegistry.sol";
import {CountryRestrictModule} from "../../src/compliance/modules/CountryRestrictModule.sol";
import {MaxOwnershipModule} from "../../src/compliance/modules/MaxOwnershipModule.sol";
import {LockupModule} from "../../src/compliance/modules/LockupModule.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {DigitalSecurityToken} from "../../src/tokens/DigitalSecurityToken.sol";
import {IDigitalSecurityToken} from "../../src/interfaces/tokens/IDigitalSecurityToken.sol";
import {IIdentityRegistry} from "../../src/interfaces/IIdentityRegistry.sol";

contract ComplianceRegistryTest is Test {
    IdentityRegistry public identityRegistry;
    ComplianceRegistry public complianceRegistry;
    CountryRestrictModule public countryModule;
    MaxOwnershipModule public maxOwnershipModule;
    LockupModule public lockupModule;
    DigitalSecurityToken public token;

    address public admin = address(0xAD01);
    address public agent = address(0xA601);
    address public minter = address(0xAA01);

    address public alice = address(0x1111);
    address public bob = address(0x2222);
    address public charlie = address(0x3333);
    address public sanctionedDave = address(0x4444);

    bytes32 public aliceId = keccak256("ALICE_IDENTITY");
    bytes32 public bobId = keccak256("BOB_IDENTITY");
    bytes32 public charlieId = keccak256("CHARLIE_IDENTITY");
    bytes32 public daveId = keccak256("DAVE_IDENTITY");

    uint16 public constant INDIA = 356;
    uint16 public constant USA = 840;
    uint16 public constant RESTRICTED_COUNTRY = 999;

    function setUp() public {
        vm.startPrank(admin);

        // 1. Deploy IdentityRegistry via proxy
        IdentityRegistry identityImpl = new IdentityRegistry();
        bytes memory idInit = abi.encodeWithSelector(IdentityRegistry.initialize.selector, admin, agent);
        ERC1967Proxy idProxy = new ERC1967Proxy(address(identityImpl), idInit);
        identityRegistry = IdentityRegistry(address(idProxy));

        // 2. Deploy ComplianceRegistry via proxy
        ComplianceRegistry compImpl = new ComplianceRegistry();
        bytes memory compInit = abi.encodeWithSelector(ComplianceRegistry.initialize.selector, admin, address(identityRegistry));
        ERC1967Proxy compProxy = new ERC1967Proxy(address(compImpl), compInit);
        complianceRegistry = ComplianceRegistry(address(compProxy));

        // 3. Deploy DigitalSecurityToken via proxy
        DigitalSecurityToken tokenImpl = new DigitalSecurityToken();
        bytes memory tokenInit = abi.encodeWithSelector(
            DigitalSecurityToken.initialize.selector,
            "Growww Reliance Equity",
            "GROWWW-RELIANCE",
            18,
            "INE002A01018",
            address(identityRegistry),
            address(complianceRegistry),
            admin
        );
        ERC1967Proxy tokenProxy = new ERC1967Proxy(address(tokenImpl), tokenInit);
        token = DigitalSecurityToken(address(tokenProxy));
        token.grantRole(token.MINTER_ROLE(), minter);

        // 4. Deploy modules
        countryModule = new CountryRestrictModule(admin, address(identityRegistry));
        countryModule.setCountryStatus(INDIA, true);
        countryModule.setCountryStatus(USA, true);

        // 10% ownership cap (1000 bps)
        maxOwnershipModule = new MaxOwnershipModule(admin, address(token), 1000);

        lockupModule = new LockupModule(admin, address(token));

        // Bind modules to ComplianceRegistry
        complianceRegistry.bindModule(address(countryModule));
        complianceRegistry.bindModule(address(maxOwnershipModule));
        complianceRegistry.bindModule(address(lockupModule));

        vm.stopPrank();

        // 5. Register Identities via agent
        vm.startPrank(agent);
        identityRegistry.registerIdentity(alice, aliceId, INDIA, 2);
        identityRegistry.registerIdentity(bob, bobId, INDIA, 2);
        identityRegistry.registerIdentity(charlie, charlieId, RESTRICTED_COUNTRY, 2);
        identityRegistry.registerIdentity(sanctionedDave, daveId, INDIA, 2);
        identityRegistry.updateSanctionStatus(daveId, true);
        vm.stopPrank();
    }

    function test_IdentityRegistry_VerificationStatus() public view {
        assertTrue(identityRegistry.isVerified(alice));
        assertTrue(identityRegistry.isVerified(bob));
        assertTrue(identityRegistry.isVerified(charlie)); // KYC tier verified, but country restricted
        assertFalse(identityRegistry.isVerified(sanctionedDave)); // Sanctioned
        assertFalse(identityRegistry.isVerified(address(0x9999))); // Not registered
    }

    function test_IdentityRegistry_BatchRegistration() public {
        address[] memory users = new address[](2);
        users[0] = address(0x5555);
        users[1] = address(0x6666);

        bytes32[] memory ids = new bytes32[](2);
        ids[0] = keccak256("USER_1");
        ids[1] = keccak256("USER_2");

        uint16[] memory countries = new uint16[](2);
        countries[0] = INDIA;
        countries[1] = USA;

        uint8[] memory tiers = new uint8[](2);
        tiers[0] = 1;
        tiers[1] = 3;

        vm.prank(agent);
        identityRegistry.batchRegisterIdentity(users, ids, countries, tiers);

        assertTrue(identityRegistry.isVerified(users[0]));
        assertTrue(identityRegistry.isVerified(users[1]));
        assertEq(identityRegistry.getInvestorCountry(users[0]), INDIA);
        assertEq(identityRegistry.getInvestorCountry(users[1]), USA);
    }

    function test_IdentityRegistry_DeleteIdentity() public {
        vm.prank(agent);
        identityRegistry.deleteIdentity(alice);

        assertFalse(identityRegistry.isVerified(alice));
        assertFalse(identityRegistry.contains(alice));
    }

    function test_IdentityRegistry_UpdateCountryAndTier() public {
        vm.startPrank(agent);
        identityRegistry.updateCountry(alice, USA);
        assertEq(identityRegistry.getInvestorCountry(alice), USA);

        identityRegistry.updateKycTier(alice, 3);
        IIdentityRegistry.InvestorClaim memory claim = identityRegistry.getInvestorClaim(alice);
        assertEq(claim.kycTier, 3);
        vm.stopPrank();
    }

    function test_Compliance_CanTransferBasic() public {
        // Mint 100,000 tokens to Alice so supply is sufficient and unlocked balance exists
        vm.prank(minter);
        token.mint(alice, 100_000e18, keccak256("BATCH_BASIC"), keccak256("PROOF_BASIC"));

        // Alice and Bob are both verified Indian investors
        assertTrue(complianceRegistry.canTransfer(alice, bob, 100e18));

        // Charlie resides in RESTRICTED_COUNTRY
        assertFalse(complianceRegistry.canTransfer(alice, charlie, 100e18));

        // Dave is sanctioned
        assertFalse(complianceRegistry.canTransfer(alice, sanctionedDave, 100e18));
    }

    function test_Compliance_MaxOwnershipModule() public {
        // Mint initial supply of 10,000 tokens to Alice
        vm.prank(minter);
        token.mint(alice, 10_000e18, keccak256("BATCH_1"), keccak256("PROOF_1"));

        // 10% cap of 10,000 is 1,000 tokens
        // Transfer 500 tokens to Bob: should pass
        assertTrue(complianceRegistry.canTransfer(alice, bob, 500e18));

        // Transfer 1,001 tokens to Bob: breaches 10% cap, should fail
        assertFalse(complianceRegistry.canTransfer(alice, bob, 1001e18));

        // Exemption test
        vm.prank(admin);
        maxOwnershipModule.setExemption(bob, true);
        assertTrue(complianceRegistry.canTransfer(alice, bob, 1001e18));
    }

    function test_Compliance_LockupModule() public {
        // Mint 10,000 tokens to Alice so 10% ownership cap is 1,000
        vm.prank(minter);
        token.mint(alice, 10_000e18, keccak256("BATCH_2"), keccak256("PROOF_2"));

        // Lock 9,500 tokens for Alice for 30 days
        vm.prank(admin);
        lockupModule.lockTokens(alice, 9_500e18, uint64(block.timestamp + 30 days));

        assertEq(lockupModule.getLockedBalance(alice), 9_500e18);

        // Transferring 500 tokens should pass (unlocked = 500, within 10% cap)
        assertTrue(complianceRegistry.canTransfer(alice, bob, 500e18));

        // Transferring 501 tokens should fail (only 500 unlocked)
        assertFalse(complianceRegistry.canTransfer(alice, bob, 501e18));

        // Advance time past lockup
        vm.warp(block.timestamp + 31 days);
        assertEq(lockupModule.getLockedBalance(alice), 0);
        assertTrue(complianceRegistry.canTransfer(alice, bob, 500e18));
    }

    function test_EndToEnd_TransferComplianceVerification() public {
        // Mint 10,000 tokens to Alice
        vm.prank(minter);
        token.mint(alice, 10_000e18, keccak256("BATCH_3"), keccak256("PROOF_3"));

        // Transfer 100 tokens from Alice to Bob
        vm.prank(alice);
        token.transfer(bob, 100e18);

        assertEq(token.balanceOf(alice), 9_900e18);
        assertEq(token.balanceOf(bob), 100e18);

        // Attempt transfer to unregistered address: reverts
        address unregistered = address(0x8888);
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSelector(IDigitalSecurityToken.RecipientNotCompliant.selector, unregistered)
        );
        token.transfer(unregistered, 50e18);

        // Attempt transfer to sanctioned Dave: reverts
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSelector(IDigitalSecurityToken.RecipientNotCompliant.selector, sanctionedDave)
        );
        token.transfer(sanctionedDave, 50e18);
    }
}
