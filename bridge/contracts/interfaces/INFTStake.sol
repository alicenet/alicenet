// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

interface INFTStake {
    function getAccumulatorScaleFactor() external view returns (uint256);

    function getTotalShares() external view returns (uint256);

    function getTotalReserveEth() external view returns (uint256);

    function getTotalReserveMadToken() external view returns (uint256);

    function estimateEthCollection(uint256 tokenID_) external view returns (uint256 payout);

    function estimateTokenCollection(uint256 tokenID_) external view returns (uint256 payout);

    function estimateExcessToken() external view returns (uint256 excess);

    function estimateExcessEth() external view returns (uint256 excess);

    function skimExcessEth(address to_) external returns (uint256 excess);

    function skimExcessToken(address to_) external returns (uint256 excess);

    function depositToken(uint8 magic_, uint256 amount_) external;

    function depositEth(uint8 magic_) external payable;

    function lockPosition(
        address caller_,
        uint256 tokenID_,
        uint256 lockDuration_
    ) external returns (uint256);

    function lockOwnPosition(uint256 tokenID_, uint256 lockDuration_) external returns (uint256);

    function lockWithdraw(uint256 tokenID_, uint256 lockDuration_) external returns (uint256);

    function mint(uint256 amount_) external returns (uint256 tokenID);

    function mintTo(
        address to_,
        uint256 amount_,
        uint256 lockDuration_
    ) external returns (uint256 tokenID);

    function burn(uint256 tokenID_) external returns (uint256 payoutEth, uint256 payoutMadToken);

    function burnTo(address to_, uint256 tokenID_)
        external
        returns (uint256 payoutEth, uint256 payoutMadToken);

    function collectEth(uint256 tokenID_) external returns (uint256 payout);

    function collectToken(uint256 tokenID_) external returns (uint256 payout);

    function collectEthTo(address to_, uint256 tokenID_) external returns (uint256 payout);

    function collectTokenTo(address to_, uint256 tokenID_) external returns (uint256 payout);

    function getPosition(uint256 tokenID_)
        external
        view
        returns (
            uint256 shares,
            uint256 freeAfter,
            uint256 withdrawFreeAfter,
            uint256 accumulatorEth,
            uint256 accumulatorToken
        );

    function getEthAccumulator() external view returns (uint256 accumulator, uint256 slush);

    function getTokenAccumulator() external view returns (uint256 accumulator, uint256 slush);
}
