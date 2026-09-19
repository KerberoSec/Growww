// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {CircuitBreakerHalt} from "../../src/governance/CircuitBreakerHalt.sol";
import {ICircuitBreakerHalt} from "../../src/interfaces/ICircuitBreakerHalt.sol";

contract CircuitBreakerHaltTest is Test {
    CircuitBreakerHalt public implementation;
    CircuitBreakerHalt public breaker;

    address public admin = address(0xAD01);
    address public timelock = address(0x7101);
    address public emergencyGuardian = address(0xEA01);
    address public surveillanceSentinel = address(0x5501);
    address public oracleSentinel = address(0x0501);
    address public depositorySentinel = address(0xDE01);

    bytes32 public isin1 = keccak256("INE002A01018"); // Reliance Industries
    bytes32 public isin2 = keccak256("INE467B01029"); // TCS
    bytes32 public sectorBanking = keccak256("SECTOR_BANKING");
    bytes32 public indexNifty50 = keccak256("NIFTY_50_COMPOSITE");

    function setUp() public {
        implementation = new CircuitBreakerHalt();

        bytes memory initData = abi.encodeWithSelector(
            CircuitBreakerHalt.initialize.selector,
            admin,
            timelock,
            emergencyGuardian
        );

        ERC1967Proxy proxy = new ERC1967Proxy(address(implementation), initData);
        breaker = CircuitBreakerHalt(address(proxy));

        vm.startPrank(admin);
        breaker.authorizeSentinel(surveillanceSentinel, breaker.SURVEILLANCE_SENTINEL_ROLE());
        breaker.authorizeSentinel(oracleSentinel, breaker.ORACLE_SENTINEL_ROLE());
        breaker.authorizeSentinel(depositorySentinel, breaker.DEPOSITORY_SENTINEL_ROLE());
        breaker.mapISINToSector(isin1, sectorBanking);
        vm.stopPrank();
    }

    function test_Initialization() public {
        assertTrue(breaker.hasRole(breaker.DEFAULT_ADMIN_ROLE(), admin));
        assertTrue(breaker.hasRole(breaker.GOVERNANCE_TIMELOCK_ROLE(), timelock));
        assertTrue(breaker.hasRole(breaker.EMERGENCY_GUARDIAN_ROLE(), emergencyGuardian));
        assertFalse(breaker.isGlobalFrozen());

        ICircuitBreakerHalt.MarketWideHaltState memory mwState = breaker.getMarketWideState();
        assertEq(uint8(mwState.sessionState), uint8(ICircuitBreakerHalt.MarketSessionState.NORMAL_TRADING));
        assertEq(breaker.defaultOracleHaltDuration(), 3600);
        assertEq(breaker.maxHaltDuration(), 30 days);
    }

    function test_Tier0_DepositoryHalt() public {
        bytes32 mismatchHash = keccak256("NSDL_CDSL_MISMATCH_BATCH_99");

        vm.prank(depositorySentinel);
        breaker.triggerDepositoryHalt(isin1, mismatchHash);

        (bool halted, ICircuitBreakerHalt.HaltTier tier, uint64 expiresAt) = breaker.isHalted(isin1);
        assertTrue(halted);
        assertEq(uint8(tier), uint8(ICircuitBreakerHalt.HaltTier.TIER_0_ISIN));
        assertEq(expiresAt, type(uint64).max);

        vm.expectRevert(
            abi.encodeWithSelector(
                ICircuitBreakerHalt.CircuitBreakerActive.selector,
                isin1,
                ICircuitBreakerHalt.HaltTier.TIER_0_ISIN,
                expiresAt
            )
        );
        breaker.requireNotHalted(isin1);
    }

    function test_Tier0_OracleDivergenceHalt() public {
        bytes32 proofHash = keccak256("ORACLE_MEDIAN_SPREAD_OVER_200BPS");

        vm.prank(oracleSentinel);
        breaker.triggerOracleDivergenceHalt(isin1, 250, proofHash);

        (bool halted, ICircuitBreakerHalt.HaltTier tier, uint64 expiresAt) = breaker.isHalted(isin1);
        assertTrue(halted);
        assertEq(uint8(tier), uint8(ICircuitBreakerHalt.HaltTier.TIER_0_ISIN));
        assertEq(expiresAt, uint64(block.timestamp + 3600));

        ICircuitBreakerHalt.ISINHaltStatus memory status = breaker.getISINHaltStatus(isin1);
        assertEq(uint8(status.reason), uint8(ICircuitBreakerHalt.HaltReason.ORACLE_DIVERGENCE));
    }

    function test_Tier0_SurveillanceVolatilityHaltAndExpiry() public {
        bytes32 anomalyHash = keccak256("SPOOFING_BURST_DETECTED");

        vm.prank(surveillanceSentinel);
        breaker.triggerSurveillanceVolatilityHalt(isin2, 1800, anomalyHash);

        (bool halted, ICircuitBreakerHalt.HaltTier tier, uint64 expiresAt) = breaker.isHalted(isin2);
        assertTrue(halted);
        assertEq(uint8(tier), uint8(ICircuitBreakerHalt.HaltTier.TIER_0_ISIN));
        assertEq(expiresAt, uint64(block.timestamp + 1800));

        // Advance block time beyond expiry (1801 seconds)
        vm.warp(block.timestamp + 1801);
        (bool haltedAfter, ICircuitBreakerHalt.HaltTier tierAfter, ) = breaker.isHalted(isin2);
        assertFalse(haltedAfter);
        assertEq(uint8(tierAfter), uint8(ICircuitBreakerHalt.HaltTier.NONE));
        breaker.requireNotHalted(isin2);
    }

    function test_Tier0_ManualHaltAndResume() public {
        bytes32 detailsHash = keccak256("REGULATORY_INSPECTION_ORDER_44");

        vm.prank(emergencyGuardian);
        breaker.haltISINManual(isin1, ICircuitBreakerHalt.HaltReason.MANUAL_GOVERNANCE, 7200, detailsHash);

        (bool halted, , ) = breaker.isHalted(isin1);
        assertTrue(halted);

        // Resume by timelock governance
        vm.prank(timelock);
        breaker.resumeISINManual(isin1);

        (bool haltedAfter, , ) = breaker.isHalted(isin1);
        assertFalse(haltedAfter);
    }

    function test_Tier1_SectorHaltAndResume() public {
        bytes32 detailsHash = keccak256("BANKING_LIQUIDITY_CASCADE");

        assertFalse(breaker.isSectorHalted(sectorBanking));

        vm.prank(emergencyGuardian);
        breaker.haltSector(sectorBanking, ICircuitBreakerHalt.HaltReason.MANUAL_GOVERNANCE, detailsHash);

        assertTrue(breaker.isSectorHalted(sectorBanking));

        // isin1 is mapped to sectorBanking, so it should report Tier 1 halt
        (bool halted, ICircuitBreakerHalt.HaltTier tier, ) = breaker.isHalted(isin1);
        assertTrue(halted);
        assertEq(uint8(tier), uint8(ICircuitBreakerHalt.HaltTier.TIER_1_SECTOR));

        // isin2 is not in sectorBanking, should not be halted
        (bool halted2, , ) = breaker.isHalted(isin2);
        assertFalse(halted2);

        // Resume sector by timelock
        vm.prank(timelock);
        breaker.resumeSector(sectorBanking);

        assertFalse(breaker.isSectorHalted(sectorBanking));
        (bool halted1After, , ) = breaker.isHalted(isin1);
        assertFalse(halted1After);
    }

    function test_Tier2_MarketWideCircuitBreakerLifecycle() public {
        // Simulated timestamp at 10:00 AM (e.g. 36000 seconds into day)
        uint64 morningTimestamp = 36000;

        vm.prank(surveillanceSentinel);
        breaker.triggerMarketWideCircuitBreaker(10, indexNifty50, 2200000, morningTimestamp);

        ICircuitBreakerHalt.MarketWideHaltState memory mwState = breaker.getMarketWideState();
        assertEq(uint8(mwState.sessionState), uint8(ICircuitBreakerHalt.MarketSessionState.VOLATILITY_HALT));
        assertEq(mwState.circuitLevel, 10);

        // All securities must report Tier 2 halt
        (bool halted, ICircuitBreakerHalt.HaltTier tier, ) = breaker.isHalted(isin1);
        assertTrue(halted);
        assertEq(uint8(tier), uint8(ICircuitBreakerHalt.HaltTier.TIER_2_MARKET_WIDE));

        (bool halted2, ICircuitBreakerHalt.HaltTier tier2, ) = breaker.isHalted(isin2);
        assertTrue(halted2);
        assertEq(uint8(tier2), uint8(ICircuitBreakerHalt.HaltTier.TIER_2_MARKET_WIDE));

        // Advance time past cooldown
        vm.warp(block.timestamp + 2701);

        // Transition to call-auction discovery
        breaker.transitionMarketWideToAuction();
        ICircuitBreakerHalt.MarketWideHaltState memory auctionState = breaker.getMarketWideState();
        assertEq(uint8(auctionState.sessionState), uint8(ICircuitBreakerHalt.MarketSessionState.CALL_AUCTION_DISCOVERY));

        // Advance time past auction discovery (900 seconds)
        vm.warp(block.timestamp + 901);

        // Resume market-wide trading
        breaker.resumeMarketWide();
        ICircuitBreakerHalt.MarketWideHaltState memory finalState = breaker.getMarketWideState();
        assertEq(uint8(finalState.sessionState), uint8(ICircuitBreakerHalt.MarketSessionState.NORMAL_TRADING));

        (bool haltedFinal, , ) = breaker.isHalted(isin1);
        assertFalse(haltedFinal);
    }

    function test_Tier2_MarketWideCircuitBreakerLevel20ClosesDay() public {
        vm.prank(surveillanceSentinel);
        breaker.triggerMarketWideCircuitBreaker(20, indexNifty50, 1800000, 36000);

        ICircuitBreakerHalt.MarketWideHaltState memory state = breaker.getMarketWideState();
        assertEq(uint8(state.sessionState), uint8(ICircuitBreakerHalt.MarketSessionState.MARKET_CLOSED_FOR_DAY));
        assertEq(state.cooldownExpiresAt, type(uint64).max);

        (bool halted, ICircuitBreakerHalt.HaltTier tier, ) = breaker.isHalted(isin1);
        assertTrue(halted);
        assertEq(uint8(tier), uint8(ICircuitBreakerHalt.HaltTier.TIER_2_MARKET_WIDE));
    }

    function test_Tier3_GlobalEmergencyFreezeAndLift() public {
        bytes32 freezeReason = keccak256("CONSORTIUM_BRIDGE_ZERO_DAY_EXPLOIT");

        vm.prank(emergencyGuardian);
        breaker.triggerGlobalEmergencyFreeze(freezeReason);

        assertTrue(breaker.isGlobalFrozen());

        // Both isin1 and isin2 must report Tier 3 Global halt
        (bool halted1, ICircuitBreakerHalt.HaltTier tier1, ) = breaker.isHalted(isin1);
        assertTrue(halted1);
        assertEq(uint8(tier1), uint8(ICircuitBreakerHalt.HaltTier.TIER_3_GLOBAL));

        (bool halted2, ICircuitBreakerHalt.HaltTier tier2, ) = breaker.isHalted(isin2);
        assertTrue(halted2);
        assertEq(uint8(tier2), uint8(ICircuitBreakerHalt.HaltTier.TIER_3_GLOBAL));

        vm.expectRevert(ICircuitBreakerHalt.GlobalFreezeActive.selector);
        breaker.requireNotHalted(isin1);

        // Lift global freeze via timelock governance
        vm.prank(timelock);
        breaker.liftGlobalEmergencyFreeze();

        assertFalse(breaker.isGlobalFrozen());
        (bool haltedAfter, , ) = breaker.isHalted(isin1);
        assertFalse(haltedAfter);
    }

    function test_HaltHierarchyPriority() public {
        // Trigger both Tier 0 and Tier 1 on isin1
        vm.prank(depositorySentinel);
        breaker.triggerDepositoryHalt(isin1, keccak256("DESYNC"));

        vm.prank(emergencyGuardian);
        breaker.haltSector(sectorBanking, ICircuitBreakerHalt.HaltReason.MANUAL_GOVERNANCE, keccak256("SECTOR_RISK"));

        // Tier 1 Sector takes priority over Tier 0 ISIN
        (, ICircuitBreakerHalt.HaltTier tierSector, ) = breaker.isHalted(isin1);
        assertEq(uint8(tierSector), uint8(ICircuitBreakerHalt.HaltTier.TIER_1_SECTOR));

        // Trigger Tier 3 Global Freeze
        vm.prank(emergencyGuardian);
        breaker.triggerGlobalEmergencyFreeze(keccak256("SYSTEMIC"));

        // Tier 3 Global takes highest priority
        (, ICircuitBreakerHalt.HaltTier tierGlobal, ) = breaker.isHalted(isin1);
        assertEq(uint8(tierGlobal), uint8(ICircuitBreakerHalt.HaltTier.TIER_3_GLOBAL));
    }

    function test_UnauthorizedSentinelReverts() public {
        address unauth = address(0xBEEF);
        bytes32 role = breaker.DEPOSITORY_SENTINEL_ROLE();

        vm.prank(unauth);
        vm.expectRevert(
            abi.encodeWithSelector(
                ICircuitBreakerHalt.UnauthorizedSentinel.selector,
                unauth,
                role
            )
        );
        breaker.triggerDepositoryHalt(isin1, keccak256("MISMATCH"));
    }
}
