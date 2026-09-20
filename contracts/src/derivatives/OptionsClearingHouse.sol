// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "../interfaces/derivatives/IOptionsClearingHouse.sol";
import "../interfaces/derivatives/IOptionsTokenFactory.sol";
import "../interfaces/derivatives/IPhysicalAndCashSettler.sol";

/**
 * @title OptionsClearingHouse
 * @notice Central on-chain clearinghouse and 100% full collateralization vault for options.
 * @dev Ingests EIP-712 signed oracle settlement prices and executes automated batch ITM settlement.
 */
contract OptionsClearingHouse is
    IOptionsClearingHouse,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    ReentrancyGuardUpgradeable,
    PausableUpgradeable,
    EIP712Upgradeable
{
    using SafeERC20 for IERC20;

    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    bytes32 public constant EMERGENCY_GUARDIAN_ROLE = keccak256("EMERGENCY_GUARDIAN_ROLE");
    bytes32 public constant ORACLE_RELAYER_ROLE = keccak256("ORACLE_RELAYER_ROLE");
    bytes32 public constant SETTLEMENT_OPERATOR_ROLE = keccak256("SETTLEMENT_OPERATOR_ROLE");

    bytes32 public constant SETTLEMENT_PRICE_TYPEHASH =
        keccak256("SettlementPriceData(bytes32 seriesId,uint256 settlementPrice,uint256 priceTimestamp,uint256 sequenceNumber)");

    IOptionsTokenFactory public optionsFactory;
    IPhysicalAndCashSettler public settler;

    uint256 public nextPositionId;
    mapping(uint256 => CollateralPosition) private _collateralPositions;
    mapping(bytes32 => uint256[]) private _seriesPositionIds;
    mapping(bytes32 => SettlementPriceData) private _settlementPrices;
    mapping(bytes32 => bool) private _priceLocked;
    mapping(address => bool) public isOracleRelayer;
    mapping(address => uint256) public override totalLockedCollateral;

    // Mapping of seriesId => total contracts exercised
    mapping(bytes32 => uint256) public totalContractsExercised;
    // Mapping of seriesId => total collateral payout disbursed
    mapping(bytes32 => uint256) public totalCollateralDisbursed;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address admin,
        address _optionsFactory,
        address _settler
    ) external initializer {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();
        __Pausable_init();
        __EIP712_init("GrowwwOptionsClearingHouse", "1.0.0");

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        _grantRole(EMERGENCY_GUARDIAN_ROLE, admin);
        _grantRole(SETTLEMENT_OPERATOR_ROLE, admin);

        optionsFactory = IOptionsTokenFactory(_optionsFactory);
        settler = IPhysicalAndCashSettler(_settler);
    }

    function setOptionsFactory(address _optionsFactory) external onlyRole(DEFAULT_ADMIN_ROLE) {
        optionsFactory = IOptionsTokenFactory(_optionsFactory);
    }

    function setSettler(address _settler) external onlyRole(DEFAULT_ADMIN_ROLE) {
        settler = IPhysicalAndCashSettler(_settler);
    }

    function setOracleRelayer(address relayer, bool whitelisted) external onlyRole(DEFAULT_ADMIN_ROLE) {
        isOracleRelayer[relayer] = whitelisted;
        if (whitelisted) {
            _grantRole(ORACLE_RELAYER_ROLE, relayer);
        } else {
            _revokeRole(ORACLE_RELAYER_ROLE, relayer);
        }
        emit OracleRelayerUpdated(relayer, whitelisted);
    }

    function pause() external onlyRole(EMERGENCY_GUARDIAN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    function writeAndMintOption(
        bytes32 seriesId,
        uint256 contractAmount
    ) external whenNotPaused nonReentrant returns (uint256 positionId) {
        if (contractAmount == 0) revert InsufficientCollateralProvided(1, 0);

        IOptionsTokenFactory.OptionSeriesParams memory series = optionsFactory.getSeriesParams(seriesId);
        if (series.state != IOptionsTokenFactory.SeriesState.ACTIVE) {
            revert InactiveOptionSeries(seriesId);
        }

        address collateralToken;
        uint256 requiredCollateral;

        if (series.optionType == IOptionsTokenFactory.OptionType.CALL) {
            collateralToken = series.underlyingToken;
            requiredCollateral = contractAmount;
        } else {
            collateralToken = series.settlementToken;
            requiredCollateral = (series.strikePrice * contractAmount) / 1e18;
        }

        // Pull 100% full collateral from writer into clearinghouse vault
        IERC20(collateralToken).safeTransferFrom(msg.sender, address(this), requiredCollateral);
        totalLockedCollateral[collateralToken] += requiredCollateral;

        positionId = ++nextPositionId;
        _collateralPositions[positionId] = CollateralPosition({
            positionId: positionId,
            seriesId: seriesId,
            writer: msg.sender,
            collateralToken: collateralToken,
            lockedAmount: requiredCollateral,
            mintedContracts: contractAmount,
            unexercisedContracts: contractAmount,
            isReleased: false
        });

        _seriesPositionIds[seriesId].push(positionId);

        // Mint ERC-1155 option tokens to the writer
        uint256 tokenId = optionsFactory.computeTokenId(
            series.underlyingToken,
            series.optionType,
            series.strikePrice,
            series.expiryTimestamp,
            series.settlementType
        );

        optionsFactory.mintOption(msg.sender, tokenId, contractAmount);

        emit CollateralDepositedAndOptionMinted(
            positionId,
            seriesId,
            msg.sender,
            collateralToken,
            requiredCollateral,
            contractAmount
        );
    }

    function submitSettlementPrice(
        bytes32 seriesId,
        SettlementPriceData calldata data,
        bytes calldata signature
    ) external whenNotPaused {
        IOptionsTokenFactory.OptionSeriesParams memory series = optionsFactory.getSeriesParams(seriesId);
        if (series.state != IOptionsTokenFactory.SeriesState.ACTIVE) {
            revert InactiveOptionSeries(seriesId);
        }
        if (block.timestamp < series.expiryTimestamp) {
            revert ExpiryNotReached(block.timestamp, series.expiryTimestamp);
        }
        if (_priceLocked[seriesId]) {
            revert ExpiryAlreadyProcessed(seriesId);
        }
        if (data.priceTimestamp > block.timestamp) {
            revert StaleOraclePrice(data.priceTimestamp, block.timestamp);
        }
        if (block.timestamp - data.priceTimestamp > 1 days) {
            revert StaleOraclePrice(data.priceTimestamp, block.timestamp);
        }

        // Verify EIP-712 signature
        bytes32 structHash = keccak256(
            abi.encode(
                SETTLEMENT_PRICE_TYPEHASH,
                data.seriesId,
                data.settlementPrice,
                data.priceTimestamp,
                data.sequenceNumber
            )
        );
        bytes32 digest = _hashTypedDataV4(structHash);
        address recoveredSigner = ECDSA.recover(digest, signature);

        if (!isOracleRelayer[recoveredSigner] && !hasRole(ORACLE_RELAYER_ROLE, msg.sender)) {
            revert OracleNotWhitelisted(recoveredSigner);
        }

        _settlementPrices[seriesId] = data;
        _priceLocked[seriesId] = true;

        (bool isITM, ) = _isITM(series, data.settlementPrice);
        if (isITM) {
            optionsFactory.updateSeriesState(seriesId, IOptionsTokenFactory.SeriesState.EXPIRED_PRICE_LOCKED);
        } else {
            optionsFactory.updateSeriesState(seriesId, IOptionsTokenFactory.SeriesState.EXPIRED_WORTHLESS_OTM);
        }

        emit SettlementPriceSubmitted(seriesId, data.settlementPrice, data.priceTimestamp, recoveredSigner);
    }

    function batchAutoExerciseAndSettle(
        BatchExerciseParams calldata params
    ) external whenNotPaused nonReentrant {
        uint256 length = params.holders.length;
        if (length != params.contractAmounts.length || length != params.costBases.length) {
            revert ArrayLengthMismatch();
        }

        IOptionsTokenFactory.OptionSeriesParams memory series = optionsFactory.getSeriesParams(params.seriesId);
        if (!_priceLocked[params.seriesId]) {
            revert InactiveOptionSeries(params.seriesId);
        }

        (bool itm, uint256 intrinsicPayoff) = _isITM(series, _settlementPrices[params.seriesId].settlementPrice);
        if (!itm) {
            revert OptionNotITM(params.seriesId, _settlementPrices[params.seriesId].settlementPrice, series.strikePrice);
        }

        uint256 tokenId = optionsFactory.computeTokenId(
            series.underlyingToken,
            series.optionType,
            series.strikePrice,
            series.expiryTimestamp,
            series.settlementType
        );

        for (uint256 i = 0; i < length; i++) {
            address holder = params.holders[i];
            uint256 amount = params.contractAmounts[i];
            uint256 costBasis = params.costBases[i];

            if (amount == 0) continue;

            // Burn option tokens from the holder
            optionsFactory.burnOption(holder, tokenId, amount);
            totalContractsExercised[params.seriesId] += amount;

            if (series.settlementType == IOptionsTokenFactory.SettlementType.CASH_SETTLED) {
                uint256 grossPayoff = (intrinsicPayoff * amount) / 1e18;
                if (series.optionType == IOptionsTokenFactory.OptionType.PUT) {
                    if (totalLockedCollateral[series.settlementToken] >= grossPayoff) {
                        totalLockedCollateral[series.settlementToken] -= grossPayoff;
                    }
                } else {
                    if (totalLockedCollateral[series.underlyingToken] >= amount) {
                        totalLockedCollateral[series.underlyingToken] -= amount;
                    }
                }
                totalCollateralDisbursed[params.seriesId] += grossPayoff;

                IERC20(series.settlementToken).safeTransfer(address(settler), grossPayoff);
                (uint256 netPayout, uint256 feeDeducted) = settler.executeCashSettlement(
                    params.seriesId,
                    holder,
                    amount,
                    intrinsicPayoff,
                    costBasis,
                    series.settlementToken
                );

                emit AutomatedExerciseExecuted(params.seriesId, holder, amount, netPayout, feeDeducted);
            } else {
                // Physical settlement
                address writer = address(this); // Settler interacts with vault tokens
                if (series.optionType == IOptionsTokenFactory.OptionType.CALL) {
                    IERC20(series.underlyingToken).safeTransfer(address(settler), amount);
                    totalLockedCollateral[series.underlyingToken] -= amount;
                } else {
                    uint256 cashReq = (series.strikePrice * amount) / 1e18;
                    IERC20(series.settlementToken).safeTransfer(address(settler), cashReq);
                    totalLockedCollateral[series.settlementToken] -= cashReq;
                }

                settler.executePhysicalDelivery(
                    params.seriesId,
                    holder,
                    writer,
                    amount,
                    series.strikePrice,
                    series.optionType,
                    costBasis,
                    series.underlyingToken,
                    series.settlementToken
                );

                emit AutomatedExerciseExecuted(params.seriesId, holder, amount, amount, 0);
            }
        }
    }

    function reclaimOTMCollateral(uint256 positionId) external nonReentrant {
        CollateralPosition storage pos = _collateralPositions[positionId];
        if (pos.isReleased) revert PositionAlreadyReleased(positionId);
        if (msg.sender != pos.writer && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }

        IOptionsTokenFactory.OptionSeriesParams memory series = optionsFactory.getSeriesParams(pos.seriesId);
        if (series.state != IOptionsTokenFactory.SeriesState.EXPIRED_WORTHLESS_OTM) {
            revert InactiveOptionSeries(pos.seriesId);
        }

        pos.isReleased = true;
        uint256 releaseAmount = pos.lockedAmount;
        totalLockedCollateral[pos.collateralToken] -= releaseAmount;

        IERC20(pos.collateralToken).safeTransfer(pos.writer, releaseAmount);

        emit CollateralUnlocked(positionId, pos.seriesId, pos.writer, releaseAmount);
    }

    function reclaimResidualCollateral(uint256 positionId) external nonReentrant {
        CollateralPosition storage pos = _collateralPositions[positionId];
        if (pos.isReleased) revert PositionAlreadyReleased(positionId);
        if (msg.sender != pos.writer && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }

        IOptionsTokenFactory.OptionSeriesParams memory series = optionsFactory.getSeriesParams(pos.seriesId);
        if (!_priceLocked[pos.seriesId]) {
            revert InactiveOptionSeries(pos.seriesId);
        }

        (, uint256 intrinsicPayoff) = _isITM(series, _settlementPrices[pos.seriesId].settlementPrice);

        uint256 payoffPerContract;
        if (series.settlementType == IOptionsTokenFactory.SettlementType.CASH_SETTLED) {
            payoffPerContract = intrinsicPayoff;
        } else {
            payoffPerContract = (series.optionType == IOptionsTokenFactory.OptionType.CALL) ? 1e18 : series.strikePrice;
        }

        uint256 totalObligation;
        if (series.optionType == IOptionsTokenFactory.OptionType.CALL && series.settlementType == IOptionsTokenFactory.SettlementType.CASH_SETTLED) {
            // Payoff in settlementToken, but writer locked underlyingToken
            // If underlying price is known, residual underlying is unlocked
            totalObligation = 0;
        } else if (series.optionType == IOptionsTokenFactory.OptionType.PUT) {
            totalObligation = (payoffPerContract * pos.mintedContracts) / 1e18;
        }

        uint256 residualAmount = 0;
        if (pos.lockedAmount > totalObligation) {
            residualAmount = pos.lockedAmount - totalObligation;
        }

        pos.isReleased = true;
        if (residualAmount > 0) {
            totalLockedCollateral[pos.collateralToken] -= residualAmount;
            IERC20(pos.collateralToken).safeTransfer(pos.writer, residualAmount);
            emit CollateralUnlocked(positionId, pos.seriesId, pos.writer, residualAmount);
        }
    }

    function getCollateralPosition(uint256 positionId) external view returns (CollateralPosition memory) {
        return _collateralPositions[positionId];
    }

    function getSettlementPrice(bytes32 seriesId) external view returns (uint256 price, uint256 timestamp, bool isLocked) {
        SettlementPriceData memory data = _settlementPrices[seriesId];
        return (data.settlementPrice, data.priceTimestamp, _priceLocked[seriesId]);
    }

    function isOptionITM(bytes32 seriesId) external view returns (bool isITM, uint256 intrinsicPayoffPerUnit) {
        IOptionsTokenFactory.OptionSeriesParams memory series = optionsFactory.getSeriesParams(seriesId);
        return _isITM(series, _settlementPrices[seriesId].settlementPrice);
    }

    function _isITM(
        IOptionsTokenFactory.OptionSeriesParams memory series,
        uint256 settlementPrice
    ) internal pure returns (bool isITM, uint256 intrinsicPayoffPerUnit) {
        if (series.optionType == IOptionsTokenFactory.OptionType.CALL) {
            if (settlementPrice > series.strikePrice) {
                return (true, settlementPrice - series.strikePrice);
            }
        } else {
            if (settlementPrice < series.strikePrice) {
                return (true, series.strikePrice - settlementPrice);
            }
        }
        return (false, 0);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    uint256[45] private __gap;
}
