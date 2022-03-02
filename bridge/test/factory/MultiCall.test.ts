import { expect } from "chai";
import { ethers } from "hardhat";
import { MockBaseContract__factory, Utils__factory } from "../../typechain-types";
import {
  CONTRACT_ADDR, DEPLOYED_PROXY,
  DEPLOYED_RAW,
  DEPLOYED_STATIC,
  DEPLOYED_TEMPLATE, DEPLOY_STATIC, DEPLOY_TEMPLATE, END_POINT, MADNET_FACTORY, MOCK, UTILS
} from '../../scripts/lib/constants';
import {
  deployFactory,
  getAccounts,
  getCreateAddress,
  getEventVar,
  getSalt,
  metaMockLogicTest,
  proxyMockLogicTest
} from "./Setup";
import { ContractFactory } from "ethers";


describe("Multicall deploy proxy", () => {
  let firstOwner: string;
  let secondOwner: string;
  let firstDelegator: string;
  let accounts: Array<string> = [];

  beforeEach(async () => {
    accounts = await getAccounts();
    firstOwner = accounts[0];
    secondOwner = accounts[1];
    firstDelegator = accounts[2];
  });

  it("multicall deploycreate, deployproxy, upgradeproxy", async () => {
    let UtilsBase = await ethers.getContractFactory(UTILS)
    let factory = await deployFactory();
    let mockFactory = await ethers.getContractFactory(MOCK);
    let endPointFactory = await ethers.getContractFactory(END_POINT);
    let Salt = getSalt();
    let mockCon = await ethers.getContractFactory(MOCK);
    let endPoint = await endPointFactory.deploy(factory.address);
    //deploy code for mock with constructor args i = 2
    let deployTX = mockCon.getDeployTransaction(2, "s");
    const factoryBase = await ethers.getContractFactory(MADNET_FACTORY);
    let transactionCount = await ethers.provider.getTransactionCount(
      factory.address
    );
    //calculate the deployCreate Address
    let expectedMockLogicAddr = getCreateAddress(
      factory.address,
      transactionCount
    );
    //encoded function call to deployCreate
    let deployCreate = factoryBase.interface.encodeFunctionData(
      "deployCreate",
      [deployTX.data]
    );
    //encoded function call to deployProxy
    let deployProxy = factoryBase.interface.encodeFunctionData(
      "deployProxy",
      [Salt]
    );
    //encoded function call to upgradeProxy
    let upgradeProxy = factoryBase.interface.encodeFunctionData(
      "upgradeProxy",
      [Salt, expectedMockLogicAddr, "0x"]
    );
    let txResponse = await factory.multiCall([
      deployCreate,
      deployProxy,
      upgradeProxy,
    ]);
    let mockLogicAddr = await getEventVar(
      txResponse,
      DEPLOYED_RAW,
      CONTRACT_ADDR
    );
    let proxyAddr = await getEventVar(txResponse, DEPLOYED_PROXY, CONTRACT_ADDR);
    expect(mockLogicAddr).to.equal(expectedMockLogicAddr);
    // console.log("MULTICALL DEPLOYPROXY, DEPLOYCREATE, UPGRADEPROXY GASUSED: ", receipt["receipt"]["gasUsed"]);
    //check the proxy behavior
    await proxyMockLogicTest(
      mockFactory,
      Salt,
      proxyAddr,
      mockLogicAddr,
      endPoint.address,
      factory.address
    );
  });

  it("multicall deploytemplate, deploystatic", async () => {
    let UtilsBase = await ethers.getContractFactory(UTILS)
    let factory = await deployFactory();
    let utilsBase: ContractFactory = await ethers.getContractFactory(UTILS);
    let mockFactory: ContractFactory = await ethers.getContractFactory(MOCK);
    let utilsContract = await utilsBase.deploy();
    let Salt = getSalt();
    //ethers instance of Mock contract abstraction
    let deployTX = mockFactory.getDeployTransaction(2, "s");
    const MadnetFactory = await ethers.getContractFactory(MADNET_FACTORY);
    //encoded function call to deployTemplate
    let deployTemplate = MadnetFactory.interface.encodeFunctionData(
      DEPLOY_TEMPLATE,
      [deployTX.data]
    );
    //encoded function call to deployStatic
    let deployStatic = MadnetFactory.interface.encodeFunctionData(
      DEPLOY_STATIC,
      [Salt, "0x"]
    );
    let txResponse = await factory.multiCall([deployTemplate, deployStatic]);
    //get the deployed template contract address from the event
    let tempSDAddr = await getEventVar(
      txResponse,
      DEPLOYED_TEMPLATE,
      CONTRACT_ADDR
    );
    //get the deployed metamorphic contract address from the event
    let metaAddr = await getEventVar(txResponse, DEPLOYED_STATIC, CONTRACT_ADDR);
    let tempCSize = await utilsContract.getCodeSize(tempSDAddr);
    let staticCSize = await utilsContract.getCodeSize(metaAddr);
    expect(tempCSize.toNumber()).to.be.greaterThan(0);
    expect(staticCSize.toNumber()).to.be.greaterThan(0);
    //test logic at deployed metamorphic location
    await metaMockLogicTest(mockFactory, metaAddr, factory.address);
    // console.log("MULTICALL DEPLOYTEMPLATE, DEPLOYSTATIC GASUSED: ", receipt["receipt"]["gasUsed"]);
  });
});

