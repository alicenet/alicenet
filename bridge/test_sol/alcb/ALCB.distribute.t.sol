// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.16;

import "forge-std/Test.sol";
import "forge-std/console.sol";
import "contracts/ALCB.sol";
import "contracts/Distribution.sol";
import "test/contract-mocks/bToken/BridgeRouterMock.sol";
import "contracts/libraries/errors/UtilityTokenErrors.sol";
import "contracts/Distribution.sol";
import "contracts/AliceNetFactory.sol";
import "test_sol/Setup.sol";

contract ALCBTest is Test {
    Setup.Fixture fixture;
    CentralBridgeRouterMock stakingRouter;
    ALCB alcb;
    AliceNetFactory aliceNetFactory;
    address randomAddress = address(0x000000000000000000000000000000000000dEaD);
    address randomAddress2 = address(0x00000000000000000000000000000000DeaDBeef);
    address zeroAddress = address(0x0000000000000000000000000000000000000000);
    uint96 marketSpread = 4;
    uint256 minALCBs = 0;
    uint256 alcbs = 0;
    uint256 etherIn = 40 ether;

    function setUp() public {
        fixture = Setup.deployFixture(vm, false, false, false);
        stakingRouter = fixture.stakingRouter;
        alcb = fixture.alcb;
        aliceNetFactory = fixture.aliceNetFactory;

        // mint some alcb
        vm.deal(randomAddress, etherIn);
        vm.prank(randomAddress);
        alcbs = alcb.mint{value: etherIn}(0);
    }

    function testDistribution(uint96 etherToSend) public {
        vm.assume(etherToSend > marketSpread);

        uint256 validatorStakingSplit = 250;
        uint256 publicStakingSplit = 250;
        uint256 liquidityProviderStakingSplit = 250;
        uint256 protocolFeeSplit = 250;
        
        updateDistributionContract(
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
        uint256 distributable = mintDistribute(etherToSend, 0);
        assertSplitsBalance(
            distributable,
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
    }

    function updateDistributionContract(
        uint256 validatorStakingSplit,
        uint256 publicStakingSplit,
        uint256 liquidityProviderStakingSplit,
        uint256 protocolFeeSplit
    ) internal returns (Distribution) {
        vm.prank(fixture.adminAddress);
        address deployAddress = fixture.aliceNetFactory.deployCreate(
            abi.encodePacked(
                type(Distribution).creationCode,
                abi.encodePacked(
                    validatorStakingSplit,
                    publicStakingSplit,
                    liquidityProviderStakingSplit,
                    protocolFeeSplit
                )
            )
        );

        vm.prank(fixture.adminAddress);
        fixture.aliceNetFactory.upgradeProxy("Distribution", deployAddress, "");
        return Distribution(deployAddress);
    }

    function mintDistribute(uint96 ethIn, uint256 distributable) internal returns (uint256) {
        // fund the address
        vm.deal(randomAddress, ethIn);
        // mock the call to use the address for the next call
        vm.prank(randomAddress);

        alcb.mint{value: ethIn}(0);
        uint256 yield = alcb.getYield();
        distributable = distributable + yield;

        alcb.distribute();
        assertEq(alcb.getYield(), 0);

        return distributable;
    }

    function assertSplitsBalance(
        uint256 excess,
        uint256 validatorStakingSplit,
        uint256 publicStakingSplit,
        uint256 liquidityProviderStakingSplit,
        uint256 protocolFeeSplit
    ) internal {
        uint256 PERCENTAGE_SCALE = 1000;
        uint256 protocolFeeShare = (excess * protocolFeeSplit) / PERCENTAGE_SCALE;
        // split remaining between validators, stakers and lp stakers
        uint256 publicStakingShare = (excess * publicStakingSplit) / PERCENTAGE_SCALE;
        uint256 lpStakingShare = (excess * liquidityProviderStakingSplit) / PERCENTAGE_SCALE;
        // then give validators the rest
        uint256 validatorStakingShare = excess -
            (protocolFeeShare + publicStakingShare + lpStakingShare);

        assertEq(address(fixture.validatorStaking).balance, validatorStakingShare);
        assertEq(address(fixture.publicStaking).balance, publicStakingShare);
        assertEq(address(fixture.liquidityProviderStaking).balance, lpStakingShare);
        assertEq(address(fixture.foundation).balance, protocolFeeShare);
    }
}
