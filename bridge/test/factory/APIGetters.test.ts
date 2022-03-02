import { ethers } from "hardhat";
import {
  deployStatic,
  deployUpgradeable,
  upgradeProxy
} from "../../scripts/lib/MadnetFactory";
import { MadnetFactory, Utils } from "../../typechain-types";
import { assert, expect } from "../chai-setup";
import {
  END_POINT,
  MOCK,
  PROXY,
  UTILS
} from '../../scripts/lib/constants';
import {
  bytes32ArrayToStringArray,
  deployFactory,
  getAccounts,
  getMetamorphicAddress
} from "./Setup";
process.env.silencer = "true";

describe("Madnetfactory API test", async () => {
  let firstOwner: string;
  let firstDelegator: string;
  let accounts: Array<string> = [];
  let utilsContract: Utils;
  let factory: MadnetFactory;

  beforeEach(async () => {
    let utilsBase = await ethers.getContractFactory(UTILS);
    accounts = await getAccounts();
    //set owner and delegator
    firstOwner = accounts[0];
    firstDelegator = accounts[1];
    utilsContract = await utilsBase.deploy();
    factory = await deployFactory();
    let cSize = await utilsContract.getCodeSize(factory.address);
    expect(cSize.toNumber()).to.be.greaterThan(0);
  });

  it("getters", async () => {
    await deployStatic(END_POINT, factory.address);
    let implAddr = await factory.implementation();
    expect(implAddr).to.not.be.undefined;
    let saltsArray = await factory.contracts();
    expect(saltsArray.length).to.be.greaterThan(0);
    let numContracts = await factory.getNumContracts();
    expect(numContracts.toNumber()).to.equal(saltsArray.length);
    let saltStrings = bytes32ArrayToStringArray(saltsArray);
    for (let i = 0; i < saltStrings.length; i++) {
      let address = await factory.lookup(saltStrings[i]);
      expect(address).to.equal(
        getMetamorphicAddress(factory.address, saltsArray[i])
      );
    }
  });

  it("deploy Upgradeable", async () => {
    let res = await deployUpgradeable(MOCK, factory.address, ["2", "s"]);
    const Proxy = await ethers.getContractFactory(PROXY)
    const proxy = Proxy.attach(res.proxyAddress)
    expect(await proxy.getImplementationAddress()).to.be.equal(res.logicAddress)
    assert(res !== undefined, "Couldn't deploy upgradable contract")
    let cSize = await utilsContract.getCodeSize(res.logicAddress);
    expect(cSize.toNumber()).to.be.greaterThan(0);
    cSize = await utilsContract.getCodeSize(res.proxyAddress);
    expect(cSize.toNumber()).to.be.greaterThan(0);
  });

  it("upgrade deployment", async () => {
    let res = await deployUpgradeable(MOCK, factory.address, ["2", "s"]);
    const Proxy = await ethers.getContractFactory(PROXY)
    const proxy = Proxy.attach(res.proxyAddress)
    expect(await proxy.getImplementationAddress()).to.be.equal(res.logicAddress)
    assert(res !== undefined, "Couldn't deploy upgradable contract")
    let res2 = await upgradeProxy(MOCK, factory.address, ["2", "s"]);
    expect(await proxy.getImplementationAddress()).to.be.equal(res2.logicAddress)
    assert(res2.logicAddress != res.logicAddress, "Logic address should be different after updateProxy!")
  });

  it("deploystatic", async () => {
    let res = await deployStatic(END_POINT, factory.address);
    let cSize = await utilsContract.getCodeSize(res.templateAddress);
    expect(cSize.toNumber()).to.be.greaterThan(0);
    cSize = await utilsContract.getCodeSize(res.metaAddress);
    expect(cSize.toNumber()).to.be.greaterThan(0);
  });
});
