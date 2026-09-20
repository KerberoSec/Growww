// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {OptionsTokenFactory} from "../../src/derivatives/OptionsTokenFactory.sol";
import {OptionsClearingHouse} from "../../src/derivatives/OptionsClearingHouse.sol";
import {PhysicalAndCashSettler} from "../../src/derivatives/PhysicalAndCashSettler.sol";
import {IOptionsTokenFactory} from "../../src/interfaces/derivatives/IOptionsTokenFactory.sol";
import {IOptionsClearingHouse} from "../../src/interfaces/derivatives/IOptionsClearingHouse.sol";
import {IPhysicalAndCashSettler} from "../../src/interfaces/derivatives/IPhysicalAndCashSettler.sol";
import {IIdentityRegistry} from "../../src/interfaces/IIdentityRegistry.sol";

contract MockOptionERC20 is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract MockIdentityRegistry is IIdentityRegistry {
    mapping(address => bool) public verified;

    function setVerified(address user, bool status) external {
        verified[user] = status;
    }

    function isVerified(address userAddress) external view override returns (bool) {
        return verified[userAddress];
    }

    // Dummy implementations for remaining interface methods
    function registerIdentity(address, bytes32, uint16, uint8) external override {}
    function batchRegisterIdentity(address[] calldata, bytes32[] calldata, uint16[] calldata, uint8[] calldata) external override {}
    function deleteIdentity(address) external override {}
    function updateSanctionStatus(bytes32, bool) external override {}
    function updateCountry(address, uint16) external override {}
    function updateKycTier(address, uint8) external override {}
    function getInvestorClaim(address) external pure override returns (InvestorClaim memory claim) {
        return claim;
    }
    function getInvestorCountry(address) external pure override returns (uint16) { return 356; }
    function getIdentityId(address) external pure override returns (bytes32) { return bytes32(0); }
    function contains(address) external pure override returns (bool) { return true; }
    function isSanctioned(bytes32) external pure override returns (bool) { return false; }
    function totalIdentities() external pure override returns (uint256) { return 0; }
}

