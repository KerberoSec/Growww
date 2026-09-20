// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "../interfaces/yield/ICopyTradingVault.sol";

/**
 * @title CopyTradingVault
 * @notice Vault contract holding follower funds with automated 10-20% high-water mark performance fee distribution.
 * @dev Enforces pro-rata share accounting, high-water mark fee deductions, and strategy execution controls.
 */
contract CopyTradingVault is
    ERC20Upgradeable,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    ReentrancyGuardUpgradeable,
    PausableUpgradeable,
    ICopyTradingVault
{
    using SafeERC20 for IERC20;

    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    bytes32 public constant EMERGENCY_GUARDIAN_ROLE = keccak256("EMERGENCY_GUARDIAN_ROLE");

    address public override masterTrader;
    address public assetToken;
    uint256 public performanceFeeBps; // 1000 - 2000 (10% - 20%)
    uint256 public maxCapacity;
    uint256 public minDeposit;
    uint256 public lockupPeriod;

    uint256 public override highWaterMark;
    uint256 public deployedStrategyFunds;

    mapping(address => FollowerPosition) private _followerPositions;
    mapping(address => bool) public approvedStrategyTargets;

    modifier onlyMasterTrader() {
        if (msg.sender != masterTrader && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedMasterTrader(msg.sender);
        }
        _;
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address admin,
        address _masterTrader,
        address _assetToken,
        string memory shareName,
        string memory shareSymbol,
        uint256 _performanceFeeBps,
        uint256 _maxCapacity,
        uint256 _minDeposit,
        uint256 _lockupPeriod
    ) external initializer {
        if (_masterTrader == address(0) || _assetToken == address(0)) revert InvalidZeroAddress();
        if (_performanceFeeBps < 1000 || _performanceFeeBps > 2000) revert InvalidPerformanceFeeBps(_performanceFeeBps);

        __ERC20_init(shareName, shareSymbol);
        __AccessControl_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();
        __Pausable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        _grantRole(EMERGENCY_GUARDIAN_ROLE, admin);

        masterTrader = _masterTrader;
        assetToken = _assetToken;
        performanceFeeBps = _performanceFeeBps;
        maxCapacity = _maxCapacity;
        minDeposit = _minDeposit;
        lockupPeriod = _lockupPeriod;

        highWaterMark = 1e18; // Base 1:1 price
    }

    function setApprovedStrategyTarget(address target, bool approved) external onlyRole(DEFAULT_ADMIN_ROLE) {
        approvedStrategyTargets[target] = approved;
    }

    function setMasterTrader(address newMaster) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newMaster == address(0)) revert InvalidZeroAddress();
        address oldMaster = masterTrader;
        masterTrader = newMaster;
        emit MasterTraderUpdated(oldMaster, newMaster);
    }

    function setPerformanceFeeBps(uint256 newBps) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newBps < 1000 || newBps > 2000) revert InvalidPerformanceFeeBps(newBps);
        uint256 oldBps = performanceFeeBps;
        performanceFeeBps = newBps;
        emit PerformanceFeeBpsUpdated(oldBps, newBps);
    }

    function setMaxCapacity(uint256 newCapacity) external onlyRole(DEFAULT_ADMIN_ROLE) {
        maxCapacity = newCapacity;
    }

    function pause() external onlyRole(EMERGENCY_GUARDIAN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    function totalAssets() public view override returns (uint256) {
        return IERC20(assetToken).balanceOf(address(this)) + deployedStrategyFunds;
    }

    function totalShares() external view override returns (uint256) {
        return totalSupply();
    }

    function currentSharePrice() public view override returns (uint256) {
        uint256 supply = totalSupply();
        if (supply == 0) return 1e18;
        return (totalAssets() * 1e18) / supply;
    }

    function convertToShares(uint256 assets) public view override returns (uint256) {
        uint256 supply = totalSupply();
        uint256 assetsTotal = totalAssets();
        if (supply == 0 || assetsTotal == 0) {
            return assets;
        }
        return (assets * supply) / assetsTotal;
    }

    function convertToAssets(uint256 shares) public view override returns (uint256) {
        uint256 supply = totalSupply();
        if (supply == 0) return shares;
        return (shares * totalAssets()) / supply;
    }

    function deposit(uint256 assets, address receiver) external override whenNotPaused nonReentrant returns (uint256 shares) {
        if (assets < minDeposit) revert DepositBelowMinimum(assets, minDeposit);
        if (receiver == address(0)) revert InvalidZeroAddress();

        uint256 expectedTotalAssets = totalAssets() + assets;
        if (maxCapacity > 0 && expectedTotalAssets > maxCapacity) {
            revert CapacityExceeded(expectedTotalAssets, maxCapacity);
        }

        shares = convertToShares(assets);
        IERC20(assetToken).safeTransferFrom(msg.sender, address(this), assets);

        _mint(receiver, shares);

        FollowerPosition storage pos = _followerPositions[receiver];
        pos.shares += shares;
        pos.depositedAmount += assets;
        pos.depositTimestamp = block.timestamp;
        if (pos.highWaterMark == 0) {
            pos.highWaterMark = currentSharePrice();
        }

        emit FollowerDeposited(receiver, assets, shares);
    }

    function withdraw(uint256 shares, address receiver) external override nonReentrant returns (uint256 assets) {
        if (receiver == address(0)) revert InvalidZeroAddress();
        FollowerPosition storage pos = _followerPositions[msg.sender];
        if (shares > pos.shares || shares > balanceOf(msg.sender)) {
            revert InsufficientShares(shares, pos.shares);
        }
        if (block.timestamp < pos.depositTimestamp + lockupPeriod) {
            revert LockupActive(pos.depositTimestamp + lockupPeriod, block.timestamp);
        }

        assets = convertToAssets(shares);
        pos.shares -= shares;
        _burn(msg.sender, shares);

        IERC20(assetToken).safeTransfer(receiver, assets);

        emit FollowerWithdrawn(msg.sender, shares, assets);
    }

    function harvestPerformanceFee() public override nonReentrant returns (uint256 feeShares) {
        uint256 price = currentSharePrice();
        if (price <= highWaterMark) {
            revert NoProfitAboveHighWaterMark(price, highWaterMark);
        }

        uint256 profitPerShare = price - highWaterMark;
        uint256 supply = totalSupply();
        uint256 grossProfit = (profitPerShare * supply) / 1e18;
        uint256 feeAmount = (grossProfit * performanceFeeBps) / 10000;

        // Fee shares minted to master trader
        feeShares = (feeAmount * 1e18) / price;
        if (feeShares > 0) {
            _mint(masterTrader, feeShares);
        }

        uint256 oldHWM = highWaterMark;
        highWaterMark = currentSharePrice();

        emit PerformanceFeeHarvested(masterTrader, grossProfit, feeAmount, highWaterMark);
        emit HighWaterMarkUpdated(oldHWM, highWaterMark);
    }

    function executeStrategy(
        address target,
        uint256 amount,
        bytes calldata data
    ) external override onlyMasterTrader whenNotPaused nonReentrant returns (bytes memory result) {
        if (!approvedStrategyTargets[target]) revert TargetNotApproved(target);
        if (amount > IERC20(assetToken).balanceOf(address(this))) {
            revert InsufficientShares(amount, IERC20(assetToken).balanceOf(address(this)));
        }

        deployedStrategyFunds += amount;
        IERC20(assetToken).safeTransfer(target, amount);

        (bool success, bytes memory returnData) = target.call(data);
        require(success, "Strategy execution failed");

        emit VaultStrategyExecuted(target, amount, data);
        return returnData;
    }

    function returnStrategyFunds(uint256 amount) external override nonReentrant {
        IERC20(assetToken).safeTransferFrom(msg.sender, address(this), amount);
        if (amount > deployedStrategyFunds) {
            deployedStrategyFunds = 0;
        } else {
            deployedStrategyFunds -= amount;
        }

        emit StrategyFundsReturned(msg.sender, amount);
    }

    function getFollowerPosition(address follower) external view override returns (FollowerPosition memory) {
        return _followerPositions[follower];
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    uint256[45] private __gap;
}
