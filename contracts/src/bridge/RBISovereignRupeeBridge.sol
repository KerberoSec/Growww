// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title WrappedDigitalRupee
 * @notice 1:1 backed wrapped settlement token for RBI Central Bank Digital Currency (eINR-W).
 */
contract WrappedDigitalRupee is ERC20, Ownable {
    address public bridgeContract;

    modifier onlyBridge() {
        require(msg.sender == bridgeContract, "Only bridge can mint/burn");
        _;
    }

    constructor(address initialOwner) ERC20("Wrapped Sovereign e-Rupee", "w-eINR") Ownable(initialOwner) {}

    function setBridge(address _bridge) external onlyOwner {
        bridgeContract = _bridge;
    }

    function mint(address to, uint256 amount) external onlyBridge {
        _mint(to, amount);
    }

    function burn(address from, uint256 amount) external onlyBridge {
        _burn(from, amount);
    }
}

/**
 * @title RBISovereignRupeeBridge
 * @notice Sovereign bridge locking e-Rupee tokens and minting 1:1 wrapped settlement tokens (w-eINR).
 * @dev Enforces RBI regulatory compliance, daily volume thresholds, and emergency circuit breakers.
 */
contract RBISovereignRupeeBridge is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    IERC20 public immutable sovereignRupee;
    WrappedDigitalRupee public immutable wrappedToken;

    uint256 public totalLockedReserve;
    uint256 public dailyLimitPaise = 50_000_000_00; // Default 50 Lakh INR per 24 hours per account
    bool public isBridgePaused;

    mapping(address => bool) public isKYCApproved;
    mapping(address => mapping(uint256 => uint256)) public dailyVolume; // user => dayTimestamp => amount
    mapping(bytes32 => bool) public processedInboundWires;

    event RupeeLocked(address indexed user, uint256 amountPaise, uint256 timestamp);
    event RupeeUnlocked(address indexed user, uint256 amountPaise, string bankTxRef, uint256 timestamp);
    event DailyLimitUpdated(uint256 oldLimit, uint256 newLimit);
    event KYCStatusSet(address indexed user, bool status);
    event EmergencyHaltToggled(bool isPaused);

    error BridgePaused();
    error KYCRequired();
    error DailyLimitExceeded(uint256 currentVolume, uint256 attemptAmount, uint256 limit);
    error InvalidAmount();
    error WireAlreadyProcessed(bytes32 wireHash);

    modifier whenNotPaused() {
        if (isBridgePaused) revert BridgePaused();
        _;
    }

    constructor(address _sovereignRupee, address initialOwner) Ownable(initialOwner) {
        require(_sovereignRupee != address(0), "Invalid sovereign rupee token");
        sovereignRupee = IERC20(_sovereignRupee);
        wrappedToken = new WrappedDigitalRupee(address(this));
        wrappedToken.setBridge(address(this));
    }

    function setKYCStatus(address user, bool status) external onlyOwner {
        isKYCApproved[user] = status;
        emit KYCStatusSet(user, status);
    }

    function setDailyLimit(uint256 newLimitPaise) external onlyOwner {
        emit DailyLimitUpdated(dailyLimitPaise, newLimitPaise);
        dailyLimitPaise = newLimitPaise;
    }

    function toggleEmergencyHalt(bool pauseState) external onlyOwner {
        isBridgePaused = pauseState;
        emit EmergencyHaltToggled(pauseState);
    }

    /**
     * @notice Deposit sovereign CBDC e-Rupee to mint 1:1 wrapped settlement tokens (w-eINR)
     */
    function lockAndMint(uint256 amountPaise) external nonReentrant whenNotPaused {
        if (amountPaise == 0) revert InvalidAmount();
        if (!isKYCApproved[msg.sender]) revert KYCRequired();

        uint256 today = block.timestamp / 1 days;
        uint256 currentDayVol = dailyVolume[msg.sender][today];
        if (currentDayVol + amountPaise > dailyLimitPaise) {
            revert DailyLimitExceeded(currentDayVol, amountPaise, dailyLimitPaise);
        }

        dailyVolume[msg.sender][today] = currentDayVol + amountPaise;
        totalLockedReserve += amountPaise;

        // Escrow sovereign tokens
        sovereignRupee.safeTransferFrom(msg.sender, address(this), amountPaise);

        // Mint wrapped tokens 1:1
        wrappedToken.mint(msg.sender, amountPaise);

        emit RupeeLocked(msg.sender, amountPaise, block.timestamp);
    }

    /**
     * @notice Burn wrapped tokens (w-eINR) to unlock and release sovereign e-Rupee
     */
    function burnAndUnlock(uint256 amountPaise, string calldata bankTxRef) external nonReentrant whenNotPaused {
        if (amountPaise == 0) revert InvalidAmount();
        if (!isKYCApproved[msg.sender]) revert KYCRequired();
        require(totalLockedReserve >= amountPaise, "Insufficient bridge reserve");

        totalLockedReserve -= amountPaise;

        // Burn wrapped tokens
        wrappedToken.burn(msg.sender, amountPaise);

        // Transfer sovereign e-Rupee back to recipient
        sovereignRupee.safeTransfer(msg.sender, amountPaise);

        emit RupeeUnlocked(msg.sender, amountPaise, bankTxRef, block.timestamp);
    }
}
