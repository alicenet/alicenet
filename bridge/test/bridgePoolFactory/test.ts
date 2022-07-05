import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import {
  AliceNetFactoryBaseErrorCodes,
  ImmutableAuthErrorCodes,
} from "../../typechain-types";
import { expect } from "../chai-setup";
import {
  Fixture,
  getContractAddressFromDeployedProxyEvent,
  getFixture,
} from "../setup";

describe("BridgePool Contract Factory", () => {
  let firstOwner: SignerWithAddress;
  let fixture: Fixture;
  let immutableAuthErrorCodesContract: ImmutableAuthErrorCodes;
  let aliceNetFactoryBaseErrorCodes: AliceNetFactoryBaseErrorCodes;

  beforeEach(async () => {
    fixture = await getFixture(false, false, false);
    [firstOwner] = await ethers.getSigners();
    const AliceNetFactoryBaseErrorCodesFactory =
      await ethers.getContractFactory("AliceNetFactoryBaseErrorCodes");
    aliceNetFactoryBaseErrorCodes =
      await AliceNetFactoryBaseErrorCodesFactory.deploy();
    await aliceNetFactoryBaseErrorCodes.deployed();
    const ImmutableAuthErrorCodesContract = await ethers.getContractFactory(
      "ImmutableAuthErrorCodes"
    );
    immutableAuthErrorCodesContract =
      await ImmutableAuthErrorCodesContract.deploy();
    await immutableAuthErrorCodesContract.deployed();
  });

  describe("Testing Access control", () => {
    it("should not deploy new BridgePool with BridgePoolFactory not being delegator", async () => {
      const reason = ethers.utils.parseBytes32String(
        await aliceNetFactoryBaseErrorCodes.ALICENETFACTORYBASE_UNAUTHORIZED()
      );
      await expect(
        fixture.bridgePoolFactory.deployNewPool(
          fixture.aToken.address,
          fixture.bToken.address
        )
      ).to.be.revertedWith(reason);
    });

    it("should not deploy new BridgePool with BridgePoolFactory trying to access factory logic directly", async () => {
      const reason = ethers.utils.parseBytes32String(
        await immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_BRIDGEPOOLFACTORY()
      );
      await expect(
        fixture.bridgePoolFactory.deployViaFactoryLogic(
          fixture.aToken.address,
          fixture.bToken.address
        )
      ).to.be.revertedWith(reason);
    });

    it("should not deploy two BridgePools with same ERC20 contract", async () => {
      const reason = ethers.utils.parseBytes32String(
        await aliceNetFactoryBaseErrorCodes.ALICENETFACTORYBASE_CODE_SIZE_ZERO()
      );
      await fixture.factory.setDelegator(fixture.bridgePoolFactory.address);
      await fixture.bridgePoolFactory.deployNewPool(
        fixture.aToken.address,
        fixture.bToken.address
      );
      await expect(
        fixture.bridgePoolFactory.deployNewPool(
          fixture.aToken.address,
          fixture.bToken.address
        )
      ).to.be.revertedWith(reason);
    });

    it("should deploy new BridgePool with BridgePoolFactory being delegator", async () => {
      await fixture.factory.setDelegator(fixture.bridgePoolFactory.address);
      const deployNewPoolTransaction =
        await fixture.bridgePoolFactory.deployNewPool(
          fixture.aToken.address,
          fixture.bToken.address
        );
      const bridgePoolAddress = await getContractAddressFromDeployedProxyEvent(
        deployNewPoolTransaction
      );
      const bridgePool = (await ethers.getContractFactory("BridgePool")).attach(
        bridgePoolAddress
      );
      await expect(
        bridgePool.deposit(1, firstOwner.address, 1, 1)
      ).to.be.revertedWith("ERC20: insufficient allowance");
    });
  });
});
