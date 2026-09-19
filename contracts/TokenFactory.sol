// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import "./DigitalSecurityToken.sol";

/**
 * @title TokenFactory
 * @dev Factory contract for issuing regulatory-compliant Digital Security Tokens (DST)
 */
contract TokenFactory {
    address public immutable admin;
    address public complianceRegistry;

    address[] public deployedTokens;
    mapping(string => address) public tokenByISIN;

    event SecurityTokenCreated(string indexed isin, address tokenAddress, string name, string symbol, uint256 supply);

    error Unauthorized();
    error TokenAlreadyExists(string isin);

    constructor(address _complianceRegistry) {
        admin = msg.sender;
        complianceRegistry = _complianceRegistry;
    }

    function createToken(
        string memory isin,
        string memory name,
        string memory symbol,
        uint8 decimals,
        uint256 initialSupply,
        address tokenOwner
    ) external returns (address) {
        if (msg.sender != admin) revert Unauthorized();
        if (tokenByISIN[isin] != address(0)) revert TokenAlreadyExists(isin);

        DigitalSecurityToken token = new DigitalSecurityToken(
            name,
            symbol,
            decimals,
            initialSupply,
            complianceRegistry,
            tokenOwner
        );

        address tokenAddr = address(token);
        tokenByISIN[isin] = tokenAddr;
        deployedTokens.push(tokenAddr);

        emit SecurityTokenCreated(isin, tokenAddr, name, symbol, initialSupply);
        return tokenAddr;
    }

    function totalTokens() external view returns (uint256) {
        return deployedTokens.length;
    }
}
