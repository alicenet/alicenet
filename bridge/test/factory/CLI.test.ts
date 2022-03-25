import { expect } from "chai";
import { BytesLike } from "ethers";
import { ethers, run } from "hardhat";
import {
  DEPLOY_CREATE,
  DEPLOY_FACTORY,
  DEPLOY_METAMORPHIC,
  DEPLOY_PROXY,
  DEPLOY_STATIC,
  DEPLOY_TEMPLATE,
  DEPLOY_UPGRADEABLE_PROXY,
  MOCK,
  MOCK_INITIALIZABLE,
  MULTI_CALL_DEPLOY_PROXY,
  MULTI_CALL_UPGRADE_PROXY,
  UPGRADE_DEPLOYED_PROXY,
} from "../../scripts/lib/constants";
import { deployFactory } from "../../scripts/lib/deployment/deploymentUtil";
import {
  DeployCreateData,
  FactoryData,
  MetaContractData,
  ProxyData,
  TemplateData,
} from "../../scripts/lib/deployment/factoryStateUtil";
import { getSalt } from "../../scripts/lib/MadnetFactory";
import {
  getAccounts,
  getMetamorphicAddress,
  predictFactoryAddress,
} from "./Setup";

describe("Cli tasks", async () => {
  let firstOwner: string;
  let firstDelegator: string;
  const accounts: Array<string> = [];

  beforeEach(async () => {
    process.env.test = "true";
    process.env.silencer = "true";
    // set owner and delegator
    firstOwner = accounts[0];
    firstDelegator = accounts[1];
  });

  it("deploys factory with cli and checks if the default factory is updated in factory state toml file", async () => {
    const accounts = await getAccounts();
    const futureFactoryAddress = await predictFactoryAddress(accounts[0]);
    const factoryData: FactoryData = await run(DEPLOY_FACTORY);
    // check if the address is the predicted
    expect(factoryData.address).to.equal(futureFactoryAddress);
  });
  // todo add init call data and check init vars
  it("deploys MockInitializable contract with deployUpgradeableProxy", async () => {
    // deploys factory using the deployFactory task
    const factoryData: FactoryData = await deployFactory(run);
    const proxyData: ProxyData = await run(DEPLOY_UPGRADEABLE_PROXY, {
      contractName: MOCK_INITIALIZABLE,
      factoryAddress: factoryData.address,
      initCallData: "14",
    });
    const expectedProxyAddress = getMetamorphicAddress(
      factoryData.address,
      ethers.utils.formatBytes32String(MOCK_INITIALIZABLE)
    );
    expect(proxyData.proxyAddress).to.equal(expectedProxyAddress);
  });
  // todo check mock logic
  it("deploys mock contract with deployStatic", async () => {
    // deploys factory using the deployFactory task
    const factoryData: FactoryData = await cliDeployFactory();
    const metaData = await cliDeployMetamorphic(MOCK, factoryData.address, undefined, [
      "2",
      "s",
    ]);
    const salt = ethers.utils.formatBytes32String("Mock");
    const expectedMetaAddr = getMetamorphicAddress(factoryData.address, salt);
    expect(metaData.metaAddress).to.equal(expectedMetaAddr);
  });

  it("deploys MockInitializable contract with deployCreate", async () => {
    const factoryData: FactoryData = await cliDeployFactory();
    const deployCreateData = await cliDeployCreate(MOCK_INITIALIZABLE,factoryData.address);
  });

  it("deploys MockInitializable with deploy create, deploys proxy, then upgrades proxy to point to MockInitializable with initCallData", async () => {
    const factoryData: FactoryData = await cliDeployFactory();
    const test = "1";
    const deployCreateData = await cliDeployCreate(MOCK_INITIALIZABLE, factoryData.address);
    const salt = (await getSalt(MOCK_INITIALIZABLE)) as string;
    const proxyData = await cliDeployProxy(salt, factoryData.address);
    const logicFactory = await ethers.getContractFactory(MOCK_INITIALIZABLE);
    const upgradedProxyData = await cliUpgradeDeployedProxy(
      MOCK_INITIALIZABLE,
      deployCreateData.address,
      factoryData.address,
      test
    );
    const mockContract = logicFactory.attach(upgradedProxyData.proxyAddress);
    const i = await mockContract.callStatic.getImut();
    expect(i.toNumber()).to.equal(parseInt(test, 10));
  });

  it("deploys mock contract with deployTemplate then deploys a metamorphic contract", async () => {
    const factoryData: FactoryData = await cliDeployFactory();
    const testVar1 = "14";
    const testVar2 = "s";
    const templateData = await cliDeployTemplate(MOCK, factoryData.address, [
      testVar1,
      testVar2,
    ]);
    const metaData = await cliDeployStatic(MOCK, factoryData.address, undefined);
    const logicFactory = await ethers.getContractFactory(MOCK);
    const mockContract = logicFactory.attach(metaData.metaAddress);
    const i = await mockContract.callStatic.getImut();
    expect(i.toNumber()).to.equal(parseInt(testVar1, 10));
    const pString = await mockContract.callStatic.getpString();
    expect(pString).to.equal(testVar2);
  });

  it("deploys mockInitializable with deployCreate, then deploy and upgrades a proxy with multiCallDeployProxy", async () => {
    const factoryData: FactoryData = await cliDeployFactory();
    const logicData = await cliDeployCreate(MOCK_INITIALIZABLE, factoryData.address);
    const proxyData = await cliMultiCallDeployProxy(
      MOCK_INITIALIZABLE,
      logicData.address,
      factoryData.address,
      "1"
    );
  });

  it("deploys mock with deployCreate", async () => {
    const factoryData: FactoryData = await cliDeployFactory();
    const factory = await deployFactory(run);
    const deployCreateData = await cliDeployCreate(MOCK, factoryData.address, ["2", "s"]);
  });
});

