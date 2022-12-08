import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers/lib/ethers";
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
  Dynamics,
  ETHDKG,
  ETHDKGAccusations,
  ETHDKGPhases,
  Lockup,
  PublicStaking,
  Snapshots,
  StakingRouterV1,
  ValidatorPool,
  ValidatorStaking,
} from "../../typechain-types";
import { BonusPool } from "../../typechain-types/contracts/BonusPool";
import {} from "../../typechain-types/contracts/libraries/ethdkg/ETHDKGPhases";
import {} from "../../typechain-types/contracts/PublicStaking";
import { RewardPool } from "../../typechain-types/contracts/RewardPool";
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
  const PUBLIC_STAKING_NAME = "APSNFT";
  const PUBLIC_STAKING_SYMBOL = "APS";
  const ALCA_DECIMALS = 18;
  const EPOCH_LENGTH = 1024;
  const LOCKUP_PERIOD = BigNumber.from(2628000);
  const LOCKUP_ENROLLMENT_PERIOD = BigNumber.from(648000);
  const TOTAL_BONUS_AMOUNT = BigNumber.from(900000000000000000000000);
  const MIN_INTERVAL_BETWEEN_SNAPSHOTS = EPOCH_LENGTH / 4;
  let deploymentConfig: DeploymentConfigWrapper;
  let contractAddresses: DeployedConstractAddresses;
  let owner: SignerWithAddress;
  let alicenetFactory: AliceNetFactory;
  let alca: ALCA;
  let dynamics: Dynamics;
  let publicStaking: PublicStaking;
  let snapshots: Snapshots;
  let validatorPool: ValidatorPool;
  let validatorStaking: ValidatorStaking;
  let eTHDKGAccusations: ETHDKGAccusations;
  let ethDKGPhases: ETHDKGPhases;
  let ethdkg: ETHDKG;
  let lockup: Lockup;
  let stakingRouterV1: StakingRouterV1;
  let rewardPool: RewardPool;
  let bonusPool: BonusPool;

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
    publicStaking = await ethers.getContractAt(
      "PublicStaking",
      contractAddresses["PublicStaking"]
    );
    snapshots = await ethers.getContractAt(
      "Snapshots",
      contractAddresses["Snapshots"]
    );
    validatorPool = await ethers.getContractAt(
      "ValidatorPool",
      contractAddresses["ValidatorPool"]
    );
    validatorStaking = await ethers.getContractAt(
      "ValidatorStaking",
      contractAddresses["ValidatorStaking"]
    );
    eTHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      contractAddresses["ETHDKGAccusations"]
    );
    ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      contractAddresses["ETHDKGPhases"]
    );
    ethdkg = await ethers.getContractAt("ETHDKG", contractAddresses["ETHDKG"]);
    lockup = await ethers.getContractAt("Lockup", contractAddresses["Lockup"]);
    stakingRouterV1 = await ethers.getContractAt(
      "StakingRouterV1",
      contractAddresses["StakingRouterV1"]
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

  describe("Post Snapshots deployment tests", () => {
    it("attempts to initialize again as factory", async () => {
      const factory = await alicenetFactory.connect(owner);
      const initialize = snapshots.interface.encodeFunctionData(
        "initialize",
        [1, 1]
      );
      await expect(
        factory.callAny(snapshots.address, 0, initialize)
      ).to.be.revertedWith("Initializable: contract is already initialized");
    });

    it("attempts to initialize with external account", async () => {
      const signers = await ethers.getSigners();
      await expect(snapshots.initialize(1, 1))
        .to.be.revertedWithCustomError(snapshots, "OnlyFactory")
        .withArgs(signers[0].address, alicenetFactory.address);
    });

    it("checks if snapshot desperation delay is set", async () => {
      let snapshotConfig: any = extractContractConfig(
        "Snapshots",
        deploymentConfig
      );
      const desperationDelay = await snapshots.getSnapshotDesperationDelay();
      expect(snapshotConfig["initializerArgs"]["desperationDelay_"]).to.not.be
        .undefined;
      expect(desperationDelay.toNumber()).to.equal(
        snapshotConfig["initializerArgs"]["desperationDelay_"]
      );
    });

    it("checks if snapshot desperation factor is set", async () => {
      let snapshotConfig: any = extractContractConfig(
        "Snapshots",
        deploymentConfig
      );
      const desperationFactor = await snapshots.getSnapshotDesperationFactor();
      expect(snapshotConfig["initializerArgs"]["desperationDelay_"]).to.not.be
        .undefined;
      expect(desperationFactor.toNumber()).to.equal(
        snapshotConfig["initializerArgs"]["desperationDelay_"]
      );
    });

    it("checks if chainID is set to 21", async () => {
      const chainID = await snapshots.getChainId();
      expect(chainID).to.equal(21);
    });

    it("checks if epoch length is set to 1024", async () => {
      const epochLength = await snapshots.getEpochLength();
      expect(epochLength).to.equal(1024);
    });

    it("checks if minimum interval between snapshots is set to 1/4 of epoch length", async () => {
      const miniInterval = await snapshots.getMinimumIntervalBetweenSnapshots();
      expect(miniInterval.toNumber()).to.eq(MIN_INTERVAL_BETWEEN_SNAPSHOTS);
    });
  });

  describe("Post PublicStaking deployment tests", () => {
    it("attempts to initialize again as factory", async () => {
      const factory = alicenetFactory.connect(owner);
      const initialize =
        publicStaking.interface.encodeFunctionData("initialize");
      await expect(
        factory.callAny(publicStaking.address, 0, initialize)
      ).to.be.revertedWith("Initializable: contract is already initialized");
    });

    it("attempts to initialize with external account", async () => {
      const signers = await ethers.getSigners();
      await expect(publicStaking.initialize())
        .to.be.revertedWithCustomError(publicStaking, "OnlyFactory")
        .withArgs(signers[0].address, alicenetFactory.address);
    });
    it("checks if public staking nft name is set to APSNFT", async () => {
      const name = await publicStaking.name();
      expect(name).to.equal(PUBLIC_STAKING_NAME);
    });
    it("checks if public staking nft symbol is set to APS", async () => {
      const symbol = await publicStaking.symbol();
      expect(symbol).to.equal(PUBLIC_STAKING_SYMBOL);
    });
  });

  describe("Post ValidatorPool deployment tests", () => {
    it("attempts to initialize again as factory", async () => {
      const factory = alicenetFactory.connect(owner);
      const initialize = validatorPool.interface.encodeFunctionData(
        "initialize",
        [1, 1, 1, 1]
      );
      await expect(
        factory.callAny(validatorPool.address, 0, initialize)
      ).to.be.revertedWith("Initializable: contract is already initialized");
    });
    it("attempts to initialize with external account", async () => {
      const signers = await ethers.getSigners();
      await expect(validatorPool.initialize(1, 1, 1, 1))
        .to.be.revertedWithCustomError(validatorPool, "OnlyFactory")
        .withArgs(signers[0].address, alicenetFactory.address);
    });
    it("attempts to use owner functions without being owner", async () => {});
    it("attempts to use owner functions without being owner", async () => {});
  });

  describe("Post ValidatorStaking deployment tests", () => {
    it("attempts to initialize again as factory", async () => {
      const factory = alicenetFactory.connect(owner);
      const initialize =
        validatorStaking.interface.encodeFunctionData("initialize");
      await expect(
        factory.callAny(validatorStaking.address, 0, initialize)
      ).to.be.revertedWith("Initializable: contract is already initialized");
    });
    it("attempts to initialize with external account", async () => {
      const signers = await ethers.getSigners();
      await expect(validatorStaking.initialize())
        .to.be.revertedWithCustomError(validatorStaking, "OnlyFactory")
        .withArgs(signers[0].address, alicenetFactory.address);
    });
    it("attempts to use owner functions without being owner", async () => {});
    it("attempts to use owner functions without being owner", async () => {});
  });

  describe("Post ETHDKG deployment tests", () => {
    it("attempts to initialize again as factory", async () => {
      const factory = alicenetFactory.connect(owner);
      const initialize = ethdkg.interface.encodeFunctionData(
        "initialize",
        [1, 1]
      );
      await expect(
        factory.callAny(ethdkg.address, 0, initialize)
      ).to.be.revertedWith("Initializable: contract is already initialized");
    });
    it("attempts to initialize with external account", async () => {
      const signers = await ethers.getSigners();
      await expect(ethdkg.initialize(1, 1))
        .to.be.revertedWithCustomError(ethdkg, "OnlyFactory")
        .withArgs(signers[0].address, alicenetFactory.address);
    });
    it("checks if the confirmation length is 6", async () => {
      const confirmationLength = await ethdkg.getConfirmationLength();
      expect(confirmationLength.toNumber()).to.equal(6);
    });
    it("checks if the confirmation length is 6", async () => {
      const confirmationLength = await ethdkg.getConfirmationLength();
      expect(confirmationLength.toNumber()).to.equal(6);
    });
  });

  describe("Post Lockup deployment tests", () => {
    it("checks if start block and endblock is 1 year", async () => {
      const startBlock = await lockup.getLockupStartBlock();
      const endBlock = await lockup.getLockupEndBlock();
      const lockingPeriod = endBlock.sub(startBlock);
      expect(endBlock.sub(startBlock)).to.equal(LOCKUP_PERIOD);
      expect(lockingPeriod).to.equal(LOCKUP_PERIOD);
    });
  });

  describe("Post RewardPool deployment tests", () => {
    it("confirms rewardPool has the correct lockup contract address", async () => {
      const rewardPoolAddress = await lockup.getRewardPoolAddress();
      rewardPool = await ethers.getContractAt("RewardPool", rewardPoolAddress);
      const lockupAddress = await rewardPool.getLockupContractAddress();
      expect(lockupAddress).to.equal(lockup.address);
    });
  });

  describe("Post BonusPool deployment tests", () => {
    it("checks if it has the correct reward pool address", async () => {
      const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
      bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);
      const rewardPoolAddress = await bonusPool.getRewardPoolAddress();
      expect(rewardPoolAddress).to.equal(rewardPool.address);
    });
    it("checks if total bonus amount is correct", async () => {
      const totalBonusAmount = await bonusPool.getTotalBonusAmount();
      expect(totalBonusAmount).to.equal(TOTAL_BONUS_AMOUNT);
    });
    it("checks if it has the correct lockup address", async () => {
      const lockupAddress = await bonusPool.getLockupContractAddress();
      expect(lockupAddress).to.equal(lockup.address);
    });
    it("checks if bonus pool received 900000000000000000000000 alca", async () => {
      const balance = await alca.balanceOf(bonusPool.address);
      expect(balance).to.equal(TOTAL_BONUS_AMOUNT);
    });
  });

  describe("Post StakingRouterV1 deployment tests", () => {
    it("check if legacy token is set correctly", async () => {
      const legacyTokenAddress = await stakingRouterV1.getLegacyTokenAddress();
      expect(legacyTokenAddress).to.equal(legacyTokenAddress);
    });
  });
});
