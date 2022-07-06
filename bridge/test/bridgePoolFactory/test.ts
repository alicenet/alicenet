import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Fixture, getContractAddressFromEventLog, getFixture } from "../setup";

describe("BridgePool Contract Factory", () => {
  let firstOwner: SignerWithAddress;
  let fixture: Fixture;

  beforeEach(async () => {
    fixture = await getFixture(true, true, false);
    [firstOwner] = await ethers.getSigners();
  });

  describe("Testing Access control", () => {
    it("should not deploy new BridgePool with BridgePoolFactory not being delegator", async () => {
      const reason = ethers.utils.parseBytes32String(
        await fixture.aliceNetFactoryBaseErrorCodes.ALICENETFACTORYBASE_UNAUTHORIZED()
      );
      await expect(
        fixture.bridgePoolFactory.deployNewPool(
          fixture.aToken.address,
          fixture.bToken.address
        )
      ).to.be.revertedWith(reason);
    });

    it("should not deploy two BridgePools with same ERC20 contract", async () => {
      const reason = ethers.utils.parseBytes32String(
        await fixture.aliceNetFactoryBaseErrorCodes.ALICENETFACTORYBASE_CODE_SIZE_ZERO()
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
          fixture.bToken.address,
          1
        );
      const eventSignature = "event BridgePoolCreated(address contractAddr)";
      const eventName = "BridgePoolCreated";
      const bridgePoolAddress = await getContractAddressFromEventLog(
        deployNewPoolTransaction,
        eventSignature,
        eventName
      );
      // Final bridgePool address
      const bridgePool = (
        await ethers.getContractFactory("BridgePoolV1")
      ).attach(bridgePoolAddress);
      await expect(
        bridgePool.deposit(1, firstOwner.address, 1, 1)
      ).to.be.revertedWith("ERC20: insufficient allowance");
    });
  });
});
