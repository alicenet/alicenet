import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { factoryCallAnyFixture, Fixture, getFixture } from "../setup";
const poolType = 1;
let fixture: Fixture;
const bridgePoolVersion = 1;
const unexistentBridgePoolVersion = 11;

describe("Testing BridgePool Factory", async () => {
  beforeEach(async () => {
    fixture = await getFixture(true, true, false);
  });

  it("Should deploy new BridgePool as factory even if public pool deployment is not enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewLocalPool",
      [poolType, ethers.constants.AddressZero, bridgePoolVersion]
    );
  });

  it("Should not deploy new BridgePool as user if public pool deployment is not enabled", async () => {
    await expect(
      fixture.bridgePoolFactory.deployNewLocalPool(
        poolType,
        ethers.constants.AddressZero,
        bridgePoolVersion
      )
    ).to.be.revertedWithCustomError(
      fixture.bridgePoolFactory,
      "PublicPoolDeploymentTemporallyDisabled"
    );
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
      [poolType, ethers.constants.AddressZero, bridgePoolVersion]
    );
  });

  it("Should deploy new BridgePool as user if public pool deployment is enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "togglePublicPoolDeployment",
      []
    );
    await fixture.bridgePoolFactory.deployNewLocalPool(
      poolType,
      ethers.constants.AddressZero,
      bridgePoolVersion
    );
  });

  it("Should not deploy two BridgePools with same ERC contract and version", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewLocalPool",
      [poolType, ethers.constants.AddressZero, bridgePoolVersion]
    );
    await expect(
      factoryCallAnyFixture(
        fixture,
        "bridgePoolFactory",
        "deployNewLocalPool",
        [poolType, ethers.constants.AddressZero, bridgePoolVersion]
      )
    ).to.be.revertedWithCustomError(
      fixture.bridgePoolFactory,
      "StaticPoolDeploymentFailed"
    );
  });

  it("Should not deploy new BridgePool with inexistent version", async () => {
    await expect(
      factoryCallAnyFixture(
        fixture,
        "bridgePoolFactory",
        "deployNewLocalPool",
        [poolType, ethers.constants.AddressZero, unexistentBridgePoolVersion]
      )
    )
      .to.be.revertedWithCustomError(
        fixture.bridgePoolFactory,
        "PoolVersionNotSupported"
      )
      .withArgs(unexistentBridgePoolVersion);
  });
});
