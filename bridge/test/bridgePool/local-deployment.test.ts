import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  factoryCallAny,
  factoryCallAnyFixture,
  Fixture,
  getFixture,
} from "../setup";
const bridgePoolTokenType = 0;
const bridgePoolChainId = 1337;
const bridgePoolValue = 0;
const bridgePoolDeployCode = "0x38585839386009f3"; // UNIVERSAL_DEPLOY_CODE

let fixture: Fixture;
const bridgePoolVersion = 1;
const unexistentBridgePoolVersion = 11;

describe("Testing BridgePool Factory", async () => {
  beforeEach(async () => {
    fixture = await getFixture(true, true, false);
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        bridgePoolTokenType,
        bridgePoolChainId,
        bridgePoolVersion,
        bridgePoolValue,
        bridgePoolDeployCode,
      ]
    );
  });

  it("Should deploy new BridgePool as factory even if public pool deployment is not enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewNativePool",
      [bridgePoolTokenType, ethers.constants.AddressZero, bridgePoolVersion]
    );
  });

  it("Should not deploy new BridgePool as user if public pool deployment is not enabled", async () => {
    await expect(
      fixture.bridgePoolFactory.deployNewNativePool(
        bridgePoolTokenType,
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
      "deployNewNativePool",
      [bridgePoolTokenType, ethers.constants.AddressZero, bridgePoolVersion]
    );
  });

  it("Should deploy new BridgePool as user if public pool deployment is enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "togglePublicPoolDeployment",
      []
    );
    await fixture.bridgePoolFactory.deployNewNativePool(
      bridgePoolTokenType,
      ethers.constants.AddressZero,
      bridgePoolVersion
    );
  });

  it("Should not deploy two BridgePools with same ERC contract and version", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewNativePool",
      [bridgePoolTokenType, ethers.constants.AddressZero, bridgePoolVersion]
    );
    await expect(
      factoryCallAnyFixture(
        fixture,
        "bridgePoolFactory",
        "deployNewNativePool",
        [bridgePoolTokenType, ethers.constants.AddressZero, bridgePoolVersion]
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
        "deployNewNativePool",
        [
          bridgePoolTokenType,
          ethers.constants.AddressZero,
          unexistentBridgePoolVersion,
        ]
      )
    )
      .to.be.revertedWithCustomError(
        fixture.bridgePoolFactory,
        "PoolVersionNotSupported"
      )
      .withArgs(unexistentBridgePoolVersion);
  });
});
