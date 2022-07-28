// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "contracts/interfaces/IBridgeRouter.sol";
import "contracts/utils/Admin.sol";
import "contracts/utils/Mutex.sol";
import "contracts/utils/MagicEthTransfer.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/libraries/math/Sigmoid.sol";
import "contracts/utils/ImmutableAuth.sol";
import {BTokenErrors} from "contracts/libraries/errors/BTokenErrors.sol";

/// @custom:salt BToken
/// @custom:deploy-type deployStatic
contract BToken is
    ERC20Upgradeable,
    Admin,
    Mutex,
    MagicEthTransfer,
    EthSafeTransfer,
    Sigmoid,
    ImmutableFactory,
    ImmutablePublicStaking,
    ImmutableValidatorStaking,
    ImmutableLiquidityProviderStaking,
    ImmutableFoundation
{
    // Struct to keep track of the amount divided between the staking contracts
    // dividends distribution.
    struct Splits {
        uint32 validatorStaking;
        uint32 publicStaking;
        uint32 liquidityProviderStaking;
        uint32 protocolFee;
    }

    // Struct to keep track of the deposits
    struct Deposit {
        uint8 accountType;
        address account;
        uint256 value;
    }

    // multiply factor for the selling/minting bonding curve
    uint256 internal constant _MARKET_SPREAD = 4;

    // Scaling factor to get the staking percentages
    uint256 internal constant _PERCENTAGE_SCALE = 1000;

    // Balance in ether that is hold in the contract after minting and burning
    uint256 internal _poolBalance;

    // Value of the percentages that will send to each staking contract. Divide
    // this value by _PERCENTAGE_SCALE = 1000 to get the corresponding percentages.
    // These values must sum to 1000.
    Splits internal _splits;

    // Monotonically increasing variable to track the BTokens deposits.
    uint256 internal _depositID;

    // Total amount of BTokens that were deposited in the AliceNet chain. The
    // BTokens deposited in the AliceNet are burned by this contract.
    uint256 internal _totalDeposited;

    // Tracks the amount of each deposit. Key is deposit id, value is amount
    // deposited.
    mapping(uint256 => Deposit) internal _deposits;

    /// @notice Event emitted when a deposit is received
    event DepositReceived(
        uint256 indexed depositID,
        uint8 indexed accountType,
        address indexed depositor,
        uint256 amount
    );

    constructor()
        Admin(msg.sender)
        Mutex()
        ImmutableFactory(msg.sender)
        ImmutablePublicStaking()
        ImmutableValidatorStaking()
        ImmutableLiquidityProviderStaking()
        ImmutableFoundation()
    {}

    function initialize() public onlyFactory initializer {
        __ERC20_init("BToken", "BOB");
        _setSplitsInternal(332, 332, 332, 4);
    }

    /// @dev sets the percentage that will be divided between all the staking
    /// contracts, must only be called by _admin
    function setSplits(
        uint256 validatorStakingSplit_,
        uint256 publicStakingSplit_,
        uint256 liquidityProviderStakingSplit_,
        uint256 protocolFee_
    ) public onlyAdmin {
        _setSplitsInternal(
            validatorStakingSplit_,
            publicStakingSplit_,
            liquidityProviderStakingSplit_,
            protocolFee_
        );
    }

    /// Distributes the yields of the BToken sale to all stakeholders: validators,
    /// public stakers, liquidity providers stakers, foundation.
    function distribute()
        public
        returns (
            uint256 minerAmount,
            uint256 stakingAmount,
            uint256 lpStakingAmount,
            uint256 foundationAmount
        )
    {
        return _distribute();
    }

    /// Deposits a BToken amount into the AliceNet blockchain. The BTokens amount is
    /// deducted from the sender and it is burned by this contract. The created
    /// deposit Id is owned by the `to_` address.
    /// @param accountType_ The type of account the to_ address must be equivalent with ( 1 for Eth native, 2 for BN )
    /// @param to_ The address of the account that will own the deposit
    /// @param amount_ The amount of BTokens to be deposited
    /// Return The deposit ID of the deposit created
    function deposit(
        uint8 accountType_,
        address to_,
        uint256 amount_
    ) public returns (uint256) {
        return _deposit(accountType_, to_, amount_);
    }

    /// Allows deposits to be minted in a virtual manner and sent to the AliceNet
    /// chain by simply emitting a Deposit event without actually minting or
    /// burning any tokens, must only be called by _admin.
    /// @param accountType_ The type of account the to_ address must be equivalent with ( 1 for Eth native, 2 for BN )
    /// @param to_ The address of the account that will own the deposit
    /// @param amount_ The amount of BTokens to be deposited
    /// Return The deposit ID of the deposit created
    function virtualMintDeposit(
        uint8 accountType_,
        address to_,
        uint256 amount_
    ) public onlyAdmin returns (uint256) {
        return _virtualDeposit(accountType_, to_, amount_);
    }

    /// Allows deposits to be minted and sent to the AliceNet chain without actually
    /// minting or burning any BTokens. This function receives ether and converts
    /// them directly into BToken and then deposit them into the AliceNet chain. This
    /// function has the same effect as calling mint (creating the tokens) + deposit
    /// (burning the tokens) functions but it costs less gas.
    /// @param accountType_ The type of account the to_ address must be equivalent with ( 1 for Eth native, 2 for BN )
    /// @param to_ The address of the account that will own the deposit
    /// @param minBTK_ The amount of BTokens to be deposited
    /// Return The deposit ID of the deposit created
    function mintDeposit(
        uint8 accountType_,
        address to_,
        uint256 minBTK_
    ) public payable returns (uint256) {
        return _mintDeposit(accountType_, to_, minBTK_, msg.value);
    }

    /// Mints BToken. This function receives ether in the transaction and
    /// converts them into BToken using a bonding price curve.
    /// @param minBTK_ Minimum amount of BToken that you wish to mint given an
    /// amount of ether. If its not possible to mint the desired amount with the
    /// current price in the bonding curve, the transaction is reverted. If the
    /// minBTK_ is met, the whole amount of ether sent will be converted in BToken.
    /// Return The number of BToken minted
    function mint(uint256 minBTK_) public payable returns (uint256 numBTK) {
        numBTK = _mint(msg.sender, msg.value, minBTK_);
        return numBTK;
    }

    /// Mints BToken. This function receives ether in the transaction and
    /// converts them into BToken using a bonding price curve.
    /// @param to_ The account to where the tokens will be minted
    /// @param minBTK_ Minimum amount of BToken that you wish to mint given an
    /// amount of ether. If its not possible to mint the desired amount with the
    /// current price in the bonding curve, the transaction is reverted. If the
    /// minBTK_ is met, the whole amount of ether sent will be converted in BToken.
    /// Return The number of BToken minted
    function mintTo(address to_, uint256 minBTK_) public payable returns (uint256 numBTK) {
        numBTK = _mint(to_, msg.value, minBTK_);
        return numBTK;
    }

    /// Burn the tokens without sending ether back to user as the normal burn
    /// function. The generated ether will be distributed in the distribute method.
    function destroyBTokens(uint256 numBTK_) public returns (bool) {
        _destroyTokens(msg.sender, numBTK_);
        return true;
    }

    /// Burn pre-approved tokens without sending ether back to an account as the
    /// normal burn function. The generated ether will be distributed in the
    /// distribute method. Mostly used to burn tokens in a third party contract
    /// without having to transfer before.
    function destroyPreApprovedBTokens(address account, uint256 numBTK_) public returns (bool) {
        ERC20Upgradeable._spendAllowance(account, msg.sender, numBTK_);
        _destroyTokens(account, numBTK_);
        return true;
    }

    function depositTokensOnBridges(
        uint256 maxTokens,
        uint16 bridgeVersion,
        bytes calldata data
    ) public payable {
        //calculate router address
        bytes32 bridgeRouterSalt = keccak256(
            bytes.concat(
                keccak256(abi.encodePacked("BridgeRouter")),
                keccak256(abi.encodePacked(bridgeVersion))
            )
        );
        address bridgeRouterAddress = getMetamorphicContractAddress(
            bridgeRouterSalt,
            _factoryAddress()
        );
        if (!_isContract(bridgeRouterAddress)) {
            revert BTokenErrors.InexistentRouterContract(bridgeRouterAddress);
        }
        //forward call to router
        uint256 bTokenFee = IBridgeRouter(bridgeRouterAddress).routeDeposit(
            msg.sender,
            maxTokens,
            data
        );
        if (msg.value > 0) {
            uint256 ethFee = _getEthToMintBTokens(totalSupply(), bTokenFee);
            if (ethFee > msg.value) {
                revert BTokenErrors.InsufficientFee(msg.value, ethFee);
            }
            uint256 refund = msg.value - ethFee;
            if (refund > 0) {
                _safeTransferEth(msg.sender, refund);
            }
        } else {
            _destroyTokens(msg.sender, bTokenFee);
        }
    }

    /// Burn BToken. This function sends ether corresponding to the amount of
    /// BTokens being burned using a bonding price curve.
    /// @param amount_ The amount of BToken being burned
    /// @param minEth_ Minimum amount ether that you expect to receive given the
    /// amount of BToken being burned. If the amount of BToken being burned
    /// worth less than this amount the transaction is reverted.
    /// Return The number of ether being received
    function burn(uint256 amount_, uint256 minEth_) public returns (uint256 numEth) {
        numEth = _burn(msg.sender, msg.sender, amount_, minEth_);
        return numEth;
    }

    /// Burn BTokens and send the ether received to other account. This
    /// function sends ether corresponding to the amount of BTokens being
    /// burned using a bonding price curve.
    /// @param to_ The account to where the ether from the burning will send
    /// @param amount_ The amount of BTokens being burned
    /// @param minEth_ Minimum amount ether that you expect to receive given the
    /// amount of BTokens being burned. If the amount of BTokens being burned
    /// worth less than this amount the transaction is reverted.
    /// Return The number of ether being received
    function burnTo(
        address to_,
        uint256 amount_,
        uint256 minEth_
    ) public returns (uint256 numEth) {
        numEth = _burn(msg.sender, to_, amount_, minEth_);
        return numEth;
    }

    /// Gets the value of the percentages that will send to each staking contract.
    /// Divide this value by _PERCENTAGE_SCALE = 1000 to get the corresponding
    /// percentages.
    function getSplits() public view returns (Splits memory) {
        return _splits;
    }

    /// Gets the latest deposit ID emitted.
    function getDepositID() public view returns (uint256) {
        return _depositID;
    }

    /// Gets the pool balance in ether.
    function getPoolBalance() public view returns (uint256) {
        return _poolBalance;
    }

    /// Gets the total amount of BTokens that were deposited in the AliceNet
    /// blockchain. Since BTokens are burned when deposited, this value will be
    /// different from the total supply of BTokens.
    function getTotalBTokensDeposited() public view returns (uint256) {
        return _totalDeposited;
    }

    /// Gets the deposited amount given a depositID.
    /// @param depositID The Id of the deposit
    function getDeposit(uint256 depositID) public view returns (Deposit memory) {
        Deposit memory d = _deposits[depositID];
        if (d.account == address(0)) {
            revert BTokenErrors.InvalidDepositId(depositID);
        }

        return d;
    }

    /// Compute how many ether will be necessary to mint an amount of BTokens in the
    /// current state of the contract. Should be used carefully if its being called
    /// outside an smart contract transaction, as the bonding curve state can change
    /// before a future transaction is sent.
    /// @param numBTK_ Amount of BTokens that we want to convert in ether
    function getLatestEthToMintBTokens(uint256 numBTK_) public view returns (uint256 numEth) {
        return _getEthToMintBTokens(totalSupply(), numBTK_);
    }

    /// Compute how many ether will be received during a BToken burn at the current
    /// bonding curve state. Should be used carefully if its being called outside an
    /// smart contract transaction, as the bonding curve state can change before a
    /// future transaction is sent.
    /// @param numBTK_ Amount of BTokens to convert in ether
    function getLatestEthFromBTokensBurn(uint256 numBTK_) public view returns (uint256 numEth) {
        return _bTokensToEth(_poolBalance, totalSupply(), numBTK_);
    }

    /// Gets an amount of BTokens that will be minted at the current state of the
    /// bonding curve. Should be used carefully if its being called outside an smart
    /// contract transaction, as the bonding curve state can change before a future
    /// transaction is sent.
    /// @param numEth_ Amount of ether to convert in BTokens
    function getLatestMintedBTokensFromEth(uint256 numEth_) public view returns (uint256) {
        return _ethToBTokens(_poolBalance, numEth_);
    }

    /// Gets the market spread (difference between the minting and burning bonding
    /// curves).
    function getMarketSpread() public pure returns (uint256) {
        return _MARKET_SPREAD;
    }

    /// Compute how many ether will be necessary to mint an amount of BTokens at a
    /// certain point in the bonding curve.
    /// @param totalSupply_ The total supply of BToken at a given moment where we
    /// want to compute the amount of ether necessary to mint.
    /// @param numBTK_ Amount of BTokens that we want to convert in ether
    function getEthToMintBTokens(uint256 totalSupply_, uint256 numBTK_)
        public
        pure
        returns (uint256 numEth)
    {
        return _getEthToMintBTokens(totalSupply_, numBTK_);
    }

    /// Compute how many ether will be received during a BToken burn.
    /// @param poolBalance_ The pool balance (in ether) at the moment
    /// that of the conversion.
    /// @param totalSupply_ The total supply of BToken at the moment
    /// that of the conversion.
    /// @param numBTK_ Amount of BTokens to convert in ether
    function getEthFromBTokensBurn(
        uint256 poolBalance_,
        uint256 totalSupply_,
        uint256 numBTK_
    ) public pure returns (uint256 numEth) {
        return _bTokensToEth(poolBalance_, totalSupply_, numBTK_);
    }

    /// Gets an amount of BTokens that will be minted given a point in the bonding
    /// curve.
    /// @param poolBalance_ The pool balance (in ether) at the moment
    /// that of the conversion.
    /// @param numEth_ Amount of ether to convert in BTokens
    function getMintedBTokensFromEth(uint256 poolBalance_, uint256 numEth_)
        public
        pure
        returns (uint256)
    {
        return _ethToBTokens(poolBalance_, numEth_);
    }

    /// Distributes the yields from the BToken minting to all stake holders.
    function _distribute()
        internal
        withLock
        returns (
            uint256 minerAmount,
            uint256 stakingAmount,
            uint256 lpStakingAmount,
            uint256 foundationAmount
        )
    {
        // make a local copy to save gas
        uint256 poolBalance = _poolBalance;

        // find all value in excess of what is needed in pool
        uint256 excess = address(this).balance - poolBalance;

        // take out protocolFee from excess and decrement excess
        foundationAmount = (excess * _splits.protocolFee) / _PERCENTAGE_SCALE;

        // split remaining between miners, stakers and lp stakers
        stakingAmount = (excess * _splits.publicStaking) / _PERCENTAGE_SCALE;
        lpStakingAmount = (excess * _splits.liquidityProviderStaking) / _PERCENTAGE_SCALE;
        // then give miners the difference of the original and the sum of the
        // stakingAmount
        minerAmount = excess - (stakingAmount + lpStakingAmount + foundationAmount);

        if (foundationAmount != 0) {
            _safeTransferEthWithMagic(IMagicEthTransfer(_foundationAddress()), foundationAmount);
        }
        if (minerAmount != 0) {
            _safeTransferEthWithMagic(IMagicEthTransfer(_validatorStakingAddress()), minerAmount);
        }
        if (stakingAmount != 0) {
            _safeTransferEthWithMagic(IMagicEthTransfer(_publicStakingAddress()), stakingAmount);
        }
        if (lpStakingAmount != 0) {
            _safeTransferEthWithMagic(
                IMagicEthTransfer(_liquidityProviderStakingAddress()),
                lpStakingAmount
            );
        }
        if (address(this).balance < poolBalance) {
            revert BTokenErrors.InvalidBalance(address(this).balance, poolBalance);
        }

        // invariants hold
        return (minerAmount, stakingAmount, lpStakingAmount, foundationAmount);
    }

    // Burn the tokens during deposits without sending ether back to user as the
    // normal burn function. The ether will be distributed in the distribute
    // method.
    function _destroyTokens(address account, uint256 numBTK_) internal returns (bool) {
        if (numBTK_ == 0) {
            revert BTokenErrors.InvalidBurnAmount(numBTK_);
        }
        _poolBalance -= _bTokensToEth(_poolBalance, totalSupply(), numBTK_);
        ERC20Upgradeable._burn(account, numBTK_);
        return true;
    }

    // Internal function that does the deposit in the AliceNet Chain, i.e emit the
    // event DepositReceived. All the BTokens sent to this function are burned.
    function _deposit(
        uint8 accountType_,
        address to_,
        uint256 amount_
    ) internal returns (uint256) {
        if (_isContract(to_)) {
            revert BTokenErrors.ContractsDisallowedDeposits(to_);
        }

        if (amount_ == 0) {
            revert BTokenErrors.DepositAmountZero();
        }

        if (!_destroyTokens(msg.sender, amount_)) {
            revert BTokenErrors.DepositBurnFail(amount_);
        }

        // copying state to save gas
        return _doDepositCommon(accountType_, to_, amount_);
    }

    // does a virtual deposit into the AliceNet Chain without actually minting or
    // burning any token.
    function _virtualDeposit(
        uint8 accountType_,
        address to_,
        uint256 amount_
    ) internal returns (uint256) {
        if (_isContract(to_)) {
            revert BTokenErrors.ContractsDisallowedDeposits(to_);
        }

        if (amount_ == 0) {
            revert BTokenErrors.DepositAmountZero();
        }

        // copying state to save gas
        return _doDepositCommon(accountType_, to_, amount_);
    }

    // Mints a virtual deposit into the AliceNet Chain without actually minting or
    // burning any token. This function converts ether sent in BTokens.
    function _mintDeposit(
        uint8 accountType_,
        address to_,
        uint256 minBTK_,
        uint256 numEth_
    ) internal returns (uint256) {
        if (_isContract(to_)) {
            revert BTokenErrors.ContractsDisallowedDeposits(to_);
        }
        if (numEth_ < _MARKET_SPREAD) {
            revert BTokenErrors.MinimumValueNotMet(numEth_, _MARKET_SPREAD);
        }

        numEth_ = numEth_ / _MARKET_SPREAD;
        uint256 amount_ = _ethToBTokens(_poolBalance, numEth_);
        if (amount_ < minBTK_) {
            revert BTokenErrors.InsufficientEth(amount_, minBTK_);
        }

        return _doDepositCommon(accountType_, to_, amount_);
    }

    function _doDepositCommon(
        uint8 accountType_,
        address to_,
        uint256 amount_
    ) internal returns (uint256) {
        uint256 depositID = _depositID + 1;
        _deposits[depositID] = _newDeposit(accountType_, to_, amount_);
        _totalDeposited += amount_;
        _depositID = depositID;
        emit DepositReceived(depositID, accountType_, to_, amount_);
        return depositID;
    }

    // Internal function that mints the BToken tokens following the bounding
    // price curve.
    function _mint(
        address to_,
        uint256 numEth_,
        uint256 minBTK_
    ) internal returns (uint256 numBTK) {
        if (numEth_ < _MARKET_SPREAD) {
            revert BTokenErrors.MinimumValueNotMet(numEth_, _MARKET_SPREAD);
        }

        numEth_ = numEth_ / _MARKET_SPREAD;
        uint256 poolBalance = _poolBalance;
        numBTK = _ethToBTokens(poolBalance, numEth_);
        if (numBTK < minBTK_) {
            revert BTokenErrors.MinimumMintNotMet(numBTK, minBTK_);
        }

        poolBalance += numEth_;
        _poolBalance = poolBalance;
        ERC20Upgradeable._mint(to_, numBTK);
        return numBTK;
    }

    // Internal function that burns the BToken tokens following the bounding
    // price curve.
    function _burn(
        address from_,
        address to_,
        uint256 numBTK_,
        uint256 minEth_
    ) internal returns (uint256 numEth) {
        if (numBTK_ == 0) {
            revert BTokenErrors.InvalidBurnAmount(numBTK_);
        }

        uint256 poolBalance = _poolBalance;
        numEth = _bTokensToEth(poolBalance, totalSupply(), numBTK_);

        if (numEth < minEth_) {
            revert BTokenErrors.MinimumBurnNotMet(numEth, minEth_);
        }

        poolBalance -= numEth;
        _poolBalance = poolBalance;
        ERC20Upgradeable._burn(from_, numBTK_);
        _safeTransferEth(to_, numEth);
        return numEth;
    }

    function _setSplitsInternal(
        uint256 validatorStakingSplit_,
        uint256 publicStakingSplit_,
        uint256 liquidityProviderStakingSplit_,
        uint256 protocolFee_
    ) internal {
        if (
            validatorStakingSplit_ +
                publicStakingSplit_ +
                liquidityProviderStakingSplit_ +
                protocolFee_ !=
            _PERCENTAGE_SCALE
        ) {
            revert BTokenErrors.SplitValueSumError();
        }
        _splits = Splits(
            uint32(validatorStakingSplit_),
            uint32(publicStakingSplit_),
            uint32(liquidityProviderStakingSplit_),
            uint32(protocolFee_)
        );
    }

    // Check if addr_ is EOA (Externally Owned Account) or a contract.
    function _isContract(address addr_) internal view returns (bool) {
        uint256 size;
        assembly {
            size := extcodesize(addr_)
        }
        return size > 0;
    }

    // Internal function that converts an ether amount into BToken tokens
    // following the bounding price curve.
    function _ethToBTokens(uint256 poolBalance_, uint256 numEth_) internal pure returns (uint256) {
        return _p(poolBalance_ + numEth_) - _p(poolBalance_);
    }

    // Internal function that converts a BToken amount into ether following the
    // bounding price curve.
    function _bTokensToEth(
        uint256 poolBalance_,
        uint256 totalSupply_,
        uint256 numBTK_
    ) internal pure returns (uint256 numEth) {
        if (totalSupply_ < numBTK_) {
            revert BTokenErrors.BurnAmountExceedsSupply(numBTK_, totalSupply_);
        }
        return _min(poolBalance_, _pInverse(totalSupply_) - _pInverse(totalSupply_ - numBTK_));
    }

    // Internal function to compute the amount of ether required to mint an amount
    // of BTokens. Inverse of the _ethToBTokens function.
    function _getEthToMintBTokens(uint256 totalSupply_, uint256 numBTK_)
        internal
        pure
        returns (uint256 numEth)
    {
        return (_pInverse(totalSupply_ + numBTK_) - _pInverse(totalSupply_)) * _MARKET_SPREAD;
    }

    function _newDeposit(
        uint8 accountType_,
        address account_,
        uint256 value_
    ) internal pure returns (Deposit memory) {
        Deposit memory d = Deposit(accountType_, account_, value_);
        return d;
    }
}
