import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  Fixture,
  getContractAddressFromEventLog,
  getFixture,
} from "../../setup";

describe("Local ERC721 BridgePool Contract Factory", () => {
  let firstOwner: SignerWithAddress;
  let user: SignerWithAddress;
  let fixture: Fixture;

  beforeEach(async () => {
    fixture = await getFixture(true, true, false);
    [firstOwner, user] = await ethers.getSigners();
  });

  it.skip("Should not deploy new BridgePool with BridgePoolFactory not being delegator", async () => {
    const reason = ethers.utils.parseBytes32String(
      await fixture.aliceNetFactoryBaseErrorCodesContract.ALICENETFACTORYBASE_UNAUTHORIZED()
    );
    await expect(
      fixture.bridgePoolFactory.deployNewPool(fixture.erc721Mock.address, 1)
    ).to.be.revertedWith(reason);
  });

  it("Should not deploy two BridgePools with same ERC20 contract", async () => {
    const reason = ethers.utils.parseBytes32String(
      await fixture.bridgePoolFactoryErrorCodesContract.BRIDGEPOOLFACTORY_UNABLE_TO_DEPLOY_BRIDGEPOOL()
    );
    await fixture.factory.setDelegator(fixture.bridgePoolFactory.address);
    await fixture.bridgePoolFactory.deployNewPool(
      fixture.erc721Mock.address,
      1
    );
    await expect(
      fixture.bridgePoolFactory.deployNewPool(fixture.erc721Mock.address, 1)
    ).to.be.revertedWith(reason);
  });

  it("Should not deploy new BridgePool with inexistent version", async () => {
    const reason = ethers.utils.parseBytes32String(
      await fixture.bridgePoolFactoryErrorCodesContract.BRIDGEPOOLFACTORY_UNEXISTENT_BRIDGEPOOL_IMPLEMENTATION_VERSION()
    );
    await fixture.factory.setDelegator(fixture.bridgePoolFactory.address);
    await expect(
      fixture.bridgePoolFactory.deployNewPool(fixture.erc721Mock.address, 11)
    ).to.be.revertedWith(reason);
  });

  it("Should deploy new BridgePool with BridgePoolFactory being delegator", async () => {
    await fixture.factory.setDelegator(fixture.bridgePoolFactory.address);
    const deployNewPoolTransaction =
      await fixture.bridgePoolFactory.deployNewPool(
        fixture.erc721Mock.address,
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
      await ethers.getContractFactory("LocalERC721BridgePoolV1")
    ).attach(bridgePoolAddress);
    await expect(
      bridgePool.deposit(1, firstOwner.address, 1, 1)
    ).to.be.revertedWith("ERC721: operator query for nonexistent token");
  });
});
