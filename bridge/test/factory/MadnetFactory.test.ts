import {
  MOCK,
  DEPLOYED_PROXY,
  DEPLOYED_RAW,
  DEPLOYED_STATIC,
  DEPLOYED_TEMPLATE,
  MOCK_INITIALIZABLE,
  END_POINT,
  CONTRACT_ADDR,
  PROXY,
  UTILS,
  MADNET_FACTORY,
} from '../../scripts/lib/constants';
import { expect } from "chai";
import { expectRevert } from "@openzeppelin/test-helpers";
import {
  checkMockInit,
  deployCreate2Initable as deployCreate2Initializable,
  deployFactory,
  expectTxSuccess,
  getAccounts,
  getCreateAddress,
  getDeployStaticArgs,
  getDeployTemplateArgs,
  getEventVar,
  getMetamorphicAddress,
  getSalt,
} from "./Setup";
import { ethers, artifacts } from "hardhat";
import { BytesLike, ContractFactory } from "ethers";
describe("Madnet Contract Factory", () => {
  let firstOwner: string;
  let secondOwner: string;
  let firstDelegator: string;
  let accounts: Array<string> = [];


  beforeEach(async () => {
    process.env.silencer = "true";
    accounts = await getAccounts();
    firstOwner = accounts[0];
    secondOwner = accounts[1];
    firstDelegator = accounts[2];
  });
//delete
  it("deploy mock", async () => {
    let UtilsBase = await ethers.getContractFactory(UTILS)
    let utils = await UtilsBase.deploy();
    let mockFactory = await ethers.getContractFactory(MOCK)
    let mock = await mockFactory.deploy(2, "s");
    let size = await utils.getCodeSize(mock.address);
    expect(size.toNumber()).to.be.greaterThan(0);
  });
//delete
  it("deploy endpoint", async () => {
    let UtilsBase = await ethers.getContractFactory(UTILS);
    let utils = await UtilsBase.deploy();
    let factory = await deployFactory();
    let endPointFactory = await ethers.getContractFactory(END_POINT)
    let endPoint = await endPointFactory.deploy(factory.address);
    let size = await utils.getCodeSize(endPoint.address);
    expect(size.toNumber()).to.be.greaterThan(0);
  });

  it("set owner", async () => {
    let factory = await deployFactory();
    //sets the second account as owner
    expect(await factory.owner()).to.equal(firstOwner);
    await factory.setOwner(accounts[1]);
    expect(await factory.owner()).to.equal(accounts[1]);
  });

  it("set delegator", async () => {
    let factory = await deployFactory();
    //sets the second account as delegator
    await factory.setDelegator(firstDelegator);
    expect(await factory.delegator()).to.equal(firstDelegator);
  });

  it("should not allow set owner via delegator", async () => {
    let factory = await deployFactory();
    let signers = await ethers.getSigners()
    let factoryBase = (await ethers.getContractFactory(MADNET_FACTORY)).connect(signers[2])
    //sets the second account as delegator
    await factory.setDelegator(firstDelegator, { from: firstOwner });
    expect(await factory.delegator()).to.equal(firstDelegator);
    factory = factoryBase.attach(factory.address)
    await expect(
      factory.setOwner(accounts[0])
    ).to.be.revertedWith("unauthorized")
  });

  it("get owner, delegator", async () => {
    let factory = await deployFactory();
    await factory.setDelegator(firstDelegator, { from: firstOwner });
    expect(await factory.delegator()).to.equal(firstDelegator);
    let owner = await factory.owner();
    expect(owner).to.equal(firstOwner);
    let delegator = await factory.delegator();
    expect(delegator).to.equal(firstDelegator);
  });

  it("deploy mock with deploytemplate as owner expect succeed", async () => {
    let factory = await deployFactory();
    //ethers instance of Mock contract abstraction
    let mockCon = await ethers.getContractFactory(MOCK);
    //deploy code for mock with constructor args i = 2
    let deployTxData = mockCon.getDeployTransaction(2, "s").data as BytesLike;
    //deploy the mock Contract to deployTemplate
    let transactionCount = await ethers.provider.getTransactionCount(
      factory.address
    );
    let expectedMockTempAddress = getCreateAddress(
      factory.address,
      transactionCount
    );
    let txResponse = await factory.deployTemplate(deployTxData);
    await expectTxSuccess(txResponse);
    let mockTempAddress = await getEventVar(
      txResponse,
      DEPLOYED_TEMPLATE,
      CONTRACT_ADDR
    );
    expect(mockTempAddress).to.equal(expectedMockTempAddress);
    // console.log("DEPLOYTEMPLATE GASUSED: ", receipt["receipt"]["gasUsed"]);
  });

  it("should not allow deploy contract with bytecode 0", async () => {
    let factory = await deployFactory();
    let Salt = getSalt();
    await expectRevert(
        factory.deployStatic(Salt, "0x"),
        "reverted with an unrecognized custom error"
      );

  });

  it("should not allow deploy static with unauthorized account", async () => {
    let signers = await ethers.getSigners()
    let factoryBase = (await ethers.getContractFactory(MADNET_FACTORY)).connect(signers[2])
    let factory = await deployFactory();
    factory = factoryBase.attach(factory.address)
    let Salt = getSalt();
    await expect(
      factory.deployStatic(Salt, "0x", { from: firstDelegator })).to.be.revertedWith("unauthorized");
  });

  it("deploy contract with deploystatic", async () => {
    let factory = await deployFactory();
    //deploy a template of the mock Initializable
    let byteCode = (await getDeployTemplateArgs(
      MOCK_INITIALIZABLE
    )) as BytesLike;
    let txResponse = await factory.deployTemplate(byteCode);
    let deployStatic = await getDeployStaticArgs(MOCK_INITIALIZABLE, [2]);
    txResponse = await factory.deployStatic(
      deployStatic.salt,
      deployStatic.initCallData
    );
    let mockInitAddr = await getEventVar(txResponse, DEPLOYED_STATIC, CONTRACT_ADDR);
    checkMockInit(mockInitAddr, 2);
  });

  it("deployproxy", async () => {
    let UtilsBase = await ethers.getContractFactory(UTILS);
    let utils = await UtilsBase.deploy();
    let factory = await deployFactory();
    let proxyBase = await artifacts.require(PROXY);
    let proxySalt = getSalt();
    //the calculated proxy address
    const expectedProxyAddr = getMetamorphicAddress(factory.address, proxySalt);
    //deploy the proxy through the factory
    let txResponse = await factory.deployProxy(proxySalt);
    //check if transaction succeeds
    await expectTxSuccess(txResponse);
    //get the deployed proxy contract address fom the DeployedProxy event
    let proxyAddr = await await getEventVar(txResponse, DEPLOYED_PROXY, CONTRACT_ADDR);
    //check if the deployed contract address match the calculated address
    expect(proxyAddr).to.equal(expectedProxyAddr);
    // console.log("DEPLOYPROXY GASUSED: ", receipt["receipt"]["gasUsed"]);
    let cSize = await utils.getCodeSize(proxyAddr);
    expect(cSize.toNumber()).to.equal(
      (proxyBase.deployedBytecode.length - 2) / 2
    );
  });

  it("should not allow deploy proxy with unauthorized account", async () => {
    let signers = await ethers.getSigners()
    let factoryBase = (await ethers.getContractFactory(MADNET_FACTORY)).connect(signers[2])
    let factory = await deployFactory();
    factory = factoryBase.attach(factory.address)
    let Salt = getSalt();
    await expectRevert(
      factory.deployProxy(Salt, { from: firstDelegator }),
      "unauthorized"
    );
  });

  it("deploycreate mock logic contract expect success", async () => {
    let factory = await deployFactory();
    //use the ethers Mock contract abstraction to generate the deploy transaction data
    let mockCon: ContractFactory = await ethers.getContractFactory(MOCK);
    //get the init code with contructor args appended
    let deployTx = mockCon.getDeployTransaction(2, "s");
    let transactionCount = await ethers.provider.getTransactionCount(
      factory.address
    );
    //deploy Mock Logic through the factory
    // 27fe1822
    let txResponse = await factory.deployCreate(deployTx.data as BytesLike);
    //check if the transaction is mined or failed
    await expectTxSuccess(txResponse);
    let dcMockAddress = await getEventVar(txResponse, DEPLOYED_RAW, CONTRACT_ADDR);
    //calculate the deployed address
    let expectedMockAddr = ethers.utils.getContractAddress({
      from: factory.address,
      nonce: transactionCount,
    });
    expect(dcMockAddress).to.equal(expectedMockAddr);
    // console.log("DEPLOYCREATE MOCK LOGIC GASUSED: ", receipt["receipt"]["gasUsed"]);
  });

  it("should not allow deploycreate mock logic contract with unauthorized account", async () => {
    let factory = await deployFactory();
    //use the ethers Mock contract abstraction to generate the deploy transaction data
    let mockCon = await ethers.getContractFactory(MOCK);
    //get the init code with contructor args appended
    let deployTx = mockCon.getDeployTransaction(2, "s");
    //deploy Mock Logic through the factory
    let signers = await ethers.getSigners()
    let factoryBase = (await ethers.getContractFactory(MADNET_FACTORY)).connect(signers[2])
    factory = factoryBase.attach(factory.address)
    let txResponse = factory.deployCreate(deployTx.data as BytesLike);
    await expect(txResponse).to.be.revertedWith("unauthorized");
  });

  it("upgrade proxy to point to mock address with the factory", async () => {
    let factory = await deployFactory();
    let mockFactory = await ethers.getContractFactory(MOCK);
    let proxySalt = getSalt();
    let txResponse = await factory.deployProxy(proxySalt);
    await expectTxSuccess(txResponse);
    let mockContract = await mockFactory.deploy(2, "s");
    txResponse = await factory.upgradeProxy(
      proxySalt,
      mockContract.address,
      "0x"
    );
    await expectTxSuccess(txResponse);
    // console.log("UPGRADE PROXY GASUSED: ", txResponse["receipt"]["gasUsed"]);
  });

  it("should not allow unauthorized account to update proxy to point to other address ", async () => {
    let factory = await deployFactory();
    let mockFactory = await ethers.getContractFactory(MOCK);
    let proxySalt = getSalt();
    let txResponse = await factory.deployProxy(proxySalt);
    await expectTxSuccess(txResponse);
    let mockContract = await mockFactory.deploy(2, "s");
    txResponse = await factory.upgradeProxy(
      proxySalt,
      mockContract.address,
      "0x"
    );
    await expectTxSuccess(txResponse);
    let signers = await ethers.getSigners()
    let factoryBase = (await ethers.getContractFactory(MADNET_FACTORY)).connect(signers[2])
    factory = factoryBase.attach(factory.address)
    // console.log("UPGRADE PROXY GASUSED: ", txResponse["receipt"]["gasUsed"]);
    await expect(factory.upgradeProxy(proxySalt, mockContract.address, "0x", {
        from: firstDelegator,
      })).to.be.revertedWith("unauthorized")
  });

  it("call setfactory in mock through proxy expect su", async () => {
    let factory = await deployFactory();
    let mockFactory = await ethers.getContractFactory(MOCK);
    let endPointFactory = await ethers.getContractFactory(END_POINT);
    let endPoint = await endPointFactory.deploy(factory.address);
    let mockContract = await mockFactory.deploy(2, "s");
    let proxySalt = getSalt();
    let txResponse = await factory.deployProxy(proxySalt);
    let proxyAddr = await getEventVar(
      txResponse,
      "DeployedProxy",
      "contractAddr"
    );
    txResponse = await factory.upgradeProxy(
      proxySalt,
      mockContract.address,
      "0x"
    );
    await expectTxSuccess(txResponse);
    //connect a Mock interface to the proxy contract
    let proxyMock = await mockFactory.attach(proxyAddr);
    txResponse = await proxyMock.setFactory(accounts[0]);
    await expectTxSuccess(txResponse);
    let mockFactoryAddress = await proxyMock.callStatic.getFactory();
    expect(mockFactoryAddress).to.equal(accounts[0]);
    // console.log("SETFACTORY GASUSED: ", txResponse["receipt"]["gasUsed"]);
    //lock the proxy
    txResponse = await proxyMock.lock();
    await expectTxSuccess(txResponse);
    // console.log("LOCK UPGRADES GASUSED: ", txResponse["receipt"]["gasUsed"]);
    let txRes = factory.upgradeProxy(proxySalt, endPoint.address, "0x");
    await expect(txRes).to.be.revertedWith("revert");
    txResponse = await proxyMock.unlock();
    await expectTxSuccess(txResponse);
    txResponse = await factory.upgradeProxy(proxySalt, endPoint.address, "0x");
    await expectTxSuccess(txResponse);
  });

  //fail on bad code
  it("should not allow deploycreate with bad code", async () => {
    let factory = await deployFactory();
    let txResponse = factory.deployCreate("0x6000");
    await expect(txResponse).to.be.revertedWith("csize0");
  });

  //fail on unauthorized with bad code
  it("should not allow deploycreate with bad code and unauthorized account", async () => {
    let factory = await deployFactory();
    let signers = await ethers.getSigners()
    let factoryBase = (await ethers.getContractFactory(MADNET_FACTORY)).connect(signers[2])
    factory = factoryBase.attach(factory.address)
    let txResponse = factory.deployCreate("0x6000");
    await expect(txResponse).to.be.revertedWith("unauthorized");
  });

  //fail on unauthorized with good code
  it("should not allow deploycreate with valid code and unauthorized account", async () => {
    let factory = await deployFactory();
    let endPointFactory = await ethers.getContractFactory(END_POINT)
    let signers = await ethers.getSigners()
    let factoryBase = (await ethers.getContractFactory(MADNET_FACTORY)).connect(signers[2])
    let factory2 = factoryBase.attach(factory.address)
    let txResponse = factory2.deployCreate(endPointFactory.bytecode);
    await expect(txResponse).to.be.revertedWith("unauthorized");
  });

  it("deploycreate2 mockinitializable", async () => {
    let factory = await deployFactory();
    let salt = getSalt();
    let txResponse = await deployCreate2Initializable(factory, salt);
    await expectTxSuccess(txResponse);
    let mockInitAddr = await await getEventVar(txResponse, DEPLOYED_RAW, CONTRACT_ADDR);
    expect(mockInitAddr).to.not.be.undefined;
    let mockInitializable = await ethers.getContractFactory(MOCK_INITIALIZABLE);
    let initCallData = mockInitializable.interface.encodeFunctionData(
      "initialize",
      [2]
    );
    txResponse = await factory.initializeContract(
      mockInitAddr,
      initCallData
    );
    await expectTxSuccess(txResponse);
    await checkMockInit(mockInitAddr, 2);
  });

  it("callany", async () => {
    let factory = await deployFactory();
    let salt = await getSalt();
    let txResponse = await deployCreate2Initializable(factory, salt);
    await expectTxSuccess(txResponse);
    let mockInitAddr = await getEventVar(txResponse, DEPLOYED_RAW, CONTRACT_ADDR);
    expect(mockInitAddr).to.not.be.undefined;
    // call data to initialize mockInitializable
    let mockInitable = await ethers.getContractFactory(MOCK_INITIALIZABLE);
    let initCallData = await mockInitable.interface.encodeFunctionData(
      "initialize",
      [2]
    );
    txResponse = await factory.callAny(mockInitAddr, 0, initCallData);
    await checkMockInit(mockInitAddr, 2);
  });

  it("delegatecallany", async () => {
    let factory = await deployFactory();
    expect(await factory.owner()).to.equal(firstOwner);
    //deploy an instance of mock logic for factory
    let mockFactoryBase = await ethers.getContractFactory("MockFactory")
    let mockFactoryInstance = await mockFactoryBase.deploy();
    //generate the call data for the factory instance
    let mfEncode = await ethers.getContractFactory("MockFactory");
    let setOwner = mfEncode.interface.encodeFunctionData("setOwner", [
      accounts[2],
    ]);
    //delegate call into the factory and change the owner
    let txResponse = await factory.delegateCallAny(
      mockFactoryInstance.address,
      setOwner
    );
    await expectTxSuccess(txResponse);
    let owner = await factory.owner();
    expect(owner).to.equal(accounts[2]);
    setOwner = await mfEncode.interface.encodeFunctionData("setOwner", [
      accounts[0],
    ]);
    let signers = await ethers.getSigners()
    let factoryBase = (await ethers.getContractFactory(MADNET_FACTORY)).connect(signers[2])
    let factory2 = factoryBase.attach(factory.address)
    txResponse = await factory2.delegateCallAny(
      mockFactoryInstance.address,
      setOwner
    );
    await expectTxSuccess(txResponse);
    owner = await factory.owner();
    expect(owner).to.equal(accounts[0]);
  });

  it("upgrade proxy through factory", async () => {
    let factory = await deployFactory();
    let endPointLockableFactory = await ethers.getContractFactory("MockEndPointLockable");
    let endPointLockable = await endPointLockableFactory.deploy(factory.address);
    let salt = getSalt();
    let expectedProxyAddr = getMetamorphicAddress(factory.address, salt);
    let txResponse = await factory.deployProxy(salt);
    await expectTxSuccess(txResponse);
    let proxyAddr = await getEventVar(txResponse, DEPLOYED_PROXY, CONTRACT_ADDR);
    expect(proxyAddr).to.equal(expectedProxyAddr);
    txResponse = await factory.upgradeProxy(salt, endPointLockable.address, "0x");
    await expectTxSuccess(txResponse);
    let proxyFactory = await ethers.getContractFactory("Proxy");
    let proxyContract = proxyFactory.attach(proxyAddr)
    let proxyImplAddress = await proxyContract.callStatic.getImplementationAddress();
    expect(proxyImplAddress).to.equal(endPointLockable.address);
  });

  it("lock proxy upgrade from factory", async () => {
    let factory = await deployFactory();
    let endPointLockableFactory = await ethers.getContractFactory("MockEndPointLockable");
    let endPointLockable = await endPointLockableFactory.deploy(factory.address);
    let salt = getSalt();
    let expectedProxyAddr = getMetamorphicAddress(factory.address, salt);
    let txResponse = await factory.deployProxy(salt);
    await expectTxSuccess(txResponse);
    let proxyAddr = await getEventVar(txResponse, DEPLOYED_PROXY, CONTRACT_ADDR);
    expect(proxyAddr).to.equal(expectedProxyAddr);
    txResponse = await factory.upgradeProxy(salt, endPointLockable.address, "0x");
    await expectTxSuccess(txResponse);
    let proxyFactory = await ethers.getContractFactory("Proxy");
    let proxyContract = proxyFactory.attach(proxyAddr)
    let proxyImplAddress = await proxyContract.callStatic.getImplementationAddress();
    expect(proxyImplAddress).to.equal(endPointLockable.address);
  });

  it("should prevent locked proxy logic from being upgraded from factory", async () => {
    let factory = await deployFactory();
    let endPointFactory = await ethers.getContractFactory(END_POINT);
    let endPoint = await endPointFactory.deploy(factory.address);
    let endPointLockableFactory = await ethers.getContractFactory("MockEndPointLockable");
    let endPointLockable = await endPointLockableFactory.deploy(factory.address);
    let salt = getSalt();
    let expectedProxyAddr = getMetamorphicAddress(factory.address, salt);
    let txResponse = await factory.deployProxy(salt);
    await expectTxSuccess(txResponse);
    let proxyAddr = await getEventVar(txResponse, DEPLOYED_PROXY, CONTRACT_ADDR);
    expect(proxyAddr).to.equal(expectedProxyAddr);
    txResponse = await factory.upgradeProxy(salt, endPointLockable.address, "0x");
    await expectTxSuccess(txResponse);
    let proxyFactory = await ethers.getContractFactory("Proxy");
    let proxy = proxyFactory.attach(proxyAddr)
    let proxyImplAddress = await proxy.callStatic.getImplementationAddress();
    expect(proxyImplAddress).to.equal(endPointLockable.address);
    let proxyContract = endPointLockableFactory.attach(proxy.address);
    let lockResponse = await proxyContract.upgradeLock();
    let receipt = await lockResponse.wait();
    expect(receipt.status).to.equal(1);
    await expect(factory.upgradeProxy(salt, endPoint.address, "0x")).to.revertedWith("reverted with an unrecognized custom error");
  });
});
