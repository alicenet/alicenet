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
const bridgePoolTokenTypeERC20 = 0;
const bridgePoolTokenTypeERC721 = 1;
const bridgePoolTokenTypeERC1155 = 2;

const bridgePoolNativeChainId = 1337;
const bridgePoolExternalChainId = 9999;
const bridgePoolValue = 0;
const bridgePoolDeployCode = "0x38585839386009f3"; // UNIVERSAL_DEPLOY_CODE
interface BridgePoolFactoryFixture extends BaseTokensFixture {
  bridgePoolFactory: BridgePoolFactory;
}
let fixture: BridgePoolFactoryFixture;
let admin: SignerWithAddress;
const bridgePoolVersion = 1;
const unexistentBridgePoolVersion = 11;

describe("Testing BridgePool Factory", async () => {
  async function deployFixture() {
    await preFixtureSetup();
    const [admin] = await ethers.getSigners();

    const baseTokenFixture = await deployFactoryAndBaseTokens(admin);
    const bridgePoolFactory = (await deployUpgradeableWithFactory(
      baseTokenFixture.factory,
      "BridgePoolFactory",
      "BridgePoolFactory"
    )) as BridgePoolFactory;
    const fixture: BridgePoolFactoryFixture = {
      ...baseTokenFixture,
      bridgePoolFactory,
    };

    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        bridgePoolTokenTypeERC20,
        bridgePoolNativeChainId,
        bridgePoolValue,
        bridgePoolDeployCode,
      ]
    );
    return { fixture, admin };
  }

  beforeEach(async function () {
    ({ fixture, admin } = await loadFixture(deployFixture));
  });

  it("Should fail to enable public pool deploying without impersonating factory", async () => {
    await expect(
      fixture.bridgePoolFactory.togglePublicPoolDeployment()
    ).to.be.revertedWithCustomError(fixture.bridgePoolFactory, "OnlyFactory");
  });

  it("Should fail to deploy pool logic without impersonating factory", async () => {
    await expect(
      fixture.bridgePoolFactory.deployPoolLogic(
        bridgePoolTokenTypeERC20,
        bridgePoolNativeChainId,
        bridgePoolValue,
        bridgePoolDeployCode
      )
    ).to.be.revertedWithCustomError(fixture.bridgePoolFactory, "OnlyFactory");
  });

  it("Should deploy new BridgePool as factory even if public pool deployment is not enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewNativePool",
      [
        bridgePoolTokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion,
      ]
    );
    const salt = await fixture.bridgePoolFactory.getBridgePoolSalt(
      ethers.constants.AddressZero,
      0,
      1337,
      1
    );
    const expectedSalt = calculateBridgePoolSalt(
      ethers.constants.AddressZero,
      0,
      1337,
      1
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

  it("Should not deploy new BridgePool as user if public pool deployment is not enabled", async () => {
    await expect(
      fixture.bridgePoolFactory.deployNewNativePool(
        bridgePoolTokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion
      )
    ).to.be.revertedWithCustomError(
      fixture.bridgePoolFactory,
      "PublicPoolDeploymentTemporallyDisabled"
    );
  });

  it("Should deploy new BridgePool as factory if public pool deployment is enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "togglePublicPoolDeployment",
      []
    );
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewNativePool",
      [
        bridgePoolTokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion,
      ]
    );
  });

  it("Should deploy new BridgePool as user if public pool deployment is enabled", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "togglePublicPoolDeployment",
      []
    );
    await fixture.bridgePoolFactory.deployNewNativePool(
      bridgePoolTokenTypeERC20,
      ethers.constants.AddressZero,
      bridgePoolVersion
    );
  });

  it("Should not deploy two BridgePools with same ERC contract and version", async () => {
    await factoryCallAnyFixture(
      fixture,
      "bridgePoolFactory",
      "deployNewNativePool",
      [
        bridgePoolTokenTypeERC20,
        ethers.constants.AddressZero,
        bridgePoolVersion,
      ]
    );
    await expect(
      factoryCallAnyFixture(
        fixture,
        "bridgePoolFactory",
        "deployNewNativePool",
        [
          bridgePoolTokenTypeERC20,
          ethers.constants.AddressZero,
          bridgePoolVersion,
        ]
      )
    ).to.be.revertedWithCustomError(
      fixture.bridgePoolFactory,
      "StaticPoolDeploymentFailed"
    );
  });

  it("Should not deploy new BridgePool with inexistent version", async () => {
    await expect(
      factoryCallAnyFixture(
        fixture,
        "bridgePoolFactory",
        "deployNewNativePool",
        [
          bridgePoolTokenTypeERC20,
          ethers.constants.AddressZero,
          unexistentBridgePoolVersion,
        ]
      )
    )
      .to.be.revertedWithCustomError(
        fixture.bridgePoolFactory,
        "PoolVersionNotSupported"
      )
      .withArgs(unexistentBridgePoolVersion);
  });

  it("Should get correct Bridge Pool Salt", async () => {
    const [bridgePoolSalt] = await callFunctionAndGetReturnValues(
      fixture.bridgePoolFactory,
      "getBridgePoolSalt",
      admin,
      [
        ethers.constants.AddressZero,
        bridgePoolTokenTypeERC20,
        bridgePoolNativeChainId,
        bridgePoolVersion,
      ]
    );
    expect(bridgePoolSalt).to.be.eq(
      getBridgePoolSalt(
        ethers.constants.AddressZero,
        bridgePoolTokenTypeERC20,
        bridgePoolNativeChainId,
        bridgePoolVersion
      )
    );
  });

  it("Should get latest pool logic version for ERC20 native", async () => {
    const latestNativeERC20Version =
      await fixture.bridgePoolFactory.getLatestPoolLogicVersion(
        bridgePoolNativeChainId,
        0
      );
    expect(latestNativeERC20Version).to.eq(bridgePoolVersion);
  });

  it("Should get latest pool logic version for ERC20 external", async () => {
    const latestNativeERC20Version =
      await fixture.bridgePoolFactory.getLatestPoolLogicVersion(1, 0);
    expect(latestNativeERC20Version).to.eq(0);
  });

  it("Should failed to deploy with incorrect deploy code", async () => {
    await expect(
      factoryCallAny(
        fixture.factory,
        fixture.bridgePoolFactory,
        "deployPoolLogic",
        [
          bridgePoolTokenTypeERC20,
          bridgePoolNativeChainId,
          bridgePoolValue,
          "0x00",
        ]
      )
    ).to.be.revertedWithCustomError(
      fixture.bridgePoolFactory,
      "FailedToDeployLogic"
    );
  });

  it("Should deploy bridge for ERC20 external", async () => {
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        bridgePoolTokenTypeERC20,
        bridgePoolExternalChainId,
        bridgePoolValue,
        bridgePoolDeployCode,
      ]
    );
  });

  it("Should deploy bridge for ERC721 external", async () => {
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        bridgePoolTokenTypeERC721,
        bridgePoolExternalChainId,
        bridgePoolValue,
        bridgePoolDeployCode,
      ]
    );
  });

  it("Should deploy bridge for ERC71155 external", async () => {
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        bridgePoolTokenTypeERC1155,
        bridgePoolExternalChainId,
        bridgePoolValue,
        bridgePoolDeployCode,
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
