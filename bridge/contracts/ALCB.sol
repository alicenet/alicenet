// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Address.sol";
import "contracts/utils/MagicEthTransfer.sol";
import "contracts/libraries/math/Sigmoid.sol";

/// @custom:salt ALCB
/// @custom:deploy-type deployCreateAndRegister
/// @custom:deploy-group alcb
/// @custom:deploy-group-index 0
contract ALCB is ERC20Permit, Ownable, Pausable, ReentrancyGuard, MagicEthTransfer, Sigmoid {
    using Address for address;

    struct ShareHolder {
        address account;
        uint32 percentage;
        bool isMagicTransfer;
    }

    // multiply factor for the selling/minting bonding curve
    uint256 internal constant _MARKET_SPREAD = 4;

    // Scaling factor to get the staking percentages
    uint256 public constant PERCENTAGE_SCALE = 1000;

    // Balance in ether that is hold in the contract after minting and burning
    uint256 internal _poolBalance;

    // Shareholders of the contract and their percentage of the yield. The sum of all percentages
    // must be 1000.
    ShareHolder[] internal _shareHolders;

    error InvalidBalance(uint256 contractBalance, uint256 poolBalance);
    error InvalidBurnAmount(uint256 amount);
    error MinimumValueNotMet(uint256 amount, uint256 minimumValue);
    error MinimumMintNotMet(uint256 amount, uint256 minimum);
    error MinimumBurnNotMet(uint256 amount, uint256 minimum);
    error BurnAmountExceedsSupply(uint256 amount, uint256 supply);
    error ShareHoldersPercentageSumError(uint256 sum, uint256 expectedSum);
    error CannotTransferToZeroAddress();

    constructor() ERC20("AliceNet Utility Token", "ALCB") ERC20Permit("ALCB") {}

    function pause() public onlyOwner {
        _pause();
    }

    function unpause() public onlyOwner {
        _unpause();
    }

    function setShareHolders(ShareHolder[] memory shareHolders_) public onlyOwner {
        uint32 sum = 0;
        for (uint256 i = 0; i < shareHolders_.length; i++) {
            if (shareHolders_[i].account == address(0)) {
                revert CannotTransferToZeroAddress();
            }
            sum += shareHolders_[i].percentage;
        }
        if (sum != PERCENTAGE_SCALE) {
            revert ShareHoldersPercentageSumError(sum, PERCENTAGE_SCALE);
        }
        delete _shareHolders;
        for (uint256 i = 0; i < shareHolders_.length; i++) {
            _shareHolders.push(shareHolders_[i]);
        }
    }

    /**
     * @notice Distributes the yields of the ALCB sale to all stakeholders
     * @return true if the method succeeded
     * */
    function distribute() public nonReentrant whenNotPaused returns (bool) {
        return _distribute();
    }

    /**
     * @notice Mints ALCB. This function receives ether in the transaction and converts them into
     * ALCB using a bonding price curve.
     * @param minBTK_ Minimum amount of ALCB that you wish to mint given an amount of ether. If
     * its not possible to mint the desired amount with the current price in the bonding curve, the
     * transaction is reverted. If the minBTK_ is met, the whole amount of ether sent will be
     * converted in ALCB.
     * @return numBTK the number of ALCB minted
     */
    function mint(uint256 minBTK_) public payable whenNotPaused returns (uint256 numBTK) {
        numBTK = _mint(msg.sender, msg.value, minBTK_);
        return numBTK;
    }

    /**
     * @notice Mints ALCB. This function receives ether in the transaction and converts them into
     * ALCB using a bonding price curve.
     * @param to_ The account to where the tokens will be minted
     * @param minBTK_ Minimum amount of ALCB that you wish to mint given an
     * amount of ether. If its not possible to mint the desired amount with the
     * current price in the bonding curve, the transaction is reverted. If the
     * minBTK_ is met, the whole amount of ether sent will be converted in ALCB.
     * @return numBTK the number of ALCB minted
     */
    function mintTo(
        address to_,
        uint256 minBTK_
    ) public payable whenNotPaused returns (uint256 numBTK) {
        numBTK = _mint(to_, msg.value, minBTK_);
        return numBTK;
    }

    /**
     * @notice Burn the tokens without sending ether back to user as the normal burn
     * function. The generated ether will be distributed in the distribute method. This function can
     * be used to charge ALCBs fees in other systems.
     * @param numBTK_ the number of ALCB to be burned
     */
    function destroyTokens(uint256 numBTK_) public {
        _destroyTokens(msg.sender, numBTK_);
    }

    /**
     * @notice Burn ALCB. This function sends ether corresponding to the amount of ALCBs being
     * burned using a bonding price curve.
     * @param amount_ The amount of ALCB being burned
     * @param minEth_ Minimum amount ether that you expect to receive given the
     * amount of ALCB being burned. If the amount of ALCB being burned
     * worth less than this amount the transaction is reverted.
     * @return numEth The number of ether being received
     * */
    function burn(uint256 amount_, uint256 minEth_) public nonReentrant returns (uint256 numEth) {
        numEth = _burn(msg.sender, msg.sender, amount_, minEth_);
        return numEth;
    }

    /**
     * @notice Burn ALCBs and send the ether received to an other account. This
     * function sends ether corresponding to the amount of ALCBs being
     * burned using a bonding price curve.
     * @param to_ The account to where the ether from the burning will be send
     * @param amount_ The amount of ALCBs being burned
     * @param minEth_ Minimum amount ether that you expect to receive given the
     * amount of ALCBs being burned. If the amount of ALCBs being burned
     * worth less than this amount the transaction is reverted.
     * @return numEth the number of ether being received
     * */
    function burnTo(
        address to_,
        uint256 amount_,
        uint256 minEth_
    ) public nonReentrant returns (uint256 numEth) {
        numEth = _burn(msg.sender, to_, amount_, minEth_);
        return numEth;
    }

    /**
     * @notice Gets the amount that can be distributed as profits to the stakeholders contracts.
     * @return The amount that can be distributed as yield
     */
    function getYield() public view returns (uint256) {
        return address(this).balance - _poolBalance;
    }

    /**
     * @notice Gets the pool balance in ether.
     * @return The pool balance in ether
     */
    function getPoolBalance() public view returns (uint256) {
        return _poolBalance;
    }

    function getShareHolders() public view returns (ShareHolder[] memory) {
        return _shareHolders;
    }

    function getShareHolderCount() public view returns (uint256) {
        return _shareHolders.length;
    }

    /**
     * @notice Compute how many ether will be necessary to mint an amount of ALCBs in the
     * current state of the contract. Should be used carefully if its being called
     * outside an smart contract transaction, as the bonding curve state can change
     * before a future transaction is sent.
     * @param numBTK_ Amount of ALCBs that we want to convert in ether
     * @return numEth the number of ether necessary to mint an amount of ALCB
     */
    function getLatestEthToMintTokens(uint256 numBTK_) public view returns (uint256 numEth) {
        return _getEthToMintTokens(totalSupply(), numBTK_);
    }

    /**
     * @notice Compute how many ether will be received during a ALCB burn at the current
     * bonding curve state. Should be used carefully if its being called outside an
     * smart contract transaction, as the bonding curve state can change before a
     * future transaction is sent.
     * @param numBTK_ Amount of ALCBs to convert in ether
     * @return numEth the amount of ether will be received during a ALCB burn at the current
     * bonding curve state
     */
    function getLatestEthFromTokensBurn(uint256 numBTK_) public view returns (uint256 numEth) {
        return _tokensToEth(_poolBalance, totalSupply(), numBTK_);
    }

    /**
     * @notice Gets an amount of ALCBs that will be minted at the current state of the
     * bonding curve. Should be used carefully if its being called outside an smart
     * contract transaction, as the bonding curve state can change before a future
     * transaction is sent.
     * @param numEth_ Amount of ether to convert in ALCBs
     * @return the amount of ALCBs that will be minted at the current state of the
     * bonding curve
     * */
    function getLatestMintedTokensFromEth(uint256 numEth_) public view returns (uint256) {
        return _ethToTokens(_poolBalance, numEth_ / _MARKET_SPREAD);
    }

    /**
     * @notice Gets the market spread (difference between the minting and burning bonding
     * curves).
     * @return the market spread (difference between the minting and burning bonding
     * curves).
     * */
    function getMarketSpread() public pure returns (uint256) {
        return _MARKET_SPREAD;
    }

    /**
     * @notice Compute how many ether will be necessary to mint an amount of ALCBs at a
     * certain point in the bonding curve.
     * @param totalSupply_ The total supply of ALCB at a given moment where we
     * want to compute the amount of ether necessary to mint.
     * @param numBTK_ Amount of ALCBs that we want to convert in ether
     * @return numEth the amount ether that will be necessary to mint an amount of ALCBs at a
     * certain point in the bonding curve
     * */
    function getEthToMintTokens(
        uint256 totalSupply_,
        uint256 numBTK_
    ) public pure returns (uint256 numEth) {
        return _getEthToMintTokens(totalSupply_, numBTK_);
    }

    /**
     * @notice Compute how many ether will be received during a ALCB burn.
     * @param poolBalance_ The pool balance (in ether) at the moment
     * that of the conversion.
     * @param totalSupply_ The total supply of ALCB at the moment
     * that of the conversion.
     * @param numBTK_ Amount of ALCBs to convert in ether
     * @return numEth the ether that will be received during a ALCB burn
     * */
    function getEthFromTokensBurn(
        uint256 poolBalance_,
        uint256 totalSupply_,
        uint256 numBTK_
    ) public pure returns (uint256 numEth) {
        return _tokensToEth(poolBalance_, totalSupply_, numBTK_);
    }

    /**
     * @notice Gets an amount of ALCBs that will be minted at given a point in the bonding
     * curve.
     * @param poolBalance_ The pool balance (in ether) at the moment
     * that of the conversion.
     * @param numEth_ Amount of ether to convert in ALCBs
     * @return the amount of ALCBs that will be minted at given a point in the bonding
     * curve.
     * */
    function getMintedTokensFromEth(
        uint256 poolBalance_,
        uint256 numEth_
    ) public pure returns (uint256) {
        return _ethToTokens(poolBalance_, numEth_ / _MARKET_SPREAD);
    }

    /// Distributes the yields from the ALCB minting to all stake holders.
    function _distribute() internal returns (bool) {
        // make a local copy to save gas
        uint256 poolBalance = _poolBalance;
        // find all value in excess of what is needed in pool
        uint256 excess = address(this).balance - poolBalance;
        if (excess == 0) {
            return true;
        }
        uint256 paidAmount = 0;
        for (uint256 i = 0; i < _shareHolders.length; i++) {
            ShareHolder memory shareHolder = _shareHolders[i];
            uint256 amount;
            // sending the remainders of the integer division to the last share holder
            if (i == _shareHolders.length - 1) {
                amount = excess - paidAmount;
            } else {
                amount = (excess * shareHolder.percentage) / PERCENTAGE_SCALE;
            }
            paidAmount += amount;
            if (shareHolder.isMagicTransfer) {
                _safeTransferEthWithMagic(IMagicEthTransfer(shareHolder.account), amount);
            } else {
                // // send the share to the share holder
                Address.sendValue(payable(shareHolder.account), amount);
            }
        }

        if (address(this).balance < poolBalance) {
            revert InvalidBalance(address(this).balance, poolBalance);
        }
        return true;
    }

    // Burn the tokens during deposits without sending ether back to user as the
    // normal burn function. The ether will be distributed in the distribute
    // method.
    function _destroyTokens(address account, uint256 numBTK_) internal {
        if (numBTK_ == 0) {
            revert InvalidBurnAmount(numBTK_);
        }
        _poolBalance -= _tokensToEth(_poolBalance, totalSupply(), numBTK_);
        ERC20._burn(account, numBTK_);
    }

    // Internal function that mints the ALCB tokens following the bounding
    // price curve.
    function _mint(
        address to_,
        uint256 numEth_,
        uint256 minBTK_
    ) internal returns (uint256 numBTK) {
        if (numEth_ < _MARKET_SPREAD) {
            revert MinimumValueNotMet(numEth_, _MARKET_SPREAD);
        }

        numEth_ = numEth_ / _MARKET_SPREAD;
        uint256 poolBalance = _poolBalance;
        numBTK = _ethToTokens(poolBalance, numEth_);
        if (numBTK < minBTK_) {
            revert MinimumMintNotMet(numBTK, minBTK_);
        }

        poolBalance += numEth_;
        _poolBalance = poolBalance;
        ERC20._mint(to_, numBTK);
        return numBTK;
    }

    // Internal function that burns the ALCB tokens following the bounding
    // price curve.
    function _burn(
        address from_,
        address to_,
        uint256 numBTK_,
        uint256 minEth_
    ) internal returns (uint256 numEth) {
        if (numBTK_ == 0) {
            revert InvalidBurnAmount(numBTK_);
        }

        uint256 poolBalance = _poolBalance;
        numEth = _tokensToEth(poolBalance, totalSupply(), numBTK_);

        if (numEth < minEth_) {
            revert MinimumBurnNotMet(numEth, minEth_);
        }

        poolBalance -= numEth;
        _poolBalance = poolBalance;
        ERC20._burn(from_, numBTK_);
        Address.sendValue(payable(to_), numEth);
        return numEth;
    }

    // Internal function that converts an ether amount into ALCB tokens
    // following the bounding price curve.
    function _ethToTokens(uint256 poolBalance_, uint256 numEth_) internal pure returns (uint256) {
        return _p(poolBalance_ + numEth_) - _p(poolBalance_);
    }

    // Internal function that converts a ALCB amount into ether following the
    // bounding price curve.
    function _tokensToEth(
        uint256 poolBalance_,
        uint256 totalSupply_,
        uint256 numBTK_
    ) internal pure returns (uint256 numEth) {
        if (totalSupply_ < numBTK_) {
            revert BurnAmountExceedsSupply(numBTK_, totalSupply_);
        }
        return _min(poolBalance_, _pInverse(totalSupply_) - _pInverse(totalSupply_ - numBTK_));
    }

    // Internal function to compute the amount of ether required to mint an amount
    // of ALCBs. Inverse of the _ethToALCBs function.
    function _getEthToMintTokens(
        uint256 totalSupply_,
        uint256 numBTK_
    ) internal pure returns (uint256 numEth) {
        return (_pInverse(totalSupply_ + numBTK_) - _pInverse(totalSupply_)) * _MARKET_SPREAD;
    }
}
