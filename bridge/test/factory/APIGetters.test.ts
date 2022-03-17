import { ethers } from "hardhat";
import { END_POINT, MOCK, PROXY, UTILS } from "../../scripts/lib/constants";
import {
  deployStatic,
  deployUpgradeable,
  upgradeProxy,
} from "../../scripts/lib/MadnetFactory";
import { MadnetFactory, Utils } from "../../typechain-types";
import { assert, expect } from "../chai-setup";
import {
  bytes32ArrayToStringArray,
  deployFactory,
  getAccounts,
  getMetamorphicAddress,
} from "./Setup";
process.env.silencer = "true";

describe("Madnetfactory API test", async () => {
  let accounts: Array<string> = [];
  let utilsContract: Utils;
  let factory: MadnetFactory;

  beforeEach(async () => {
    const utilsBase = await ethers.getContractFactory(UTILS);
    accounts = await getAccounts();
    // set owner and delegator
    utilsContract = await utilsBase.deploy();
    factory = await deployFactory();
    const cSize = await utilsContract.getCodeSize(factory.address);
    expect(cSize.toNumber()).to.be.greaterThan(0);
  });

  it("getters", async () => {
    await deployStatic(END_POINT, factory.address);
    const implAddr = await factory.getImplementation();
    expect(implAddr).to.not.equal(undefined);
    const saltsArray = await factory.contracts();
    expect(saltsArray.length).to.be.greaterThan(0);
    const numContracts = await factory.getNumContracts();
    expect(numContracts.toNumber()).to.equal(saltsArray.length);
    const saltStrings = bytes32ArrayToStringArray(saltsArray);
    for (let i = 0; i < saltStrings.length; i++) {
      const address = await factory.lookup(
        ethers.utils.formatBytes32String(saltStrings[i])
      );
      expect(address).to.equal(
        getMetamorphicAddress(factory.address, saltsArray[i])
      );
    }
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
    const Proxy = await ethers.getContractFactory(PROXY);
    const proxy = Proxy.attach(res.proxyAddress);
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

  it("deploystatic", async () => {
    const res = await deployStatic(END_POINT, factory.address);
    let cSize = await utilsContract.getCodeSize(res.templateAddress);
    expect(cSize.toNumber()).to.be.greaterThan(0);
    cSize = await utilsContract.getCodeSize(res.metaAddress);
    expect(cSize.toNumber()).to.be.greaterThan(0);
  });
});
