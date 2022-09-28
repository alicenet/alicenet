import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  callFunctionAndGetReturnValues,
  factoryCallAny,
  factoryCallAnyFixture,
  Fixture,
  getBridgePoolSalt,
  getFixture,
} from "../setup";
const bridgePoolTokenType = 0;
const bridgePoolChainId = 1337;
const bridgePoolValue = 0;
const bridgePoolDeployCode = "0x38585839386009f3"; // UNIVERSAL_DEPLOY_CODE

let fixture: Fixture;
let admin: SignerWithAddress;
const bridgePoolVersion = 1;
const unexistentBridgePoolVersion = 11;

describe("Testing BridgePool Factory", async () => {
  async function deployFixture() {
    const fixture = await getFixture(true, true, false);
    const [admin] = await ethers.getSigners();
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        bridgePoolTokenType,
        bridgePoolChainId,
        bridgePoolValue,
        bridgePoolDeployCode,
      ]
    );
    return { fixture, admin };
  }

  beforeEach(async function () {
    ({ fixture, admin } = await loadFixture(deployFixture));
  });

  it("Should fail to enable public pool deploying without impersonating factory", async () => {
    await expect(
      fixture.bridgePoolFactory.togglePublicPoolDeployment()
    ).to.be.revertedWithCustomError(fixture.bridgePoolFactory, "OnlyFactory");
  });

  it("Should fail to deploy pool logic without impersonating factory", async () => {
    await expect(
      fixture.bridgePoolFactory.deployPoolLogic(
        bridgePoolTokenType,
        bridgePoolChainId,
        bridgePoolValue,
        bridgePoolDeployCode
      )
    ).to.be.revertedWithCustomError(fixture.bridgePoolFactory, "OnlyFactory");
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

  it("Should get correct Bridge Pool Salt", async () => {
    const [bridgePoolSalt] = await callFunctionAndGetReturnValues(
      fixture.bridgePoolFactory,
      "getBridgePoolSalt",
      admin,
      [
        ethers.constants.AddressZero,
        bridgePoolTokenType,
        bridgePoolChainId,
        bridgePoolVersion,
      ]
    );
    expect(bridgePoolSalt).to.be.eq(
      getBridgePoolSalt(
        ethers.constants.AddressZero,
        bridgePoolTokenType,
        bridgePoolChainId,
        bridgePoolVersion
      )
    );
  });
});