contract OptionsClearingTest is Test {
    OptionsTokenFactory public factoryImpl;
    OptionsTokenFactory public factory;
    OptionsClearingHouse public clearingHouseImpl;
    OptionsClearingHouse public clearingHouse;
    PhysicalAndCashSettler public settlerImpl;
    PhysicalAndCashSettler public settler;

    MockOptionERC20 public underlyingToken;
    MockOptionERC20 public settlementToken;
    MockIdentityRegistry public identityRegistry;

    address public admin = address(0xAD01);
    address public treasury = address(0x7801);
    address public coreSgf = address(0x5601);
    address public ipf = address(0x1901);

    uint256 public oraclePrivateKey = 0xA11CE;
    address public oracleRelayer;

    address public writer = address(0x1001);
    address public holder = address(0x1002);
    address public nonKycUser = address(0x1003);

    bytes32 public callCashSeriesId;
    uint256 public callCashTokenId;
    bytes32 public putCashSeriesId;
    uint256 public putCashTokenId;
    bytes32 public callPhysicalSeriesId;
    uint256 public callPhysicalTokenId;

    uint256 public strikePrice = 100e18; // 100 eINR per share
    uint256 public expiry;

    function setUp() public {
        oracleRelayer = vm.addr(oraclePrivateKey);

        underlyingToken = new MockOptionERC20("Growww RWA Equity", "RWA-EQ");
        settlementToken = new MockOptionERC20("Tokenized eINR", "eINR");
        identityRegistry = new MockIdentityRegistry();

        identityRegistry.setVerified(writer, true);
        identityRegistry.setVerified(holder, true);
        identityRegistry.setVerified(admin, true);

        // Deploy Factory
        factoryImpl = new OptionsTokenFactory();
        bytes memory factoryInit = abi.encodeWithSelector(
            OptionsTokenFactory.initialize.selector,
            admin,
            address(0),
            address(identityRegistry),
            "https://derivatives.growww.in/options/{id}.json"
        );
        factory = OptionsTokenFactory(address(new ERC1967Proxy(address(factoryImpl), factoryInit)));

        // Deploy Settler
        settlerImpl = new PhysicalAndCashSettler();
        bytes memory settlerInit = abi.encodeWithSelector(
            PhysicalAndCashSettler.initialize.selector,
            admin,
            address(0),
            treasury,
            coreSgf,
            ipf
        );
        settler = PhysicalAndCashSettler(address(new ERC1967Proxy(address(settlerImpl), settlerInit)));

        // Deploy ClearingHouse
        clearingHouseImpl = new OptionsClearingHouse();
        bytes memory chInit = abi.encodeWithSelector(
            OptionsClearingHouse.initialize.selector,
            admin,
            address(factory),
            address(settler)
        );
        clearingHouse = OptionsClearingHouse(address(new ERC1967Proxy(address(clearingHouseImpl), chInit)));

        // Link addresses
        vm.startPrank(admin);
        factory.setClearingHouse(address(clearingHouse));
        settler.setClearingHouse(address(clearingHouse));
        clearingHouse.setOracleRelayer(oracleRelayer, true);
        vm.stopPrank();

        // Setup test balances
        underlyingToken.mint(writer, 10_000e18);
        settlementToken.mint(writer, 1_000_000e18);
        underlyingToken.mint(holder, 10_000e18);
        settlementToken.mint(holder, 1_000_000e18);

        vm.prank(writer);
        underlyingToken.approve(address(clearingHouse), type(uint256).max);
        vm.prank(writer);
        settlementToken.approve(address(clearingHouse), type(uint256).max);

        vm.prank(holder);
        underlyingToken.approve(address(settler), type(uint256).max);
        vm.prank(holder);
        settlementToken.approve(address(settler), type(uint256).max);

        expiry = block.timestamp + 7 days;

        // Create Series
        (callCashSeriesId, callCashTokenId) = factory.createOptionSeries(
            address(underlyingToken),
            address(settlementToken),
            IOptionsTokenFactory.OptionType.CALL,
            IOptionsTokenFactory.SettlementType.CASH_SETTLED,
            strikePrice,
            expiry,
            false
        );

        (putCashSeriesId, putCashTokenId) = factory.createOptionSeries(
            address(underlyingToken),
            address(settlementToken),
            IOptionsTokenFactory.OptionType.PUT,
            IOptionsTokenFactory.SettlementType.CASH_SETTLED,
            strikePrice,
            expiry,
            false
        );

        (callPhysicalSeriesId, callPhysicalTokenId) = factory.createOptionSeries(
            address(underlyingToken),
            address(settlementToken),
            IOptionsTokenFactory.OptionType.CALL,
            IOptionsTokenFactory.SettlementType.PHYSICAL_DELIVERY,
            strikePrice,
            expiry,
            false
        );
    }

    function test_SeriesCreationAndDeterministicIds() public {
        bytes32 expectedSeriesId = factory.computeSeriesId(
            address(underlyingToken),
            IOptionsTokenFactory.OptionType.CALL,
            strikePrice,
            expiry,
            IOptionsTokenFactory.SettlementType.CASH_SETTLED
        );
        assertEq(callCashSeriesId, expectedSeriesId);
        assertTrue(factory.isSeriesActive(callCashSeriesId));

        IOptionsTokenFactory.OptionSeriesParams memory params = factory.getSeriesParams(callCashSeriesId);
        assertEq(params.underlyingToken, address(underlyingToken));
        assertEq(params.settlementToken, address(settlementToken));
        assertEq(params.strikePrice, strikePrice);
        assertEq(params.expiryTimestamp, expiry);
    }

    function test_RevertOnDuplicateSeries() public {
        vm.expectRevert(abi.encodeWithSelector(IOptionsTokenFactory.SeriesAlreadyExists.selector, callCashSeriesId));
        factory.createOptionSeries(
            address(underlyingToken),
            address(settlementToken),
            IOptionsTokenFactory.OptionType.CALL,
            IOptionsTokenFactory.SettlementType.CASH_SETTLED,
            strikePrice,
            expiry,
            false
        );
    }

    function test_CoveredCallLockupAndMinting() public {
        uint256 contracts = 10e18; // 10 option contracts
        vm.prank(writer);
        uint256 posId = clearingHouse.writeAndMintOption(callCashSeriesId, contracts);

        assertEq(posId, 1);
        assertEq(underlyingToken.balanceOf(address(clearingHouse)), contracts);
        assertEq(clearingHouse.totalLockedCollateral(address(underlyingToken)), contracts);
        assertEq(factory.balanceOf(writer, callCashTokenId), contracts);

        IOptionsClearingHouse.CollateralPosition memory pos = clearingHouse.getCollateralPosition(posId);
        assertEq(pos.writer, writer);
        assertEq(pos.lockedAmount, contracts);
        assertEq(pos.mintedContracts, contracts);
        assertFalse(pos.isReleased);
    }

    function test_CashSecuredPutLockupAndMinting() public {
        uint256 contracts = 5e18; // 5 option contracts
        uint256 requiredCash = (strikePrice * contracts) / 1e18; // 500 eINR
        vm.prank(writer);
        uint256 posId = clearingHouse.writeAndMintOption(putCashSeriesId, contracts);

        assertEq(posId, 1);
        assertEq(settlementToken.balanceOf(address(clearingHouse)), requiredCash);
        assertEq(clearingHouse.totalLockedCollateral(address(settlementToken)), requiredCash);
        assertEq(factory.balanceOf(writer, putCashTokenId), contracts);
    }

    function test_SecondaryTransferKYCCompliance() public {
        uint256 contracts = 10e18;
        vm.prank(writer);
        clearingHouse.writeAndMintOption(callCashSeriesId, contracts);

        // Transfer to verified holder succeeds
        vm.prank(writer);
        factory.safeTransferFrom(writer, holder, callCashTokenId, 5e18, "");
        assertEq(factory.balanceOf(holder, callCashTokenId), 5e18);

        // Transfer to non-KYC user reverts
        vm.prank(writer);
        vm.expectRevert(abi.encodeWithSelector(IOptionsTokenFactory.NonKYCRecipient.selector, nonKycUser));
        factory.safeTransferFrom(writer, nonKycUser, callCashTokenId, 5e18, "");
    }

    function _signSettlementPrice(
        bytes32 seriesId,
        uint256 price,
        uint256 timestamp,
        uint256 seq
    ) internal view returns (bytes memory) {
        bytes32 structHash = keccak256(
            abi.encode(
                clearingHouse.SETTLEMENT_PRICE_TYPEHASH(),
                seriesId,
                price,
                timestamp,
                seq
            )
        );

        bytes32 domainSeparator = keccak256(
            abi.encode(
                keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
                keccak256(bytes("GrowwwOptionsClearingHouse")),
                keccak256(bytes("1.0.0")),
                block.chainid,
                address(clearingHouse)
            )
        );

        bytes32 digest = keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(oraclePrivateKey, digest);
        return abi.encodePacked(r, s, v);
    }

    function test_OraclePriceSubmissionAndCashITMSettlement() public {
        uint256 contracts = 10e18;
        vm.prank(writer);
        clearingHouse.writeAndMintOption(callCashSeriesId, contracts);

        // Transfer options to holder
        vm.prank(writer);
        factory.safeTransferFrom(writer, holder, callCashTokenId, contracts, "");

        // Advance past expiry
        vm.warp(expiry + 1);

        // Underlying price settled at 120 eINR (Strike 100 -> Intrinsic 20 eINR per unit)
        uint256 settlementPrice = 120e18;
        IOptionsClearingHouse.SettlementPriceData memory data = IOptionsClearingHouse.SettlementPriceData({
            seriesId: callCashSeriesId,
            settlementPrice: settlementPrice,
            priceTimestamp: block.timestamp,
            sequenceNumber: 1
        });

        bytes memory signature = _signSettlementPrice(callCashSeriesId, settlementPrice, block.timestamp, 1);
        clearingHouse.submitSettlementPrice(callCashSeriesId, data, signature);

        (bool isITM, uint256 intrinsic) = clearingHouse.isOptionITM(callCashSeriesId);
        assertTrue(isITM);
        assertEq(intrinsic, 20e18);

        // Fund clearing house with settlement cash for Cash Settled Call intrinsic payout
        settlementToken.mint(address(clearingHouse), 200e18);

        // Batch auto exercise
        address[] memory holders = new address[](1);
        holders[0] = holder;
        uint256[] memory amounts = new uint256[](1);
        amounts[0] = contracts;
        uint256[] memory costBases = new uint256[](1);
        costBases[0] = 5e18 * 10;

        uint256 holderBalBefore = settlementToken.balanceOf(holder);

        clearingHouse.batchAutoExerciseAndSettle(
            IOptionsClearingHouse.BatchExerciseParams({
                seriesId: callCashSeriesId,
                holders: holders,
                contractAmounts: amounts,
                costBases: costBases
            })
        );

        assertEq(factory.balanceOf(holder, callCashTokenId), 0);
        // Payoff: 20 eINR * 10 contracts = 200 eINR net (0 fee at launch)
        assertEq(settlementToken.balanceOf(holder) - holderBalBefore, 200e18);
    }

    function test_OTMExpirationAndCollateralReclamation() public {
        uint256 contracts = 10e18;
        vm.prank(writer);
        uint256 posId = clearingHouse.writeAndMintOption(callCashSeriesId, contracts);

        vm.warp(expiry + 1);

        // Underlying price settled at 80 eINR (Strike 100 -> Call is OTM)
        uint256 settlementPrice = 80e18;
        IOptionsClearingHouse.SettlementPriceData memory data = IOptionsClearingHouse.SettlementPriceData({
            seriesId: callCashSeriesId,
            settlementPrice: settlementPrice,
            priceTimestamp: block.timestamp,
            sequenceNumber: 1
        });

        bytes memory signature = _signSettlementPrice(callCashSeriesId, settlementPrice, block.timestamp, 1);
        clearingHouse.submitSettlementPrice(callCashSeriesId, data, signature);

        (bool isITM, ) = clearingHouse.isOptionITM(callCashSeriesId);
        assertFalse(isITM);

        uint256 writerBalBefore = underlyingToken.balanceOf(writer);
        vm.prank(writer);
        clearingHouse.reclaimOTMCollateral(posId);

        assertEq(underlyingToken.balanceOf(writer) - writerBalBefore, contracts);
        assertEq(clearingHouse.totalLockedCollateral(address(underlyingToken)), 0);

        // Cannot reclaim twice
        vm.prank(writer);
        vm.expectRevert(abi.encodeWithSelector(IOptionsClearingHouse.PositionAlreadyReleased.selector, posId));
        clearingHouse.reclaimOTMCollateral(posId);
    }

    function test_PhysicalDeliverySettlement() public {
        uint256 contracts = 2e18;
        vm.prank(writer);
        clearingHouse.writeAndMintOption(callPhysicalSeriesId, contracts);

        vm.prank(writer);
        factory.safeTransferFrom(writer, holder, callPhysicalTokenId, contracts, "");

        vm.warp(expiry + 1);

        uint256 settlementPrice = 150e18; // ITM Call
        IOptionsClearingHouse.SettlementPriceData memory data = IOptionsClearingHouse.SettlementPriceData({
            seriesId: callPhysicalSeriesId,
            settlementPrice: settlementPrice,
            priceTimestamp: block.timestamp,
            sequenceNumber: 1
        });

        bytes memory signature = _signSettlementPrice(callPhysicalSeriesId, settlementPrice, block.timestamp, 1);
        clearingHouse.submitSettlementPrice(callPhysicalSeriesId, data, signature);

        address[] memory holders = new address[](1);
        holders[0] = holder;
        uint256[] memory amounts = new uint256[](1);
        amounts[0] = contracts;
        uint256[] memory costBases = new uint256[](1);
        costBases[0] = 0;

        uint256 holderUnderlyingBefore = underlyingToken.balanceOf(holder);
        uint256 holderCashBefore = settlementToken.balanceOf(holder);

        clearingHouse.batchAutoExerciseAndSettle(
            IOptionsClearingHouse.BatchExerciseParams({
                seriesId: callPhysicalSeriesId,
                holders: holders,
                contractAmounts: amounts,
                costBases: costBases
            })
        );

        // Holder received 2 underlying shares and paid 200 strike cash
        assertEq(underlyingToken.balanceOf(holder) - holderUnderlyingBefore, 2e18);
        assertEq(holderCashBefore - settlementToken.balanceOf(holder), 200e18);
    }
}
