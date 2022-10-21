import { expect } from "chai";
import { ethers } from "hardhat";
import { encodeMultiCallArgs } from "../../scripts/lib/alicenetTasks";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  DEPLOYED_PROXY,
  DEPLOYED_RAW,
  END_POINT,
  MOCK,
} from "../../scripts/lib/constants";
import {
  deployFactory,
  getCreateAddress,
  getEventVar,
  getSalt,
  proxyMockLogicTest,
} from "./Setup";

describe("Multicall deploy proxy", () => {
  it("multicall deploycreate, deployproxy, upgradeproxy", async () => {
    const factory = await deployFactory();
    const mockFactory = await ethers.getContractFactory(MOCK);
    const endPointFactory = await ethers.getContractFactory(END_POINT);
    const Salt = getSalt();
    const mockCon = await ethers.getContractFactory(MOCK);
    const endPoint = await endPointFactory.deploy();
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
      encodeMultiCallArgs(factory.address, 0, deployCreate),
      encodeMultiCallArgs(factory.address, 0, deployProxy),
      encodeMultiCallArgs(factory.address, 0, upgradeProxy),
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
});
