// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/interfaces/IStakingToken.sol";
import "contracts/BonusPool.sol";
// holds all Eth that is part of reserved amount of rewards
// on base positions
// holds all AToken that is part of reserved amount of rewards
// on base positions
contract RewardPool {
    error CallerNotLocking();
    error CallerNotLockingOrBonus();
    error EthSendFailure();

    uint256 internal constant _unitOne = 10 ^ 18;

    uint256 _tokenBalance;
    uint256 _ethBalance;
    IStakingToken internal immutable _alca;
    address internal immutable _locking;
    BonusPool internal immutable _bonusPool;

    constructor(address alca_) {
        BonusPool bp = new BonusPool(msg.sender);
        _bonusPool = bp;
        _locking = msg.sender;
        IStakingToken st = IStakingToken(alca_);
        _alca = st;
    }

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

    function tokenBalance() public view returns (uint256) {
        return _tokenBalance;
    }

    function ethBalance() public view returns(uint256) {
        return _ethBalance;
    }

    function deposit(uint256 numTokens_) public payable onlyLockingOrBonus {
        _tokenBalance += numTokens_;
        _ethBalance += msg.value;
    }

    function payout(uint256 total_, uint256 shares_) public onlyLocking returns (uint256, uint256) {
        uint256 pe = (_ethBalance * shares_ * _unitOne) / (_unitOne * total_);
        uint256 pt = (_tokenBalance * shares_ * _unitOne) / (_unitOne * total_);
        _alca.transfer(_locking, pt);
        _safeSendEth(payable(_locking), pe);
        return (pt, pe);
    }

    function _safeSendEth(address payable acct_, uint256 val_) internal {
        if (val_ == 0) {
            return;
        }
        bool ok;
        (ok, ) = acct_.call{value: val_}("");
        if (!ok) {
            revert EthSendFailure();
        }
    }
}