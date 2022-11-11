import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BytesLike } from "ethers";
import { artifacts, ethers, run } from "hardhat";
import {
  MOCK,
  MOCK_INITIALIZABLE,
  TASK_DEPLOY_CONTRACTS,
  TASK_DEPLOY_CREATE,
  TASK_DEPLOY_FACTORY,
  TASK_DEPLOY_PROXY,
  TASK_DEPLOY_UPGRADEABLE_PROXY,
  TASK_FULL_MULTI_CALL_DEPLOY_PROXY,
  TASK_MULTI_CALL_DEPLOY_METAMORPHIC,
  TASK_MULTI_CALL_DEPLOY_PROXY,
  TASK_MULTI_CALL_UPGRADE_PROXY,
  TASK_UPGRADE_DEPLOYED_PROXY,
} from "../../scripts/lib/constants";
import { getBytes32Salt } from "../../scripts/lib/deployment/deploymentUtil";
import {
  DeployCreateData,
  FactoryData,
  MetaContractData,
  ProxyData,
} from "../../scripts/lib/deployment/factoryStateUtil";
import { expect } from "../chai-setup";
import { BaseTokensFixture, getBaseTokensFixture } from "../setup";
import { getMetamorphicAddress, predictFactoryAddress } from "./Setup";

describe("Cli tasks", async () => {
  beforeEach(async () => {
    process.env.test = "true";
    process.env.silencer = "true";
  });

  it("deploys factory with cli and checks if the default factory is updated in factory state toml file", async () => {
    const signers = await ethers.getSigners();
    const legacyToken = await ethers.deployContract("LegacyToken");
    const futureFactoryAddress = await predictFactoryAddress(
      signers[0].address
    );
    const factoryData: FactoryData = await run(TASK_DEPLOY_FACTORY, {
      waitConfirmation: 0,
      constructorArgs: [legacyToken.address],
    });
    // check if the address is the predicted
    expect(factoryData.address).to.equal(futureFactoryAddress);
  });

  describe("factory functions", async () => {
    let fixture: BaseTokensFixture;
    beforeEach(async () => {
      fixture = await loadFixture(getBaseTokensFixture);
    });

    it("deploys MockInitializable contract with deployUpgradeableProxy", async () => {
      const proxyData: ProxyData = await run(TASK_DEPLOY_UPGRADEABLE_PROXY, {
        contractName: MOCK_INITIALIZABLE,
        factoryAddress: fixture.factory.address,
        initCallData: "14",
      });
      const expectedProxyAddress = getMetamorphicAddress(
        fixture.factory.address,
        ethers.utils.formatBytes32String(MOCK_INITIALIZABLE)
      );
      expect(proxyData.proxyAddress).to.equal(expectedProxyAddress);
    });

    it("deploys MockInitializable contract with deployCreate", async () => {
      const nonce = await ethers.provider.getTransactionCount(
        fixture.factory.address
      );
      const expectedAddress = ethers.utils.getContractAddress({
        from: fixture.factory.address,
        nonce,
      });
      const deployCreateData = await cliDeployCreate(
        MOCK_INITIALIZABLE,
        fixture.factory.address
      );
      expect(deployCreateData.address).to.equal(expectedAddress);
    });

    it("deploys MockInitializable with deploy create, deploys proxy, then upgrades proxy to point to MockInitializable with initCallData", async () => {
      const test = "1";
      const deployCreateData = await cliDeployCreate(
        MOCK_INITIALIZABLE,
        fixture.factory.address
      );
      const salt = await getBytes32Salt(MOCK_INITIALIZABLE, artifacts, ethers);
      const expectedProxyAddress = getMetamorphicAddress(
        fixture.factory.address,
        salt
      );
      const proxyData = await cliDeployProxy(salt, fixture.factory.address);
      expect(proxyData.proxyAddress).to.equal(expectedProxyAddress);
      const logicFactory = await ethers.getContractFactory(MOCK_INITIALIZABLE);
      const upgradedProxyData = await cliUpgradeDeployedProxy(
        MOCK_INITIALIZABLE,
        deployCreateData.address,
        fixture.factory.address,
        test
      );
      const mockContract = logicFactory.attach(upgradedProxyData.proxyAddress);
      const i = await mockContract.callStatic.getImut();
      expect(i.toNumber()).to.equal(parseInt(test, 10));
    });

    it("deploys mockInitializable with deployCreate, then deploy and upgrades a proxy with multiCallDeployProxy", async () => {
      const logicData = await cliDeployCreate(
        MOCK_INITIALIZABLE,
        fixture.factory.address
      );
      const salt = await getBytes32Salt(MOCK_INITIALIZABLE, artifacts, ethers);
      const expectedProxyAddress = getMetamorphicAddress(
        fixture.factory.address,
        salt
      );
      const proxyData = await cliMultiCallDeployProxy(
        MOCK_INITIALIZABLE,
        logicData.address,
        fixture.factory.address,
        "1"
      );
      expect(proxyData.proxyAddress).to.equal(expectedProxyAddress);
    });

    it("deploys mock with deployCreate", async () => {
      const deployCreateData = await cliDeployCreate(
        MOCK,
        fixture.factory.address,
        ["2", "s"]
      );
      expect(deployCreateData.address).to.not.equal(
        ethers.constants.AddressZero
      );
    });

    it("deploys MockBaseContract with fullMultiCallDeployProxy", async () => {
      const proxyData = await cliFullMultiCallDeployProxy(
        MOCK,
        fixture.factory.address,
        undefined,
        undefined,
        ["2", "s"]
      );
      const salt = await getBytes32Salt(MOCK, artifacts, ethers);
      const expectedProxyAddress = getMetamorphicAddress(
        fixture.factory.address,
        salt
      );
      expect(proxyData.proxyAddress).to.equal(expectedProxyAddress);
    });

    xit("deploys all contracts in deploymentList", async () => {
      await cliDeployContracts();
    });
  });
});