export async function cliDeployUpgradeableProxy(
  contractName: string,
  factoryAddress?: string,
  initCallData?: string,
  constructorArgs?: Array<string>
): Promise<ProxyData> {
  return await run(DEPLOY_UPGRADEABLE_PROXY, {
    contractName,
    factoryAddress,
    initCallData,
    constructorArgs,
  });
}

export async function cliDeployMetamorphic(
  contractName: string,
  factoryAddress?: string,
  initCallData?: string,
  constructorArgs?: Array<string>
): Promise<MetaContractData> {
  return await run(DEPLOY_METAMORPHIC, {
    contractName,
    factoryAddress,
    initCallData,
    constructorArgs,
  });
}

export async function cliDeployCreate(
  contractName: string,
  factoryAddress?: string,
  constructorArgs?: Array<string>
): Promise<DeployCreateData> {
  return await run(DEPLOY_CREATE, {
    contractName,
    factoryAddress,
    constructorArgs,
  });
}

export async function cliUpgradeDeployedProxy(
  contractName: string,
  logicAddress: string,
  factoryAddress?: string,
  initCallData?: string
): Promise<ProxyData> {
  return await run(UPGRADE_DEPLOYED_PROXY, {
    contractName,
    logicAddress,
    factoryAddress,
    initCallData,
  });
}

export async function cliDeployTemplate(
  contractName: string,
  factoryAddress?: string,
  constructorArgs?: Array<string>
): Promise<TemplateData> {
  return await run(DEPLOY_TEMPLATE, {
    contractName,
    factoryAddress,
    constructorArgs,
  });
}

export async function cliDeployStatic(
  contractName: string,
  factoryAddress?: string,
  initCallData?: Array<string>
): Promise<MetaContractData> {
  return await run(DEPLOY_STATIC, {
    contractName,
    factoryAddress,
    initCallData,
  });
}

export async function cliMultiCallDeployProxy(
  contractName: string,
  logicAddress: string,
  factoryAddress?: string,
  initCallData?: string,
  salt?: string
): Promise<ProxyData> {
  return await run(MULTI_CALL_DEPLOY_PROXY, {
    contractName,
    logicAddress,
    factoryAddress,
    initCallData,
    salt,
  });
}

export async function cliMultiCallUpgradeProxy(
  contractName: string,
  factoryAddress?: BytesLike,
  initCallData?: BytesLike,
  salt?: BytesLike,
  constructorArgs?: Array<string>
): Promise<ProxyData> {
  return await run(MULTI_CALL_UPGRADE_PROXY, {
    contractName,
    factoryAddress,
    initCallData,
    salt,
    constructorArgs,
  });
}

export async function cliDeployFactory() {
  return await run(DEPLOY_FACTORY);
}

export async function cliDeployProxy(
  salt: string,
  factoryAddress?: string
): Promise<ProxyData> {
  return await run(DEPLOY_PROXY, {
    salt,
    factoryAddress,
  });
}
