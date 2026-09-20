// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {DigitalSecurityToken} from "../../src/tokens/DigitalSecurityToken.sol";
import {TokenRedemption} from "../../src/redemption/TokenRedemption.sol";
import {ITokenRedemption} from "../../src/interfaces/ITokenRedemption.sol";
import {IIdentityRegistry} from "../../src/interfaces/IIdentityRegistry.sol";

contract MockRedemptionIdentityRegistry is IIdentityRegistry {
    mapping(address => bool) private _verified;

    function setVerified(address user, bool status) external {
        _verified[user] = status;
    }

    function isVerified(address userAddress) external view override returns (bool) {
        return _verified[userAddress];
    }

    function registerIdentity(address, bytes32, uint16, uint8) external override {}
    function batchRegisterIdentity(address[] calldata, bytes32[] calldata, uint16[] calldata, uint8[] calldata) external override {}
    function deleteIdentity(address) external override {}
    function updateSanctionStatus(bytes32, bool) external override {}
    function updateCountry(address, uint16) external override {}
    function updateKycTier(address, uint8) external override {}
    function getInvestorClaim(address) external pure override returns (InvestorClaim memory) {
        return InvestorClaim(bytes32(0), 0, 0, false, 0);
    }
    function getInvestorCountry(address) external pure override returns (uint16) { return 0; }
    function getIdentityId(address) external pure override returns (bytes32) { return bytes32(0); }
    function contains(address) external pure override returns (bool) { return false; }
    function isSanctioned(bytes32) external pure override returns (bool) { return false; }
    function totalIdentities() external pure override returns (uint256) { return 0; }
}

