// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.16;

import "forge-std/Test.sol";
import "forge-std/console.sol";
import "contracts/ALCB.sol";
import "test/contract-mocks/bToken/BridgeRouterMock.sol";
import "contracts/libraries/errors/UtilityTokenErrors.sol";

contract ALCBTest is Test {
    CentralBridgeRouterMock stakingRouter;
    ALCB alcb;
    address randomAddress = address(0x000000000000000000000000000000000000dEaD);
    address randomAddress2 = address(0x00000000000000000000000000000000DeaDBeef);
    address zeroAddress = address(0x0000000000000000000000000000000000000000);
    uint96 marketSpread = 4;
    uint256 minALCBs = 0;
    uint256 alcbs = 0;
    uint256 etherIn = 40 ether;

    function setUp() public {
        stakingRouter = new CentralBridgeRouterMock(1000);
        alcb = new ALCB(address(stakingRouter));

        // mint some alcb
        vm.deal(randomAddress, etherIn);
        vm.prank(randomAddress);
        alcbs = alcb.mint{value: etherIn}(0);
    }

    function testBurn() public {
        uint256 remaining = 100 ether;
        uint256 burnQuantity = alcbs - remaining;
        uint256 minEth = 0;
        vm.prank(randomAddress);
        uint256 ethReturned = alcb.burn(burnQuantity, minEth);
        assertEq(9751261920046697614, ethReturned);
    }

    function testBurnFuzz(uint96 etherToSend) public {
        vm.assume(etherToSend > marketSpread);
        
        // fund the address
        vm.deal(randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        // call the mint function
        uint256 alcbMinted = alcb.mint{value: etherToSend}(0);
        uint256 poolBalanceBefore = alcb.getPoolBalance();
        uint256 totalSupplyBefore = alcb.totalSupply();
        uint256 addressBalanceBefore = alcb.balanceOf(randomAddress);

        uint256 minEth = 0;
        vm.prank(randomAddress);
        uint256 ethReturned = alcb.burn(alcbMinted, minEth);

        assertEq(alcb.balanceOf(randomAddress), addressBalanceBefore - alcbMinted);
        assertEq(alcb.getPoolBalance(), poolBalanceBefore - ethReturned);
        assertEq(alcb.totalSupply(), totalSupplyBefore - alcbMinted);
    }
}
