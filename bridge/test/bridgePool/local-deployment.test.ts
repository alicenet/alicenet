import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";
import { maxEth, maxTokens, tokenTypes } from "./setup";
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

      it("Should not deploy two BridgePools with same ERC721 contract and version", async () => {
        const reason = ethers.utils.parseBytes32String(
          await fixture.bridgeRouterErrorCodesContract.BRIDGEROUTER_UNABLE_TO_DEPLOY_BRIDGEPOOL()
        );
        await fixture.bridgeRouter.deployNewLocalPool(
          run.options.poolType,
          fixture[run.options.ercContractName].address,
          1
        );
        await expect(
          fixture.bridgeRouter.deployNewLocalPool(
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            1
          )
        ).to.be.revertedWith(reason);
      });

      it("Should not deploy new BridgePool with inexistent version", async () => {
        const reason = ethers.utils.parseBytes32String(
          await fixture.bridgeRouterErrorCodesContract.BRIDGEROUTER_UNEXISTENT_BRIDGEPOOL_IMPLEMENTATION_VERSION()
        );
        await fixture.factory.setDelegator(fixture.bridgeRouter.address);
        await expect(
          fixture.bridgeRouter.deployNewLocalPool(
            run.options.poolType,
            fixture[run.options.ercContractName].address,
            11
          )
        ).to.be.revertedWith(reason);
      });

      it("Should deploy new BridgePool with correct parameters", async () => {
        await fixture.bridgeRouter.deployNewLocalPool(
          run.options.poolType,
          fixture[run.options.ercContractName].address,
          1
        );
        await expect(
          fixture.bToken
            .connect(user)
            .payAndDeposit(maxEth, maxTokens, encodedDepositCallData, {
              value: valueSent,
            })
        ).to.be.revertedWith(run.options.errorReason);
      });
    }
  );
});
