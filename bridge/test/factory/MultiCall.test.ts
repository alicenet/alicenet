import { expect } from "chai";
import { ContractFactory } from "ethers";
import { ethers } from "hardhat";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  DEPLOYED_PROXY,
  DEPLOYED_RAW,
  DEPLOYED_STATIC,
  DEPLOYED_TEMPLATE,
  DEPLOY_STATIC,
  DEPLOY_TEMPLATE,
  END_POINT,
  MOCK,
  UTILS,
} from "../../scripts/lib/constants";
import {
  deployFactory,
  getCreateAddress,
  getEventVar,
  getSalt,
  metaMockLogicTest,
  proxyMockLogicTest,
} from "./Setup";

describe("Multicall deploy proxy", () => {
  it("multicall deploycreate, deployproxy, upgradeproxy", async () => {
    const factory = await deployFactory();
    const mockFactory = await ethers.getContractFactory(MOCK);
    const endPointFactory = await ethers.getContractFactory(END_POINT);
    const Salt = getSalt();
    const mockCon = await ethers.getContractFactory(MOCK);
    const endPoint = await endPointFactory.deploy(factory.address);
    // deploy code for mock with constructor args i = 2
    const deployTX = mockCon.getDeployTransaction(2, "s");
    const factoryBase = await ethers.getContractFactory(ALICENET_FACTORY);
    const transactionCount = await ethers.provider.getTransactionCount(
      factory.address
    );
    // calculate the deployCreate Address
    const expectedMockLogicAddr = getCreateAddress(
      factory.address,
      transactionCount
    );
    // encoded function call to deployCreate
    const deployCreate = factoryBase.interface.encodeFunctionData(
      "deployCreate",
      [deployTX.data]
    );
    // encoded function call to deployProxy
    const deployProxy = factoryBase.interface.encodeFunctionData(
      "deployProxy",
      [Salt]
    );
    // encoded function call to upgradeProxy
    const upgradeProxy = factoryBase.interface.encodeFunctionData(
      "upgradeProxy",
      [Salt, expectedMockLogicAddr, "0x"]
    );
    const txResponse = await factory.multiCall([
      deployCreate,
      deployProxy,
      upgradeProxy,
    ]);
    const mockLogicAddr = await getEventVar(
      txResponse,
      DEPLOYED_RAW,
      CONTRACT_ADDR
    );
    const proxyAddr = await getEventVar(
      txResponse,
      DEPLOYED_PROXY,
      CONTRACT_ADDR
    );
    expect(mockLogicAddr).to.equal(expectedMockLogicAddr);
    // console.log("MULTICALL DEPLOYPROXY, DEPLOYCREATE, UPGRADEPROXY GASUSED: ", receipt["receipt"]["gasUsed"]);
    // check the proxy behavior
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
    const factory = await deployFactory();
    const utilsBase: ContractFactory = await ethers.getContractFactory(UTILS);
    const mockFactory: ContractFactory = await ethers.getContractFactory(MOCK);
    const utilsContract = await utilsBase.deploy();
    const Salt = getSalt();
    // ethers instance of Mock contract abstraction
    const deployTX = mockFactory.getDeployTransaction(2, "s");
    const AliceNetFactory = await ethers.getContractFactory(ALICENET_FACTORY);
    // encoded function call to deployTemplate
    const deployTemplate = AliceNetFactory.interface.encodeFunctionData(
      DEPLOY_TEMPLATE,
      [deployTX.data]
    );
    // encoded function call to deployStatic
    const deployStatic = AliceNetFactory.interface.encodeFunctionData(
      DEPLOY_STATIC,
      [Salt, "0x"]
    );
    const txResponse = await factory.multiCall([deployTemplate, deployStatic]);
    // get the deployed template contract address from the event
    const tempSDAddr = await getEventVar(
      txResponse,
      DEPLOYED_TEMPLATE,
      CONTRACT_ADDR
    );
    // get the deployed metamorphic contract address from the event
    const metaAddr = await getEventVar(
      txResponse,
      DEPLOYED_STATIC,
      CONTRACT_ADDR
    );
    const tempCSize = await utilsContract.getCodeSize(tempSDAddr);
    const staticCSize = await utilsContract.getCodeSize(metaAddr);
    expect(tempCSize.toNumber()).to.be.greaterThan(0);
    expect(staticCSize.toNumber()).to.be.greaterThan(0);
    // test logic at deployed metamorphic location
    await metaMockLogicTest(mockFactory, metaAddr, factory.address);
    // console.log("MULTICALL DEPLOYTEMPLATE, DEPLOYSTATIC GASUSED: ", receipt["receipt"]["gasUsed"]);
  });
});
