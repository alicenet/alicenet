import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BytesLike, ContractFactory } from "ethers";
import { artifacts, ethers } from "hardhat";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  END_POINT,
  EVENT_DEPLOYED_PROXY,
  EVENT_DEPLOYED_RAW,
  MOCK,
  MOCK_INITIALIZABLE,
  PROXY,
  UTILS,
} from "../utils/constants";
import { AliceNetFactory } from "../../../typechain-types";
import {
  checkMockInit,
  deployCreate2Initable as deployCreate2Initializable,
  deployFactory,
  expectTxSuccess,
  getAccounts,
  getEventVar,
  getMetamorphicAddress,
  getSalt,
} from "./Setup";
import { expect } from "chai";
describe("AliceNet Contract Factory", () => {
  let firstOwner: string;
  let firstDelegator: string;
  let accounts: Array<string> = [];
  let factory: AliceNetFactory;
  beforeEach(async () => {
    process.env.silencer = "true";
    accounts = await getAccounts();
    firstOwner = accounts[0];
    firstDelegator = accounts[2];
    factory = await loadFixture(deployFactory);
  });

  describe("set and get owner", () => {
    it("set owner", async () => {
      // sets the second account as owner
      expect(await factory.owner()).to.equal(firstOwner);
      await factory.setOwner(accounts[1]);
      expect(await factory.owner()).to.equal(accounts[1]);
    });
    it("get owner", async () => {
      const owner = await factory.owner();
      expect(owner).to.equal(firstOwner);
    });
  });

  describe("external contract deployment tests", () => {
    it("deploys a mock contract with deployCreateAndRegister", async () => {
      const salt = ethers.encodeBytes32String("test");
      const mockFactory = await ethers.getContractFactory(MOCK);
      const mockDeployBytecode = mockFactory.getDeployTransaction(2, "s");
      // .data as BytesLike;
      // before deployment the salt will be zero
      const actualMockAddressBeforeDeploy = await factory.lookup(salt);
      expect(actualMockAddressBeforeDeploy).to.equal(ethers.ZeroAddress);
      const tx = await factory.deployCreateAndRegister(
        (await mockDeployBytecode).data,
        salt
      );
      const expectedMockAddress = await getEventVar(
        tx,
        EVENT_DEPLOYED_RAW,
        CONTRACT_ADDR
      );
      await tx.wait();
      const actualLookupAddress = await factory.lookup(salt);
      expect(actualLookupAddress).to.equal(expectedMockAddress);
    });

    it("attempts to deployCreateAndRegister as firstDelegator", async () => {
      const salt = ethers.encodeBytes32String("test");
      const mockFactory = await ethers.getContractFactory(MOCK);
      const mockDeployBytecode = mockFactory.getDeployTransaction(2, "s");
      // .data as BytesLike;
      await expect(
        factory
          .connect(await ethers.provider.getSigner(firstDelegator))
          .deployCreateAndRegister((await mockDeployBytecode).data, salt)
      ).to.be.revertedWithCustomError(factory, "Unauthorized");
    });

    it("attempts to deploy 2 contracts with the same salt", async () => {
      const salt = ethers.encodeBytes32String("test");
      const mockFactory = await ethers.getContractFactory(MOCK);
      const mockDeployBytecode = mockFactory.getDeployTransaction(2, "s");
      // .data as BytesLike;
      const tx = await factory.deployCreateAndRegister(
        (await mockDeployBytecode).data,
        salt
      );
      const expectedMockAddress = await getEventVar(
        tx,
        EVENT_DEPLOYED_RAW,
        CONTRACT_ADDR
      );
      await tx.wait();
      const actualLookupAddress = await factory.lookup(salt);
      expect(actualLookupAddress).to.equal(expectedMockAddress);
      await expect(
        factory.deployCreateAndRegister((await mockDeployBytecode).data, salt)
      )
        .to.be.revertedWithCustomError(factory, "SaltAlreadyInUse")
        .withArgs(salt);
    });
  });
  describe("addNewExternalContract tests", async () => {
    it("deploys a contract from an eoa and records the address to the externalContractRegistry with salt", async () => {
      const salt = ethers.encodeBytes32String("test");
      const mockEndPointBase = await ethers.getContractFactory("MockEndPoint");
      const mockEndPoint = await mockEndPointBase.deploy();
      const lookupBefore = await factory.lookup(salt);
      const txResponse = await factory.addNewExternalContract(
        salt,
        await mockEndPoint.getAddress()
      );
      await txResponse.wait();
      const lookupAfter = await factory.lookup(salt);
      expect(lookupBefore).to.not.equal(lookupAfter);
      expect(lookupAfter).to.equal(await mockEndPoint.getAddress());
    });
    it("attempts to add a eoa address as a contract", async () => {
      const salt = ethers.encodeBytes32String("test");
      await expect(
        factory.addNewExternalContract(salt, firstOwner)
      ).to.be.revertedWithCustomError(factory, "CodeSizeZero");
    });

    it("attempts to add external contract as non owner", async () => {
      const salt = ethers.encodeBytes32String("test");
      const mockEndPointBase = await ethers.getContractFactory("MockEndPoint");
      const mockEndPoint = await mockEndPointBase.deploy();
      const signers = await ethers.getSigners();
      const txResponse = factory
        .connect(signers[1])
        .addNewExternalContract(salt, await mockEndPoint.getAddress());
      await expect(txResponse).to.be.revertedWithCustomError(
        factory,
        "Unauthorized"
      );
    });
    it("should not allow changing address of registered salt", async () => {
      const salt = ethers.encodeBytes32String("test");
      const mockEndPointBase = await ethers.getContractFactory("MockEndPoint");
      const mockEndPoint = await mockEndPointBase.deploy();
      let txResponse = factory.addNewExternalContract(
        salt,
        await mockEndPoint.getAddress()
      );
      await (await txResponse).wait();
      const mockEndPoint2 = await mockEndPointBase.deploy();
      txResponse = factory.addNewExternalContract(
        salt,
        await mockEndPoint2.getAddress()
      );
      await expect(txResponse).to.be.revertedWithCustomError(
        factory,
        "SaltAlreadyInUse"
      );
    });

    it("should not allow register a contract with already deployed proxy with salt", async () => {
      const salt = ethers.encodeBytes32String("test");
      const mockFactory = await ethers.getContractFactory(MOCK);
      const mockDeployBytecode = mockFactory.getDeployTransaction(2, "s");
      // .data as BytesLike;
      await factory.deployCreateAndRegister(
        (await mockDeployBytecode).data,
        salt
      );
      const mockEndPoint2 = await mockFactory.deploy(1, "e");
      const txResponse = factory.addNewExternalContract(
        salt,
        await mockEndPoint2.getAddress()
      );
      await expect(txResponse).to.be.revertedWithCustomError(
        factory,
        "SaltAlreadyInUse"
      );
    });
    it("should not allow deploy proxy salt already registered", async () => {
      const salt = ethers.encodeBytes32String("test");
      const mockFactory = await ethers.getContractFactory(MOCK);
      const mockEndPoint = await mockFactory.deploy(1, "e");
      await factory.addNewExternalContract(
        salt,
        await mockEndPoint.getAddress()
      );
      const mockDeployBytecode = mockFactory.getDeployTransaction(2, "s");
      // .data as BytesLike;
      await expect(
        factory.deployCreateAndRegister((await mockDeployBytecode).data, salt)
      ).to.be.revertedWithCustomError(factory, "SaltAlreadyInUse");
    });
  });

  describe("deployProxy test", async () => {

    it("should not allow deploy proxy with unauthorized account", async () => {
      const signers = await ethers.getSigners();
      const factoryBase = (
        await ethers.getContractFactory(ALICENET_FACTORY)
      ).connect(signers[2]);
      let factory = await deployFactory();
      factory = factoryBase.attach(await factory.getAddress());
      const Salt = getSalt();

      await expect(
        factory.deployProxy(Salt, { from: firstDelegator })
      ).to.be.revertedWithCustomError(factory, `Unauthorized`);
    });
  });

  describe("deployCreate tests", () => {
    it("deploycreate mock logic contract with factory", async () => {
      const factory = await deployFactory();
      // use the ethers Mock contract abstraction to generate the deploy transaction state
      const mockCon: ContractFactory = await ethers.getContractFactory(MOCK);
      // get the init code with constructor args appended
      const deployTx = await mockCon.getDeployTransaction(2, "s");
      const transactionCount = await ethers.provider.getTransactionCount(
        await factory.getAddress()
      );
      // deploy Mock Logic through the factory
      // 27fe1822
      const txResponse = await factory.deployCreate(deployTx.data as BytesLike);
      // check if the transaction is mined or failed
      await expectTxSuccess(txResponse);
      const dcMockAddress = await getEventVar(
        txResponse,
        EVENT_DEPLOYED_RAW,
        CONTRACT_ADDR
      );
      // calculate the deployed address
      const expectedMockAddr = ethers.getCreateAddress({
        from: await factory.getAddress(),
        nonce: transactionCount,
      });
      expect(dcMockAddress).to.equal(expectedMockAddr);
      // console.log("DEPLOYCREATE MOCK LOGIC GASUSED: ", receipt["receipt"]["gasUsed"]);
    });
    it("call setfactory in mock through proxy expect su", async () => {
      const factory = await deployFactory();
      const mockFactory = await ethers.getContractFactory(MOCK);
      const endPointFactory = await ethers.getContractFactory(END_POINT);
      const endPoint = await endPointFactory.deploy();
      const mockContract = await mockFactory.deploy(2, "s");
      const proxySalt = getSalt();
      let txResponse = await factory.deployProxy(proxySalt);
      const proxyAddr = await getEventVar(
        txResponse,
        "DeployedProxy",
        "contractAddr"
      );
      txResponse = await factory.upgradeProxy(
        proxySalt,
        await mockContract.getAddress(),
        "0x"
      );
      await expectTxSuccess(txResponse);
      // connect a Mock interface to the proxy contract
      const proxyMock = await mockFactory.attach(proxyAddr);
      txResponse = await proxyMock.setFactory(accounts[0]);
      await expectTxSuccess(txResponse);
      const mockFactoryAddress = await proxyMock.getFactory.staticCall();
      expect(mockFactoryAddress).to.equal(accounts[0]);
      // console.log("SETFACTORY GASUSED: ", txResponse["receipt"]["gasUsed"]);
      // lock the proxy
      txResponse = await proxyMock.lock();
      await expectTxSuccess(txResponse);
      // console.log("LOCK UPGRADES GASUSED: ", txResponse["receipt"]["gasUsed"]);
      const txRes = factory.upgradeProxy(
        proxySalt,
        await endPoint.getAddress(),
        "0x"
      );
      await expect(txRes).to.be.revertedWith("update locked");
      txResponse = await proxyMock.unlock();
      await expectTxSuccess(txResponse);
      txResponse = await factory.upgradeProxy(
        proxySalt,
        await endPoint.getAddress(),
        "0x"
      );
      await expectTxSuccess(txResponse);
    });
    // fail on bad code
    it("should not allow deploycreate with bad code", async () => {
      const factory = await deployFactory();
      const txResponse = factory.deployCreate("0x6000");

      await expect(txResponse).to.be.revertedWithCustomError(
        factory,
        `CodeSizeZero`
      );
    });
    // fail on unauthorized with bad code
    it("should not allow deploycreate with bad code and unauthorized account", async () => {
      let factory = await deployFactory();
      const signers = await ethers.getSigners();
      const factoryBase = (
        await ethers.getContractFactory(ALICENET_FACTORY)
      ).connect(signers[2]);
      factory = factoryBase.attach(await factory.getAddress());
      const txResponse = factory.deployCreate("0x6000");
      await expect(txResponse).to.be.revertedWithCustomError(
        factory,
        `Unauthorized`
      );
    });
    // fail on unauthorized with good code
    it("should not allow deploycreate with valid code and unauthorized account", async () => {
      const factory = await deployFactory();
      const endPointFactory = await ethers.getContractFactory(END_POINT);
      const signers = await ethers.getSigners();
      const factoryBase = (
        await ethers.getContractFactory(ALICENET_FACTORY)
      ).connect(signers[2]);
      const factory2 = factoryBase.attach(await factory.getAddress());
      const bytecode = await endPointFactory.getDeployTransaction();
      const txResponse = factory2.deployCreate(bytecode.data);
      await expect(txResponse).to.be.revertedWithCustomError(
        factory,
        `Unauthorized`
      );
    });

    it("should not allow deploycreate mock logic contract with unauthorized account", async () => {
      let factory = await deployFactory();
      // use the ethers Mock contract abstraction to generate the deploy transaction state
      const mockCon = await ethers.getContractFactory(MOCK);
      // get the init code with constructor args appended
      const deployTx = await mockCon.getDeployTransaction(2, "s");
      // deploy Mock Logic through the factory
      const signers = await ethers.getSigners();
      const factoryBase = (
        await ethers.getContractFactory(ALICENET_FACTORY)
      ).connect(signers[2]);
      factory = factoryBase.attach(await factory.getAddress());
      const txResponse = factory.deployCreate(deployTx.data);

      await expect(txResponse).to.be.revertedWithCustomError(
        factory,
        `Unauthorized`
      );
    });
  });

  describe("deployCreate2 tests", () => {
    it("deploycreate2 mockinitializable", async () => {
      const factory = await deployFactory();
      const salt = getSalt();
      let txResponse = await deployCreate2Initializable(factory, salt);
      await expectTxSuccess(txResponse);
      const mockInitAddr = await await getEventVar(
        txResponse,
        EVENT_DEPLOYED_RAW,
        CONTRACT_ADDR
      );
      expect(mockInitAddr).to.not.be.equals(undefined);
      const mockInitializable =
        await ethers.getContractFactory(MOCK_INITIALIZABLE);
      const initCallData = mockInitializable.interface.encodeFunctionData(
        "initialize",
        [2]
      );
      txResponse = await factory.initializeContract(mockInitAddr, initCallData);
      await expectTxSuccess(txResponse);
      await checkMockInit(mockInitAddr, 2);
    });
  });

  it("deploys a initializable contract and uses callany to intialize", async () => {
    const factory = await deployFactory();
    const salt = getSalt();
    let txResponse = await deployCreate2Initializable(factory, salt);
    await expectTxSuccess(txResponse);
    const mockInitAddr = await getEventVar(
      txResponse,
      EVENT_DEPLOYED_RAW,
      CONTRACT_ADDR
    );
    expect(mockInitAddr).to.not.be.equals(undefined);
    // call state to initialize mockInitializable
    const mockInitable = await ethers.getContractFactory(MOCK_INITIALIZABLE);
    const initCallData = mockInitable.interface.encodeFunctionData(
      "initialize",
      [2]
    );
    txResponse = await factory.callAny(mockInitAddr, 0, initCallData);
    await checkMockInit(mockInitAddr, 2);
  });

  describe("proxy upgrade tests", async () => {
    it("upgrade proxy to point to mock address with the factory", async () => {
      const factory = await deployFactory();
      const mockFactory = await ethers.getContractFactory(MOCK);
      const proxySalt = getSalt();
      let txResponse = await factory.deployProxy(proxySalt);
      await expectTxSuccess(txResponse);
      const mockContract = await mockFactory.deploy(2, "s");
      txResponse = await factory.upgradeProxy(
        proxySalt,
        await mockContract.getAddress(),
        "0x"
      );
      await expectTxSuccess(txResponse);
    });

    it("should not allow unauthorized account to update proxy to point to other address ", async () => {
      let factory = await deployFactory();
      const mockFactory = await ethers.getContractFactory(MOCK);
      const proxySalt = getSalt();
      let txResponse = await factory.deployProxy(proxySalt);
      await expectTxSuccess(txResponse);
      const mockContract = await mockFactory.deploy(2, "s");
      txResponse = await factory.upgradeProxy(
        proxySalt,
        await mockContract.getAddress(),
        "0x"
      );
      await expectTxSuccess(txResponse);
      const signers = await ethers.getSigners();
      const factoryBase = (
        await ethers.getContractFactory(ALICENET_FACTORY)
      ).connect(signers[2]);
      factory = factoryBase.attach(await factory.getAddress());
      await expect(
        factory.upgradeProxy(proxySalt, await mockContract.getAddress(), "0x", {
          from: firstDelegator,
        })
      ).to.be.revertedWithCustomError(factory, `Unauthorized`);
    });
  });

  describe("lookup tests", async () => {
    it("should return the correct proxy address", async () => {
      const salt = getSalt();
      const expectedProxyAddr = getMetamorphicAddress(
        await factory.getAddress(),
        salt
      );
      const txResponse = await factory.deployProxy(salt);
      await expectTxSuccess(txResponse);
      const proxyAddr = await getEventVar(
        txResponse,
        EVENT_DEPLOYED_PROXY,
        CONTRACT_ADDR
      );
      expect(proxyAddr).to.equal(expectedProxyAddr);
      const proxyAddrFromLookup = await factory.lookup(salt);
      expect(proxyAddrFromLookup).to.equal(proxyAddr);
    });

    it("get ALCA Address from lookup", async () => {
      const salt = ethers.encodeBytes32String("ALCA");
      let alcaAddress = await factory.lookup(salt);
      const alca = await ethers.getContractAt("ALCA", alcaAddress);
      const legacyToken = await alca.getLegacyTokenAddress();
      const alcaBase = await ethers.getContractFactory("ALCA");
      const alcaDeployCode = await alcaBase.getDeployTransaction(legacyToken);
      // .data as BytesLike;
      const alcaHash = ethers.keccak256(alcaDeployCode.data);
      const expectedALCAAddress = ethers.getCreate2Address(
        await factory.getAddress(),
        salt,
        alcaHash
      );
      alcaAddress = await factory.lookup(salt);
      expect(alcaAddress).to.equal(expectedALCAAddress);
    });
  });

  it("deploys a mock contract, calls payMe from factory with callAny", async () => {
    const factory = await deployFactory();
    const mockFactory = await ethers.getContractFactory(MOCK);
    const mock = await mockFactory.deploy(2, "s");
    const callData = mockFactory.interface.encodeFunctionData("payMe");
    await factory.callAny(await mock.getAddress(), 2, callData, {
      value: 2,
    });
  });

  it("check callAny returns", async () => {
    const factory = await deployFactory();
    const mockFactory = await ethers.getContractFactory(MOCK);
    const mock = await mockFactory.deploy(2, "s");
    const callData = mockFactory.interface.encodeFunctionData("getVar");
    await mock.setV(1);
    let returnValue = await factory.callAny.staticCall(
      await mock.getAddress(),
      0,
      callData
    );
    expect(returnValue).to.be.equal(
      "0x0000000000000000000000000000000000000000000000000000000000000001"
    );
    const abicoder = new ethers.AbiCoder();
    expect(abicoder.decode(["uint256"], returnValue)).to.be.deep.equal([
      BigInt(1),
    ]);
    await mock.setV(50);
    returnValue = await factory.callAny.staticCall(
      await mock.getAddress(),
      0,
      callData
    );
    expect(returnValue).to.be.equal(
      "0x0000000000000000000000000000000000000000000000000000000000000032"
    );
    expect(abicoder.decode(["uint256"], returnValue)).to.be.deep.equal([
      BigInt(50),
    ]);
  });
});
