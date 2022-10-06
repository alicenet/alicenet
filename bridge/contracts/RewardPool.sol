// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "contracts/interfaces/IStakingNFT.sol";
import "contracts/BonusPool.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/lockup/AccessControlled.sol";

/**
 * @notice RewardPool holds all ether and ALCA that is part of reserved amount
 * of rewards on base positions. deployed by lockup
 */
contract RewardPool is AccessControlled, EthSafeTransfer, ERC20SafeTransfer {
    address internal immutable _alca;
    address internal immutable _lockupContract;
    address internal immutable _bonusPool;
    uint256 internal _ethReserve;
    uint256 internal _tokenReserve;

    constructor(
        address alca_,
        address aliceNetFactory_,
        uint256 totalBonusAmount_
    ) {
        _bonusPool = address(
            new BonusPool(aliceNetFactory_, msg.sender, address(this), totalBonusAmount_)
        );
        _lockupContract = msg.sender;
        _alca = alca_;
    }

    function getBonusPoolAddress() public view returns (address) {
        return _getBonusPoolAddress();
    }

    function getLockupContractAddress() public view returns (address) {
        return _getLockupContractAddress();
    }

    function getTokenReserve() public view returns (uint256) {
        return _tokenReserve;
    }

    function getEthReserve() public view returns (uint256) {
        return _ethReserve;
    }

    function deposit(uint256 numTokens_) public payable onlyLockupOrBonus {
        _tokenReserve += numTokens_;
        _ethReserve += msg.value;
    }

    function payout(
        uint256 total_,
        uint256 shares_,
        bool isLastPosition
    ) public onlyLockup returns (uint256, uint256) {
        uint256 proportionalEth;
        uint256 proportionalTokens;
        // last position gets gets any remainder left on this contract
        if (isLastPosition) {
            proportionalEth = address(this).balance;
            proportionalTokens = IERC20(_alca).balanceOf(address(this));
        } else {
            proportionalEth = (_ethReserve * shares_) / total_;
            proportionalTokens = (_tokenReserve * shares_) / total_;
        }
        _safeTransferERC20(IERC20Transferable(_alca), _lockupContract, proportionalTokens);
        _safeTransferEth(payable(_lockupContract), proportionalEth);
        return (proportionalTokens, proportionalEth);
    }

    function _getLockupContractAddress() internal view override returns (address) {
        return _lockupContract;
    }

    function _getBonusPoolAddress() internal view override returns (address) {
        return _bonusPool;
    }

    function _getRewardPoolAddress() internal view override returns (address) {
        return address(this);
    }
}
