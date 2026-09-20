// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "./IOptionsTokenFactory.sol";

interface IPhysicalAndCashSettler {
    struct SettlementExecutionReceipt {
        bytes32 executionId;
        bytes32 seriesId;
        address holder;
        address writer;
        IOptionsTokenFactory.SettlementType settlementType;
        uint256 contractAmount;
        uint256 grossPayoutOrAssetAmount;
        uint256 feeDeducted;
        uint256 netDisbursed;
        bytes32 taxProofHash;
        uint256 timestamp;
    }

    // Events
    event CashSettlementExecuted(
        bytes32 indexed executionId,
        bytes32 indexed seriesId,
        address indexed holder,
        uint256 contractAmount,
        uint256 grossPayoff,
        uint256 feeDeducted,
        uint256 netPayout,
        bytes32 taxProofHash
    );
    event PhysicalSettlementExecuted(
        bytes32 indexed executionId,
        bytes32 indexed seriesId,
        address indexed holder,
        address writer,
        uint256 contractAmount,
        uint256 underlyingDelivered,
        uint256 cashExchanged,
        uint256 feeDeducted,
        bytes32 taxProofHash
    );
    event PlatformFeeCollected(
        bytes32 indexed seriesId,
        address indexed holder,
        uint256 turnoverAmount,
        uint256 feeAmount,
        address treasuryVault,
        address coreSgfVault,
        address ipfVault
    );
    event FeeVaultsUpdated(address indexed treasury, address indexed coreSgf, address indexed ipf);

    // Custom Errors
    error UnauthorizedClearingHouse(address caller);
    error InvalidFeeVaultAddress();
    error TransferFailed(address token, address from, address to, uint256 amount);
    error MathOverflowOrUnderflow();
    error ZeroContractsSettled();

    // Settlement Operations
    function executeCashSettlement(
        bytes32 seriesId,
        address holder,
        uint256 contractAmount,
        uint256 intrinsicPayoffPerUnit,
        uint256 costBasisTotal,
        address settlementToken
    ) external returns (uint256 netPayout, uint256 feeDeducted);

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
    ) external returns (SettlementExecutionReceipt memory receipt);

    // Mathematical Calculation Views
    function computeTransactionFee(
        uint256 turnoverAmount
    ) external view returns (uint256 feeAmount, uint256 treasuryShare, uint256 coreSgfShare, uint256 ipfShare);

    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf);
    function clearingHouse() external view returns (address);
}
