// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "forge-std/Test.sol";
import "forge-std/console.sol";
import "contracts/ALCB.sol";
import "test/contract-mocks/bToken/BridgeRouterMock.sol";
import "contracts/libraries/errors/UtilityTokenErrors.sol";
import "test_sol/Setup.sol";

contract ALCBTest is Test {
    CentralBridgeRouterMock private _stakingRouter;
    ALCB private _alcb;
    address private _randomAddress = address(0x000000000000000000000000000000000000dEaD);
    address private _randomAddress2 = address(0x00000000000000000000000000000000DeaDBeef);
    address private _zeroAddress = address(0x0000000000000000000000000000000000000000);
    uint96 private _marketSpread = 4;
    uint256 private _minALCBs = 0;
    uint256 private _alcbs = 0;
    uint256 private _etherIn = 40 ether;

    function setUp() public {
        Setup.BaseTokensFixture memory fixture = Setup.deployFactoryAndBaseTokens(vm);
        _stakingRouter = fixture.stakingRouter;
        _alcb = fixture.alcb;

        // mint some alcb
        vm.deal(_randomAddress, _etherIn);
        vm.prank(_randomAddress);
        _alcbs = _alcb.mint{value: _etherIn}(0);
    }

    function testBurn() public {
        uint256 addressEthBalanceBefore = _randomAddress.balance;
        uint256 remaining = 100 ether;
        uint256 burnQuantity = _alcbs - remaining;
        uint256 minEth = 0;
        vm.prank(_randomAddress);
        uint256 ethReturned = _alcb.burn(burnQuantity, minEth);
        assertEq(9751261920046697614, ethReturned);
        assertEq(_randomAddress.balance, addressEthBalanceBefore + ethReturned);
        assertEq(_alcb.balanceOf(_randomAddress), remaining);
    }

    function testBurnTo() public {
        uint256 addressEthBalanceBefore = _randomAddress2.balance;
        uint256 remaining = 100 ether;
        uint256 burnQuantity = _alcbs - remaining;
        uint256 minEth = 0;
        vm.prank(_randomAddress);
        uint256 ethReturned = _alcb.burnTo(_randomAddress2, burnQuantity, minEth);
        assertEq(9751261920046697614, ethReturned);
        assertEq(_randomAddress2.balance, addressEthBalanceBefore + ethReturned);
        assertEq(_alcb.balanceOf(_randomAddress), remaining);
        assertEq(_alcb.totalSupply(), remaining);
    }

    function testBurnMoreThanSupplyFails() public {
        uint256 burnQuantity = _alcbs + 1;
        uint256 minEth = 0;
        vm.expectRevert(
            abi.encodeWithSelector(
                UtilityTokenErrors.BurnAmountExceedsSupply.selector,
                burnQuantity,
                _alcbs
            )
        );
        vm.prank(_randomAddress);
        _alcb.burn(burnQuantity, minEth);
    }

    function testBurnFullSupplySucceeds() public {
        uint256 addressEthBalanceBefore = _randomAddress.balance;
        uint256 burnQuantity = _alcb.totalSupply();
        uint256 minEth = 0;

        vm.prank(_randomAddress);

        uint256 ethReturned = _alcb.burn(burnQuantity, minEth);

        assertEq(_randomAddress.balance, addressEthBalanceBefore + ethReturned);
        assertEq(_alcb.balanceOf(_randomAddress), 0);
        assertEq(_alcb.totalSupply(), 0);
    }

    function testBurnZeroFails() public {
        uint256 burnQuantity = 0;
        uint256 minEth = 0;
        vm.expectRevert(
            abi.encodeWithSelector(UtilityTokenErrors.InvalidBurnAmount.selector, burnQuantity)
        );
        vm.prank(_randomAddress);
        _alcb.burn(burnQuantity, minEth);
    }

    function testBurnFuzz(uint96 etherToSend) public {
        vm.assume(etherToSend > _marketSpread);

        // fund the address
        vm.deal(_randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        // call the mint function
        uint256 alcbMinted = _alcb.mint{value: etherToSend}(0);
        uint256 poolBalanceBefore = _alcb.getPoolBalance();
        uint256 totalSupplyBefore = _alcb.totalSupply();
        uint256 addressBalanceBefore = _alcb.balanceOf(_randomAddress);
        uint256 addressEthBalanceBefore = _randomAddress.balance;
        uint256 minEth = 0;
        vm.prank(_randomAddress);
        uint256 ethReturned = _alcb.burn(alcbMinted, minEth);

        assertEq(_alcb.balanceOf(_randomAddress), addressBalanceBefore - alcbMinted);
        assertEq(_alcb.getPoolBalance(), poolBalanceBefore - ethReturned);
        assertEq(_alcb.totalSupply(), totalSupplyBefore - alcbMinted);
        assertEq(_randomAddress.balance, addressEthBalanceBefore + ethReturned);
    }

    function testBurnToFuzz(uint96 etherToSend, address destinationAddress) public {
        vm.assume(destinationAddress != _zeroAddress);
        vm.assume(etherToSend > _marketSpread);

        // fund the address
        vm.deal(_randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        // call the mint function
        uint256 alcbMinted = _alcb.mint{value: etherToSend}(0);
        uint256 poolBalanceBefore = _alcb.getPoolBalance();
        uint256 totalSupplyBefore = _alcb.totalSupply();
        uint256 addressBalanceBefore = _alcb.balanceOf(_randomAddress);
        uint256 addressEthBalanceBefore = destinationAddress.balance;
        uint256 minEth = 0;
        vm.prank(_randomAddress);
        uint256 ethReturned = _alcb.burnTo(destinationAddress, alcbMinted, minEth);

        assertEq(_alcb.balanceOf(_randomAddress), addressBalanceBefore - alcbMinted);
        assertEq(_alcb.getPoolBalance(), poolBalanceBefore - ethReturned);
        assertEq(_alcb.totalSupply(), totalSupplyBefore - alcbMinted);
        assertEq(destinationAddress.balance, addressEthBalanceBefore + ethReturned);
    }
}
