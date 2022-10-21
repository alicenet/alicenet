import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import {
  deployUpgradeable,
  upgradeProxy,
} from "../../scripts/lib/alicenetFactory";
import { MOCK, PROXY, UTILS } from "../../scripts/lib/constants";
import { AliceNetFactory, Utils } from "../../typechain-types";
import { assert, expect } from "../chai-setup";
import { deployFactory } from "./Setup";
process.env.silencer = "true";

describe("AliceNetfactory API test", async () => {
  let utilsContract: Utils;
  let factory: AliceNetFactory;

  async function deployFixture() {
    const utilsBase = await ethers.getContractFactory(UTILS);
    const utilsContract = await utilsBase.deploy();
    const factory = await deployFactory();
    return { utilsContract, factory };
  }

  beforeEach(async () => {
    ({ utilsContract, factory } = await loadFixture(deployFixture));

    const cSize = await utilsContract.getCodeSize(factory.address);
    expect(cSize.toNumber()).to.be.greaterThan(0);
  });

  it("deploy Upgradeable", async () => {
    const res = await deployUpgradeable(MOCK, factory.address, ["2", "s"]);
    const Proxy = await ethers.getContractFactory(PROXY);
    const proxy = Proxy.attach(res.proxyAddress);
    expect(await proxy.getImplementationAddress()).to.be.equal(
      res.logicAddress
    );
    assert(res !== undefined, "Couldn't deploy upgradable contract");
    let cSize = await utilsContract.getCodeSize(res.logicAddress);
    expect(cSize.toNumber()).to.be.greaterThan(0);
    cSize = await utilsContract.getCodeSize(res.proxyAddress);
    expect(cSize.toNumber()).to.be.greaterThan(0);
  });

  it("upgrade deployment", async () => {
    const res = await deployUpgradeable(MOCK, factory.address, ["2", "s"]);
    const proxy = await ethers.getContractAt(PROXY, res.proxyAddress);
    expect(await proxy.getImplementationAddress()).to.be.equal(
      res.logicAddress
    );
    assert(res !== undefined, "Couldn't deploy upgradable contract");
    const res2 = await upgradeProxy(MOCK, factory.address, ["2", "s"]);
    expect(await proxy.getImplementationAddress()).to.be.equal(
      res2.logicAddress
    );
    assert(
      res2.logicAddress !== res.logicAddress,
      "Logic address should be different after updateProxy!"
    );
  });
});
