// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {DigitalSecurityToken} from "../../src/tokens/DigitalSecurityToken.sol";
import {TokenFactory} from "../../src/tokens/TokenFactory.sol";
import {IDigitalSecurityToken} from "../../src/interfaces/tokens/IDigitalSecurityToken.sol";
import {IIdentityRegistry} from "../../src/interfaces/IIdentityRegistry.sol";

contract MockIdentityRegistry is IIdentityRegistry {
    mapping(address => bool) private _verified;

    function setVerified(address user, bool status) external {
        _verified[user] = status;
    }

    function isVerified(address userAddress) external view override returns (bool) {
        return _verified[userAddress];
    }

    function registerIdentity(address, bytes32, uint16, uint8) external override {}

    function batchRegisterIdentity(
        address[] calldata,
        bytes32[] calldata,
        uint16[] calldata,
        uint8[] calldata
    ) external override {}

    function deleteIdentity(address) external override {}

    function updateSanctionStatus(bytes32, bool) external override {}

    function updateCountry(address, uint16) external override {}

    function updateKycTier(address, uint8) external override {}

    function getInvestorClaim(address) external pure override returns (InvestorClaim memory) {
        return InvestorClaim(bytes32(0), 0, 0, false, 0);
    }

    function getInvestorCountry(address) external pure override returns (uint16) {
        return 0;
    }

    function getIdentityId(address) external pure override returns (bytes32) {
        return bytes32(0);
    }

    function contains(address) external pure override returns (bool) {
        return false;
    }

    function isSanctioned(bytes32) external pure override returns (bool) {
        return false;
    }

    function totalIdentities() external pure override returns (uint256) {
        return 0;
    }
}

contract DigitalSecurityTokenTest is Test {
    DigitalSecurityToken public implementation;
    DigitalSecurityToken public token;
    MockIdentityRegistry public registry;
    TokenFactory public factory;

    address public admin = address(0xAD01);
    address public minter = address(0xA001);
    address public burner = address(0xB001);
    address public investor1 = address(0x1001);
    address public investor2 = address(0x1002);
    address public stranger = address(0x9999);

    string public constant ISIN = "INE002A01018";
    bytes32 public constant BATCH_ID = keccak256("BATCH_001");
    bytes32 public constant PROOF_HASH = keccak256("PROOF_001");

    function setUp() public {
        registry = new MockIdentityRegistry();
        implementation = new DigitalSecurityToken();

        bytes memory initData = abi.encodeWithSelector(
            DigitalSecurityToken.initialize.selector,
            "Growww Reliance Industries",
            "GROWWW-RELIANCE",
            18,
            ISIN,
            address(registry),
            address(0),
            admin
        );

        ERC1967Proxy proxy = new ERC1967Proxy(address(implementation), initData);
        token = DigitalSecurityToken(address(proxy));

        bytes32 minterRole = token.MINTER_ROLE();
        bytes32 burnerRole = token.BURNER_ROLE();

        vm.startPrank(admin);
        token.grantRole(minterRole, minter);
        token.grantRole(burnerRole, burner);
        vm.stopPrank();

        registry.setVerified(investor1, true);
        registry.setVerified(investor2, true);
    }

    function test_Initialization() public {
        assertEq(token.name(), "Growww Reliance Industries");
        assertEq(token.symbol(), "GROWWW-RELIANCE");
        assertEq(token.decimals(), 18);
        assertEq(token.isin(), ISIN);
        assertEq(token.isinHash(), keccak256(bytes(ISIN)));
        assertTrue(token.hasRole(token.DEFAULT_ADMIN_ROLE(), admin));
        assertTrue(token.hasRole(token.MINTER_ROLE(), admin));
        assertTrue(token.hasRole(token.MINTER_ROLE(), minter));
        assertEq(token.totalSupply(), 0);
    }

    function test_InitializeRevertsOnZeroAdmin() public {
        bytes memory initData = abi.encodeWithSelector(
            DigitalSecurityToken.initialize.selector,
            "Token",
            "TKN",
            18,
            ISIN,
            address(registry),
            address(0),
            address(0)
        );
        vm.expectRevert(IDigitalSecurityToken.InvalidZeroAddress.selector);
        new ERC1967Proxy(address(implementation), initData);
    }

    function test_MintSucceeds() public {
        vm.prank(minter);
        token.mint(investor1, 1000e18, BATCH_ID, PROOF_HASH);

        assertEq(token.balanceOf(investor1), 1000e18);
        assertEq(token.totalSupply(), 1000e18);
    }

    function test_MintRevertsWithoutRole() public {
        vm.prank(stranger);
        vm.expectRevert();
        token.mint(investor1, 1000e18, BATCH_ID, PROOF_HASH);
    }

    function test_MintRevertsForNonVerifiedRecipient() public {
        vm.prank(minter);
        vm.expectRevert(abi.encodeWithSelector(IDigitalSecurityToken.RecipientNotCompliant.selector, stranger));
        token.mint(stranger, 1000e18, BATCH_ID, PROOF_HASH);
    }

    function test_MintRevertsForZeroAmount() public {
        vm.prank(minter);
        vm.expectRevert(IDigitalSecurityToken.InvalidMintAmount.selector);
        token.mint(investor1, 0, BATCH_ID, PROOF_HASH);
    }

    function test_MintRevertsForFrozenRecipient() public {
        vm.prank(admin);
        token.freezeAddress(investor1);

        vm.prank(minter);
        vm.expectRevert(abi.encodeWithSelector(IDigitalSecurityToken.AddressIsFrozen.selector, investor1));
        token.mint(investor1, 1000e18, BATCH_ID, PROOF_HASH);
    }

    function test_BatchMintSucceeds() public {
        address[] memory recipients = new address[](2);
        recipients[0] = investor1;
        recipients[1] = investor2;

        uint256[] memory amounts = new uint256[](2);
        amounts[0] = 500e18;
        amounts[1] = 750e18;

        vm.prank(minter);
        token.batchMint(recipients, amounts, BATCH_ID, PROOF_HASH);

        assertEq(token.balanceOf(investor1), 500e18);
        assertEq(token.balanceOf(investor2), 750e18);
        assertEq(token.totalSupply(), 1250e18);
    }

    function test_BatchMintRevertsOnArrayLengthMismatch() public {
        address[] memory recipients = new address[](2);
        recipients[0] = investor1;
        recipients[1] = investor2;

        uint256[] memory amounts = new uint256[](1);
        amounts[0] = 500e18;

        vm.prank(minter);
        vm.expectRevert(
            abi.encodeWithSelector(IDigitalSecurityToken.ArrayLengthMismatch.selector, 2, 1)
        );
        token.batchMint(recipients, amounts, BATCH_ID, PROOF_HASH);
    }

    function test_BatchMintRevertsOnZeroAmount() public {
        address[] memory recipients = new address[](1);
        recipients[0] = investor1;

        uint256[] memory amounts = new uint256[](1);
        amounts[0] = 0;

        vm.prank(minter);
        vm.expectRevert(IDigitalSecurityToken.InvalidMintAmount.selector);
        token.batchMint(recipients, amounts, BATCH_ID, PROOF_HASH);
    }

    function test_BurnSucceeds() public {
        vm.prank(minter);
        token.mint(investor1, 1000e18, BATCH_ID, PROOF_HASH);

        bytes32 redemptionId = keccak256("REDEMPTION_001");
        vm.prank(burner);
        token.burn(investor1, 400e18, redemptionId);

        assertEq(token.balanceOf(investor1), 600e18);
        assertEq(token.totalSupply(), 600e18);
    }

    function test_FreezeAndUnfreeze() public {
        assertFalse(token.isFrozen(investor1));

        vm.prank(admin);
        token.freezeAddress(investor1);
        assertTrue(token.isFrozen(investor1));

        vm.prank(admin);
        token.unfreezeAddress(investor1);
        assertFalse(token.isFrozen(investor1));
    }

    function test_SetIdentityRegistry() public {
        address newRegistry = address(0xBEEF);
        vm.prank(admin);
        token.setIdentityRegistry(newRegistry);
        assertEq(token.identityRegistry(), newRegistry);
    }

    function test_SetIdentityRegistryRevertsOnZeroAddress() public {
        vm.prank(admin);
        vm.expectRevert(IDigitalSecurityToken.InvalidZeroAddress.selector);
        token.setIdentityRegistry(address(0));
    }

    function test_TotalSupplyEqualsBalanceSum(uint96 amount1, uint96 amount2) public {
        vm.assume(amount1 > 0 && amount2 > 0);

        vm.prank(minter);
        token.mint(investor1, amount1, BATCH_ID, PROOF_HASH);

        vm.prank(minter);
        token.mint(investor2, amount2, BATCH_ID, PROOF_HASH);

        uint256 sumBalances = token.balanceOf(investor1) + token.balanceOf(investor2);
        assertEq(token.totalSupply(), sumBalances);
    }
}

