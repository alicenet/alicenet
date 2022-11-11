import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BytesLike } from "ethers";
import { ethers, run } from "hardhat";
import {
  MOCK,
  MOCK_INITIALIZABLE,
  MOCK_INITIALIZABLE_WITH_CONSTRUCTOR,
  MOCK_WITH_CONSTRUCTOR,
} from "../../scripts/lib/constants";
import {
  DeployCreateData,
  FactoryData,
  MetaContractData,
  ProxyData,
} from "../../scripts/lib/deployment/interfaces";
import { AliceNetFactory } from "../../typechain-types";
import { expect } from "../chai-setup";
import { getMetamorphicAddress } from "./Setup";

async function deployFixture() {
  const accounts = await ethers.getSigners();
  const factoryData: FactoryData = await cliDeployFactory(accounts[1].address);
  // check if the address is the predicted
  return await ethers.getContractAt("AliceNetFactory", factoryData.address);
}

describe("Cli tasks", () => {
  let factory: AliceNetFactory;
  describe("factory deployment error scenarios", () => {
    it("should fail if legacyTokenAddress param is not an address", async () => {
      try {
        await cliDeployFactory("large inflateable banana");
      } catch (error: any) {
        expect(error.message).to.equal("legacyTokenAddress is not an address");
      }
    });
  });
  describe("with successfull factory deployment ", () => {
    beforeEach(async () => {
      process.env.test = "true";
      process.env.silencer = "true";
      factory = await loadFixture(deployFixture);
    });

    describe("deploy-upgradeable-proxy", () => {
      it("deploys a mock contract with a initializer function with npx hardhat, and checks if initializer arg is set correctly", async () => {
        const expectedInitVal = 14;
        const proxyData: ProxyData = await run("deploy-upgradeable-proxy", {
          contractName: MOCK_INITIALIZABLE,
          factoryAddress: factory.address,
          initializerArgs: `${expectedInitVal}`,
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

      it("deploys a mock contract with a constructor function with npx hardhat, and checks if initializer arg is set correctly", async () => {
        const expectedConstructorVal = 14;
        const proxyData: ProxyData = await run("deploy-upgradeable-proxy", {
          contractName: MOCK_WITH_CONSTRUCTOR,
          factoryAddress: factory.address,
          constructorArgs: `${expectedConstructorVal}`,
        });
        const mockWithConstructor = await ethers.getContractAt(
          MOCK_WITH_CONSTRUCTOR,
          proxyData.proxyAddress
        );
        const initval = await mockWithConstructor.constructorValue();
        const expectedProxyAddress = getMetamorphicAddress(
          factory.address,
          ethers.utils.formatBytes32String(MOCK_WITH_CONSTRUCTOR)
        );
        expect(proxyData.proxyAddress).to.equal(expectedProxyAddress);
        expect(initval).to.equal(expectedConstructorVal);
      });

      it("fails to deploy proxy without initializer args when initializer args required", async () => {
        try {
          await run("deploy-upgradeable-proxy", {
            contractName: MOCK_INITIALIZABLE_WITH_CONSTRUCTOR,
            factoryAddress: factory.address,
            constructorArgs: `123`,
          });
        } catch (error: any) {
          expect(error.message).to.equal(
            `initializerArgs must be specified for contract: ${MOCK_INITIALIZABLE_WITH_CONSTRUCTOR}`
          );
        }
      });

      it("fails to deploy proxy without constructor args when constructor args required", async () => {
        try {
          await run("deploy-upgradeable-proxy", {
            contractName: MOCK_INITIALIZABLE_WITH_CONSTRUCTOR,
            factoryAddress: factory.address,
            initializerArgs: "123",
          });
        } catch (error: any) {
          expect(error.message).to.equal(
            `constructorArgs must be specified for contract: ${MOCK_INITIALIZABLE_WITH_CONSTRUCTOR}`
          );
        }
      });

      it("deploys MockInitializable with deploy create, deploys proxy, then upgrades proxy to point to MockInitializable with initializerArgs", async () => {
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

    it("deploys mock with deployCreate", async () => {
      const deployCreateData = await cliDeployCreate(MOCK, factory.address, [
        "2",
        "s",
      ]);
      expect(deployCreateData.address).to.not.equal(
        ethers.constants.AddressZero
      );
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
  return await run("deploy-contracts", {
    factoryAddress,
    inputFolder,
  });
}

export async function cliFullMultiCallDeployProxy(
  contractName: string,
  factoryAddress: string,
  initializerArgs?: string,
  outputFolder?: string,
  constructorArgs?: Array<string>
): Promise<ProxyData> {
  return await run("full-multi-call-deploy-proxy", {
    contractName,
    factoryAddress,
    initializerArgs,
    outputFolder,
    constructorArgs,
  });
}

export async function cliMultiCallDeployMetamorphic(
  contractName: string,
  factoryAddress: string,
  initializerArgs?: string,
  outputFolder?: string,
  constructorArgs?: Array<string>
): Promise<MetaContractData> {
  return await run("multi-call-deploy-metamorphic", {
    contractName,
    factoryAddress,
    initializerArgs,
    outputFolder,
    constructorArgs,
  });
}

export async function cliDeployUpgradeableProxy(
  contractName: string,
  factoryAddress: string,
  initializerArgs?: string,
  constructorArgs?: Array<any>
): Promise<ProxyData> {
  return await run("deploy-upgradeable-proxy", {
    contractName,
    factoryAddress,
    initializerArgs,
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
  initializerArgs?: BytesLike,
  salt?: BytesLike,
  constructorArgs?: Array<string>
): Promise<ProxyData> {
  return await run("multi-call-upgrade-proxy", {
    contractName,
    factoryAddress,
    initializerArgs,
    salt,
    constructorArgs,
  });
}

export async function cliDeployFactory(
  legacyTokenAddress?: string,
  outputFolder?: string
) {
  return await run("deploy-factory", {
    legacyTokenAddress,
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
