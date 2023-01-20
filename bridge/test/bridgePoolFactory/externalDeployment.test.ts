import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { BridgePoolFactory } from "../../typechain-types/contracts/BridgePoolFactory";
import { expect } from "../chai-setup";
import {
  BaseTokensFixture,
  callFunctionAndGetReturnValues,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  factoryCallAny,
  factoryCallAnyFixture,
  getBridgePoolSalt,
  preFixtureSetup,
} from "../setup";
// const nativeBridgePool = 0;
const externalBridgePool = 1;
const tokenTypeERC20 = 0;
const tokenTypeERC721 = 1;
const tokenTypeERC1155 = 2;
const bridgePoolNativeChainId = 1337;
const bridgePoolExternalChainId = 9999;
const bridgePoolDeployCode = "0x38585839386009f3"; // UNIVERSAL_DEPLOY_CODE
interface BridgePoolFactoryFixture extends BaseTokensFixture {
  bridgePoolFactory: BridgePoolFactory;
}
let fixture: BridgePoolFactoryFixture;
let baseTokenFixture: BaseTokensFixture;
let admin: SignerWithAddress;
const bridgePoolVersion = 1;
const unexistentBridgePoolVersion = 11;

describe("Testing BridgePool Factory", async () => {
  async function deployFixture() {
    await preFixtureSetup();
    const [admin] = await ethers.getSigners();
    baseTokenFixture = await deployFactoryAndBaseTokens(admin);
    const bridgePoolFactory = (await deployUpgradeableWithFactory(
      baseTokenFixture.factory,
      "BridgePoolFactory",
      "BridgePoolFactory"
    )) as BridgePoolFactory;
    fixture = {
      ...baseTokenFixture,
      bridgePoolFactory,
    };

    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        externalBridgePool,
        tokenTypeERC20,
        bridgePoolVersion,
        bridgePoolDeployCode,
      ]
    );
    return { fixture, admin };
  }

  beforeEach(async function () {
    ({ fixture, admin } = await loadFixture(deployFixture));
  });

  it("Should deploy new external bridge pool as factory even if public pool deployment is not enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewExternalPool",
      [
        tokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion,
        bridgePoolExternalChainId,
        "0x",
      ]
    );
    const salt = await fixture.bridgePoolFactory.getBridgePoolSalt(
      ethers.constants.AddressZero,
      tokenTypeERC20,
      bridgePoolNativeChainId,
      bridgePoolVersion
    );
    const expectedSalt = calculateBridgePoolSalt(
      ethers.constants.AddressZero,
      tokenTypeERC20,
      bridgePoolNativeChainId,
      bridgePoolVersion
    );
    expect(salt).to.eq(expectedSalt);
    const bridgePoolAddress =
      await fixture.bridgePoolFactory.lookupBridgePoolAddress(salt);
    const expectedAddress = calculateBridgePoolAddress(
      fixture.bridgePoolFactory.address,
      salt
    );
    expect(bridgePoolAddress).to.eq(expectedAddress);
  });

  it("Should not deploy new external bridge pool as user if public pool deployment is not enabled", async () => {
    await expect(
      fixture.bridgePoolFactory.deployNewExternalPool(
        tokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion,
        bridgePoolExternalChainId,
        "0x"
      )
    ).to.be.revertedWithCustomError(
      fixture.bridgePoolFactory,
      "PublicPoolDeploymentTemporallyDisabled"
    );
  });

  it("Should deploy new external bridge pool as factory if public pool deployment is enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "togglePublicPoolDeployment",
      []
    );
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewExternalPool",
      [
        tokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion,
        bridgePoolExternalChainId,
        "0x",
      ]
    );
  });

  it("Should deploy new external bridge pool as user if public pool deployment is enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "togglePublicPoolDeployment",
      []
    );
    await fixture.bridgePoolFactory.deployNewExternalPool(
      tokenTypeERC20,
      ethers.constants.AddressZero,
      bridgePoolVersion,
      bridgePoolExternalChainId,
      "0x"
    );
  });

  it("Should not deploy two external bridge pools with same ERC contract and version", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewExternalPool",
      [
        tokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion,
        bridgePoolExternalChainId,
        "0x",
      ]
    );
    await expect(
      factoryCallAnyFixture(
        fixture,
        "bridgePoolFactory",
        "deployNewExternalPool",
        [
          tokenTypeERC20,
          ethers.constants.AddressZero,
          bridgePoolVersion,
          bridgePoolExternalChainId,
          "0x",
        ]
      )
    ).to.be.revertedWithCustomError(
      fixture.bridgePoolFactory,
      "StaticPoolDeploymentFailed"
    );
  });

  it("Should not deploy new external bridge pool with inexistent version", async () => {
    await expect(
      factoryCallAnyFixture(
        fixture,
        "bridgePoolFactory",
        "deployNewExternalPool",
        [
          tokenTypeERC20,
          ethers.constants.AddressZero,
          unexistentBridgePoolVersion,
          bridgePoolExternalChainId,
          "0x",
        ]
      )
    ).to.be.revertedWithCustomError(
      fixture.bridgePoolFactory,
      "PoolLogicNotSupported"
    );
  });

  it("Should get correct Bridge Pool Salt", async () => {
    const [bridgePoolSalt] = await callFunctionAndGetReturnValues(
      fixture.bridgePoolFactory,
      "getBridgePoolSalt",
      admin,
      [
        ethers.constants.AddressZero,
        tokenTypeERC20,
        bridgePoolExternalChainId,
        bridgePoolVersion,
      ]
    );
    expect(bridgePoolSalt).to.be.eq(
      getBridgePoolSalt(
        ethers.constants.AddressZero,
        tokenTypeERC20,
        bridgePoolExternalChainId,
        bridgePoolVersion
      )
    );
  });

  it("Should get latest pool logic version for ERC20 external", async () => {
    const latestExternalERC20Version =
      await fixture.bridgePoolFactory.getLatestPoolLogicVersion(
        externalBridgePool,
        tokenTypeERC20
      );
    expect(latestExternalERC20Version).to.eq(bridgePoolVersion);
  });

  it("should fail to deploy new logic for the same type of pool and version", async () => {
    let txResponse = factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        externalBridgePool,
        tokenTypeERC20,
        bridgePoolVersion,
        bridgePoolDeployCode,
      ]
    );
    expect(txResponse)
      .to.be.revertedWithCustomError(
        fixture.bridgePoolFactory,
        "PoolLogicAlreadyDeployed"
      )
      .withArgs(externalBridgePool, tokenTypeERC20, bridgePoolVersion);
  });

  it("Should deploy bridge pool logic for ERC721 external", async () => {
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        externalBridgePool,
        tokenTypeERC721,
        bridgePoolVersion,
        bridgePoolDeployCode,
      ]
    );
  });

  it("Should deploy bridge pool logic for ERC71155 external", async () => {
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        externalBridgePool,
        tokenTypeERC1155,
        bridgePoolVersion,
        bridgePoolDeployCode,
      ]
    );
  });

  it("Should keep slots after Bridge Pool Factory upgrade", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "togglePublicPoolDeployment",
      []
    );
    // Deploy new Pool Factory supporting new token type ERC777
    const updatedBridgePoolFactory = await (
      await (
        await ethers.getContractFactory("BridgePoolFactoryERC777Mock")
      ).deploy()
    ).deployed();
    // Upgrade BridgePoolFactory proxy with the address of the new deployment
    await fixture.factory.upgradeProxy(
      ethers.utils.formatBytes32String("BridgePoolFactory"),
      updatedBridgePoolFactory.address,
      "0x"
    );
    expect(await updatedBridgePoolFactory.getNativePoolType()).to.be.equals(0);
    // Slot for ERC20 logic should still be there after upgrade
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewExternalPool",
      [
        tokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion,
        bridgePoolExternalChainId,
        "0x",
      ]
    );
  });
});

export const calculateBridgePoolSalt = (
  tokenContractAddress: string,
  tokenType: number,
  chainID: number,
  version: number
): string => {
  const addr: string = ethers.utils
    .keccak256(tokenContractAddress)
    .substring(2);
  const type: string = ethers.utils
    .keccak256(ethers.utils.hexlify(tokenType))
    .substring(2);
  let chainId: string = ethers.utils.hexZeroPad(
    ethers.utils.hexlify(chainID),
    32
  );
  chainId = ethers.utils.keccak256(chainId).substring(2);
  let versionhash: string = ethers.utils.hexZeroPad(
    ethers.utils.hexlify(version),
    2
  );
  versionhash = ethers.utils.keccak256(versionhash).substring(2);
  const preSalt: string = "0x" + addr + type + chainId + versionhash;
  return ethers.utils.keccak256(preSalt);
};

export const calculateBridgePoolAddress = (
  factoryAddress: string,
  salt: string
): string => {
  const initCode = "0x5880818283335afa3d82833e3d82f3";
  return ethers.utils.getCreate2Address(
    factoryAddress,
    salt,
    ethers.utils.keccak256(initCode)
  );
};
