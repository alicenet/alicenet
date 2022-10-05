// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/interfaces/IStakingToken.sol";
import "contracts/interfaces/IStakingNFT.sol";
import "contracts/BonusPool.sol";
import "contracts/RewardPool.sol";

// TODO PREVENT ACCEPTING POSITIONS THAT HAVE WITHDRAW LOCK ENFORCED ON PROFITS
// UNLESS THEY ARE LESS THAN THE LOCKING DURATION AWAY OR THIS WILL CAUSE A
// FAILURE IN THE EXIT LOGIC
//
//ASSUMPTIONS:
// NO LOCAL ACCUMULATORS
// ONLY ONE POSITION PER ACCOUNT
// EARLY EXIT ALLOWED WITH PENALTY
//
// STATE MODEL
//
// PRELOCK  
//          INITIAL STATE AT DEPLOY
//          ENROLLMENT PERIOD OF CONTRACT
//          THIS IS THE ONLY PHASE IN WHICH NEW LOCKED POSITIONS MAY BE FORMED
//          CAN NOT USE FINAL PAYOUT METHODS
//          REWARDS ARE ALLOWED TO BE COLLECTED 80/20 SPLIT
//
// INLOCK 
//          TRANSITION VIA BLOCK NUMBER
//          REWARDS ARE ALLOWED TO BE COLLECTED 80/20 SPLIT
//          NO NEW NEW POSITIONS
//          PARTIAL WITHDRAW ALLOWED
//          CAN NOT USE FINAL PAYOUT METHODS
//
// POSTLOCK PAYOUTUNSAFE
//          TRANSITION VIA BLOCK NUMBER
//          NO NEW ENROLLMENTS
//          NO PARTIAL WITHDRAWS
//          CAN NOT USE FINAL PAYOUT METHODS
//          REWARDS ARE NOT ALLOWED TO BE COLLECTED
//
// POSTLOCK PAYOUTSAFE
//          TRANSITION VIA COMPLETION OF AGGREGATEPROFITS METHOD CALLS
//          ONLY METHOD ALLOWED IS THE AGGREGATEPROFITS METHOD CALL
//          SHOULD PAYOUT PROPORTIONATE REWARDS IN FULL TO CALLERS OF UNLOCK
//          SHOULD TRANSFER BACK TO CALLER POSSESSION OF STAKED POSITION NFT DURING METHOD UNLOCK
contract Locking {
    event EarlyExit(address to_, uint256 tokenID_);
    event NewLockup(address from_, uint256 tokenID_);

    error ContractDoesNotOwnTokenID(uint256 tokenID_);
    error TokenIDAlreadyClaimed(uint256 tokenID_);
    error MultipleLocksFromSingleAccount(uint256 exisingTokenID_);
    error EthSendFailure();
    error TokenSendFailure();
    error InsufficientALCAForEarlyExit();
    error RegistrationOver();
    error LockupOver();
    error UserHasNoPosition();
    error PreLockStateRequired();
    error PostLockStateNotAllowed();
    error PostLockStateRequired();
    error PayoutUnsafe();
    error AddressHasNotPositionLinked();

    uint8 internal constant _statePrelock = 0;
    uint8 internal constant _stateInLock = 1;
    uint8 internal constant _statePostLock = 2;

    uint256 internal constant _unitOne = 10 ^ 18;
    uint256 internal constant _fractionReserved = _unitOne / 5;

    // TODO: replace with immut lib where possible
    IStakingNFT internal immutable _publicStaking;
    RewardPool internal immutable _rewardPool;
    IStakingToken internal immutable _alca;

    // Total Locked describes the total number of ALCA
    // in this contract. Since no accumulators are used
    // this is tracked to allow proportionate payouts.
    uint256 internal _totalLocked;

    // _ownerOf tracks who is the owner of a tokenID
    // locked in this contract
    mapping(uint256 => address) internal _ownerOf;
    // _tokenOf is the inverse of ownerOf and returns the
    // owner given the tokenID
    // users are only allowed 1 position per account
    mapping(address => uint256) internal _tokenOf;
    
    // maps and index to a tokenID for iterable counting
    // stop iterating when token id is zero
    // must use tail insert to delete or else pagination will
    // end early
    mapping(uint256 => uint256) internal _tokenIDs;
    // lookup index by ID
    mapping(uint256 => uint256) internal _reverseTokenIDs;
    // tracks the number of tokenIDs this contract holds
    uint256 internal _lenTokenIDs;

    // payout amounts for all users to prevent the
    mapping(address => uint256) public rewardEth;
    // send of tokens from blocking mass exit
    // only written in post lock unsafe payout phase
    // only read during final cash out by users
    mapping(address => uint256) public rewardTokens;

    // block on which lock starts
    uint256 public immutable startBlock;
    // block on which lock ends
    uint256 public immutable endBlock;

    // determines if payout logic may be triggered
    // all profits must be collected for all accounts
    // first
    bool public payoutSafe;

    // offset for pagination of token profits
    // stored here so many people may call in
    // parellel and still do useful work
    uint256 internal _tokenIDOffset;

    constructor(
        address alca_,
        address publicStaking_,
        uint256 startBlock_,
        uint256 lockDuration_
    ) {
        // give infinite approval to bonuspool for alca
        _publicStaking = IStakingNFT(publicStaking_);
        RewardPool rp = new RewardPool(alca_);
        _rewardPool = rp;
        IStakingToken st = IStakingToken(alca_);
        _alca = st;
        startBlock = startBlock_;
        endBlock = startBlock_ + lockDuration_;
    }

    modifier onlyPreLock() {
        if (_getState() != _statePrelock) {
            revert PreLockStateRequired();
        }
        _;
    }

    modifier onlyPostLock() {
        if (_getState() != _statePostLock) {
            revert PostLockStateRequired();
        }
        _;
    }

    modifier onlyBeforePostLock() {
        if (_getState() == _statePostLock) {
            revert PostLockStateNotAllowed();
        }
        _;
    }

    modifier onlyPayoutSafe() {
        if (!payoutSafe) {
            revert PayoutUnsafe();
        }
        _;
    }

    modifier onlyPayoutUnSafe() {
        if (!payoutSafe) {
            revert PayoutUnsafe();
        }
        _;
    }

    function ownerOf(uint256 tokenID_) public view returns (address payable) {
        return payable(_ownerOf[tokenID_]);
    }

    function tokenOf(address acct_) public view returns (uint256) {
        return _tokenOf[acct_];
    }

    function lockFromApproval(uint256 tokenID_) public onlyPreLock {
        _unclaimedOrRevert(tokenID_);
        _lock(tokenID_, msg.sender);
        // interact last
        _publicStaking.transferFrom(msg.sender, address(this), tokenID_);
        // check as post condition
        _lockingOwnsOrRevert(tokenID_);
    }

    function lockFromTransfer(uint256 tokenID_, address tokenOwner_) public onlyPreLock {
        _unclaimedOrRevert(tokenID_);
        _lockingOwnsOrRevert(tokenID_);
        _lock(tokenID_, tokenOwner_);
    }

    function collectAllProfits()
        public
        onlyBeforePostLock
        returns (uint256 payoutAToken, uint256 payoutEth)
    {
        uint256 tokenID = tokenOf(msg.sender);
        if (tokenID == 0) {
            revert UserHasNoPosition();
        }
        return _collectAllProfits(_payableSender(), tokenID);
    }

    function unlockEarly(uint256 exitValue_)
        public
        onlyBeforePostLock
        returns (uint256 payoutEth, uint256 payoutAToken)
    {
        // get tokenID of caller
        uint256 tokenID = tokenOf(msg.sender);
        if (tokenID == 0) {
            revert UserHasNoPosition();
        }
        // get the number of shares and check validity
        uint256 shares = _getNumShares(tokenID);
        if (shares < exitValue_) {
            revert InsufficientALCAForEarlyExit();
        }
        // burn the existing position
        (payoutEth, payoutAToken) = _publicStaking.burn(tokenID);

        // blank old record
        _ownerOf[tokenID] = address(0);
        // create placeholder
        uint256 newTokenID;
        // find shares delta and mint new position
        uint256 remainingShares = shares - exitValue_;
        if (exitValue_ < shares) {
            // burn profits contain staked position... so sub it out
            payoutAToken = payoutAToken - shares;
            newTokenID = _publicStaking.mint(remainingShares);
            // set new records
            _ownerOf[newTokenID] = msg.sender;
            _replaceTokenID(tokenID, newTokenID);
        } else {
            _delTokenID(tokenID);
            // set new records
            // TODO ENSURE WE ACCOUNT FOR STAKED POSITION DIFFERENT THAN
            // REWARDS
        }
        // safe because newTokenId is zero if shares == exitValue
        _tokenOf[msg.sender] = newTokenID;
        // cleanup total shares and payout profits less reserve
        _totalLocked -= exitValue_;
        _distributeAllProfits(_payableSender(), payoutAToken, payoutEth);
    }

    // we must iterate all positions and dump profits before we are able to
    // allow withdraws. There is an admin over-ride in case we get stuck
    // in a bad position, but this still needs to be fixed also since it
    // could be used to attack this thing... maybe worth while to do a payout
    // contract from which they must claim?
    function aggregateProfits() public onlyPayoutUnSafe onlyPostLock {
        // get some gas cost tracking setup
        uint256 gasStart = _gas();
        uint256 gasLoop;
        // start index where we left off plus one
        uint256 i = _tokenIDOffset + 1;
        // scary for loop... but it will exit when one of following is true
        // the gas remaining is less than 5x the estimated per iteration cost
        // the iterator is done
        for (; ; i++) {
            uint256 tokenID;
            {
                bool ok;
                (tokenID, ok) = _getTokenIDAtIndex(i);
                if (!ok) {
                    // if we get here, iteration of array is dene
                    // we can move on with life and set payoutSafe
                    // since all payouts have been recorded
                    payoutSafe = true;
                    break;
                }
            }
            address payable acct = ownerOf(tokenID);
            _collectAllProfits(acct, tokenID);
            {
                uint256 gasrem = _gas();
                if (gasLoop == 0) {
                    // RECORD GAS ITERATION ESTIMATE IF NOT DONE
                    gasLoop = gasStart - gasrem;
                    // GIVE 5X MULTI ON IT TO ENSURE EVEN AN
                    // OVERPRICED ELEMENT BY 2X THE NORMAL COST WILL STILL PASS
                    gasLoop = 5 * gasLoop;
                    // ACCOUNTS FOR STATE WRITES ON EXIT
                    gasLoop = gasLoop + 10000;
                } else if (gasrem <= gasLoop) {
                    // IF WE ARE BELOW CUTOFF BREAK
                    break;
                }
            }
        }
        _tokenIDOffset = i;
    }

    function unlock()
        public
        onlyPostLock
        onlyPayoutSafe
        returns (uint256 payoutToken, uint256 payoutEth)
    {
        uint256 tokenID = tokenOf(msg.sender);
        uint256 shares = _getNumShares(tokenID);
        uint256 localTotalLocked = _totalLocked;
        //decrement the totalLocked counter
        localTotalLocked -= shares;
        // TODO UPDATE ALL MAPPINGS LOCALLY BEFORE EXTERNAL INTERACTIONS
        //update the global variable with new totalLocked value
        _totalLocked = localTotalLocked;
        //delete tokenID from iterable tokenID mapping
        _delTokenID(tokenID);
        delete(_tokenOf[msg.sender]);
        delete(_ownerOf[tokenID]);
        (payoutToken, payoutEth) = _rewardPool.payout(localTotalLocked, shares);
        // dist shares
        
        // xfer token back to user
        // _publicStaking.
    }

    function _lock(uint256 tokenID_, address tokenOwner_) internal onlyPreLock {
        uint256 shares = _getNumShares(tokenID_);
        _totalLocked += shares;
        _tokenOf[tokenOwner_] = tokenID_;
        _ownerOf[tokenID_] = tokenOwner_;
        _newTokenID(tokenID_);
        emit NewLockup(tokenOwner_, tokenID_);
    }

    function _getNumShares(uint256 tokenID_) internal view returns (uint256 shares) {
        (shares, , , , ) = _publicStaking.getPosition(tokenID_);
    }

    function _collectAllProfits(address payable acct_, uint256 tokenID_)
        internal
        returns (uint256 payoutToken, uint256 payoutEth)
    {
        (payoutToken, payoutEth) = _publicStaking.collectAllProfits(tokenID_);
        return _distributeAllProfits(acct_, payoutToken, payoutEth);
    }

    function _distributeAllProfits(
        address payable acct_,
        uint256 payoutAToken_,
        uint256 payoutEth_
    ) internal returns (uint256 usrPayoutToken, uint256 usrPayoutEth) {
        uint8 state = _getState();
        bool localPayoutSafe = payoutSafe;
        usrPayoutEth = payoutEth_;
        usrPayoutToken = payoutAToken_;
        if (localPayoutSafe && state == _statePostLock) {
            // case of we are sending out final pay based on request
            // just pay all
            usrPayoutEth += rewardEth[acct_];
            usrPayoutToken += rewardTokens[acct_];
            rewardEth[acct_] = 0;
            rewardTokens[acct_] = 0;
            _safeSendToken(acct_, usrPayoutToken);
            _safeSendEth(acct_, usrPayoutEth);
             return (usrPayoutToken, usrPayoutEth);
        }
        // implies !payoutSafe and state one of [preLock, inLock]
        // hold back reserves and fund over deltas
        uint256 reservedEth = (payoutEth_ * _fractionReserved) / _unitOne;
        usrPayoutEth = payoutEth_ - reservedEth;
        uint256 reservedToken = (payoutAToken_ * _fractionReserved) / _unitOne;
        usrPayoutToken = payoutAToken_ - reservedToken;
        // send tokens to reward pool
        _safeSendToken(address(_rewardPool), reservedToken);
        _rewardPool.deposit{value: reservedEth}(reservedToken); //todo send eth here reservedEth
        // either store to map or send to user
        if (!localPayoutSafe && state == _statePostLock) {
            // we should not send here and should instead track to local mapping
            // as otherwise a single bad user could block exit operations for all
            // other users by making the send to thier account fail via a contract
            rewardEth[acct_] += usrPayoutEth;
            rewardTokens[acct_] += usrPayoutToken;
            return (usrPayoutToken, usrPayoutEth);
        } else {
            _safeSendEth(acct_, usrPayoutEth);
            _safeSendToken(acct_, usrPayoutToken);
        }
        return (usrPayoutToken, usrPayoutEth);
    }

    function _safeSendToken(address acct_, uint256 val_) internal {
        if (val_ == 0) {
            return;
        }
        bool ok = _alca.transfer(acct_, val_);
        if (!ok) {
            revert TokenSendFailure();
        }
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

    function _payableSender() internal view returns (address payable) {
        return payable(msg.sender);
    }

    function _getTokenIDAtIndex(uint256 index_) internal view returns (uint256 tokenID, bool ok) {
        tokenID = _tokenIDs[index_];
        return (tokenID, tokenID > 0);
    }

    function _newTokenID(uint256 tokenID_) internal {
        uint256 index = _lenTokenIDs + 1;
        _tokenIDs[index] = tokenID_;
        _reverseTokenIDs[tokenID_] = index;
        _lenTokenIDs = index;
    }

    function _replaceTokenID(uint256 oldID_, uint256 newID_) internal {
        uint256 index = _reverseTokenIDs[oldID_];
        _reverseTokenIDs[oldID_] = 0;
        _tokenIDs[index] = newID_;
        _reverseTokenIDs[newID_] = index;
    }

    function _delTokenID(uint256 tokenID_) internal {
        uint256 tlen = _lenTokenIDs;
        if (tlen == 0) {
            return;
        }
        if (tlen == 1) {
            uint256 index = _reverseTokenIDs[tokenID_];
            _reverseTokenIDs[tokenID_] = 0;
            _tokenIDs[index] = 0;
            _lenTokenIDs = 0;
            return;
        }
        // pop the tail
        uint256 tailTokenID = _tokenIDs[tlen];
        _tokenIDs[tlen] = 0;
        _reverseTokenIDs[tailTokenID] = 0;
        _lenTokenIDs -= 1;
        if (tailTokenID == tokenID_) {
            // element was tail, so we are done
            return;
        }
        // use swap logic to re-insert tail over other position
        _replaceTokenID(tokenID_, tailTokenID);
    }

    function _gas() internal view returns (uint256 remains) {
        assembly {
            remains := gas()
        }
    }

    function _lockingOwnsOrRevert(uint256 tokenID_) internal view {
        if (_publicStaking.ownerOf(tokenID_) != address(this)) {
            revert ContractDoesNotOwnTokenID(tokenID_);
        }
    }

    function _unclaimedOrRevert(uint256 tokenID_) internal view {
        if (ownerOf(tokenID_) == address(0)) {
            revert TokenIDAlreadyClaimed(tokenID_);
        }
    }

    function _claimedOrRevert(uint256 tokenID_) internal view {
        if (ownerOf(tokenID_) != address(0)) {
            revert AddressHasNotPositionLinked();
        }
    }

    function _noTokenIDOrRevert(address acct_) internal view {
        uint256 tid = tokenOf(acct_);
        if (tid != 0) {
            revert MultipleLocksFromSingleAccount(tid);
        }
    }

    function _getState() internal view returns (uint8) {
        if (block.number < startBlock) {
            return _statePrelock;
        }
        if (block.number < endBlock) {
            return _stateInLock;
        }
        return _statePostLock;
    }
}
