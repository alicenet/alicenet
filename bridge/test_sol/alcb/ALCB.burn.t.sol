// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "forge-std/Test.sol";
import "forge-std/console.sol";
import "contracts/ALCB.sol";
import "test/contract-mocks/bToken/BridgeRouterMock.sol";
import "contracts/libraries/errors/UtilityTokenErrors.sol";
import "test_sol/Setup.sol";

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
        Setup.BaseTokensFixture memory fixture = Setup.deployFactoryAndBaseTokens(vm);
        stakingRouter = fixture.stakingRouter;
        alcb = fixture.alcb;

        // mint some alcb
        vm.deal(randomAddress, etherIn);
        vm.prank(randomAddress);
        alcbs = alcb.mint{value: etherIn}(0);
    }

    function testBurn() public {
        uint256 addressEthBalanceBefore = randomAddress.balance;
        uint256 remaining = 100 ether;
        uint256 burnQuantity = alcbs - remaining;
        uint256 minEth = 0;
        vm.prank(randomAddress);
        uint256 ethReturned = alcb.burn(burnQuantity, minEth);
        assertEq(9751261920046697614, ethReturned);
        assertEq(randomAddress.balance, addressEthBalanceBefore + ethReturned);
        assertEq(alcb.balanceOf(randomAddress), remaining);
    }

    function testBurnTo() public {
        uint256 addressEthBalanceBefore = randomAddress2.balance;
        uint256 remaining = 100 ether;
        uint256 burnQuantity = alcbs - remaining;
        uint256 minEth = 0;
        vm.prank(randomAddress);
        uint256 ethReturned = alcb.burnTo(randomAddress2, burnQuantity, minEth);
        assertEq(9751261920046697614, ethReturned);
        assertEq(randomAddress2.balance, addressEthBalanceBefore + ethReturned);
        assertEq(alcb.balanceOf(randomAddress), remaining);
        assertEq(alcb.totalSupply(), remaining);
    }

    function testBurnMoreThanSupplyFails() public {
        uint256 burnQuantity = alcbs + 1;
        uint256 minEth = 0;
        vm.expectRevert(
            abi.encodeWithSelector(
                UtilityTokenErrors.BurnAmountExceedsSupply.selector,
                burnQuantity,
                alcbs
            )
        );
        vm.prank(randomAddress);
        alcb.burn(burnQuantity, minEth);
    }

    function testBurnFullSupplySucceeds() public {
        uint256 addressEthBalanceBefore = randomAddress.balance;
        uint256 burnQuantity = alcb.totalSupply();
        uint256 minEth = 0;

        vm.prank(randomAddress);

        uint256 ethReturned = alcb.burn(burnQuantity, minEth);

        assertEq(randomAddress.balance, addressEthBalanceBefore + ethReturned);
        assertEq(alcb.balanceOf(randomAddress), 0);
        assertEq(alcb.totalSupply(), 0);
    }

    function testBurnZeroFails() public {
        uint256 burnQuantity = 0;
        uint256 minEth = 0;
        vm.expectRevert(
            abi.encodeWithSelector(UtilityTokenErrors.InvalidBurnAmount.selector, burnQuantity)
        );
        vm.prank(randomAddress);
        alcb.burn(burnQuantity, minEth);
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
        uint256 addressEthBalanceBefore = randomAddress.balance;
        uint256 minEth = 0;
        vm.prank(randomAddress);
        uint256 ethReturned = alcb.burn(alcbMinted, minEth);

        assertEq(alcb.balanceOf(randomAddress), addressBalanceBefore - alcbMinted);
        assertEq(alcb.getPoolBalance(), poolBalanceBefore - ethReturned);
        assertEq(alcb.totalSupply(), totalSupplyBefore - alcbMinted);
        assertEq(randomAddress.balance, addressEthBalanceBefore + ethReturned);
    }

    function testBurnToFuzz(uint96 etherToSend, address destinationAddress) public {
        vm.assume(destinationAddress != zeroAddress);
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
        uint256 addressEthBalanceBefore = destinationAddress.balance;
        uint256 minEth = 0;
        vm.prank(randomAddress);
        uint256 ethReturned = alcb.burnTo(destinationAddress, alcbMinted, minEth);

        assertEq(alcb.balanceOf(randomAddress), addressBalanceBefore - alcbMinted);
        assertEq(alcb.getPoolBalance(), poolBalanceBefore - ethReturned);
        assertEq(alcb.totalSupply(), totalSupplyBefore - alcbMinted);
        assertEq(destinationAddress.balance, addressEthBalanceBefore + ethReturned);
    }
}
