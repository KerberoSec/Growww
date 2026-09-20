// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";

import {ISecuritiesAMM} from "../interfaces/amm/ISecuritiesAMM.sol";
import {IIdentityRegistry} from "../interfaces/IIdentityRegistry.sol";
import {IComplianceRegistry} from "../interfaces/IComplianceRegistry.sol";
import {IDigitalSecurityToken} from "../interfaces/tokens/IDigitalSecurityToken.sol";

/**
 * @title SecuritiesAMM
 * @notice Constant-product (XYK) Automated Market Maker for ERC-3643 compliant digital securities.
 * @dev Enforces investor identity verification, sanction checks, and transfer compliance on all trading.
 */
contract SecuritiesAMM is ERC20, ReentrancyGuard, Pausable, Ownable, ISecuritiesAMM {
    using SafeERC20 for IERC20;

    // --- Constants ---
    uint256 public constant MINIMUM_LIQUIDITY = 1000;
    uint256 public constant MAX_FEE_BPS = 300; // Max 3.00%
    uint256 public constant BPS_DENOMINATOR = 10000;

    // --- Token Pair ---
    address public immutable override token0;
    address public immutable override token1;

    // --- Compliance Registries ---
    address public override identityRegistry;
    address public override complianceRegistry;

    // --- Fee Configuration ---
    uint256 public override feeBps;          // Total fee in BPS (e.g. 30 = 0.30%)
    uint256 public override protocolFeeBps;  // Protocol portion of fee (e.g. 5 = 0.05%)
    address public override feeCollector;

    // --- Reserves ---
    uint112 private _reserve0;
    uint112 private _reserve1;
    uint32 private _blockTimestampLast;

    modifier checkCompliance(address account) {
        _validateKYC(account);
        _;
    }

    constructor(
        address _tokenA,
        address _tokenB,
        address _identityRegistry,
        address _complianceRegistry,
        address _feeCollector,
        uint256 _initialFeeBps,
        uint256 _initialProtocolFeeBps,
        address initialOwner
    )
        ERC20("Growww Securities LP Token", "G-LP")
        Ownable(initialOwner)
    {
        if (_tokenA == address(0) || _tokenB == address(0)) revert InvalidZeroAddress();
        if (_tokenA == _tokenB) revert IdenticalTokens();
        if (_initialFeeBps > MAX_FEE_BPS) revert ExcessiveFeeBps(_initialFeeBps, MAX_FEE_BPS);
        if (_initialProtocolFeeBps > _initialFeeBps) revert ExcessiveFeeBps(_initialProtocolFeeBps, _initialFeeBps);

        if (_tokenA < _tokenB) {
            token0 = _tokenA;
            token1 = _tokenB;
        } else {
            token0 = _tokenB;
            token1 = _tokenA;
        }

        identityRegistry = _identityRegistry;
        complianceRegistry = _complianceRegistry;
        feeCollector = _feeCollector;
        feeBps = _initialFeeBps;
        protocolFeeBps = _initialProtocolFeeBps;
    }

    // =========================================================================
    // View Functions
    // =========================================================================

    function getReserves()
        external
        view
        override
        returns (uint256 reserve0, uint256 reserve1, uint32 blockTimestampLast)
    {
        reserve0 = _reserve0;
        reserve1 = _reserve1;
        blockTimestampLast = _blockTimestampLast;
    }

    function quote(
        uint256 amountA,
        uint256 reserveA,
        uint256 reserveB
    ) public pure override returns (uint256 amountB) {
        if (amountA == 0) revert InvalidAmount();
        if (reserveA == 0 || reserveB == 0) revert InsufficientLiquidityBurned();
        amountB = (amountA * reserveB) / reserveA;
    }

    function getAmountOut(
        uint256 amountIn,
        address tokenIn
    ) public view override returns (uint256 amountOut, uint256 fee) {
        if (amountIn == 0) revert InsufficientInputAmount();
        if (tokenIn != token0 && tokenIn != token1) revert InvalidZeroAddress();

        uint256 reserveIn = tokenIn == token0 ? _reserve0 : _reserve1;
        uint256 reserveOut = tokenIn == token0 ? _reserve1 : _reserve0;
        if (reserveIn == 0 || reserveOut == 0) revert InsufficientLiquidityBurned();

        fee = (amountIn * feeBps) / BPS_DENOMINATOR;
        uint256 amountInWithFee = amountIn - fee;
        uint256 numerator = amountInWithFee * reserveOut;
        uint256 denominator = reserveIn + amountInWithFee;
        amountOut = numerator / denominator;
    }

    // =========================================================================
    // Liquidity Management
    // =========================================================================

    function addLiquidity(
        uint256 amount0Desired,
        uint256 amount1Desired,
        uint256 amount0Min,
        uint256 amount1Min,
        address to,
        uint256 deadline
    )
        external
        override
        nonReentrant
        whenNotPaused
        checkCompliance(msg.sender)
        checkCompliance(to)
        returns (uint256 amount0, uint256 amount1, uint256 lpShares)
    {
        if (block.timestamp > deadline) revert DeadlineExpired(deadline, block.timestamp);
        if (to == address(0)) revert InvalidZeroAddress();

        (uint256 currentRes0, uint256 currentRes1) = (_reserve0, _reserve1);

        if (currentRes0 == 0 && currentRes1 == 0) {
            amount0 = amount0Desired;
            amount1 = amount1Desired;
        } else {
            uint256 amount1Optimal = quote(amount0Desired, currentRes0, currentRes1);
            if (amount1Optimal <= amount1Desired) {
                if (amount1Optimal < amount1Min) revert SlippageExceeded(amount1Min, amount1Optimal);
                amount0 = amount0Desired;
                amount1 = amount1Optimal;
            } else {
                uint256 amount0Optimal = quote(amount1Desired, currentRes1, currentRes0);
                assert(amount0Optimal <= amount0Desired);
                if (amount0Optimal < amount0Min) revert SlippageExceeded(amount0Min, amount0Optimal);
                amount0 = amount0Optimal;
                amount1 = amount1Desired;
            }
        }

        IERC20(token0).safeTransferFrom(msg.sender, address(this), amount0);
        IERC20(token1).safeTransferFrom(msg.sender, address(this), amount1);

        uint256 totalLpSupply = totalSupply();
        if (totalLpSupply == 0) {
            uint256 initialLiquidity = Math.sqrt(amount0 * amount1);
            if (initialLiquidity <= MINIMUM_LIQUIDITY) revert InsufficientLiquidityMinted();
            lpShares = initialLiquidity - MINIMUM_LIQUIDITY;
            _mint(address(0x000000000000000000000000000000000000dEaD), MINIMUM_LIQUIDITY);
        } else {
            lpShares = Math.min(
                (amount0 * totalLpSupply) / currentRes0,
                (amount1 * totalLpSupply) / currentRes1
            );
        }

        if (lpShares == 0) revert InsufficientLiquidityMinted();
        _mint(to, lpShares);

        _updateReserves(IERC20(token0).balanceOf(address(this)), IERC20(token1).balanceOf(address(this)));
        emit LiquidityAdded(msg.sender, amount0, amount1, lpShares);
    }

    function removeLiquidity(
        uint256 lpShares,
        uint256 amount0Min,
        uint256 amount1Min,
        address to,
        uint256 deadline
    )
        external
        override
        nonReentrant
        whenNotPaused
        checkCompliance(msg.sender)
        checkCompliance(to)
        returns (uint256 amount0, uint256 amount1)
    {
        if (block.timestamp > deadline) revert DeadlineExpired(deadline, block.timestamp);
        if (to == address(0)) revert InvalidZeroAddress();
        if (lpShares == 0) revert InvalidAmount();

        uint256 totalLpSupply = totalSupply();
        uint256 balance0 = IERC20(token0).balanceOf(address(this));
        uint256 balance1 = IERC20(token1).balanceOf(address(this));

        amount0 = (lpShares * balance0) / totalLpSupply;
        amount1 = (lpShares * balance1) / totalLpSupply;

        if (amount0 == 0 || amount1 == 0) revert InsufficientLiquidityBurned();
        if (amount0 < amount0Min) revert SlippageExceeded(amount0Min, amount0);
        if (amount1 < amount1Min) revert SlippageExceeded(amount1Min, amount1);

        _burn(msg.sender, lpShares);

        _safeTransferSecurity(token0, to, amount0);
        _safeTransferSecurity(token1, to, amount1);

        _updateReserves(IERC20(token0).balanceOf(address(this)), IERC20(token1).balanceOf(address(this)));
        emit LiquidityRemoved(msg.sender, amount0, amount1, lpShares, to);
    }

    // =========================================================================
    // Trading / Swaps
    // =========================================================================

    function swapExactTokensForTokens(
        uint256 amountIn,
        uint256 amountOutMin,
        address tokenIn,
        address to,
        uint256 deadline
    )
        external
        override
        nonReentrant
        whenNotPaused
        checkCompliance(msg.sender)
        checkCompliance(to)
        returns (uint256 amountOut)
    {
        if (block.timestamp > deadline) revert DeadlineExpired(deadline, block.timestamp);
        if (to == address(0)) revert InvalidZeroAddress();
        if (tokenIn != token0 && tokenIn != token1) revert InvalidZeroAddress();

        (uint256 calculatedOut, uint256 totalFee) = getAmountOut(amountIn, tokenIn);
        if (calculatedOut < amountOutMin) revert SlippageExceeded(amountOutMin, calculatedOut);
        amountOut = calculatedOut;

        address tokenOut = tokenIn == token0 ? token1 : token0;

        // Protocol fee routing
        if (protocolFeeBps > 0 && feeCollector != address(0)) {
            uint256 protocolFee = (amountIn * protocolFeeBps) / BPS_DENOMINATOR;
            if (protocolFee > 0) {
                IERC20(tokenIn).safeTransferFrom(msg.sender, feeCollector, protocolFee);
            }
            IERC20(tokenIn).safeTransferFrom(msg.sender, address(this), amountIn - protocolFee);
        } else {
            IERC20(tokenIn).safeTransferFrom(msg.sender, address(this), amountIn);
        }

        // Transfer output tokens to recipient
        _safeTransferSecurity(tokenOut, to, amountOut);

        uint256 balance0 = IERC20(token0).balanceOf(address(this));
        uint256 balance1 = IERC20(token1).balanceOf(address(this));

        // Enforce constant product invariant
        uint256 lpFeeBps = protocolFeeBps > 0 && feeCollector != address(0)
            ? (feeBps >= protocolFeeBps ? feeBps - protocolFeeBps : 0)
            : feeBps;

        uint256 balance0Adjusted = tokenIn == token0
            ? (balance0 * BPS_DENOMINATOR) - (amountIn * lpFeeBps)
            : balance0 * BPS_DENOMINATOR;
        uint256 balance1Adjusted = tokenIn == token1
            ? (balance1 * BPS_DENOMINATOR) - (amountIn * lpFeeBps)
            : balance1 * BPS_DENOMINATOR;

        if (balance0Adjusted * balance1Adjusted < uint256(_reserve0) * uint256(_reserve1) * (BPS_DENOMINATOR ** 2)) {
            revert KInvariantViolated(
                balance0Adjusted * balance1Adjusted,
                uint256(_reserve0) * uint256(_reserve1) * (BPS_DENOMINATOR ** 2)
            );
        }

        _updateReserves(balance0, balance1);
        emit Swap(msg.sender, to, tokenIn, amountIn, amountOut, totalFee);
    }

    function sync() external override nonReentrant {
        _updateReserves(IERC20(token0).balanceOf(address(this)), IERC20(token1).balanceOf(address(this)));
    }

    // =========================================================================
    // Admin Controls
    // =========================================================================

    function setFeeRate(uint256 newFeeBps, uint256 newProtocolFeeBps) external onlyOwner {
        if (newFeeBps > MAX_FEE_BPS) revert ExcessiveFeeBps(newFeeBps, MAX_FEE_BPS);
        if (newProtocolFeeBps > newFeeBps) revert ExcessiveFeeBps(newProtocolFeeBps, newFeeBps);
        emit FeeRateUpdated(feeBps, newFeeBps);
        feeBps = newFeeBps;
        protocolFeeBps = newProtocolFeeBps;
    }

    function setFeeCollector(address newFeeCollector) external onlyOwner {
        if (newFeeCollector == address(0)) revert InvalidZeroAddress();
        emit FeeCollectorUpdated(feeCollector, newFeeCollector);
        feeCollector = newFeeCollector;
    }

    function setIdentityRegistry(address newRegistry) external onlyOwner {
        emit IdentityRegistryUpdated(identityRegistry, newRegistry);
        identityRegistry = newRegistry;
    }

    function setComplianceRegistry(address newRegistry) external onlyOwner {
        complianceRegistry = newRegistry;
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    // =========================================================================
    // Internal Helpers
    // =========================================================================

    function _updateReserves(uint256 balance0, uint256 balance1) internal {
        require(balance0 <= type(uint112).max && balance1 <= type(uint112).max, "OVERFLOW");
        _reserve0 = uint112(balance0);
        _reserve1 = uint112(balance1);
        _blockTimestampLast = uint32(block.timestamp % 2**32);
        emit ReservesSynced(_reserve0, _reserve1);
    }

    function _validateKYC(address account) internal view {
        if (identityRegistry != address(0)) {
            IIdentityRegistry idReg = IIdentityRegistry(identityRegistry);
            if (!idReg.isVerified(account)) {
                revert CallerNotCompliant(account);
            }
            bytes32 identityId = idReg.getIdentityId(account);
            if (identityId != bytes32(0) && idReg.isSanctioned(identityId)) {
                revert CallerNotCompliant(account);
            }
        }
    }

    function _safeTransferSecurity(address token, address to, uint256 amount) internal {
        // If the token is a digital security token, check frozen status and compliance
        try IDigitalSecurityToken(token).isFrozen(to) returns (bool frozen) {
            if (frozen) revert RecipientNotCompliant(to);
        } catch {}

        if (complianceRegistry != address(0)) {
            try IComplianceRegistry(complianceRegistry).canTransfer(address(this), to, amount) returns (bool canTx) {
                if (!canTx) revert RecipientNotCompliant(to);
            } catch {}
        }

        IERC20(token).safeTransfer(to, amount);
    }
}