export async function cliDeployContracts(
  factoryAddress?: string,
  inputFolder?: string
) {
  return await run(TASK_DEPLOY_CONTRACTS, {
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
  return await run(TASK_FULL_MULTI_CALL_DEPLOY_PROXY, {
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
  return await run(TASK_MULTI_CALL_DEPLOY_METAMORPHIC, {
    contractName,
    factoryAddress,
    initCallData,
    outputFolder,
    constructorArgs,
    waitConfirmation: 0,
  });
}

export async function cliDeployUpgradeableProxy(
  contractName: string,
  factoryAddress: string,
  initCallData?: string,
  constructorArgs?: Array<string>
): Promise<ProxyData> {
  return await run(TASK_DEPLOY_UPGRADEABLE_PROXY, {
    contractName,
    factoryAddress,
    initCallData,
    constructorArgs,
    waitConfirmation: 0,
  });
}

export async function cliDeployCreate(
  contractName: string,
  factoryAddress: string,
  constructorArgs?: Array<string>
): Promise<DeployCreateData> {
  return await run(TASK_DEPLOY_CREATE, {
    contractName,
    factoryAddress,
    constructorArgs,
    waitConfirmation: 0,
  });
}

export async function cliUpgradeDeployedProxy(
  contractName: string,
  logicAddress: string,
  factoryAddress: string,
  initCallData?: string
): Promise<ProxyData> {
  return await run(TASK_UPGRADE_DEPLOYED_PROXY, {
    contractName,
    logicAddress,
    factoryAddress,
    initCallData,
    waitConfirmation: 0,
  });
}

export async function cliMultiCallDeployProxy(
  contractName: string,
  logicAddress: string,
  factoryAddress: string,
  initCallData?: string,
  salt?: string
): Promise<ProxyData> {
  return await run(TASK_MULTI_CALL_DEPLOY_PROXY, {
    contractName,
    logicAddress,
    factoryAddress,
    initCallData,
    salt,
  });
}

export async function cliMultiCallUpgradeProxy(
  contractName: string,
  factoryAddress: BytesLike,
  initCallData?: BytesLike,
  salt?: BytesLike,
  constructorArgs?: Array<string>
): Promise<ProxyData> {
  return await run(TASK_MULTI_CALL_UPGRADE_PROXY, {
    contractName,
    factoryAddress,
    initCallData,
    salt,
    constructorArgs,
    waitConfirmation: 0,
  });
}

export async function cliDeployFactory(
  constructorArgs: string[],
  outputFolder?: string
) {
  return await run(TASK_DEPLOY_FACTORY, {
    outputFolder,
    constructorArgs,
    waitConfirmation: 0,
  });
}

export async function cliDeployProxy(
  salt: string,
  factoryAddress: string
): Promise<ProxyData> {
  return await run(TASK_DEPLOY_PROXY, {
    salt,
    factoryAddress,
    waitConfirmation: 0,
  });
}
