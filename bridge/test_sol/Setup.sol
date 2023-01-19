// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.16;

import "contracts/AliceNetFactory.sol";
import "test/contract-mocks/legacyToken/LegacyToken.sol";
import "test/contract-mocks/bToken/BridgeRouterMock.sol";
import "contracts/ALCA.sol";
import "contracts/ALCB.sol";
import "contracts/PublicStaking.sol";
import "contracts/ValidatorStaking.sol";
import "contracts/LiquidityProviderStaking.sol";
import "contracts/Foundation.sol";
import "test/contract-mocks/validatorPool/ValidatorPoolMock.sol";
import "contracts/ValidatorPool.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/ALCAMinter.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/interfaces/IStakingNFTDescriptor.sol";
import "contracts/InvalidTxConsumptionAccusation.sol";
import "contracts/MultipleProposalAccusation.sol";
import "contracts/Distribution.sol";
import "contracts/Dynamics.sol";
import "contracts/libraries/ethdkg/ETHDKGAccusations.sol";
import "contracts/StakingPositionDescriptor.sol";
import "contracts/libraries/ethdkg/ETHDKGPhases.sol";
import "test/contract-mocks/ethdkg/ETHDKGMock.sol";
import "contracts/ETHDKG.sol";
import "test/contract-mocks/snapshots/SnapshotsMock.sol";
import "contracts/Snapshots.sol";
import "contracts/ALCABurner.sol";
import {StdStorage, Vm} from "forge-std/Components.sol";

