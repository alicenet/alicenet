import { expect } from "chai";
import { ethers } from "hardhat";
import { ALICENET_FACTORY } from "../../scripts/lib/constants";
import { DeploymentConfigWrapper } from "../../scripts/lib/deployment/interfaces";
import {
  extractNameFromFullyQualifiedName,
  readDeploymentConfig,
  readJSON,
} from "../../scripts/lib/deployment/utils";
import { AliceNetFactory } from "../../typechain-types";

interface DeployedConstractAddresses {
  [key: string]: string;
}
function extractContractConfig(
  contractName: string,
  deploymentConfig: DeploymentConfigWrapper
) {
  for (let fullyQualifiedName in deploymentConfig) {
    const name = extractNameFromFullyQualifiedName(fullyQualifiedName);
    if (name === contractName) {
      return deploymentConfig[fullyQualifiedName];
    }
  }
}

describe("contract post deployment tests", () => {
  let deploymentConfig: DeploymentConfigWrapper;
  let contractAddresses: DeployedConstractAddresses;
  before(async () => {
    deploymentConfig = await readDeploymentConfig(
      "../scripts/base-files/deploymentConfig.json"
    );
    contractAddresses = readJSON(
      "../scripts/base-files/deployedContracts.json"
    );
  });

  let AlicenetFactory: AliceNetFactory;
  it("checks if factory is by connecting to it", async () => {
    AlicenetFactory = await ethers.getContractAt(
      ALICENET_FACTORY,
      contractAddresses[ALICENET_FACTORY]
    );
    const expectedALCAAddress = contractAddresses["ALCA"];
    const alcaAddress = await AlicenetFactory.getALCAAddress();
    expect(alcaAddress).to.eq(expectedALCAAddress);
  });

  it("attempts to use owner functions without being owner", async () => {});
});
