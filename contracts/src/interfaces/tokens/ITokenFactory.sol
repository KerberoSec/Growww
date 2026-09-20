// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title ITokenFactory
 * @notice Factory interface for deploying standardized upgradeable DigitalSecurityToken instances.
 */
interface ITokenFactory {
    event TokenDeployed(
        string indexed isin,
        address indexed tokenAddress,
        address indexed tokenProxy,
        string name,
        string symbol
    );

    function deploySecurityToken(
        string calldata name,
        string calldata symbol,
        string calldata isin,
        address identityRegistry,
        address compliance,
        address adminMultisig
    ) external returns (address proxyAddress);

    function getTokenByISIN(string calldata isin) external view returns (address);
}
