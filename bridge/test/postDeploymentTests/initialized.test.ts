import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { ALICENET_FACTORY } from "../../scripts/lib/constants";
import { DeploymentConfigWrapper } from "../../scripts/lib/deployment/interfaces";
import {
  extractNameFromFullyQualifiedName,
  readDeploymentConfig,
  readJSON,
} from "../../scripts/lib/deployment/utils";
import {
  ALCA,
  AliceNetFactory,
  Distribution,
  Dynamics,
  ETHDKG,
  ETHDKGAccusations,
  ETHDKGPhases,
  Foundation,
  Governance,
  InvalidTxConsumptionAccusation,
  LiquidityProviderStaking,
  Lockup,
  MultipleProposalAccusation,
  PublicStaking,
  Snapshots,
  StakingPositionDescriptor,
  StakingRouterV1,
  ValidatorPool,
  ValidatorStaking,
} from "../../typechain-types";
import {} from "../../typechain-types/contracts/libraries/ethdkg/ETHDKGPhases";
import {} from "../../typechain-types/contracts/PublicStaking";
import {} from "../../typechain-types/contracts/StakingPositionDescriptor";
import {} from "../../typechain-types/contracts/ValidatorPool";
import {} from "../../typechain-types/contracts/ValidatorStaking";

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
  const LEGACY_TOKEN_ADDRESS = "0x5B09A0371C1DA44A8E24D36Bf5DEb1141a84d875";
  const ALCA_NAME = "AliceNet Staking Token";
  const ALCA_SYMBOL = "ALCA";
  const ALCA_DECIMALS = 18;
  let deploymentConfig: DeploymentConfigWrapper;
  let contractAddresses: DeployedConstractAddresses;
  let owner: SignerWithAddress;
  let alicenetFactory: AliceNetFactory;
  let alca: ALCA;
  let dynamics: Dynamics;
  let foundation: Foundation;
  let governance: Governance;
  let invalidTxConsumptionAccusation: InvalidTxConsumptionAccusation;
  let liquidityProviderStaking: LiquidityProviderStaking;
  let multipleProposalAccusation: MultipleProposalAccusation;
  let publicStaking: PublicStaking;
  let snapshots: Snapshots;
  let stakingPositionDescriptor: StakingPositionDescriptor;
  let validatorPool: ValidatorPool;
  let validatorStaking: ValidatorStaking;
  let distribution: Distribution;
  let eTHDKGAccusations: ETHDKGAccusations;
  let eTHDKGPhases: ETHDKGPhases;
  let eTHDKG: ETHDKG;
  let lockup: Lockup;
  let stakingRouterV1: StakingRouterV1;

  before(async () => {
    deploymentConfig = await readDeploymentConfig(
      "../scripts/base-files/deploymentConfig.json"
    );
    contractAddresses = readJSON(
      "../scripts/base-files/deployedContracts.json"
    );
    owner = await ethers.getImpersonatedSigner(
      "0xDAEEaEB740a2218C9a06C16c0Fe2C7f19Ac7b0cA"
    );
    alicenetFactory = await ethers.getContractAt(
      ALICENET_FACTORY,
      contractAddresses[ALICENET_FACTORY]
    );
    alca = await ethers.getContractAt("ALCA", contractAddresses["ALCA"]);
    dynamics = await ethers.getContractAt(
      "Dynamics",
      contractAddresses["Dynamics"]
    );
    foundation = await ethers.getContractAt(
      "Foundation",
      contractAddresses["Foundation"]
    );
    governance = await ethers.getContractAt(
      "Governance",
      contractAddresses["Governance"]
    );
    invalidTxConsumptionAccusation = await ethers.getContractAt(
      "InvalidTxConsumptionAccusation",
      contractAddresses["InvalidTxConsumptionAccusation"]
    );
  });
  describe("post factory deployment tests", () => {
    it("checks if factory is deployed by connecting to it and getting the ALCA address", async () => {
      const expectedALCAAddress = contractAddresses["ALCA"];
      const alcaAddress = await alicenetFactory.getALCAAddress();
      expect(alcaAddress).to.eq(expectedALCAAddress);
    });
    it("attempts to claim ownership of factory from none owner account", async () => {});

    it("gets proxy template address from factory", async () => {
      const proxyTemplateAddress =
        await alicenetFactory.callStatic.getImplementation();
      expect(proxyTemplateAddress).to.not.equal(ethers.constants.AddressZero);
    });

    it("checks if the owner is 0xDAEEaEB740a2218C9a06C16c0Fe2C7f19Ac7b0cA", async () => {
      const ownerAddress = await alicenetFactory.owner();
      expect(ownerAddress).to.eq(owner.address);
    });

    it("gets the ALCA Initialization hash", async () => {
      let alcaCreationCodeHash =
        await alicenetFactory.getALCACreationCodeHash();
      const alcaBase = await ethers.getContractFactory("ALCA");
      const deployTx = alcaBase.getDeployTransaction(LEGACY_TOKEN_ADDRESS);
      const expectedALCACreationCodeHash = ethers.utils.keccak256(
        deployTx.data!
      );
      expect(alcaCreationCodeHash).to.equal(expectedALCACreationCodeHash);
      const salt = ethers.utils.formatBytes32String("ALCA");
      const expectedALCAAddress = ethers.utils.getCreate2Address(
        alicenetFactory.address,
        expectedALCACreationCodeHash,
        salt
      );
      const lookupALCAAddress = await alicenetFactory.lookup(salt);
      const alcaAddress = await alicenetFactory.getALCAAddress();
      expect(alcaAddress).to.equal(lookupALCAAddress);
      expect(alcaAddress).to.equal(expectedALCAAddress);
      expect(alcaAddress).to.equal(contractAddresses["ALCA"]);
    });
  });

  it("checks alca legacy token address 0x5B09A0371C1DA44A8E24D36Bf5DEb1141a84d875", async () => {
    const legacyTokenAddress = await alca.getLegacyTokenAddress();
    expect(legacyTokenAddress).to.equal(LEGACY_TOKEN_ADDRESS);
  });

  it("checks alca symbol, name, and decimal", async () => {
    const symbol = await alca.symbol();
    const name = await alca.name();
    const decimals = await alca.decimals();
    expect(symbol).to.equal(ALCA_SYMBOL);
    expect(name).to.equal(ALCA_NAME);
    expect(decimals).to.equal(ALCA_DECIMALS);
  });
  it("checks if early stage migration is on", async () => {
    const isEarlyStageMigration = await alca.isEarlyStageMigration();
    expect(isEarlyStageMigration).to.be.true;
  });
  describe("Post dynamics deployment tests", () => {
    it("attempts to call initializer as non factory", async () => {
      const signers = await ethers.getSigners();
      await expect(dynamics.initialize())
        .to.be.revertedWithCustomError(dynamics, "OnlyFactory")
        .withArgs(signers[0].address, alicenetFactory.address);
    });
    it("attempts to initialize again after deployment", async () => {
      await expect(dynamics.initialize()).to.be.revertedWithCustomError(
        dynamics,
        "AlreadyInitialized"
      );
    });
    it("gets the latest dynamic values", async () => {
      const latestDynamicVals = await dynamics.getLatestDynamicValues();
    });
    it("gets the latest cononical version", async () => {
      const latestCononicalVersion = await dynamics.getLatestAliceNetVersion();
    });
  });

  describe("Post foundation deployment tests", () => {
    it("attempts to use owner functions without being owner", async () => {});
    it("attempts to use owner functions without being owner", async () => {});
    it("attempts to use owner functions without being owner", async () => {});
    it("attempts to use owner functions without being owner", async () => {});
  });

  describe("Post governance deployment tests", () => {
    it("attempts to use owner functions without being owner", async () => {});
    it("attempts to use owner functions without being owner", async () => {});
    it("attempts to use owner functions without being owner", async () => {});
    it("attempts to use owner functions without being owner", async () => {});
  });
});
