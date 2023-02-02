// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "forge-std/Test.sol";
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
    uint256 private _marketSpread = 4;

    function setUp() public {
        Setup.BaseTokensFixture memory fixture = Setup.deployFactoryAndBaseTokens(vm);
        _stakingRouter = fixture.stakingRouter;
        _alcb = fixture.alcb;
    }

    function testMintsAlcbFromEtherValueToSender() public {
        uint256 totalSupplyBefore = _alcb.totalSupply();
        uint256 etherToSend = 4 ether;
        // fund the address
        vm.deal(_randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        // call the mint function
        uint256 alcbMinted = _alcb.mint{value: etherToSend}(0);

        // check the amount of alcb minted
        assertEq(402028731704364116575, alcbMinted);
        assertEq(_alcb.balanceOf(_randomAddress), alcbMinted);
        assertEq(_alcb.getPoolBalance(), etherToSend / _marketSpread);
        assertEq(_alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintToMintsAlcbFromEtherValueToSpecifiedAddress() public {
        uint256 totalSupplyBefore = _alcb.totalSupply();
        uint256 etherToSend = 4 ether;
        // fund the address
        vm.deal(_randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        // call the mint function
        uint256 alcbMinted = _alcb.mintTo{value: etherToSend}(_randomAddress2, 0);

        // check the amount of alcb minted
        assertEq(402028731704364116575, alcbMinted);
        assertEq(_alcb.balanceOf(_randomAddress2), alcbMinted);
        assertEq(_alcb.getPoolBalance(), etherToSend / _marketSpread);
        assertEq(_alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintToBigEthAmount() public {
        uint256 totalSupplyBefore = _alcb.totalSupply();
        uint256 etherToSend = 70000000000 ether;
        // fund the address
        vm.deal(_randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        // call the mint function
        uint256 alcbMinted = _alcb.mintTo{value: etherToSend}(_randomAddress2, 0);

        // check the amount of alcb minted
        assertEq(70001004975246203818081563855, alcbMinted);
        assertEq(_alcb.balanceOf(_randomAddress2), alcbMinted);
        assertEq(_alcb.getPoolBalance(), etherToSend / _marketSpread);
        assertEq(_alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintToFuzz(uint96 etherToSend, address destinationAddress) public {
        vm.assume(destinationAddress != _zeroAddress);
        vm.assume(etherToSend > _marketSpread);

        uint256 totalSupplyBefore = _alcb.totalSupply();

        // fund the address
        vm.deal(_randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        // call the mint function
        uint256 alcbMinted = _alcb.mintTo{value: etherToSend}(destinationAddress, 0);

        // check the amount of alcb minted
        assertEq(_alcb.balanceOf(destinationAddress), alcbMinted);
        assertEq(_alcb.getPoolBalance(), etherToSend / _marketSpread);
        assertEq(_alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintFuzz(uint96 etherToSend) public {
        vm.assume(etherToSend > _marketSpread);
        uint256 totalSupplyBefore = _alcb.totalSupply();

        // fund the address
        vm.deal(_randomAddress, etherToSend);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        // call the mint function
        uint256 alcbMinted = _alcb.mint{value: etherToSend}(0);

        // check the amount of alcb minted
        assertEq(_alcb.balanceOf(_randomAddress), alcbMinted);
        assertEq(_alcb.getPoolBalance(), etherToSend / _marketSpread);
        assertEq(_alcb.totalSupply(), totalSupplyBefore + alcbMinted);
    }

    function testMintAlcbWithMinValueNotMetReverts() public {
        vm.expectRevert(
            abi.encodeWithSelector(UtilityTokenErrors.MinimumValueNotMet.selector, 0, _marketSpread)
        );
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        _alcb.mint{value: 0 ether}(0);
    }

    function testMintAlcbToZeroAddressReverts() public {
        // fund the address
        vm.deal(_randomAddress, 4 ether);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        vm.expectRevert(bytes("ERC20: mint to the zero address"));

        _alcb.mintTo{value: 4 ether}(_zeroAddress, 0);
    }
}
