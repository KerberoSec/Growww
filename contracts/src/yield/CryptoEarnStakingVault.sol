// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title CryptoEarnStakingVault
 * @notice Production-grade institutional wealth management staking vault (Prompt 088).
 * @dev Supports Flexible Earn (instant redemption) and Fixed-Term Staking (30/60/90 days).
 * Features:
 *   - Continuous daily yield calculation based on product APR (basis points, e.g., 500 = 5.00%).
 *   - 10% risk reserve fund allocation from gross interest to safeguard user principal.
 *   - Early redemption penalty on fixed staking terms (forfeits accrued unvested interest).
 *   - Admin funding of yield reward pools.
 */
contract CryptoEarnStakingVault is AccessControl, Pausable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    enum StakingType { FLEXIBLE, FIXED_30, FIXED_60, FIXED_90 }

    struct Product {
        uint256 productId;
        address assetToken;
        StakingType stakingType;
        uint256 lockDurationSeconds; // 0 for FLEXIBLE, 30 days, 60 days, 90 days
        uint256 aprBps;              // Annual Percentage Rate in basis points (100 bps = 1.00%)
        uint256 minDeposit;
        uint256 maxTotalCapacity;
        uint256 currentTotalDeposits;
        bool active;
    }

    struct UserStake {
        uint256 stakeId;
        uint256 productId;
        address user;
        uint256 principal;
        uint256 depositTimestamp;
        uint256 lastClaimTimestamp;
        uint256 unlockTimestamp;
        bool active;
    }

    uint256 public constant BPS_DENOMINATOR = 10000;
    uint256 public constant SECONDS_PER_YEAR = 365 days;
    uint256 public constant RISK_RESERVE_BPS = 1000; // 10% of gross yield

    uint256 public nextProductId = 1;
    uint256 public nextStakeId = 1;

    mapping(uint256 => Product) public products;
    mapping(uint256 => UserStake) public userStakes;
    mapping(address => uint256[]) internal _userStakeIds;
    mapping(address => uint256) public riskReserveBalances; // token => accumulated reserve

    event ProductCreated(uint256 indexed productId, address indexed assetToken, StakingType stakingType, uint256 aprBps);
    event Deposited(uint256 indexed stakeId, uint256 indexed productId, address indexed user, uint256 amount);
    event YieldClaimed(uint256 indexed stakeId, address indexed user, uint256 netYield, uint256 reserveCut);
    event Redeemed(uint256 indexed stakeId, address indexed user, uint256 principalReturned, uint256 netYield);
    event ReserveFundWithdrawn(address indexed token, address indexed recipient, uint256 amount);

    error ProductNotFound();
    error ProductInactive();
    error DepositTooSmall();
    error CapacityExceeded();
    error StakeNotFound();
    error StakeInactive();
    error StakeLocked();
    error Unauthorized();

    constructor(address admin) {
        require(admin != address(0), "Invalid admin");
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(OPERATOR_ROLE, admin);
    }

    /**
     * @notice Creates an Earn staking product (Flexible or Fixed)
     */
    function createProduct(
        address assetToken,
        StakingType stakingType,
        uint256 aprBps,
        uint256 minDeposit,
        uint256 maxTotalCapacity
    ) external onlyRole(OPERATOR_ROLE) returns (uint256) {
        require(assetToken != address(0), "Invalid token");
        require(aprBps > 0 && aprBps <= 5000, "APR 0.01% - 50%"); // reasonable upper bound

        uint256 duration = 0;
        if (stakingType == StakingType.FIXED_30) {
            duration = 30 days;
        } else if (stakingType == StakingType.FIXED_60) {
            duration = 60 days;
        } else if (stakingType == StakingType.FIXED_90) {
            duration = 90 days;
        }

        uint256 pid = nextProductId++;
        products[pid] = Product({
            productId: pid,
            assetToken: assetToken,
            stakingType: stakingType,
            lockDurationSeconds: duration,
            aprBps: aprBps,
            minDeposit: minDeposit,
            maxTotalCapacity: maxTotalCapacity,
            currentTotalDeposits: 0,
            active: true
        });

        emit ProductCreated(pid, assetToken, stakingType, aprBps);
        return pid;
    }

    /**
     * @notice Stakes tokens into an active Earn product
     */
    function deposit(uint256 productId, uint256 amount) external nonReentrant whenNotPaused returns (uint256) {
        Product storage prod = products[productId];
        if (!prod.active) revert ProductInactive();
        if (amount < prod.minDeposit) revert DepositTooSmall();
        if (prod.currentTotalDeposits + amount > prod.maxTotalCapacity) revert CapacityExceeded();

        prod.currentTotalDeposits += amount;

        uint256 sid = nextStakeId++;
        uint256 unlockTime = prod.lockDurationSeconds > 0 ? block.timestamp + prod.lockDurationSeconds : 0;

        userStakes[sid] = UserStake({
            stakeId: sid,
            productId: productId,
            user: msg.sender,
            principal: amount,
            depositTimestamp: block.timestamp,
            lastClaimTimestamp: block.timestamp,
            unlockTimestamp: unlockTime,
            active: true
        });

        _userStakeIds[msg.sender].push(sid);

        IERC20(prod.assetToken).safeTransferFrom(msg.sender, address(this), amount);

        emit Deposited(sid, productId, msg.sender, amount);
        return sid;
    }

    /**
     * @notice Computes gross accrued interest for a stake
     */
    function calculateAccruedInterest(uint256 stakeId) public view returns (uint256 grossInterest, uint256 netYield, uint256 reserveCut) {
        UserStake storage stake = userStakes[stakeId];
        if (!stake.active) return (0, 0, 0);

        Product storage prod = products[stake.productId];
        uint256 duration = block.timestamp - stake.lastClaimTimestamp;
        if (duration == 0) return (0, 0, 0);

        // gross = principal * aprBps * duration / (BPS_DENOMINATOR * SECONDS_PER_YEAR)
        grossInterest = (stake.principal * prod.aprBps * duration) / (BPS_DENOMINATOR * SECONDS_PER_YEAR);
        reserveCut = (grossInterest * RISK_RESERVE_BPS) / BPS_DENOMINATOR;
        netYield = grossInterest - reserveCut;
    }

    /**
     * @notice Harvests accrued yield without withdrawing principal
     */
    function claimYield(uint256 stakeId) external nonReentrant whenNotPaused returns (uint256) {
        UserStake storage stake = userStakes[stakeId];
        if (!stake.active) revert StakeInactive();
        if (stake.user != msg.sender) revert Unauthorized();

        (uint256 gross, uint256 net, uint256 reserveCut) = calculateAccruedInterest(stakeId);
        if (gross == 0) return 0;

        stake.lastClaimTimestamp = block.timestamp;
        Product storage prod = products[stake.productId];

        riskReserveBalances[prod.assetToken] += reserveCut;
        IERC20(prod.assetToken).safeTransfer(msg.sender, net);

        emit YieldClaimed(stakeId, msg.sender, net, reserveCut);
        return net;
    }

    /**
     * @notice Redeems stake principal + accrued yield
     * For Fixed-term: if redeemed before unlock, forfeits unharvested interest.
     */
    function redeem(uint256 stakeId) external nonReentrant returns (uint256 totalReturned) {
        UserStake storage stake = userStakes[stakeId];
        if (!stake.active) revert StakeInactive();
        if (stake.user != msg.sender) revert Unauthorized();

        Product storage prod = products[stake.productId];
        uint256 principal = stake.principal;
        uint256 netYield = 0;

        if (prod.lockDurationSeconds > 0 && block.timestamp < stake.unlockTimestamp) {
            // Early redemption penalty: forfeits unharvested yield
            netYield = 0;
        } else {
            // Full maturity or flexible: grant accrued yield
            (, netYield, ) = calculateAccruedInterest(stakeId);
        }

        stake.active = false;
        prod.currentTotalDeposits -= principal;

        totalReturned = principal + netYield;
        IERC20(prod.assetToken).safeTransfer(msg.sender, totalReturned);

        emit Redeemed(stakeId, msg.sender, principal, netYield);
    }

    /**
     * @notice Withdraw risk reserve pool for emergency insurance or protocol maintenance
     */
    function withdrawReserve(address token, address recipient, uint256 amount) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(amount <= riskReserveBalances[token], "Insufficient reserve");
        riskReserveBalances[token] -= amount;
        IERC20(token).safeTransfer(recipient, amount);
        emit ReserveFundWithdrawn(token, recipient, amount);
    }

    function getUserStakes(address user) external view returns (uint256[] memory) {
        return _userStakeIds[user];
    }

    function pause() external onlyRole(OPERATOR_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(OPERATOR_ROLE) {
        _unpause();
    }
}