library Setup {
    address private constant adminAddress = address(0xdeAD00000000000000000000000000000000dEAd);
    struct BaseTokensFixture {
        LegacyToken legacyToken;
        CentralBridgeRouterMock stakingRouter;
        AliceNetFactory aliceNetFactory;
        ALCA alca;
        ALCB alcb;
        PublicStaking publicStaking;
        address adminAddress;
    }

    struct Fixture {
        LegacyToken legacyToken;
        CentralBridgeRouterMock stakingRouter;
        AliceNetFactory aliceNetFactory;
        ALCA alca;
        ALCB alcb;
        PublicStaking publicStaking;
        ALCAMinter alcaMinter;
        ALCABurner alcaBurner;
        ValidatorStaking validatorStaking;
        LiquidityProviderStaking liquidityProviderStaking;
        Foundation foundation;
        IValidatorPool validatorPool;
        ISnapshots snapshots;
        IETHDKG ethdkg;
        IStakingNFTDescriptor stakingPositionDescriptor;
        InvalidTxConsumptionAccusation invalidTxConsumptionAccusation;
        MultipleProposalAccusation multipleProposalAccusation;
        Distribution distribution;
        Dynamics dynamics;
        address adminAddress;
    }

    function deployFactoryAndBaseTokens(Vm vm) public returns (BaseTokensFixture memory fixture) {
        LegacyToken legacyToken = new LegacyToken();
        CentralBridgeRouterMock stakingRouter = new CentralBridgeRouterMock(1000);
        vm.prank(adminAddress);
        AliceNetFactory aliceNetFactory = new AliceNetFactory(address(legacyToken));
        ALCA alca = ALCA(aliceNetFactory.getALCAAddress());

        vm.prank(adminAddress);
        address alcbAddress = aliceNetFactory.deployCreateAndRegister(
            abi.encodePacked(type(ALCB).creationCode, abi.encodePacked([address(stakingRouter)])),
            "ALCB"
        );
        ALCB alcb = ALCB(alcbAddress);

        (, address proxyMultipleProposalAccusationAddress) = deployUpgradeableWithFactory(
            vm,
            adminAddress,
            aliceNetFactory,
            abi.encodePacked(type(PublicStaking).creationCode),
            "PublicStaking",
            ""
        );

        PublicStaking publicStaking = PublicStaking(proxyMultipleProposalAccusationAddress);

        vm.prank(adminAddress);
        aliceNetFactory.callAny(
            address(publicStaking),
            0,
            abi.encodeWithSelector(publicStaking.initialize.selector)
        );

        fixture = BaseTokensFixture({
            legacyToken: legacyToken,
            stakingRouter: stakingRouter,
            aliceNetFactory: aliceNetFactory,
            alca: alca,
            alcb: alcb,
            publicStaking: publicStaking,
            adminAddress: adminAddress
        });
    }

    function deployFixture(
        Vm vm,
        bool mockValidatorPool,
        bool mockSnapshots,
        bool mockETHDKG
    ) public returns (Fixture memory fixture) {
        BaseTokensFixture memory baseTokensFixture = deployFactoryAndBaseTokens(vm);

        fixture.legacyToken = baseTokensFixture.legacyToken;
        fixture.stakingRouter = baseTokensFixture.stakingRouter;
        fixture.aliceNetFactory = baseTokensFixture.aliceNetFactory;
        fixture.alca = baseTokensFixture.alca;
        fixture.alcb = baseTokensFixture.alcb;
        fixture.publicStaking = baseTokensFixture.publicStaking;
        fixture.adminAddress = baseTokensFixture.adminAddress;

        {
            (, address proxyValidatorStakingAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(ValidatorStaking).creationCode),
                "ValidatorStaking",
                abi.encodeWithSelector(ValidatorStaking.initialize.selector)
            );
            fixture.validatorStaking = ValidatorStaking(proxyValidatorStakingAddress);
        }
        {
            (, address proxyLiquidityProviderStakingAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(LiquidityProviderStaking).creationCode),
                "LiquidityProviderStaking",
                abi.encodeWithSelector(LiquidityProviderStaking.initialize.selector)
            );
            fixture.liquidityProviderStaking = LiquidityProviderStaking(
                proxyLiquidityProviderStakingAddress
            );
        }
        {
            (, address proxyFoundationAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(Foundation).creationCode),
                "Foundation",
                abi.encodeWithSelector(Foundation.initialize.selector)
            );

            fixture.foundation = Foundation(proxyFoundationAddress);
        }

        {
            (, address proxyValidatorPoolAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                mockValidatorPool
                    ? abi.encodePacked(type(ValidatorPoolMock).creationCode)
                    : abi.encodePacked(type(ValidatorPool).creationCode),
                "ValidatorPool",
                mockValidatorPool
                    ? abi.encodeWithSelector(ValidatorPoolMock.initialize.selector)
                    : abi.encodeWithSelector(
                        ValidatorPool.initialize.selector,
                        [20000 ether, 10, 3 ether, 8192]
                    )
            );
            fixture.validatorPool = IValidatorPool(proxyValidatorPoolAddress);
        }

        {
            deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(ETHDKGAccusations).creationCode),
                "ETHDKGAccusations",
                ""
            );
        }
        {
            (, address proxyStakingPositionDescriptorAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(StakingPositionDescriptor).creationCode),
                "StakingPositionDescriptor",
                ""
            );

            fixture.stakingPositionDescriptor = IStakingNFTDescriptor(
                proxyStakingPositionDescriptorAddress
            );
        }
        {
            deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(ETHDKGPhases).creationCode),
                "ETHDKGPhases",
                ""
            );
        }
        {
            (, address proxyETHDKGAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                mockETHDKG
                    ? abi.encodePacked(type(ETHDKGMock).creationCode)
                    : abi.encodePacked(type(ETHDKG).creationCode),
                "ETHDKG",
                mockETHDKG
                    ? abi.encodeWithSelector(ETHDKGMock.initialize.selector)
                    : abi.encodeWithSelector(
                        ETHDKG.initialize.selector,
                        abi.encodePacked(uint256(40), uint256(6))
                    )
            );

            fixture.ethdkg = IETHDKG(proxyETHDKGAddress);
        }
        {
            (, address proxySnapshotsAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                mockSnapshots
                    ? abi.encodePacked(
                        type(SnapshotsMock).creationCode,
                        abi.encodePacked(uint256(10), uint256(40))
                    )
                    : abi.encodePacked(
                        type(Snapshots).creationCode,
                        abi.encodePacked(uint256(1), uint256(1024))
                    ),
                "Snapshots",
                mockSnapshots
                    ? abi.encodeWithSelector(
                        SnapshotsMock.initialize.selector,
                        abi.encodePacked(uint256(10), uint256(40))
                    )
                    : abi.encodeWithSelector(
                        Snapshots.initialize.selector,
                        abi.encodePacked(uint256(10), uint256(40))
                    )
            );

            fixture.snapshots = ISnapshots(proxySnapshotsAddress);
        }

        {
            (, address deployALCAMinterAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(ALCAMinter).creationCode),
                "ALCAMinter",
                ""
            );
            ALCAMinter alcaMinter = ALCAMinter(deployALCAMinterAddress);

            vm.prank(fixture.adminAddress);
            baseTokensFixture.aliceNetFactory.callAny(
                address(alcaMinter),
                0,
                abi.encodeWithSelector(
                    alcaMinter.mint.selector,
                    abi.encodePacked(
                        address(baseTokensFixture.aliceNetFactory),
                        uint256(100000000 ether)
                    )
                )
            );

            fixture.alcaMinter = alcaMinter;
        }

        {
            (, address proxyALCABurnerAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(ALCABurner).creationCode),
                "ALCABurner",
                ""
            );

            fixture.alcaBurner = ALCABurner(proxyALCABurnerAddress);
        }

        {
            (, address proxyInvalidTxConsumptionAccusationAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(InvalidTxConsumptionAccusation).creationCode),
                keccak256(
                    abi.encodePacked(
                        keccak256("InvalidTxConsumptionAccusation"),
                        keccak256("Accusation")
                    )
                ),
                ""
            );

            fixture.invalidTxConsumptionAccusation = InvalidTxConsumptionAccusation(
                proxyInvalidTxConsumptionAccusationAddress
            );
        }

        {
            (, address proxyMultipleProposalAccusationAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(MultipleProposalAccusation).creationCode),
                keccak256(
                    abi.encodePacked(
                        keccak256("MultipleProposalAccusation"),
                        keccak256("Accusation")
                    )
                ),
                ""
            );

            fixture.multipleProposalAccusation = MultipleProposalAccusation(
                proxyMultipleProposalAccusationAddress
            );
        }

        {
            (, address proxyDistributionAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(
                    type(Distribution).creationCode,
                    abi.encodePacked(uint256(332), uint256(332), uint256(332), uint256(4))
                ),
                "Distribution",
                ""
            );

            fixture.distribution = Distribution(proxyDistributionAddress);
        }
        {
            (, address proxyDynamicsAddress) = deployUpgradeableWithFactory(
                vm,
                fixture.adminAddress,
                baseTokensFixture.aliceNetFactory,
                abi.encodePacked(type(Dynamics).creationCode),
                "Dynamics",
                abi.encodeWithSelector(
                    Dynamics.initialize.selector,
                    abi.encodePacked(uint256(4000))
                )
            );

            fixture.dynamics = Dynamics(proxyDynamicsAddress);
        }
    }

    function deployUpgradeableWithFactory(
        Vm vm,
        address adminAddress_,
        AliceNetFactory factory,
        bytes memory creationCode,
        bytes32 salt,
        bytes memory initCallData
    ) public returns (address deployAddress, address proxyAddress) {
        vm.prank(adminAddress_);
        deployAddress = factory.deployCreate(creationCode);

        vm.prank(adminAddress_);
        proxyAddress = factory.deployProxy(salt);

        vm.prank(adminAddress_);
        factory.upgradeProxy(salt, deployAddress, initCallData);
    }
}
