// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.16;

import "contracts/AliceNetFactory.sol";
import "test/contract-mocks/legacyToken/LegacyToken.sol";
import "test/contract-mocks/bToken/BridgeRouterMock.sol";
import "contracts/ALCA.sol";
import "contracts/ALCB.sol";
import "contracts/PublicStaking.sol";

library Setup {
    struct BaseTokensFixture {
        LegacyToken legacyToken;
        CentralBridgeRouterMock stakingRouter;
        AliceNetFactory aliceNetFactory;
        ALCA alca;
        ALCB alcb;
        PublicStaking publicStaking;
    }

    function deployFactoryAndBaseTokens() public returns (BaseTokensFixture memory fixture) {
        LegacyToken legacyToken = new LegacyToken();
        CentralBridgeRouterMock stakingRouter = new CentralBridgeRouterMock(1000);
        AliceNetFactory aliceNetFactory = new AliceNetFactory(address(legacyToken));
        ALCA alca = ALCA(aliceNetFactory.getALCAAddress());

        // deploy ALCB and register it
        bytes memory creationCode = abi.encodePacked(
            type(ALCB).creationCode,
            abi.encodePacked([address(stakingRouter)])
        );

        // format the ALCB name string into bytes32
        bytes32 name = "ALCB";

        address alcbAddress = aliceNetFactory.deployCreateAndRegister(creationCode, name);
        ALCB alcb = ALCB(alcbAddress);

        bytes memory publicStakingCreationCode = abi.encodePacked(type(PublicStaking).creationCode);

        address publicStakingAddress = aliceNetFactory.deployCreate(publicStakingCreationCode);
        PublicStaking publicStaking = PublicStaking(publicStakingAddress);
        bytes memory initializePublicStaking = abi.encodeWithSelector(
            publicStaking.initialize.selector
        );
        aliceNetFactory.callAny(publicStakingAddress, 0, initializePublicStaking);

        fixture = BaseTokensFixture({
            legacyToken: legacyToken,
            stakingRouter: stakingRouter,
            aliceNetFactory: aliceNetFactory,
            alca: alca,
            alcb: alcb,
            publicStaking: publicStaking
        });
    }
}
