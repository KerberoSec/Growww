// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "../interfaces/derivatives/IPhysicalAndCashSettler.sol";
import "../interfaces/derivatives/IOptionsTokenFactory.sol";

/**
 * @title PhysicalAndCashSettler
 * @notice Modular settlement engine executing atomic physical delivery or cash intrinsic payout.
 * @dev Enforces the canonical Growww 0.00% platform fee at launch and records FIFO tax proof attestations.
 */
contract PhysicalAndCashSettler is
    IPhysicalAndCashSettler,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    ReentrancyGuardUpgradeable
{
    using SafeERC20 for IERC20;

    bytes32 public constant CLEARING_HOUSE_ROLE = keccak256("CLEARING_HOUSE_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    address public treasuryVault;
    address public coreSgfVault;
    address public ipfVault;
    address public override clearingHouse;

    // Fee in basis points (0 bps = 0.00% at launch as per Growww Zero Fee platform model)
    uint256 public feeBps;

    // Split bps: 70% treasury, 20% coreSgf, 10% ipf (out of 10000)
    uint256 public treasurySplitBps;
    uint256 public coreSgfSplitBps;
    uint256 public ipfSplitBps;

    modifier onlyClearingHouse() {
        if (msg.sender != clearingHouse && !hasRole(CLEARING_HOUSE_ROLE, msg.sender)) {
            revert UnauthorizedClearingHouse(msg.sender);
        }
        _;
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address admin,
        address _clearingHouse,
        address _treasuryVault,
        address _coreSgfVault,
        address _ipfVault
    ) external initializer {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);

        clearingHouse = _clearingHouse;
        if (_clearingHouse != address(0)) {
            _grantRole(CLEARING_HOUSE_ROLE, _clearingHouse);
        }

        treasuryVault = _treasuryVault;
        coreSgfVault = _coreSgfVault;
        ipfVault = _ipfVault;

        feeBps = 0; // 0.00% at launch
        treasurySplitBps = 7000;
        coreSgfSplitBps = 2000;
        ipfSplitBps = 1000;
    }

    function setClearingHouse(address _clearingHouse) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (clearingHouse != address(0)) {
            _revokeRole(CLEARING_HOUSE_ROLE, clearingHouse);
        }
        clearingHouse = _clearingHouse;
        if (_clearingHouse != address(0)) {
            _grantRole(CLEARING_HOUSE_ROLE, _clearingHouse);
        }
    }

    function setFeeVaults(
        address _treasuryVault,
        address _coreSgfVault,
        address _ipfVault
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (_treasuryVault == address(0) || _coreSgfVault == address(0) || _ipfVault == address(0)) {
            revert InvalidFeeVaultAddress();
        }
        treasuryVault = _treasuryVault;
        coreSgfVault = _coreSgfVault;
        ipfVault = _ipfVault;
        emit FeeVaultsUpdated(_treasuryVault, _coreSgfVault, _ipfVault);
    }

    function setFeeParameters(
        uint256 _feeBps,
        uint256 _treasurySplitBps,
        uint256 _coreSgfSplitBps,
        uint256 _ipfSplitBps
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(_treasurySplitBps + _coreSgfSplitBps + _ipfSplitBps == 10000, "Split must equal 10000");
        feeBps = _feeBps;
        treasurySplitBps = _treasurySplitBps;
        coreSgfSplitBps = _coreSgfSplitBps;
        ipfSplitBps = _ipfSplitBps;
    }

    function computeTransactionFee(
        uint256 turnoverAmount
    ) public view returns (uint256 feeAmount, uint256 treasuryShare, uint256 coreSgfShare, uint256 ipfShare) {
        if (feeBps == 0 || turnoverAmount == 0) {
            return (0, 0, 0, 0);
        }
        feeAmount = (turnoverAmount * feeBps) / 10000;
        treasuryShare = (feeAmount * treasurySplitBps) / 10000;
        coreSgfShare = (feeAmount * coreSgfSplitBps) / 10000;
        ipfShare = feeAmount - treasuryShare - coreSgfShare;
    }

    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf) {
        return (treasuryVault, coreSgfVault, ipfVault);
    }

    function executeCashSettlement(
        bytes32 seriesId,
        address holder,
        uint256 contractAmount,
        uint256 intrinsicPayoffPerUnit,
        uint256 costBasisTotal,
        address settlementToken
    ) external onlyClearingHouse nonReentrant returns (uint256 netPayout, uint256 feeDeducted) {
        if (contractAmount == 0) revert ZeroContractsSettled();

        uint256 grossPayoff = (intrinsicPayoffPerUnit * contractAmount) / 1e18;
        (uint256 fee, uint256 treasuryShare, uint256 coreSgfShare, uint256 ipfShare) = computeTransactionFee(grossPayoff);

        feeDeducted = fee;
        netPayout = grossPayoff - fee;

        bytes32 taxProofHash = keccak256(
            abi.encode(seriesId, holder, contractAmount, grossPayoff, costBasisTotal, block.timestamp)
        );
        bytes32 executionId = keccak256(
            abi.encodePacked(seriesId, holder, contractAmount, block.timestamp, block.number)
        );

        if (netPayout > 0) {
            IERC20(settlementToken).safeTransfer(holder, netPayout);
        }

        if (fee > 0) {
            if (treasuryShare > 0) IERC20(settlementToken).safeTransfer(treasuryVault, treasuryShare);
            if (coreSgfShare > 0) IERC20(settlementToken).safeTransfer(coreSgfVault, coreSgfShare);
            if (ipfShare > 0) IERC20(settlementToken).safeTransfer(ipfVault, ipfShare);

            emit PlatformFeeCollected(seriesId, holder, grossPayoff, fee, treasuryVault, coreSgfVault, ipfVault);
        }

        emit CashSettlementExecuted(
            executionId,
            seriesId,
            holder,
            contractAmount,
            grossPayoff,
            feeDeducted,
            netPayout,
            taxProofHash
        );
    }

    function executePhysicalDelivery(
        bytes32 seriesId,
        address holder,
        address writer,
        uint256 contractAmount,
        uint256 strikePrice,
        IOptionsTokenFactory.OptionType optionType,
        uint256 costBasisTotal,
        address underlyingToken,
        address settlementToken
    ) external onlyClearingHouse nonReentrant returns (SettlementExecutionReceipt memory receipt) {
        if (contractAmount == 0) revert ZeroContractsSettled();

        uint256 cashExchanged = (strikePrice * contractAmount) / 1e18;
        uint256 underlyingDelivered = contractAmount;

        (uint256 fee, uint256 treasuryShare, uint256 coreSgfShare, uint256 ipfShare) = computeTransactionFee(cashExchanged);
        uint256 netCashToWriter = cashExchanged - fee;

        bytes32 taxProofHash = keccak256(
            abi.encode(seriesId, holder, writer, contractAmount, cashExchanged, costBasisTotal, block.timestamp)
        );
        bytes32 executionId = keccak256(
            abi.encodePacked(seriesId, holder, writer, contractAmount, block.timestamp, block.number)
        );

        if (optionType == IOptionsTokenFactory.OptionType.CALL) {
            // Holder pays strike cash -> writer gets netCashToWriter, settler/vault gets fee
            IERC20(settlementToken).safeTransferFrom(holder, address(this), cashExchanged);
            IERC20(settlementToken).safeTransfer(writer, netCashToWriter);

            // Holder receives underlying tokens from clearinghouse
            IERC20(underlyingToken).safeTransfer(holder, underlyingDelivered);
        } else {
            // PUT: Holder delivers underlying equity -> writer receives underlying equity
            IERC20(underlyingToken).safeTransferFrom(holder, writer, underlyingDelivered);

            // Holder receives strike cash from clearinghouse vault minus fee
            uint256 netCashToHolder = cashExchanged - fee;
            IERC20(settlementToken).safeTransfer(holder, netCashToHolder);
        }

        if (fee > 0) {
            if (treasuryShare > 0) IERC20(settlementToken).safeTransfer(treasuryVault, treasuryShare);
            if (coreSgfShare > 0) IERC20(settlementToken).safeTransfer(coreSgfVault, coreSgfShare);
            if (ipfShare > 0) IERC20(settlementToken).safeTransfer(ipfVault, ipfShare);

            emit PlatformFeeCollected(seriesId, holder, cashExchanged, fee, treasuryVault, coreSgfVault, ipfVault);
        }

        emit PhysicalSettlementExecuted(
            executionId,
            seriesId,
            holder,
            writer,
            contractAmount,
            underlyingDelivered,
            cashExchanged,
            fee,
            taxProofHash
        );

        receipt = SettlementExecutionReceipt({
            executionId: executionId,
            seriesId: seriesId,
            holder: holder,
            writer: writer,
            settlementType: IOptionsTokenFactory.SettlementType.PHYSICAL_DELIVERY,
            contractAmount: contractAmount,
            grossPayoutOrAssetAmount: cashExchanged,
            feeDeducted: fee,
            netDisbursed: (optionType == IOptionsTokenFactory.OptionType.CALL) ? netCashToWriter : cashExchanged - fee,
            taxProofHash: taxProofHash,
            timestamp: block.timestamp
        });
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    uint256[45] private __gap;
}