contract TokenRedemptionTest is Test {
    DigitalSecurityToken public token;
    TokenRedemption public redemption;
    MockRedemptionIdentityRegistry public registry;

    address public admin = address(0xAD01);
    address public relayer = address(0xBE01);
    uint256 public custodianKey;
    address public custodian;
    address public investor1 = address(0x1001);
    address public unauthorized = address(0x9999);

    string public constant ISIN = "INE002A01018";
    bytes32 public constant DEMAT_REF = keccak256("NSDL_DIS_REF_12345");
    uint64 public constant DEFAULT_TIMEOUT = 86400; // 24 hours

    function setUp() public {
        (custodian, custodianKey) = makeAddrAndKey("custodian");

        // Deploy Identity Registry
        registry = new MockRedemptionIdentityRegistry();
        registry.setVerified(investor1, true);

        // Deploy DigitalSecurityToken
        DigitalSecurityToken tokenImpl = new DigitalSecurityToken();
        bytes memory tokenInit = abi.encodeWithSelector(
            DigitalSecurityToken.initialize.selector,
            "Growww Reliance Equity",
            "GROWWW-RELIANCE",
            18,
            ISIN,
            address(registry),
            address(0),
            admin
        );
        ERC1967Proxy tokenProxy = new ERC1967Proxy(address(tokenImpl), tokenInit);
        token = DigitalSecurityToken(address(tokenProxy));

        // Deploy TokenRedemption
        TokenRedemption redImpl = new TokenRedemption();
        bytes memory redInit = abi.encodeWithSelector(
            TokenRedemption.initialize.selector,
            admin,
            relayer,
            DEFAULT_TIMEOUT
        );
        ERC1967Proxy redProxy = new ERC1967Proxy(address(redImpl), redInit);
        redemption = TokenRedemption(address(redProxy));

        // Grant BURNER_ROLE on token to redemption contract
        vm.startPrank(admin);
        token.grantRole(token.BURNER_ROLE(), address(redemption));
        token.grantRole(token.MINTER_ROLE(), admin);
        // Mint initial tokens to investor1
        token.mint(investor1, 1000 ether, keccak256("BATCH_001"), keccak256("PROOF_001"));

        // Setup redemption contract configuration
        redemption.setSupportedToken(address(token), true);
        redemption.grantRole(redemption.CUSTODIAN_ROLE(), custodian);
        vm.stopPrank();

        // Register redemption contract as verified in identity registry so it can hold escrowed tokens
        registry.setVerified(address(redemption), true);
    }

    function test_Initialization() public {
        assertTrue(redemption.hasRole(redemption.DEFAULT_ADMIN_ROLE(), admin));
        assertTrue(redemption.hasRole(redemption.RELAYER_ROLE(), relayer));
        assertTrue(redemption.hasRole(redemption.CUSTODIAN_ROLE(), custodian));
        assertTrue(redemption.supportedTokens(address(token)));
        assertEq(redemption.defaultTimeoutSeconds(), DEFAULT_TIMEOUT);
    }

    function test_RequestRedemptionEscrow() public {
        uint256 redeemAmount = 100 ether;

        vm.startPrank(investor1);
        token.approve(address(redemption), redeemAmount);
        bytes32 rId = redemption.requestRedemption(address(token), redeemAmount);
        vm.stopPrank();

        assertEq(token.balanceOf(investor1), 900 ether);
        assertEq(token.balanceOf(address(redemption)), 100 ether);

        ITokenRedemption.RedemptionRequest memory req = redemption.getRequest(rId);
        assertEq(req.investor, investor1);
        assertEq(req.tokenAddress, address(token));
        assertEq(req.amount, redeemAmount);
        assertEq(uint8(req.status), uint8(ITokenRedemption.RedemptionStatus.ESCROWED));
    }

    function test_ExecuteRedemptionWithRelayerRole() public {
        uint256 redeemAmount = 100 ether;

        vm.startPrank(investor1);
        token.approve(address(redemption), redeemAmount);
        bytes32 rId = redemption.requestRedemption(address(token), redeemAmount);
        vm.stopPrank();

        uint256 supplyBefore = token.totalSupply();

        // Relayer executes redemption without signature
        vm.prank(relayer);
        redemption.executeRedemptionWithProof(rId, DEMAT_REF, "");

        assertEq(token.balanceOf(address(redemption)), 0);
        assertEq(token.totalSupply(), supplyBefore - redeemAmount);

        ITokenRedemption.RedemptionRequest memory req = redemption.getRequest(rId);
        assertEq(uint8(req.status), uint8(ITokenRedemption.RedemptionStatus.SETTLED_BURNED));
        assertEq(req.dematDebitRef, DEMAT_REF);
    }

    function test_ExecuteRedemptionWithCustodianSignature() public {
        uint256 redeemAmount = 250 ether;

        vm.startPrank(investor1);
        token.approve(address(redemption), redeemAmount);
        bytes32 rId = redemption.requestRedemption(address(token), redeemAmount);
        vm.stopPrank();

        // Create EIP-712 signature from custodian
        bytes32 structHash = keccak256(
            abi.encode(redemption.REDEMPTION_TYPEHASH(), rId, DEMAT_REF, block.timestamp)
        );
        bytes32 domainSeparator = keccak256(
            abi.encode(
                keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
                keccak256(bytes("GrowwwRedemptionManager")),
                keccak256(bytes("1")),
                block.chainid,
                address(redemption)
            )
        );
        bytes32 digest = keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(custodianKey, digest);
        bytes memory sig = abi.encodePacked(r, s, v);

        // Caller can be anyone if valid custodian signature is provided
        vm.prank(unauthorized);
        redemption.executeRedemptionWithProof(rId, DEMAT_REF, sig);

        ITokenRedemption.RedemptionRequest memory req = redemption.getRequest(rId);
        assertEq(uint8(req.status), uint8(ITokenRedemption.RedemptionStatus.SETTLED_BURNED));
        assertEq(token.balanceOf(address(redemption)), 0);
    }

    function test_ExecuteBatchRedemption() public {
        uint256 amount1 = 50 ether;
        uint256 amount2 = 75 ether;

        vm.startPrank(investor1);
        token.approve(address(redemption), amount1 + amount2);
        bytes32 rId1 = redemption.requestRedemption(address(token), amount1);
        bytes32 rId2 = redemption.requestRedemption(address(token), amount2);
        vm.stopPrank();

        bytes32[] memory ids = new bytes32[](2);
        ids[0] = rId1;
        ids[1] = rId2;

        bytes32[] memory refs = new bytes32[](2);
        refs[0] = keccak256("DEMAT_REF_1");
        refs[1] = keccak256("DEMAT_REF_2");

        bytes[] memory sigs = new bytes[](2);
        sigs[0] = "";
        sigs[1] = "";

        vm.prank(relayer);
        redemption.executeBatchRedemption(ids, refs, sigs);

        ITokenRedemption.RedemptionRequest memory req1 = redemption.getRequest(rId1);
        ITokenRedemption.RedemptionRequest memory req2 = redemption.getRequest(rId2);

        assertEq(uint8(req1.status), uint8(ITokenRedemption.RedemptionStatus.SETTLED_BURNED));
        assertEq(uint8(req2.status), uint8(ITokenRedemption.RedemptionStatus.SETTLED_BURNED));
        assertEq(token.balanceOf(address(redemption)), 0);
    }

    function test_CancelExpiredRedemption() public {
        uint256 redeemAmount = 100 ether;

        vm.startPrank(investor1);
        token.approve(address(redemption), redeemAmount);
        bytes32 rId = redemption.requestRedemption(address(token), redeemAmount);
        vm.stopPrank();

        // Warp past expiry SLA
        vm.warp(block.timestamp + DEFAULT_TIMEOUT + 10);

        // Cannot execute expired redemption
        vm.prank(relayer);
        vm.expectRevert();
        redemption.executeRedemptionWithProof(rId, DEMAT_REF, "");

        // Anyone or investor can trigger SLA cancellation
        redemption.cancelExpiredRedemption(rId);

        ITokenRedemption.RedemptionRequest memory req = redemption.getRequest(rId);
        assertEq(uint8(req.status), uint8(ITokenRedemption.RedemptionStatus.CANCELLED));

        // Investor got refunded
        assertEq(token.balanceOf(investor1), 1000 ether);
        assertEq(token.balanceOf(address(redemption)), 0);
    }

    function test_CancelNonExpiredRedemptionReverts() public {
        uint256 redeemAmount = 100 ether;

        vm.startPrank(investor1);
        token.approve(address(redemption), redeemAmount);
        bytes32 rId = redemption.requestRedemption(address(token), redeemAmount);
        vm.stopPrank();

        // Try to cancel before SLA expiry
        vm.expectRevert();
        redemption.cancelExpiredRedemption(rId);
    }

    function test_AdminCancelRedemption() public {
        uint256 redeemAmount = 100 ether;

        vm.startPrank(investor1);
        token.approve(address(redemption), redeemAmount);
        bytes32 rId = redemption.requestRedemption(address(token), redeemAmount);
        vm.stopPrank();

        vm.prank(admin);
        redemption.adminCancelRedemption(rId, "CUSTODIAL_SETTLEMENT_REJECTED");

        ITokenRedemption.RedemptionRequest memory req = redemption.getRequest(rId);
        assertEq(uint8(req.status), uint8(ITokenRedemption.RedemptionStatus.CANCELLED));
        assertEq(token.balanceOf(investor1), 1000 ether);
    }

    function test_UnauthorizedCallerWithoutSignatureReverts() public {
        uint256 redeemAmount = 100 ether;

        vm.startPrank(investor1);
        token.approve(address(redemption), redeemAmount);
        bytes32 rId = redemption.requestRedemption(address(token), redeemAmount);
        vm.stopPrank();

        vm.prank(unauthorized);
        vm.expectRevert();
        redemption.executeRedemptionWithProof(rId, DEMAT_REF, "");
    }

    function testFuzz_InvariantSupplyReductionAndEscrow(uint96 rawAmount) public {
        vm.assume(rawAmount > 0 && rawAmount <= 1000 ether);
        uint256 amount = uint256(rawAmount);

        uint256 supplyStart = token.totalSupply();

        vm.startPrank(investor1);
        token.approve(address(redemption), amount);
        bytes32 rId = redemption.requestRedemption(address(token), amount);
        vm.stopPrank();

        assertEq(token.balanceOf(address(redemption)), amount);
        assertEq(token.totalSupply(), supplyStart);

        vm.prank(relayer);
        redemption.executeRedemptionWithProof(rId, DEMAT_REF, "");

        assertEq(token.balanceOf(address(redemption)), 0);
        assertEq(token.totalSupply(), supplyStart - amount);
    }
}
