// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IComplianceRegistry
 * @notice Interface for modular compliance registry orchestrating transfer validation modules.
 */
interface IComplianceRegistry {
    event ModuleBound(address indexed moduleAddress);
    event ModuleUnbound(address indexed moduleAddress);

    error ModuleAlreadyBound(address moduleAddress);
    error ModuleNotBound(address moduleAddress);
    error InvalidModuleAddress();

    function canTransfer(address from, address to, uint256 amount) external view returns (bool);
    function transferred(address from, address to, uint256 amount) external;
    function created(address to, uint256 amount) external;
    function destroyed(address from, uint256 amount) external;

    function bindModule(address moduleAddress) external;
    function unbindModule(address moduleAddress) external;
    function getModules() external view returns (address[] memory);
}

/**
 * @title IComplianceModule
 * @notice Pluggable compliance module interface for specific transfer restrictions.
 */
interface IComplianceModule {
    function moduleCheck(address from, address to, uint256 amount, address complianceRegistry) external view returns (bool);
    function moduleTransferAction(address from, address to, uint256 amount) external;
    function moduleMintAction(address to, uint256 amount) external;
    function moduleBurnAction(address from, uint256 amount) external;
}
