// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

// import "contracts/StakeNFT.sol";

// abstract contract BaseMock {
//     StakeNFT public stakeNFT;
//     MadTokenMock public madToken;

//     receive() external payable virtual {}

//     function setTokens(MadTokenMock madToken_, StakeNFT stakeNFT_) public {
//         stakeNFT = stakeNFT_;
//         madToken = madToken_;
//     }

//     function mint(uint256 amount_) public returns (uint256) {
//         return stakeNFT.mint(amount_);
//     }

//     function mintTo(
//         address to_,
//         uint256 amount_,
//         uint256 duration_
//     ) public returns (uint256) {
//         return stakeNFT.mintTo(to_, amount_, duration_);
//     }

//     function burn(uint256 tokenID) public returns (uint256, uint256) {
//         return stakeNFT.burn(tokenID);
//     }

//     function burnTo(address to_, uint256 tokenID) public returns (uint256, uint256) {
//         return stakeNFT.burnTo(to_, tokenID);
//     }

//     function approve(address who, uint256 amount_) public returns (bool) {
//         return madToken.approve(who, amount_);
//     }

//     function depositToken(uint256 amount_) public {
//         stakeNFT.depositToken(42, amount_);
//     }

//     function depositEth(uint256 amount_) public {
//         stakeNFT.depositEth{value: amount_}(42);
//     }

//     function collectToken(uint256 tokenID_) public returns (uint256 payout) {
//         return stakeNFT.collectToken(tokenID_);
//     }

//     function collectEth(uint256 tokenID_) public returns (uint256 payout) {
//         return stakeNFT.collectEth(tokenID_);
//     }

//     function approveNFT(address to, uint256 tokenID_) public {
//         return stakeNFT.approve(to, tokenID_);
//     }

//     function setApprovalForAll(address to, bool approve_) public {
//         return stakeNFT.setApprovalForAll(to, approve_);
//     }

//     function transferFrom(
//         address from,
//         address to,
//         uint256 tokenID_
//     ) public {
//         return stakeNFT.transferFrom(from, to, tokenID_);
//     }

//     function safeTransferFrom(
//         address from,
//         address to,
//         uint256 tokenID_,
//         bytes calldata data
//     ) public {
//         return stakeNFT.safeTransferFrom(from, to, tokenID_, data);
//     }

//     function lockWithdraw(uint256 tokenID, uint256 lockDuration) public returns (uint256) {
//         return stakeNFT.lockWithdraw(tokenID, lockDuration);
//     }
// }
