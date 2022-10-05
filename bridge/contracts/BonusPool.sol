// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/interfaces/IStakingToken.sol";
import "contracts/BonusPool.sol";


// holds all AToken that is held in escrow for locking bonuses
contract BonusPool {
    // holds in a single staked position that is owned locally
    // determines payouts from _totalLocked on Locking
    // for each token locked at termination

    error EthSendFailure();
    address internal immutable _factory; // can be brought in from locking via reward and into here
    uint256 public constant tokenWeiPerToken = 1; // one token 10^-18 per token not valid but placeholder
    uint256 internal constant _unitOne = 10 ^ 18;

    constructor(address factory_) {
        _factory = factory_;
    }

    // func terminate
    // reap underlying position and all profits
    // take in total staked and tokenPerToken
    // calc percent of position owed to Locking
    // calc percent reward owed to Locking
    // send eth and tokens to Reward
    // invoke mintTo factory on remaining ALCA
    // send remaining eth to factory
    // self destruct

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