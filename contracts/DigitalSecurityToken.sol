// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

/**
 * @title DigitalSecurityToken
 * @dev ERC-3643 / ERC-20 compliant permissioned token for Real World Assets (RWA)
 * under SEBI / IFSCA regulatory sandbox frameworks.
 */
contract DigitalSecurityToken {
    string public name;
    string public symbol;
    uint8 public immutable decimals;
    uint256 public totalSupply;
    address public immutable factory;
    address public complianceRegistry;
    bool public isPaused;

    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;
    mapping(address => bool) public isKYCVerified;
    mapping(address => bool) public isFrozen;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event IdentityVerified(address indexed investor);
    event AddressFrozen(address indexed target, bool frozen);

    error Unauthorized();
    error TransferPaused();
    error IdentityNotVerified(address investor);
    error AccountFrozen(address account);
    error InsufficientBalance();
    error InsufficientAllowance();

    constructor(
        string memory _name,
        string memory _symbol,
        uint8 _decimals,
        uint256 _initialSupply,
        address _complianceRegistry,
        address _owner
    ) {
        name = _name;
        symbol = _symbol;
        decimals = _decimals;
        factory = msg.sender;
        complianceRegistry = _complianceRegistry;
        
        isKYCVerified[_owner] = true;
        balanceOf[_owner] = _initialSupply;
        totalSupply = _initialSupply;
        emit Transfer(address(0), _owner, _initialSupply);
    }

    modifier onlyCompliant(address from, address to) {
        if (isPaused) revert TransferPaused();
        if (isFrozen[from] || isFrozen[to]) revert AccountFrozen(from);
        if (!isKYCVerified[from] && from != address(0)) revert IdentityNotVerified(from);
        if (!isKYCVerified[to] && to != address(0)) revert IdentityNotVerified(to);
        _;
    }

    function setKYCStatus(address investor, bool status) external {
        if (msg.sender != factory && msg.sender != complianceRegistry) revert Unauthorized();
        isKYCVerified[investor] = status;
        emit IdentityVerified(investor);
    }

    function setFreezeStatus(address account, bool frozen) external {
        if (msg.sender != factory && msg.sender != complianceRegistry) revert Unauthorized();
        isFrozen[account] = frozen;
        emit AddressFrozen(account, frozen);
    }

    function setPaused(bool _paused) external {
        if (msg.sender != factory && msg.sender != complianceRegistry) revert Unauthorized();
        isPaused = _paused;
    }

    function transfer(address to, uint256 amount) external onlyCompliant(msg.sender, to) returns (bool) {
        if (balanceOf[msg.sender] < amount) revert InsufficientBalance();
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        emit Transfer(msg.sender, to, amount);
        return true;
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        allowance[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external onlyCompliant(from, to) returns (bool) {
        if (balanceOf[from] < amount) revert InsufficientBalance();
        if (allowance[from][msg.sender] < amount) revert InsufficientAllowance();
        allowance[from][msg.sender] -= amount;
        balanceOf[from] -= amount;
        balanceOf[to] += amount;
        emit Transfer(from, to, amount);
        return true;
    }
}
