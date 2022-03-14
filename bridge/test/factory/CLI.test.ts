import { expect } from "chai";
import { BytesLike } from "ethers";
import { config, ethers, run } from "hardhat";
import { TASK_DEPLOY_RUN_DEPLOY } from "hardhat-deploy";
import { DEPLOY_CREATE, DEPLOY_FACTORY, DEPLOY_METAMORPHIC, DEPLOY_PROXY, DEPLOY_STATIC, DEPLOY_TEMPLATE, DEPLOY_UPGRADEABLE_PROXY, INITIALIZER, MOCK, MOCK_INITIALIZABLE, MULTI_CALL_DEPLOY_PROXY, MULTI_CALL_UPGRADE_PROXY, UPGRADE_DEPLOYED_PROXY } from "../../scripts/lib/constants";
import { deployFactory } from "../../scripts/lib/deployment/deploymentUtil";
import { DeployCreateData, getDefaultFactoryAddress, MetaContractData, ProxyData, TemplateData } from "../../scripts/lib/deployment/factoryStateUtil";
import { getSalt } from "../../scripts/lib/MadnetFactory";
import { getAccounts, getMetamorphicAddress, predictFactoryAddress } from "./Setup";


describe("Cli tasks", async () => {
  let firstOwner: string;
  let firstDelegator: string;
  let accounts: Array<string> = [];
  
  beforeEach(async () => {
    process.env.test = "true";
    process.env.silencer = "true";
    //set owner and delegator
    firstOwner = accounts[0];
    firstDelegator = accounts[1];
  });

  it("deploys factory with cli and checks if the default factory is updated in factory state toml file", async () => {
    let accounts = await getAccounts();
    let futureFactoryAddress = await predictFactoryAddress(accounts[0]);
    let factoryAddress = await run(DEPLOY_FACTORY);
    //check if the address is the predicted
    expect(factoryAddress).to.equal(futureFactoryAddress);
  });
  //todo add init call data and check init vars
  it("deploys MockInitializable contract with deployUpgradeableProxy", async () => {
    //deploys factory using the deployFactory task 
    let factory = await deployFactory(run);
    let proxyData: ProxyData = await run(DEPLOY_UPGRADEABLE_PROXY, {
      contractName: MOCK_INITIALIZABLE
    })
    let expectedProxyAddress = getMetamorphicAddress(factory, ethers.utils.formatBytes32String(MOCK_INITIALIZABLE))
    expect(proxyData.proxyAddress).to.equal(expectedProxyAddress) 
  });
  //todo check mock logic
  it("deploys mock contract with deployStatic", async () => {
    //deploys factory using the deployFactory task 
    let factory = await cliDeployFactory()
    let metaData = await cliDeployMetamorphic(MOCK, undefined, undefined, ["2", "s"])
    let salt = ethers.utils.formatBytes32String("Mock")
    let expectedMetaAddr = getMetamorphicAddress(factory, salt)
    expect(metaData.metaAddress).to.equal(expectedMetaAddr)
  });

  it("deploys MockInitializable contract with deployCreate", async () => {
    await cliDeployFactory()
    let deployCreateData = await cliDeployCreate(MOCK_INITIALIZABLE);
  });

  it("deploys MockInitializable with deploy create, deploys proxy, then upgrades proxy to point to MockInitializable with initCallData", async () => {
    await cliDeployFactory()
    let test = "1"
    let deployCreateData = await cliDeployCreate(MOCK_INITIALIZABLE);
    let salt= await getSalt(MOCK_INITIALIZABLE) as string
    let proxyData = await cliDeployProxy(salt)
    let logicFactory = await ethers.getContractFactory(MOCK_INITIALIZABLE)
    let upgradedProxyData = await cliUpgradeDeployedProxy(MOCK_INITIALIZABLE, deployCreateData.address, undefined, test)
    let mockContract = logicFactory.attach(upgradedProxyData.proxyAddress)
    let i = await mockContract.callStatic.getImut()
    expect(i.toNumber()).to.equal(parseInt(test, 10));
  });

  it("deploys mock contract with deployTemplate then deploys a metamorphic contract", async () => {
    await cliDeployFactory()
    let testVar1 = "1"
    let testVar2 = "s"
    let templateData = await cliDeployTemplate(MOCK, undefined, [testVar1, testVar2]);
    let metaData = await cliDeployStatic(MOCK, undefined, undefined);
    let logicFactory = await ethers.getContractFactory(MOCK)
    let mockContract = logicFactory.attach(metaData.metaAddress)
    let i = await mockContract.callStatic.getImut()
    expect(i.toNumber()).to.equal(parseInt(testVar1, 10))
    let pString = await mockContract.callStatic.getpString()
    expect(pString).to.equal(testVar2)
  });

  it("deploys mockInitializable with deployCreate, then deploy and upgrades a proxy with multiCallDeployProxy", async () => {
    await cliDeployFactory()
    let logicData = await cliDeployCreate(MOCK_INITIALIZABLE)
    let proxyData = await cliMultiCallDeployProxy(MOCK_INITIALIZABLE, logicData.address, undefined, "1")
    
  });

  it("deploys mock with deployCreate", async () => {
    await cliDeployFactory()
    let factory = await deployFactory(run);
    let deployCreateData = await cliDeployCreate(MOCK, undefined, ["2", "s"])
  });

});

