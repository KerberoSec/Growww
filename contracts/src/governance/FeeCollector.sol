// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

interface IBurnableToken is IERC20 {
    function burn(uint256 amount) external;
}

/**
 * @title FeeCollector
 * @notice Dynamic Fee Burning & Revenue Sharing Smart Contract for Growww / NBSE.
 * @dev Enforces automated 70/20/10 fee split (Treasury / Market Maker Rebates / Insurance Fund) and token burn.
 */
contract FeeCollector is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    IERC20 public immutable feeAsset; // eINR or USDT
    IBurnableToken public immutable utilityToken; // Native exchange utility token

    address public treasuryWallet;
    address public rebatePoolWallet;
    address public insuranceFundWallet;

    // Fee distribution shares in bps (Total 10,000 = 100%)
    uint256 public constant TREASURY_BPS = 7000;    // 70%
    uint256 public constant REBATE_POOL_BPS = 2000; // 20%
    uint256 public constant INSURANCE_BPS = 1000;   // 10%

    uint256 public totalFeesCollected;
    uint256 public totalTokensBurned;

    mapping(address => bool) public authorizedSettlers;

    event FeeCollected(address indexed source, uint256 amount, uint256 timestamp);
    event RevenueDistributed(uint256 treasuryAmount, uint256 rebateAmount, uint256 insuranceAmount);
    event TokensBurned(uint256 amountBurned, uint256 timestamp);
    event WalletUpdated(string walletType, address newWallet);

    error UnauthorizedSettler(address caller);
    error InvalidAmount();
    error InvalidAddress();

    modifier onlySettlerOrOwner() {
        if (!authorizedSettlers[msg.sender] && msg.sender != owner()) {
            revert UnauthorizedSettler(msg.sender);
        }
        _;
    }

    constructor(
        address _feeAsset,
        address _utilityToken,
        address _treasury,
        address _rebatePool,
        address _insuranceFund,
        address initialOwner
    ) Ownable(initialOwner) {
        if (_feeAsset == address(0) || _treasury == address(0) || _rebatePool == address(0) || _insuranceFund == address(0)) {
            revert InvalidAddress();
        }

        feeAsset = IERC20(_feeAsset);
        utilityToken = IBurnableToken(_utilityToken);
        treasuryWallet = _treasury;
        rebatePoolWallet = _rebatePool;
        insuranceFundWallet = _insuranceFund;
    }

    function setSettlerAuthorization(address settler, bool authorized) external onlyOwner {
        authorizedSettlers[settler] = authorized;
    }

    function updateWallets(address _treasury, address _rebatePool, address _insuranceFund) external onlyOwner {
        if (_treasury != address(0)) {
            treasuryWallet = _treasury;
            emit WalletUpdated("TREASURY", _treasury);
        }
        if (_rebatePool != address(0)) {
            rebatePoolWallet = _rebatePool;
            emit WalletUpdated("REBATE_POOL", _rebatePool);
        }
        if (_insuranceFund != address(0)) {
            insuranceFundWallet = _insuranceFund;
            emit WalletUpdated("INSURANCE_FUND", _insuranceFund);
        }
    }

    /**
     * @notice Deposit collected transaction fees and execute 70/20/10 revenue distribution
     */
    function depositAndDistributeFees(uint256 amount) external onlySettlerOrOwner nonReentrant {
        if (amount == 0) revert InvalidAmount();

        totalFeesCollected += amount;
        feeAsset.safeTransferFrom(msg.sender, address(this), amount);

        uint256 treasuryShare = (amount * TREASURY_BPS) / 10000;
        uint256 rebateShare = (amount * REBATE_POOL_BPS) / 10000;
        uint256 insuranceShare = amount - treasuryShare - rebateShare;

        feeAsset.safeTransfer(treasuryWallet, treasuryShare);
        feeAsset.safeTransfer(rebatePoolWallet, rebateShare);
        feeAsset.safeTransfer(insuranceFundWallet, insuranceShare);

        emit FeeCollected(msg.sender, amount, block.timestamp);
        emit RevenueDistributed(treasuryShare, rebateShare, insuranceShare);
    }

    /**
     * @notice Programmatic token burn mechanism triggered upon reaching volume milestones
     */
    function executeTokenBurn(uint256 burnAmount) external onlyOwner nonReentrant {
        if (burnAmount == 0) revert InvalidAmount();
        require(address(utilityToken) != address(0), "No utility token registered");

        totalTokensBurned += burnAmount;
        utilityToken.burn(burnAmount);

        emit TokensBurned(burnAmount, block.timestamp);
    }
}
