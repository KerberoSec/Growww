// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {ITokenFactory} from "../interfaces/tokens/ITokenFactory.sol";
import {DigitalSecurityToken} from "./DigitalSecurityToken.sol";

/**
 * @title TokenFactory
 * @notice Factory deploying standardized upgradeable DigitalSecurityToken instances for new equity ISINs.
 */
contract TokenFactory is AccessControl, ITokenFactory {
    bytes32 public constant DEPLOYER_ROLE = keccak256("DEPLOYER_ROLE");

    address public immutable tokenImplementation;
    mapping(string => address) private _tokensByISIN;

    error TokenAlreadyExists(string isin, address existingToken);
    error InvalidZeroAddress();

    constructor(address tokenImplementation_, address admin) {
        if (tokenImplementation_ == address(0) || admin == address(0)) revert InvalidZeroAddress();
        tokenImplementation = tokenImplementation_;
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(DEPLOYER_ROLE, admin);
    }

    /**
     * @notice Deploys a new upgradeable DigitalSecurityToken instance for a specified ISIN.
     */
    function deploySecurityToken(
        string calldata name,
        string calldata symbol,
        string calldata isin,
        address identityRegistry,
        address compliance,
        address adminMultisig
    ) external override onlyRole(DEPLOYER_ROLE) returns (address proxyAddress) {
        if (_tokensByISIN[isin] != address(0)) {
            revert TokenAlreadyExists(isin, _tokensByISIN[isin]);
        }
        if (adminMultisig == address(0)) revert InvalidZeroAddress();

        bytes memory initData = abi.encodeWithSelector(
            DigitalSecurityToken.initialize.selector,
            name,
            symbol,
            18,
            isin,
            identityRegistry,
            compliance,
            adminMultisig
        );

        ERC1967Proxy proxy = new ERC1967Proxy(tokenImplementation, initData);
        proxyAddress = address(proxy);

        _tokensByISIN[isin] = proxyAddress;

        emit TokenDeployed(isin, tokenImplementation, proxyAddress, name, symbol);
    }

    /**
     * @notice Retrieves the deployed token proxy address for a given ISIN string.
     */
    function getTokenByISIN(string calldata isin) external view override returns (address) {
        return _tokensByISIN[isin];
    }
}
