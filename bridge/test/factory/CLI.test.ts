import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BytesLike } from "ethers";
import { ethers, run } from "hardhat";
import { MOCK, MOCK_INITIALIZABLE } from "../../scripts/lib/constants";
import {
  DeployCreateData,
  FactoryData,
  MetaContractData,
  ProxyData,
} from "../../scripts/lib/deployment/factoryStateUtil";
import { AliceNetFactory } from "../../typechain-types";
import { expect } from "../chai-setup";
import { getMetamorphicAddress } from "./Setup";

async function deployFixture() {
  const accounts = await ethers.getSigners();
  const factoryData: FactoryData = await run("deploy-factory", {
    constructorArgs: [accounts[1].address],
  });
  // check if the address is the predicted
  return await ethers.getContractAt("AliceNetFactory", factoryData.address);
}

describe("Cli tasks", () => {
  let factory: AliceNetFactory;
  beforeEach(async () => {
    process.env.test = "true";
    process.env.silencer = "true";
    factory = await loadFixture(deployFixture);
  });

  it("deploys a mock contract with a initializer function with npx hardhat deploy-upgradeable-proxy, and checks if initializer arg is set correctly", async () => {
    const expectedInitVal = 14;
    const proxyData: ProxyData = await run("deploy-upgradeable-proxy", {
      contractName: MOCK_INITIALIZABLE,
      factoryAddress: factory.address,
      initCallData: `${expectedInitVal}`,
    });
    const mockInitializable = await ethers.getContractAt(
      MOCK_INITIALIZABLE,
      proxyData.proxyAddress
    );
    const initval = await mockInitializable.getImut();
    const expectedProxyAddress = getMetamorphicAddress(
      factory.address,
      ethers.utils.formatBytes32String(MOCK_INITIALIZABLE)
    );
    expect(proxyData.proxyAddress).to.equal(expectedProxyAddress);
    expect(initval).to.equal(expectedInitVal);
  });

  it("attempts to deploy proxy without initializer args with npx hardhat deploy-upgradeable-proxy", async () => {
    try {
      await run("deploy-upgradeable-proxy", {
        contractName: MOCK_INITIALIZABLE,
        factoryAddress: factory.address,
      });
    } catch (error: any) {
      expect(error.reason).to.equal("types/values length mismatch");
      expect(error.code).to.equal("INVALID_ARGUMENT");
    }
  });

  it("deploys MockInitializable contract with deployCreate", async () => {
    const nonce = await ethers.provider.getTransactionCount(factory.address);
    const expectedAddress = ethers.utils.getContractAddress({
      from: factory.address,
      nonce,
    });
    const deployCreateData = await cliDeployCreate(
      MOCK_INITIALIZABLE,
      factory.address
    );
    expect(deployCreateData.address).to.equal(expectedAddress);
  });

  it("deploys MockInitializable with deploy create, deploys proxy, then upgrades proxy to point to MockInitializable with initCallData", async () => {
    const testInitArg = "1";
    const logicContractBase = await ethers.getContractFactory(
      MOCK_INITIALIZABLE
    );
    const proxyData = await cliDeployUpgradeableProxy(
      MOCK_INITIALIZABLE,
      factory.address,
      testInitArg
    );
    const mockContract = logicContractBase.attach(proxyData.proxyAddress);
    const i = await mockContract.callStatic.getImut();
    expect(i.toNumber()).to.equal(parseInt(testInitArg, 10));
  });

  it("deploys mock with deployCreate", async () => {
    const deployCreateData = await cliDeployCreate(MOCK, factory.address, [
      "2",
      "s",
    ]);
    expect(deployCreateData.address).to.not.equal(ethers.constants.AddressZero);
  });

  xit("deploys all contracts in deploymentList", async () => {
    await cliDeployContracts();
  });
});

export async function cliDeployContracts(
  factoryAddress?: string,
  inputFolder?: string
) {
  return await run("deploy-contracts", {
    factoryAddress,
    inputFolder,
  });
}

export async function cliFullMultiCallDeployProxy(
  contractName: string,
  factoryAddress: string,
  initCallData?: string,
  outputFolder?: string,
  constructorArgs?: Array<string>
): Promise<ProxyData> {
  return await run("full-multi-call-deploy-proxy", {
    contractName,
    factoryAddress,
    initCallData,
    outputFolder,
    constructorArgs,
  });
}

export async function cliMultiCallDeployMetamorphic(
  contractName: string,
  factoryAddress: string,
  initCallData?: string,
  outputFolder?: string,
  constructorArgs?: Array<string>
): Promise<MetaContractData> {
  return await run("multi-call-deploy-metamorphic", {
    contractName,
    factoryAddress,
    initCallData,
    outputFolder,
    constructorArgs,
  });
}

export async function cliDeployUpgradeableProxy(
  contractName: string,
  factoryAddress: string,
  initCallData?: string,
  constructorArgs?: Array<any>
): Promise<ProxyData> {
  return await run("deploy-upgradeable-proxy", {
    contractName,
    factoryAddress,
    initCallData,
    constructorArgs,
  });
}

export async function cliDeployCreate(
  contractName: string,
  factoryAddress: string,
  constructorArgs?: Array<string>
): Promise<DeployCreateData> {
  return await run("deploy-create", {
    contractName,
    factoryAddress,
    constructorArgs,
  });
}

export async function cliMultiCallUpgradeProxy(
  contractName: string,
  factoryAddress: BytesLike,
  initCallData?: BytesLike,
  salt?: BytesLike,
  constructorArgs?: Array<string>
): Promise<ProxyData> {
  return await run("multi-call-upgrade-proxy", {
    contractName,
    factoryAddress,
    initCallData,
    salt,
    constructorArgs,
  });
}

export async function cliDeployFactory(outputFolder?: string) {
  return await run("deploy-factory", {
    outputFolder,
  });
}

export async function cliDeployOnlyProxy(
  salt: string,
  factoryAddress: string
): Promise<ProxyData> {
  return await run("deploy-only-proxy", {
    salt,
    factoryAddress,
  });
}
