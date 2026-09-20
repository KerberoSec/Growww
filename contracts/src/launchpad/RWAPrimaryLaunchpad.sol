// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

interface IRWAToken {
    function transfer(address to, uint256 amount) external returns (bool);
    function isKYCVerified(address account) external view returns (bool);
}

/**
 * @title RWAPrimaryLaunchpad
 * @notice Primary issuance capital formation engine for tokenized Real-World Assets (Treasury Bills, Real Estate, Debt).
 * @dev Supports Fixed-Price subscriptions, Dutch Auctions, KYC gating, soft-cap thresholds, and automated refunds.
 */
contract RWAPrimaryLaunchpad is AccessControl, Pausable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    enum IssuanceType { FIXED_PRICE, DUTCH_AUCTION }
    enum IssuanceStatus { SCHEDULED, ACTIVE, SUCCESSFUL, FAILED, CANCELLED }

    struct Issuance {
        uint256 id;
        address issuer;
        address rwaToken;
        address paymentToken; // e.g. USDT or eINR
        IssuanceType issuanceType;
        uint256 totalTokensOffered;
        uint256 tokensSubscribed;
        uint256 fixedPricePerToken;    // In paymentToken wei (for FIXED_PRICE)
        uint256 auctionStartPrice;     // For DUTCH_AUCTION
        uint256 auctionFloorPrice;     // For DUTCH_AUCTION
        uint256 startTime;
        uint256 endTime;
        uint256 minSoftCapTokens;      // Minimum threshold to consider issuance successful
        uint256 totalFundsRaised;
        IssuanceStatus status;
    }

    struct Subscription {
        uint256 tokensCommitted;
        uint256 fundsDeposited;
        bool claimed;
        bool refunded;
    }

    uint256 public nextIssuanceId = 1;
    mapping(uint256 => Issuance) public issuances;
    mapping(uint256 => mapping(address => Subscription)) public subscriptions;

    event IssuanceCreated(
        uint256 indexed issuanceId,
        address indexed issuer,
        address rwaToken,
        IssuanceType issuanceType,
        uint256 totalTokensOffered,
        uint256 startTime,
        uint256 endTime
    );

    event Subscribed(
        uint256 indexed issuanceId,
        address indexed investor,
        uint256 tokensAmount,
        uint256 fundsPaid
    );

    event IssuanceFinalized(uint256 indexed issuanceId, IssuanceStatus finalStatus, uint256 totalRaised);
    event TokensClaimed(uint256 indexed issuanceId, address indexed investor, uint256 tokensDelivered);
    event RefundClaimed(uint256 indexed issuanceId, address indexed investor, uint256 fundsReturned);

    error InvalidTimeWindow();
    error InvalidCapOrOffer();
    error IssuanceNotActive();
    error IdentityNotVerified();
    error ExceedsAvailableTokens();
    error IssuanceNotFinalized();
    error IssuanceDidNotSucceed();
    error IssuanceDidNotFail();
    error AlreadyClaimedOrRefunded();
    error NothingToClaim();

    constructor(address admin) {
        require(admin != address(0), "Invalid admin");
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(OPERATOR_ROLE, admin);
    }

    /**
     * @notice Creates a new RWA primary market issuance
     */
    function createIssuance(
        address rwaToken,
        address paymentToken,
        IssuanceType issuanceType,
        uint256 totalTokensOffered,
        uint256 pricePerToken,
        uint256 auctionStartPrice,
        uint256 auctionFloorPrice,
        uint256 startTime,
        uint256 endTime,
        uint256 minSoftCapTokens
    ) external onlyRole(OPERATOR_ROLE) whenNotPaused returns (uint256) {
        if (startTime >= endTime || startTime < block.timestamp) revert InvalidTimeWindow();
        if (totalTokensOffered == 0 || minSoftCapTokens > totalTokensOffered) revert InvalidCapOrOffer();
        if (rwaToken == address(0) || paymentToken == address(0)) revert InvalidCapOrOffer();

        uint256 id = nextIssuanceId++;
        issuances[id] = Issuance({
            id: id,
            issuer: msg.sender,
            rwaToken: rwaToken,
            paymentToken: paymentToken,
            issuanceType: issuanceType,
            totalTokensOffered: totalTokensOffered,
            tokensSubscribed: 0,
            fixedPricePerToken: pricePerToken,
            auctionStartPrice: auctionStartPrice,
            auctionFloorPrice: auctionFloorPrice,
            startTime: startTime,
            endTime: endTime,
            minSoftCapTokens: minSoftCapTokens,
            totalFundsRaised: 0,
            status: IssuanceStatus.ACTIVE
        });

        // Escrow offered RWA tokens from issuer into launchpad contract
        IERC20(rwaToken).safeTransferFrom(msg.sender, address(this), totalTokensOffered);

        emit IssuanceCreated(id, msg.sender, rwaToken, issuanceType, totalTokensOffered, startTime, endTime);
        return id;
    }

    /**
     * @notice Returns effective price at current timestamp (for Dutch Auction or Fixed Price)
     */
    function getCurrentPrice(uint256 issuanceId) public view returns (uint256) {
        Issuance storage iss = issuances[issuanceId];
        if (iss.issuanceType == IssuanceType.FIXED_PRICE) {
            return iss.fixedPricePerToken;
        }

        if (block.timestamp <= iss.startTime) {
            return iss.auctionStartPrice;
        }
        if (block.timestamp >= iss.endTime) {
            return iss.auctionFloorPrice;
        }

        uint256 elapsed = block.timestamp - iss.startTime;
        uint256 duration = iss.endTime - iss.startTime;
        uint256 priceDrop = (iss.auctionStartPrice - iss.auctionFloorPrice) * elapsed / duration;
        return iss.auctionStartPrice - priceDrop;
    }

    /**
     * @notice Subscribes to primary issuance; locks payment tokens in escrow
     */
    function subscribe(uint256 issuanceId, uint256 tokenAmount) external nonReentrant whenNotPaused {
        Issuance storage iss = issuances[issuanceId];
        if (iss.status != IssuanceStatus.ACTIVE) revert IssuanceNotActive();
        if (block.timestamp < iss.startTime || block.timestamp > iss.endTime) revert IssuanceNotActive();
        if (iss.tokensSubscribed + tokenAmount > iss.totalTokensOffered) revert ExceedsAvailableTokens();

        // Enforce ERC-3643 KYC verification check
        try IRWAToken(iss.rwaToken).isKYCVerified(msg.sender) returns (bool verified) {
            if (!verified) revert IdentityNotVerified();
        } catch {
            // Token does not support isKYCVerified; proceed
        }

        uint256 unitPrice = getCurrentPrice(issuanceId);
        uint256 totalCost = (tokenAmount * unitPrice) / 1e18;

        iss.tokensSubscribed += tokenAmount;
        iss.totalFundsRaised += totalCost;

        Subscription storage sub = subscriptions[issuanceId][msg.sender];
        sub.tokensCommitted += tokenAmount;
        sub.fundsDeposited += totalCost;

        IERC20(iss.paymentToken).safeTransferFrom(msg.sender, address(this), totalCost);

        emit Subscribed(issuanceId, msg.sender, tokenAmount, totalCost);
    }

    /**
     * @notice Finalizes issuance after end timestamp
     */
    function finalizeIssuance(uint256 issuanceId) external nonReentrant {
        Issuance storage iss = issuances[issuanceId];
        if (iss.status != IssuanceStatus.ACTIVE) revert IssuanceNotActive();
        if (block.timestamp <= iss.endTime && iss.tokensSubscribed < iss.totalTokensOffered) {
            revert IssuanceNotActive(); // Still running
        }

        if (iss.tokensSubscribed >= iss.minSoftCapTokens) {
            iss.status = IssuanceStatus.SUCCESSFUL;
            // Transfer funds to issuer
            IERC20(iss.paymentToken).safeTransfer(iss.issuer, iss.totalFundsRaised);
            // Return any unsold tokens to issuer
            uint256 unsold = iss.totalTokensOffered - iss.tokensSubscribed;
            if (unsold > 0) {
                IERC20(iss.rwaToken).safeTransfer(iss.issuer, unsold);
            }
        } else {
            iss.status = IssuanceStatus.FAILED;
            // Return all RWA tokens to issuer
            IERC20(iss.rwaToken).safeTransfer(iss.issuer, iss.totalTokensOffered);
        }

        emit IssuanceFinalized(issuanceId, iss.status, iss.totalFundsRaised);
    }

    /**
     * @notice Investor claims tokenized RWA allocations upon successful campaign
     */
    function claimTokens(uint256 issuanceId) external nonReentrant {
        Issuance storage iss = issuances[issuanceId];
        if (iss.status != IssuanceStatus.SUCCESSFUL) revert IssuanceDidNotSucceed();

        Subscription storage sub = subscriptions[issuanceId][msg.sender];
        if (sub.claimed || sub.refunded) revert AlreadyClaimedOrRefunded();
        if (sub.tokensCommitted == 0) revert NothingToClaim();

        sub.claimed = true;
        IERC20(iss.rwaToken).safeTransfer(msg.sender, sub.tokensCommitted);

        emit TokensClaimed(issuanceId, msg.sender, sub.tokensCommitted);
    }

    /**
     * @notice Investor claims full refund if issuance soft-cap was not reached
     */
    function claimRefund(uint256 issuanceId) external nonReentrant {
        Issuance storage iss = issuances[issuanceId];
        if (iss.status != IssuanceStatus.FAILED && iss.status != IssuanceStatus.CANCELLED) {
            revert IssuanceDidNotFail();
        }

        Subscription storage sub = subscriptions[issuanceId][msg.sender];
        if (sub.claimed || sub.refunded) revert AlreadyClaimedOrRefunded();
        if (sub.fundsDeposited == 0) revert NothingToClaim();

        sub.refunded = true;
        IERC20(iss.paymentToken).safeTransfer(msg.sender, sub.fundsDeposited);

        emit RefundClaimed(issuanceId, msg.sender, sub.fundsDeposited);
    }

    function pause() external onlyRole(OPERATOR_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(OPERATOR_ROLE) {
        _unpause();
    }
}
