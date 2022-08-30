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

tokenTypes.forEach(function (run) {
  describe(
    "Testing BridgePool deployment for tokenType " + run.it,
    async () => {
      beforeEach(async () => {
        fixture = await getFixture(true, true, false);
        [, user] = await ethers.getSigners();
        depositCallData = {
          ERCContract: fixture[run.options.ercContractName].address,
          tokenType: run.options.poolType,
          number: valueOrId,
          chainID: 1337,
          poolVersion: 1,
        };
        encodedDepositCallData = ethers.utils.defaultAbiCoder.encode(
          [
            "tuple(address ERCContract, uint8 tokenType, uint256 number, uint256 chainID, uint16 poolVersion)",
          ],
          [depositCallData]
        );
      });

      it.only("Should deploy new BridgePool as factory if public pool deployment is not enabled", async () => {
        await factoryCallAnyFixture(
          fixture,
          "bridgePoolFactory",
          "deployNewLocalPool",
          [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            1,
          ]
        );
        /*         await expect(
                  fixture.bToken
                    .connect(user)
                    .payAndDeposit(maxEth, maxTokens, encodedDepositCallData, {
                      value: valueSent,
                    })
                ).to.be.revertedWith(run.options.errorReason); */
      });

      it("Should not deploy new BridgePool as user if public pool deployment is not enabled", async () => {
        const reason = ethers.utils.parseBytes32String(
          await fixture.bridgeRouterErrorCodesContract.BRIDGEROUTER_POOL_DEPLOYMENT_TEMPORALLY_DISABLED()
        );
        await expect(
          fixture.bridgeRouter.deployNewLocalPool(
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            1
          )
        ).to.be.revertedWith(reason);
      });

      it("Should deploy new BridgePool as user if public pool deployment is enabled", async () => {
        await factoryCallAnyFixture(
          fixture,
          "bridgeRouter",
          "togglePublicPoolDeployment",
          []
        );
        await fixture.bridgeRouter.deployNewLocalPool(
          run.options.poolType,
          fixture[run.options.ercContractName].address,
          1
        );
        /*         await expect(
                  fixture.bToken
                    .connect(user)
                    .payAndDeposit(maxEth, maxTokens, encodedDepositCallData, {
                      value: valueSent,
                    })
                ).to.be.revertedWith(run.options.errorReason); */
      });

      it("Should not deploy two BridgePools with same ERC contract and version", async () => {
        const reason = ethers.utils.parseBytes32String(
          await fixture.bridgeRouterErrorCodesContract.BRIDGEROUTER_UNABLE_TO_DEPLOY_BRIDGEPOOL()
        );
        await factoryCallAnyFixture(
          fixture,
          "bridgeRouter",
          "deployNewLocalPool",
          [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            1,
          ]
        );
        await expect(
          factoryCallAnyFixture(fixture, "bridgeRouter", "deployNewLocalPool", [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            1,
          ])
        ).to.be.revertedWith(reason);
      });

      it("Should not deploy new BridgePool with inexistent version", async () => {
        const reason = ethers.utils.parseBytes32String(
          await fixture.bridgeRouterErrorCodesContract.BRIDGEROUTER_UNEXISTENT_BRIDGEPOOL_IMPLEMENTATION_VERSION()
        );
        await expect(
          factoryCallAnyFixture(fixture, "bridgeRouter", "deployNewLocalPool", [
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            11,
          ])
        ).to.be.revertedWith(reason);
      });
    }
  );
});
