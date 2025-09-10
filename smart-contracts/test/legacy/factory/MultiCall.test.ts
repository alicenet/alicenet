import { expect } from "chai";
import { ethers } from "hardhat";
import { encodeMultiCallArgs } from "../utils/alicenetFactory";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  END_POINT,
  EVENT_DEPLOYED_PROXY,
  EVENT_DEPLOYED_RAW,
  MOCK,
} from "../utils/constants";
import { AliceNetFactory } from "../../../typechain-types";
import {
  deployFactory,
  getEventVar,
  getSalt,
  proxyMockLogicTest,
} from "./Setup";

describe("Multicall deploy proxy", () => {
  let factory: AliceNetFactory;
  beforeEach(async () => {
    factory = await deployFactory();
  });

  it("multicall deploycreate, deployproxy, upgradeproxy", async () => {
    const mockFactory = await ethers.getContractFactory(MOCK);
    const endPointFactory = await ethers.getContractFactory(END_POINT);
    const Salt = getSalt();
    const mockCon = await ethers.getContractFactory(MOCK);
    const endPoint = await endPointFactory.deploy();
    // deploy code for mock with constructor args i = 2
    const deployTX = await mockCon.getDeployTransaction(2, "s");
    const factoryBase = await ethers.getContractFactory(ALICENET_FACTORY);
    const transactionCount = await ethers.provider.getTransactionCount(
      await factory.getAddress()
    );
    // calculate the deployCreate Address
    const expectedMockLogicAddr = ethers.getCreateAddress({
      from: await factory.getAddress(),
      nonce: transactionCount,
    });
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
      encodeMultiCallArgs(await factory.getAddress(), BigInt(0), deployCreate),
      encodeMultiCallArgs(await factory.getAddress(), BigInt(0), deployProxy),
      encodeMultiCallArgs(await factory.getAddress(), BigInt(0), upgradeProxy),
    ]);
    const mockLogicAddr = await getEventVar(
      txResponse,
      EVENT_DEPLOYED_RAW,
      CONTRACT_ADDR
    );
    const proxyAddr = await getEventVar(
      txResponse,
      EVENT_DEPLOYED_PROXY,
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
      await endPoint.getAddress(),
      await factory.getAddress()
    );
  });

  it("check multicall returns", async () => {
    const mockFactory = await ethers.getContractFactory(MOCK);
    const mock = await mockFactory.deploy(2, "s");
    // encoded function call to deployCreate
    const setVar1 = mock.interface.encodeFunctionData("setV", [1]);
    // encoded function call to deployProxy
    const getVar = mock.interface.encodeFunctionData("getVar");
    const setVar2 = mock.interface.encodeFunctionData("setV", [2]);
    const returnValue = await factory.multiCall.staticCall([
      encodeMultiCallArgs(await mock.getAddress(), BigInt(0), setVar1),
      encodeMultiCallArgs(await mock.getAddress(), BigInt(0), getVar),
      encodeMultiCallArgs(await mock.getAddress(), BigInt(0), setVar2),
      encodeMultiCallArgs(await mock.getAddress(), BigInt(0), getVar),
    ]);
    expect(returnValue).to.be.deep.equal([
      "0x",
      "0x0000000000000000000000000000000000000000000000000000000000000001",
      "0x",
      "0x0000000000000000000000000000000000000000000000000000000000000002",
    ]);
    const abicoder = new ethers.AbiCoder();
    expect(abicoder.decode(["uint256"], returnValue[1])).to.be.deep.equals([
      BigInt(1),
    ]);
    expect(abicoder.decode(["uint256"], returnValue[3])).to.be.deep.equals([
      BigInt(2),
    ]);
  });
});
