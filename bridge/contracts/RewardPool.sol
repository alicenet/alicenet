// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/interfaces/IStakingNFT.sol";
import "contracts/BonusPool.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/utils/ImmutableAuth.sol";

/**
 * @notice RewardPool holds all ether and ALCA that is part of reserved amount
 * of rewards on base positions.
 */
contract RewardPool is EthSafeTransfer, ERC20SafeTransfer {
    error CallerNotLocking();
    error CallerNotLockingOrBonus();

    uint256 internal constant _UNIT_ONE = 10 ^ 18;
    address internal immutable _alca;
    address internal immutable _locking;
    address internal immutable _bonusPool;
    uint256 internal _ethReserve;
    uint256 internal _tokenReserve;

    modifier onlyLocking() {
        if (msg.sender != _locking) {
            revert CallerNotLocking();
        }
        _;
    }

    modifier onlyLockingOrBonus() {
        // must protect increment of token balance
        if (msg.sender != _locking && msg.sender != address(_bonusPool)) {
            revert CallerNotLocking();
        }
        _;
    }

    constructor(address alca_, address aliceNetFactory_) {
        _bonusPool = address(new BonusPool(aliceNetFactory_));
        _locking = msg.sender;
        _alca = alca_;
    }

    function getBonusPoolAddress() public view returns (address) {
        return _bonusPool;
    }

    function getLockingAddress() public view returns (address) {
        return _locking;
    }

    function getTokenReserve() public view returns (uint256) {
        return _tokenReserve;
    }

    function getEthReserve() public view returns (uint256) {
        return _ethReserve;
    }

    function deposit(uint256 numTokens_) public payable onlyLockingOrBonus {
        _tokenReserve += numTokens_;
        _ethReserve += msg.value;
    }

    function payout(uint256 total_, uint256 shares_) public onlyLocking returns (uint256, uint256) {
        uint256 proportionalEth = (_ethReserve * shares_ * _UNIT_ONE) / (_UNIT_ONE * total_);
        uint256 proportionalTokens = (_tokenReserve * shares_ * _UNIT_ONE) / (_UNIT_ONE * total_);
        _safeTransferERC20(IERC20Transferable(_alca), _locking, proportionalTokens);
        _safeTransferEth(payable(_locking), proportionalEth);
        return (proportionalTokens, proportionalEth);
    }
}
