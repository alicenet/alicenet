// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/utils/EthSafeTransfer.sol";

/**
 * @notice This contract holds all ALCA that is held in escrow for locking
 * bonuses. All ALCA is hold into a single staked position that is owned
 * locally.
 */
contract BonusPool is ImmutableAToken, ImmutablePublicStaking, EthSafeTransfer, ERC721Holder {

    error BonusTokenAlreadyCreated();

    // one token 10^-18 per token not valid but placeholder
    uint256 public constant TOKEN_WEI_PER_TOKEN = 1;
    uint256 internal constant _UNIT_ONE = 10 ^ 18;
    uint256 internal immutable _bonusAmount;
    uint256 ORIGINALSTAKEDAMOUNT;
    uint256 internal _tokenID;

    constructor(address aliceNetFactory_, uint256 bonusAmount_)
        ImmutableFactory(aliceNetFactory_)
        ImmutableAToken()
        ImmutablePublicStaking()
    {
        _bonusAmount = bonusAmount_;
    }

    function createBonusStakedPosition() public onlyFactory {
        if (_tokenID != 0) {
            revert BonusTokenAlreadyCreated();
        }

    }

    // TODO: function to stake or receive position with the the amount of ALCA

    // TODO: determines payouts from _totalLocked on Locking for each token locked
    // at termination

    function terminate(uint256 total) onlyRewards {
        payoutToken, payoutEth = IPublicStaking(_publicStakingAddress).burn()
        // reap underlying position and all profits
        // take in total staked and tokenPerToken
        PORTIONALCA = totalALCA SHARES LOCKED * TOKEN_WEI_PER_TOKEN / _UNIT_ONE // THIS IS PART WE PAY BACK
        BONUSALCA = PAYOUTTOKEN - ORIGINALSTAKEDAMOUNT // PORIONATE AMOUNT OF THIS BACK
        GOES TO FACTORY ALCA AMOUNT = ORIGINALSTAKEEDAMOUNT - PORTIONALCA

        // calc percent of position owed to Locking
        // calc percent reward owed to Locking
        // send eth and tokens to Reward
        // invoke mintTo factory on remaining ALCA
        // send remaining eth to factory
        // self destruct
    }
}
