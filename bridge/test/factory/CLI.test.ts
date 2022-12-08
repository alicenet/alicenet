import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { artifacts, ethers, run } from "hardhat";
import {
  MOCK,
  MOCK_INITIALIZABLE,
  MOCK_INITIALIZABLE_WITH_CONSTRUCTOR,
  MOCK_WITH_CONSTRUCTOR,
} from "../../scripts/lib/constants";
import {
  DeployCreateData,
  FactoryData,
  ProxyData,
} from "../../scripts/lib/deployment/interfaces";
import { getBytes32Salt } from "../../scripts/lib/deployment/utils";
import { AliceNetFactory } from "../../typechain-types";
import { expect } from "../chai-setup";
import { getMetamorphicAddress, predictFactoryAddress } from "./Setup";

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
        const proxyData: ProxyData = await cliDeployUpgradeableProxy(
          MOCK_INITIALIZABLE,
          factory.address,
          `${expectedInitVal}`
        );
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
        const proxyData: ProxyData = await cliDeployUpgradeableProxy(
          MOCK_WITH_CONSTRUCTOR,
          factory.address,
          undefined,
          `${expectedConstructorVal}` // constructor args
        );
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
          await cliDeployUpgradeableProxy(
            MOCK_INITIALIZABLE_WITH_CONSTRUCTOR,
            factory.address,
            undefined,
            `123` // constructor args,
          );
        } catch (error: any) {
          expect(error.message).to.equal(
            `initializerArgs must be specified for contract: ${MOCK_INITIALIZABLE_WITH_CONSTRUCTOR}`
          );
        }
      });

      it("fails to deploy proxy without constructor args when constructor args required", async () => {
        try {
          await cliDeployUpgradeableProxy(
            MOCK_INITIALIZABLE_WITH_CONSTRUCTOR,
            factory.address,
            "123" // initializer args
          );
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

    it("deploys factory with cli and checks if the default factory is updated in factory state toml file", async () => {
      const signers = await ethers.getSigners();
      const legacyToken = await ethers.deployContract("LegacyToken");
      const futureFactoryAddress = await predictFactoryAddress(
        signers[0].address
      );
      const factoryData: FactoryData = await cliDeployFactory(
        legacyToken.address
      );
      // check if the address is the predicted
      expect(factoryData.address).to.equal(futureFactoryAddress);
    });

    it("deploys mockInitializable with deployCreate, then deploy and upgrades a proxy with deploy-upgradeable-proxy", async () => {
      await cliDeployCreate(MOCK_INITIALIZABLE, factory.address);
      const salt = await getBytes32Salt(MOCK_INITIALIZABLE, artifacts, ethers);
      const expectedProxyAddress = getMetamorphicAddress(factory.address, salt);
      const proxyData = await cliDeployUpgradeableProxy(
        MOCK_INITIALIZABLE,
        factory.address,
        "1"
      );
      expect(proxyData.proxyAddress).to.equal(expectedProxyAddress);
    });

    xit("deploys all contracts in deploymentList", async () => {
      await cliDeployContracts();
    });
  });
});

async function cliDeployContracts(
  factoryAddress?: string,
  inputFolder?: string
) {
  return await run("deploy-contracts", {
    factoryAddress,
    inputFolder,
  });
}

async function cliDeployUpgradeableProxy(
  contractName: string,
  factoryAddress: string,
  initializerArgs?: string,
  constructorArgs?: string
): Promise<ProxyData> {
  return await run("deploy-upgradeable-proxy", {
    contractName,
    factoryAddress,
    initializerArgs,
    constructorArgs,
    waitConfirmation: 0,
  });
}

async function cliDeployCreate(
  contractName: string,
  factoryAddress: string,
  constructorArgs?: Array<string>
): Promise<DeployCreateData> {
  return await run("deploy-create", {
    contractName,
    factoryAddress,
    constructorArgs,
    waitConfirmation: 0,
  });
}

// async function cliUpgradeDeployedProxy(
//   contractName: string,
//   logicAddress: string,
//   factoryAddress: string,
//   initializerArgs?: string
// ): Promise<ProxyData> {
//   return await run("upgrade-proxy", {
//     contractName,
//     logicAddress,
//     factoryAddress,
//     initializerArgs,
//     waitConfirmation: 0,
//   });
// }

async function cliDeployFactory(
  legacyTokenAddress?: string,
  outputFolder?: string
) {
  return await run("deploy-factory", {
    legacyTokenAddress,
    outputFolder,
    waitConfirmation: 0,
  });
}

// async function cliDeployOnlyProxy(
//   salt: string,
//   factoryAddress: string
// ): Promise<ProxyData> {
//   return await run("deploy-only-proxy", {
//     salt,
//     factoryAddress,
//     waitConfirmation: 0,
//   });
// }
