// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.16;

import "forge-std/Test.sol";
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
    uint256 marketSpread = 4;

    function setUp() public {
        Setup.BaseTokensFixture memory fixture = Setup.deployFactoryAndBaseTokens(vm);
        stakingRouter = fixture.stakingRouter;
        alcb = fixture.alcb;
    }

    function testMintsAlcbFromEtherValueToSender() public {
        uint256 totalSupplyBefore = alcb.totalSupply();
        uint256 etherToSend = 4 ether;
        // fund the address
        vm.deal(randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        // call the mint function
        uint256 alcbMinted = alcb.mint{value: etherToSend}(0);

        // check the amount of alcb minted
        assertEq(402028731704364116575, alcbMinted);
        assertEq(alcb.balanceOf(randomAddress), alcbMinted);
        assertEq(alcb.getPoolBalance(), etherToSend / marketSpread);
        assertEq(alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintToMintsAlcbFromEtherValueToSpecifiedAddress() public {
        uint256 totalSupplyBefore = alcb.totalSupply();
        uint256 etherToSend = 4 ether;
        // fund the address
        vm.deal(randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        // call the mint function
        uint256 alcbMinted = alcb.mintTo{value: etherToSend}(randomAddress2, 0);

        // check the amount of alcb minted
        assertEq(402028731704364116575, alcbMinted);
        assertEq(alcb.balanceOf(randomAddress2), alcbMinted);
        assertEq(alcb.getPoolBalance(), etherToSend / marketSpread);
        assertEq(alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintToBigEthAmount() public {
        uint256 totalSupplyBefore = alcb.totalSupply();
        uint256 etherToSend = 70000000000 ether;
        // fund the address
        vm.deal(randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        // call the mint function
        uint256 alcbMinted = alcb.mintTo{value: etherToSend}(randomAddress2, 0);

        // check the amount of alcb minted
        assertEq(70001004975246203818081563855, alcbMinted);
        assertEq(alcb.balanceOf(randomAddress2), alcbMinted);
        assertEq(alcb.getPoolBalance(), etherToSend / marketSpread);
        assertEq(alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintToFuzz(uint96 etherToSend, address destinationAddress) public {
        vm.assume(destinationAddress != zeroAddress);
        vm.assume(etherToSend > marketSpread);

        uint256 totalSupplyBefore = alcb.totalSupply();

        // fund the address
        vm.deal(randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        // call the mint function
        uint256 alcbMinted = alcb.mintTo{value: etherToSend}(destinationAddress, 0);

        // check the amount of alcb minted
        assertEq(alcb.balanceOf(destinationAddress), alcbMinted);
        assertEq(alcb.getPoolBalance(), etherToSend / marketSpread);
        assertEq(alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintFuzz(uint96 etherToSend) public {
        vm.assume(etherToSend > marketSpread);
        uint256 totalSupplyBefore = alcb.totalSupply();

        // fund the address
        vm.deal(randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        // call the mint function
        uint256 alcbMinted = alcb.mint{value: etherToSend}(0);

        // check the amount of alcb minted
        assertEq(alcb.balanceOf(randomAddress), alcbMinted);
        assertEq(alcb.getPoolBalance(), etherToSend / marketSpread);
        assertEq(alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintAlcbWithMinValueNotMetReverts() public {
        vm.expectRevert(
            abi.encodeWithSelector(UtilityTokenErrors.MinimumValueNotMet.selector, 0, marketSpread)
        );
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        alcb.mint{value: 0 ether}(0);
    }

    function testMintAlcbToZeroAddressReverts() public {
        // fund the address
        vm.deal(randomAddress, 4 ether);
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        vm.expectRevert(bytes("ERC20: mint to the zero address"));

        alcb.mintTo{value: 4 ether}(zeroAddress, 0);
    }
}
