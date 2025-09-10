// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.22;

import "@openzeppelin/contracts5/token/ERC20/extensions/ERC20Permit.sol";
import "@openzeppelin/contracts5/access/Ownable.sol";
import "@openzeppelin/contracts5/utils/Pausable.sol";
import "@openzeppelin/contracts5/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts5/utils/Address.sol";
import "contracts/legacy/utils/MagicEthTransfer.sol";

contract ALCB is ERC20Permit, Ownable, Pausable, ReentrancyGuard, MagicEthTransfer {
    using Address for address;

    struct ShareHolder {
        address account;
        uint32 percentage;
        bool isMagicTransfer;
    }

    // Scaling factor to get the staking percentages
    uint256 public constant PERCENTAGE_SCALE = 1000;

    // Market spread controls how much the minting and burning conversion rates differ. Factor
    // values are percentages when divided by 1000. For example, a value of 800 will mean
    // 80% of the ether will be returned to the user and 20% will be distributed to the shareholders.
    uint128 public marketSpreadFactor = 800;

    // Factor that controls the split between of the contract balance that will be used to cover the
    // operational costs and will be distributed to the shareholders. Factor values are percentages
    // when divided by 1000. For example, a value of 1000 will mean that 100% of the ether will
    // be distributed to the operational costs and 0% will be distributed to the shareholders.
    uint128 public operationalCostFactor = 1000;

    // Tracks the pool balance
    uint256 public poolBalance;

    // The wallet that will receive the operational costs
    address public operationalWallet;

    // Shareholders of the contract and their percentage of the yield. The sum of all values
    // must be 1000.
    ShareHolder[] public shareHolders;

    error InvalidBurnAmount(uint256 amount);
    error ShareHoldersPercentageSumError(uint256 sum, uint256 expectedSum);
    error MarketSpreadFactorError(uint128 marketSpreadFactor, uint256 maxFactor);
    error OperationalFactorError(uint128 operationalCostFactor, uint256 maxFactor);
    error CannotTransferToZeroAddress();
    error PercentagesCannotBeZero();

    constructor(
        address operationalWallet_,
        ShareHolder[] memory shareHolders_
    ) ERC20("AliceNet Utility Token", "ALCB") ERC20Permit("ALCB") Ownable(msg.sender) {
        operationalWallet = operationalWallet_;
        _setShareHolders(shareHolders_);
    }

    /**
     * @notice Pauses the contract. This function can only be called by the owner of the contract.
     */
    function pauseMinting() public onlyOwner {
        _pause();
    }

    /**
     * @notice Unpauses the contract. This function can only be called by the owner of the contract.
     */
    function unpauseMinting() public onlyOwner {
        _unpause();
    }

    /**
     * @notice Set the market spread factor. The market spread controls how much the minting and
     * burning conversion rates differ. Factor values are percentages when divided by 1000.
     * @param marketSpreadFactor_ The new market spread factor
     */
    function setMarketSpreadFactor(uint128 marketSpreadFactor_) public onlyOwner {
        if (marketSpreadFactor_ > PERCENTAGE_SCALE) {
            revert MarketSpreadFactorError(marketSpreadFactor_, PERCENTAGE_SCALE);
        }
        marketSpreadFactor = marketSpreadFactor_;
    }

    /**
     * @notice Set the operational cost factor. The operational cost factor controls the split
     * between of the contract balance that will be used to cover the operational costs and will be
     * distributed to the shareholders. Factor values are percentages when divided by 1000.
     * @param operationalCostFactor_ The new operational cost factor
     */
    function setOperationalCostFactor(uint128 operationalCostFactor_) public onlyOwner {
        if (operationalCostFactor_ > PERCENTAGE_SCALE) {
            revert OperationalFactorError(operationalCostFactor_, PERCENTAGE_SCALE);
        }
        operationalCostFactor = operationalCostFactor_;
    }

    /**
     * @notice Set the share holders of the contract. The share holders are the accounts that will
     * be receiving the yields coming from the ALCB burning and destroying.
     */
    function setShareHolders(ShareHolder[] memory shareHolders_) public onlyOwner {
        _setShareHolders(shareHolders_);
    }

    /**
     * @notice Set the operational wallet. The operational wallet is the account that will receive
     * the operational costs.
     * @param operationalWallet_ The account that will receive the operational costs
     */
    function setOperationalWallet(address operationalWallet_) public onlyOwner {
        operationalWallet = operationalWallet_;
    }

    /**
     * @notice Mints ALCB. This function receives ether in the transaction and converts them into
     * ALCB.
     */
    function mint() public payable whenNotPaused {
        return _mintToken(msg.sender, msg.value);
    }

    /**
     * @notice Mints ALCB. This function receives ether in the transaction and converts them into
     * ALCB.
     * @param to_ The account to where the tokens will be minted
     */
    function mintTo(address to_) public payable whenNotPaused {
        return _mintToken(to_, msg.value);
    }

    /**
     * @notice Burn the tokens without sending ether back to user as the normal burn
     * function. The ether will be distributed to the stakeholders and to the operational wallet.
     * @param amount_ the number of ALCB to be burned
     */
    function destroyTokens(uint256 amount_) public nonReentrant {
        return _burnToken(msg.sender, operationalWallet, amount_, true);
    }

    /**
     * @notice Burn ALCBs and send the ether received to the sender after deducting the market
     * spread.
     * @param amount_ The amount of ALCB being burned
     * */
    function burn(uint256 amount_) public nonReentrant {
        return _burnToken(msg.sender, msg.sender, amount_, false);
    }

    /**
     * @notice Burn ALCBs and send the ether received to a chosen account after deducting the market
     * spread.
     * @param to_ The account to where the ether from the burning will be send
     * @param amount_ The amount of ALCBs being burned
     * */
    function burnTo(address to_, uint256 amount_) public nonReentrant {
        return _burnToken(msg.sender, to_, amount_, false);
    }

    /**
     * @notice Get the share holders of the contract. The share holders are the accounts that will
     * be receiving the yields coming from the ALCB burning and destroying
     * @return the share holders of the contract
     */
    function getShareHolders() public view returns (ShareHolder[] memory) {
        return shareHolders;
    }

    /**
     * @notice Get the share holders count. The share holders are the accounts that will
     * be receiving the yields coming from the ALCB burning and destroying
     * @return the share holders count
     */
    function getShareHolderCount() public view returns (uint256) {
        return shareHolders.length;
    }

    // Internal function that sets the share holders of the contract. The share holders are the
    // accounts that will be receiving the yields coming from the ALCB burning and destroying.
    function _setShareHolders(ShareHolder[] memory shareHolders_) internal {
        uint32 sum = 0;
        for (uint256 i = 0; i < shareHolders_.length; i++) {
            if (shareHolders_[i].account == address(0)) {
                revert CannotTransferToZeroAddress();
            }
            if (shareHolders_[i].percentage == 0) {
                revert PercentagesCannotBeZero();
            }
            sum += shareHolders_[i].percentage;
        }
        if (sum != PERCENTAGE_SCALE) {
            revert ShareHoldersPercentageSumError(sum, PERCENTAGE_SCALE);
        }
        delete shareHolders;
        for (uint256 i = 0; i < shareHolders_.length; i++) {
            shareHolders.push(shareHolders_[i]);
        }
    }

    // Internal function that mints the ALCB tokens following the bounding price curve.
    function _mintToken(address to_, uint256 amountEth_) internal {
        poolBalance += amountEth_;
        ERC20._mint(to_, amountEth_);
    }

    // Internal function that burns the ALCB tokens, and sends the ether to the user after deducting
    // the market spread.
    function _burnToken(address from_, address to_, uint256 amount_, bool isToDestroy) internal {
        if (amount_ == 0) {
            return;
        }
        if (amount_ > poolBalance) {
            revert InvalidBurnAmount(amount_);
        }
        poolBalance -= amount_;
        ERC20._burn(from_, amount_);
        // When isToDestroy == FALSE, tokens are being burned and ether is split between
        // address to_ and shareHolders as determined by the marketSpreadFactor.
        // operationalEth tracks the value sent to address to_.
        uint256 operationalEth = (amount_ * marketSpreadFactor) / PERCENTAGE_SCALE;
        // When isToDestroy == TRUE, tokens are being destroyed and ether is split between
        // operationalWallet and shareHolders as determined by the operationalCostFactor.
        // operationalEth tracks the value sent to operationalWallet.
        if (isToDestroy) {
            operationalEth = (amount_ * operationalCostFactor) / PERCENTAGE_SCALE;
        }
        _distributeToShareHolders(amount_ - operationalEth);
        Address.sendValue(payable(to_), operationalEth);
    }

    /// Distributes the yields from the ALCB burning to all stake holders.
    function _distributeToShareHolders(uint256 distributeAmount_) internal {
        if (distributeAmount_ == 0) {
            return;
        }
        uint256 paidAmount = 0;
        for (uint256 i = 0; i < shareHolders.length; i++) {
            ShareHolder memory shareHolder = shareHolders[i];
            uint256 amount;
            // sending the remainders of the integer division to the last share holder
            if (i != shareHolders.length - 1) {
                amount = (distributeAmount_ * shareHolder.percentage) / PERCENTAGE_SCALE;
            } else {
                amount = distributeAmount_ - paidAmount;
            }
            paidAmount += amount;
            if (shareHolder.isMagicTransfer) {
                _safeTransferEthWithMagic(IMagicEthTransfer(shareHolder.account), amount);
            } else {
                // send the share to the share holder
                Address.sendValue(payable(shareHolder.account), amount);
            }
        }
    }
}