contract TokenFactoryTest is Test {
    DigitalSecurityToken public implementation;
    TokenFactory public factory;
    MockIdentityRegistry public registry;

    address public admin = address(0xAD01);
    address public deployer = address(0xDE01);

    string public constant ISIN_A = "INE002A01018";
    string public constant ISIN_B = "INE009A01021";

    function setUp() public {
        registry = new MockIdentityRegistry();
        implementation = new DigitalSecurityToken();
        factory = new TokenFactory(address(implementation), admin);

        bytes32 deployerRole = factory.DEPLOYER_ROLE();
        vm.prank(admin);
        factory.grantRole(deployerRole, deployer);
    }

    function test_DeploySecurityToken() public {
        vm.prank(deployer);
        address proxyAddr = factory.deploySecurityToken(
            "Growww Reliance Industries",
            "GROWWW-RELIANCE",
            ISIN_A,
            address(registry),
            address(0),
            admin
        );

        assertTrue(proxyAddr != address(0));
        assertEq(factory.getTokenByISIN(ISIN_A), proxyAddr);

        DigitalSecurityToken deployed = DigitalSecurityToken(proxyAddr);
        assertEq(deployed.isin(), ISIN_A);
        assertEq(deployed.symbol(), "GROWWW-RELIANCE");
        assertTrue(deployed.hasRole(deployed.DEFAULT_ADMIN_ROLE(), admin));
    }

    function test_DeploySecurityTokenRevertsOnDuplicateISIN() public {
        vm.prank(deployer);
        factory.deploySecurityToken(
            "Growww Reliance Industries",
            "GROWWW-RELIANCE",
            ISIN_A,
            address(registry),
            address(0),
            admin
        );

        vm.prank(deployer);
        vm.expectRevert();
        factory.deploySecurityToken(
            "Growww Reliance Industries Dup",
            "GROWWW-RELIANCE-DUP",
            ISIN_A,
            address(registry),
            address(0),
            admin
        );
    }

    function test_DeploySecurityTokenRevertsWithoutRole() public {
        vm.prank(address(0x9999));
        vm.expectRevert();
        factory.deploySecurityToken(
            "Unauthorized Token",
            "UNAUTH",
            ISIN_B,
            address(registry),
            address(0),
            admin
        );
    }

    function test_GetTokenByISINReturnsZeroForUnknown() public {
        assertEq(factory.getTokenByISIN("UNKNOWN_ISIN"), address(0));
    }
}
