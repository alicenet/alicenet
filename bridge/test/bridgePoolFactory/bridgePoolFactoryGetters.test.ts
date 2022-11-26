import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { BridgePoolFactory } from "../../typechain-types/contracts/BridgePoolFactory";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  factoryCallAny,
  preFixtureSetup,
} from "../setup";
const nativeBridgePool = 0;
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

async function deployFixture() {
  await preFixtureSetup();
  const [admin] = await ethers.getSigners();
  baseTokenFixture = await deployFactoryAndBaseTokens(admin);
  //deploy bridgePoolFactory with alicenet factory
  const bridgePoolFactory = (await deployUpgradeableWithFactory(
    baseTokenFixture.factory,
    "BridgePoolFactory",
    "BridgePoolFactory"
  )) as BridgePoolFactory;
  fixture = {
    ...baseTokenFixture,
    bridgePoolFactory,
  };
  return { fixture, admin };
}

describe("Testing bridge pool factory getter functions", async () => {
  beforeEach(async function () {
    ({ fixture, admin } = await loadFixture(deployFixture));
  });

  it("gets the alicenet factory address from the bridge pool factory", async () => {
    const alicenetFactoryAddress =
      await fixture.bridgePoolFactory.getRegistryAddress();
    expect(alicenetFactoryAddress).to.equal(fixture.factory.address);
  });

  it("gets the latest pool version for erc20", async () => {
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        nativeBridgePool,
        tokenTypeERC20,
        bridgePoolVersion,
        bridgePoolDeployCode,
      ]
    );
    const latestPoolVersion =
      await fixture.bridgePoolFactory.getLatestPoolLogicVersion(
        nativeBridgePool,
        tokenTypeERC20
      );

    expect(latestPoolVersion).to.equal(bridgePoolVersion);
  });

  it("attempt to get nonexistent erc20 version", async () => {
    const latestPoolVersion =
      fixture.bridgePoolFactory.getLatestPoolLogicVersion(
        nativeBridgePool,
        tokenTypeERC20
      );
    await expect(latestPoolVersion)
      .to.be.revertedWithCustomError(
        fixture.bridgePoolFactory,
        "LogicVersionDoesNotExist"
      )
      .withArgs(0, 0);
  });

  it("attempts to get nonexistent version pool with version ERC721", async () => {
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        nativeBridgePool,
        tokenTypeERC20,
        bridgePoolVersion,
        bridgePoolDeployCode,
      ]
    );
    const latestPoolVersion =
      fixture.bridgePoolFactory.getLatestPoolLogicVersion(
        nativeBridgePool,
        tokenTypeERC721
      );

    await expect(latestPoolVersion)
      .to.be.revertedWithCustomError(
        fixture.bridgePoolFactory,
        "LogicVersionDoesNotExist"
      )
      .withArgs(0, 1);
  });
});
