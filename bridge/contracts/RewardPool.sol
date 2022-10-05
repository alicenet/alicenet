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

    function deposit(uint256 numTokens_) public payable onlyLockingOrBonus {
        _tokenBalance += numTokens_;
    }

    function payout(uint256 total_, uint256 shares_) public onlyLocking returns (uint256, uint256) {
        uint256 pe = (address(this).balance * shares_ * _unitOne) / (_unitOne * total_);
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

    // should not be needed, but I am paranoid
    function forceBalanceCheck() public {
        uint256 bal = _alca.balanceOf(address(this));
        uint256 localTokenBalance = _tokenBalance;
        if (localTokenBalance != bal) {
            _tokenBalance = bal;
        }
    }
}