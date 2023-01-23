// SPDX-License-Identifier: MIT
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
    Setup.Fixture private _fixture;
    CentralBridgeRouterMock private _stakingRouter;
    ALCB private _alcb;
    AliceNetFactory private _aliceNetFactory;
    address private _randomAddress = address(0x000000000000000000000000000000000000dEaD);
    address private _randomAddress2 = address(0x00000000000000000000000000000000DeaDBeef);
    address private _zeroAddress = address(0x0000000000000000000000000000000000000000);
    uint96 private _marketSpread = 4;
    uint256 private _minALCBs = 0;
    uint256 private _alcbs = 0;
    uint256 private _etherIn = 40 ether;

    function setUp() public {
        _fixture = Setup.deployFixture(vm, false, false, false);
        _stakingRouter = _fixture.stakingRouter;
        _alcb = _fixture.alcb;
        _aliceNetFactory = _fixture.aliceNetFactory;

        // mint some alcb
        vm.deal(_randomAddress, _etherIn);
        vm.prank(_randomAddress);
        _alcbs = _alcb.mint{value: _etherIn}(0);
    }

    function testDistributionWithoutFoundation(uint96 etherToSend) public {
        vm.assume(etherToSend > _marketSpread);

        uint256 validatorStakingSplit = 350;
        uint256 publicStakingSplit = 350;
        uint256 liquidityProviderStakingSplit = 300;
        uint256 protocolFeeSplit = 0;

        _updateDistributionContract(
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
        uint256 distributable = _mintDistribute(etherToSend, 0);
        _assertSplitsBalance(
            distributable,
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
    }

    function testDistributionWithoutLiquidityProviderStaking(uint96 etherToSend) public {
        vm.assume(etherToSend > _marketSpread);

        uint256 validatorStakingSplit = 350;
        uint256 publicStakingSplit = 350;
        uint256 liquidityProviderStakingSplit = 0;
        uint256 protocolFeeSplit = 300;

        _updateDistributionContract(
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
        uint256 distributable = _mintDistribute(etherToSend, 0);
        _assertSplitsBalance(
            distributable,
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
    }

    function testDistributionWithoutPublicStaking(uint96 etherToSend) public {
        vm.assume(etherToSend > _marketSpread);

        uint256 validatorStakingSplit = 350;
        uint256 publicStakingSplit = 0;
        uint256 liquidityProviderStakingSplit = 350;
        uint256 protocolFeeSplit = 300;

        _updateDistributionContract(
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
        uint256 distributable = _mintDistribute(etherToSend, 0);
        _assertSplitsBalance(
            distributable,
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
    }

    function testDistributionWithoutValidatorStaking(uint96 etherToSend) public {
        vm.assume(etherToSend > _marketSpread);

        uint256 validatorStakingSplit = 0;
        uint256 publicStakingSplit = 350;
        uint256 liquidityProviderStakingSplit = 350;
        uint256 protocolFeeSplit = 300;

        _updateDistributionContract(
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
        uint256 distributable = _mintDistribute(etherToSend, 0);
        _assertSplitsBalance(
            distributable,
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
    }

    function testStandardDistribution(uint96 etherToSend) public {
        vm.assume(etherToSend > _marketSpread);

        uint256 validatorStakingSplit = 250;
        uint256 publicStakingSplit = 250;
        uint256 liquidityProviderStakingSplit = 250;
        uint256 protocolFeeSplit = 250;

        _updateDistributionContract(
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
        uint256 distributable = _mintDistribute(etherToSend, 0);
        _assertSplitsBalance(
            distributable,
            validatorStakingSplit,
            publicStakingSplit,
            liquidityProviderStakingSplit,
            protocolFeeSplit
        );
    }

    function _updateDistributionContract(
        uint256 validatorStakingSplit,
        uint256 publicStakingSplit,
        uint256 liquidityProviderStakingSplit,
        uint256 protocolFeeSplit
    ) internal returns (Distribution) {
        vm.prank(_fixture.adminAddress);
        address deployAddress = _fixture.aliceNetFactory.deployCreate(
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

        vm.prank(_fixture.adminAddress);
        _fixture.aliceNetFactory.upgradeProxy("Distribution", deployAddress, "");
        return Distribution(deployAddress);
    }

    function _mintDistribute(uint96 ethIn, uint256 distributable) internal returns (uint256) {
        // fund the address
        vm.deal(_randomAddress, ethIn);
        // mock the call to use the address for the next call
        vm.prank(_randomAddress);

        _alcb.mint{value: ethIn}(0);
        uint256 yield = _alcb.getYield();
        distributable = distributable + yield;

        _alcb.distribute();
        assertEq(_alcb.getYield(), 0);

        return distributable;
    }

    function _assertSplitsBalance(
        uint256 excess,
        uint256 validatorStakingSplit,
        uint256 publicStakingSplit,
        uint256 liquidityProviderStakingSplit,
        uint256 protocolFeeSplit
    ) internal {
        uint256 percentageScale = 1000;
        uint256 protocolFeeShare = (excess * protocolFeeSplit) / percentageScale;
        // split remaining between validators, stakers and lp stakers
        uint256 publicStakingShare = (excess * publicStakingSplit) / percentageScale;
        uint256 lpStakingShare = (excess * liquidityProviderStakingSplit) / percentageScale;
        // then give validators the rest
        uint256 validatorStakingShare = excess -
            (protocolFeeShare + publicStakingShare + lpStakingShare);

        assertEq(address(_fixture.validatorStaking).balance, validatorStakingShare);
        assertEq(address(_fixture.publicStaking).balance, publicStakingShare);
        assertEq(address(_fixture.liquidityProviderStaking).balance, lpStakingShare);
        assertEq(address(_fixture.foundation).balance, protocolFeeShare);
    }
}