export async function cliDeployUpgradeableProxy(contractName: string, factoryAddress?: string, initCallData?: string, constructorArgs?: Array<string>): Promise<ProxyData>{
  return await run(DEPLOY_UPGRADEABLE_PROXY, {
    contractName: contractName,
    factoryAddress: factoryAddress,
    initCallData: initCallData,
    constructorArgs: constructorArgs,
  });
}

export async function cliDeployMetamorphic(contractName: string, factoryAddress?: string, initCallData?: string, constructorArgs?: Array<string>): Promise<MetaContractData>{
  return await run(DEPLOY_METAMORPHIC, {
    contractName: contractName,
    factoryAddress: factoryAddress,
    initCallData: initCallData,
    constructorArgs: constructorArgs,
  });
}

export async function cliDeployCreate(contractName: string, factoryAddress?: string, constructorArgs?: Array<string>): Promise<DeployCreateData>{
  return await run(DEPLOY_CREATE, {
    contractName: contractName,
    factoryAddress: factoryAddress,
    constructorArgs: constructorArgs
  })
}

export async function cliUpgradeDeployedProxy(contractName: string, logicAddress: string, factoryAddress?: string, initCallData?: string): Promise<ProxyData> {
  return await run(UPGRADE_DEPLOYED_PROXY, {
    contractName: contractName,
    logicAddress: logicAddress,
    factoryAddress: factoryAddress,
    initCallData: initCallData
  });
}

export async function cliDeployTemplate(contractName: string, factoryAddress?: string, constructorArgs?: Array<string>): Promise<TemplateData>{
  return await run(DEPLOY_TEMPLATE, {
    contractName: contractName,
    factoryAddress: factoryAddress,
    constructorArgs: constructorArgs
  });
}

export async function cliDeployStatic(contractName: string, factoryAddress?: string, initCallData?: Array<string>): Promise<MetaContractData>{
  return await run(DEPLOY_STATIC, {
    contractName: contractName,
    factoryAddress: factoryAddress,
    initCallData: initCallData
  });
}

export async function cliMultiCallDeployProxy(contractName: string, logicAddress:string, factoryAddress?: string, initCallData?: string, salt?: string): Promise<ProxyData>{
  return await run(MULTI_CALL_DEPLOY_PROXY, {
    contractName: contractName,
    logicAddress: logicAddress,
    factoryAddress: factoryAddress,
    initCallData: initCallData,
    salt: salt,
  });
}

export async function cliMultiCallUpgradeProxy(contractName: string, factoryAddress?: BytesLike, initCallData?: BytesLike, salt?: BytesLike, constructorArgs?: Array<string>): Promise<ProxyData>{
  return await run(MULTI_CALL_UPGRADE_PROXY, {
    contractName: contractName,
    factoryAddress: factoryAddress,
    initCallData: initCallData,
    salt: salt,
    constructorArgs: constructorArgs,
  });
}

export async function cliDeployFactory(){
  return await run(DEPLOY_FACTORY)
}

export async function cliDeployProxy(salt: string, factoryAddress?: string): Promise<ProxyData>{
  return await run(DEPLOY_PROXY, {
    salt: salt,
    factoryAddress: factoryAddress
  })
}

