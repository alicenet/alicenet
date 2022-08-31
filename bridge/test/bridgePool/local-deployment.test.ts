import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { factoryCallAnyFixture, Fixture, getFixture } from "../setup";
import { tokenTypes } from "./setup";
let user: SignerWithAddress;
let fixture: Fixture;
let depositCallData: any;
let encodedDepositCallData: string;
const valueSent = ethers.utils.parseEther("1.0");
const valueOrId = 100; // value if ERC20 , tokenId otherwise
const bridgePoolVersion = 1;
const unexistentBridgePoolVersion = 11;


tokenTypes.forEach(function (run) {
  describe(
    "Testing BridgePool deployment for tokenType " + run.it,
    async () => {
      beforeEach(async () => {
        fixture = await getFixture(true, true, false);
        [, user] = await ethers.getSigners();
        depositCallData = {
          ERCContract: fixture[run.options.ercContractName].address,
          destinationAccountType: 1, // 1 for secp256k1, 2 for bn128
          destinationAccount: ethers.constants.AddressZero,
          tokenType: run.options.poolType,
          number: valueOrId,
          chainID: 1337,
          poolVersion: 1,
        };
        encodedDepositCallData = ethers.utils.defaultAbiCoder.encode(
          [
            "tuple(address ERCContract, uint8 destinationAccountType, address destinationAccount, uint8 tokenType, uint256 number, uint256 chainID, uint16 poolVersion)",
          ],
          [depositCallData]
        );
      });

      it("Should deploy new BridgePool as factory if public pool deployment is not enabled", async () => {
        await factoryCallAnyFixture(
          fixture,
          "bridgePoolFactory",
          "deployNewLocalPool",
          [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            bridgePoolVersion,
          ]
        );
        await
          fixture.bToken
            .connect(user)
            .depositTokensOnBridges(bridgePoolVersion, encodedDepositCallData, {
              value: valueSent
            })
      });

      it("Should not deploy new BridgePool as user if public pool deployment is not enabled", async () => {
        await expect(
          fixture.bridgePoolFactory.deployNewLocalPool(
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            1
          )).to.be.revertedWithCustomError(fixture.bridgePoolFactory,'PublicPoolDeploymentTemporallyDisabled')
      });

      it("Should deploy new BridgePool as factory if public pool deployment is enabled", async () => {
        await factoryCallAnyFixture(
          fixture,
          "bridgePoolFactory",
          "togglePublicPoolDeployment",
          []
        );
        await factoryCallAnyFixture(
          fixture,
          "bridgePoolFactory",
          "deployNewLocalPool",
          [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            bridgePoolVersion,
          ]
        );
        await
          fixture.bToken
            .connect(user)
            .depositTokensOnBridges(bridgePoolVersion, encodedDepositCallData, {
              value: valueSent
            })
      });

      it("Should not deploy new BridgePool as user if public pool deployment is enabled", async () => {
        await factoryCallAnyFixture(
          fixture,
          "bridgePoolFactory",
          "togglePublicPoolDeployment",
          []
        );
        await fixture.bridgePoolFactory.deployNewLocalPool(
          run.options.poolType,
          fixture[run.options.ercContractName].address,
          1
        );
        await
          fixture.bToken
            .connect(user)
            .depositTokensOnBridges(bridgePoolVersion, encodedDepositCallData, {
              value: valueSent
            })
      });


      it("Should not deploy two BridgePools with same ERC contract and version", async () => {
        await factoryCallAnyFixture(
          fixture,
          "bridgePoolFactory",
          "deployNewLocalPool",
          [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            bridgePoolVersion,
          ]
        );
        await expect(
          factoryCallAnyFixture(fixture, "bridgePoolFactory", "deployNewLocalPool", [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            bridgePoolVersion,
          ])
        ).to.be.revertedWithCustomError(fixture.bridgePoolFactory,"StaticPoolDeploymentFailed");
      });

      it("Should not deploy new BridgePool with inexistent version", async () => {
        await expect(
          factoryCallAnyFixture(fixture, "bridgePoolFactory", "deployNewLocalPool", [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            unexistentBridgePoolVersion,
          ])
          ).to.be.revertedWithCustomError(fixture.bridgePoolFactory,"PoolVersionNotSupported").withArgs(unexistentBridgePoolVersion);;
        });
    }
  );
});
