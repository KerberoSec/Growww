// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title VirtualFaucet
 * @notice Testnet Demo Trading Faucet for Growww / NBSE
 * @dev Credits paper trading accounts with 10,000 virtual USDT and 1 virtual BTC
 */
contract VirtualFaucet {
    address public owner;
    
    // Auto-credit amounts (in 18 decimals / 8 decimals)
    uint256 public constant DEFAULT_USDT_GRANT = 10_000 * 10**18;
    uint256 public constant DEFAULT_BTC_GRANT = 1 * 10**8;
    uint256 public constant COOLDOWN_PERIOD = 24 hours;

    mapping(address => uint256) public lastDripTimestamp;
    mapping(address => uint256) public totalDrippedUsdt;
    mapping(address => uint256) public totalDrippedBtc;

    event FaucetClaimed(
        address indexed recipient,
        uint256 usdtAmount,
        uint256 btcAmount,
        uint256 timestamp
    );

    constructor() {
        owner = msg.sender;
    }

    /**
     * @notice Claim demo faucet allowance
     */
    function claimDemoFunds() external returns (bool) {
        require(
            block.timestamp >= lastDripTimestamp[msg.sender] + COOLDOWN_PERIOD,
            "Cooldown: Wait 24 hours between claims"
        );

        lastDripTimestamp[msg.sender] = block.timestamp;
        totalDrippedUsdt[msg.sender] += DEFAULT_USDT_GRANT;
        totalDrippedBtc[msg.sender] += DEFAULT_BTC_GRANT;

        emit FaucetClaimed(msg.sender, DEFAULT_USDT_GRANT, DEFAULT_BTC_GRANT, block.timestamp);
        return true;
    }

    /**
     * @notice Admin-assisted initial auto-credit on demo account creation
     */
    function autoCreditUser(address recipient) external returns (bool) {
        require(msg.sender == owner, "Only owner");
        require(lastDripTimestamp[recipient] == 0, "Account already credited");

        lastDripTimestamp[recipient] = block.timestamp;
        totalDrippedUsdt[recipient] = DEFAULT_USDT_GRANT;
        totalDrippedBtc[recipient] = DEFAULT_BTC_GRANT;

        emit FaucetClaimed(recipient, DEFAULT_USDT_GRANT, DEFAULT_BTC_GRANT, block.timestamp);
        return true;
    }

    function canClaim(address user) external view returns (bool, uint256 timeRemaining) {
        uint256 nextEligible = lastDripTimestamp[user] + COOLDOWN_PERIOD;
        if (block.timestamp >= nextEligible) {
            return (true, 0);
        }
        return (false, nextEligible - block.timestamp);
    }
}
