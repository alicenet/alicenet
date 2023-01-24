import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { BridgePoolFactory } from "../../typechain-types/contracts/BridgePoolFactory";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  getImpersonatedSigner,
  preFixtureSetup,
} from "../setup";
const tokenTypeERC20 = 0;
const tokenTypeERC721 = 1;
const bridgePoolDeployCode = "0x38585839386009f3"; // UNIVERSAL_DEPLOY_CODE
interface BridgePoolFactoryFixture extends BaseTokensFixture {
  bridgePoolFactory: BridgePoolFactory;
}
let fixture: BridgePoolFactoryFixture;
let baseTokenFixture: BaseTokensFixture;
let asFactory: SignerWithAddress;
const bridgePoolVersion = 1;
const bridgePoolTokenType = 0; // ERC20
const bridgePoolType = 0; // Native
const bridgePoolChainId = 1337; // Native Chain for tests
const bridgePoolValue = 0;

async function deployFixture() {
  await preFixtureSetup();
  const [admin] = await ethers.getSigners();
  baseTokenFixture = await deployFactoryAndBaseTokens(admin);
  // deploy bridgePoolFactory with alicenet factory
  const bridgePoolFactory = (await deployUpgradeableWithFactory(
    baseTokenFixture.factory,
    "BridgePoolFactory",
    "BridgePoolFactory"
  )) as BridgePoolFactory;
  const nativeERCBridgePoolBaseErrors = await (
    await ethers.getContractFactory("BridgePoolFactoryErrors")
  ).deploy();
  fixture = {
    ...baseTokenFixture,
    bridgePoolFactory,
    nativeERCBridgePoolBaseErrors,
  };
  asFactory = await getImpersonatedSigner(fixture.factory.address);
  return { fixture, admin };
}

describe("Testing bridge pool factory getter functions", async () => {
  beforeEach(async function () {
    await loadFixture(deployFixture);
  });

  it("gets the alicenet factory address from the bridge pool factory", async () => {
    const alicenetFactoryAddress =
      await fixture.bridgePoolFactory.getRegistryAddress();
    expect(alicenetFactoryAddress).to.equal(fixture.factory.address);
  });

  it("gets the latest pool version for erc20", async () => {
    await fixture.bridgePoolFactory
      .connect(asFactory)
      .deployPoolLogic(
        bridgePoolType,
        bridgePoolTokenType,
        bridgePoolVersion,
        bridgePoolDeployCode,
        bridgePoolValue
      );
    const latestPoolVersion =
      await fixture.bridgePoolFactory.getLatestPoolLogicVersion(
        bridgePoolChainId,
        bridgePoolTokenType
      );

    expect(latestPoolVersion).to.equal(bridgePoolVersion);
  });

  it("attempt to get nonexistent erc20 version", async () => {
    await expect(
      fixture.bridgePoolFactory.getLatestPoolLogicVersion(
        bridgePoolChainId,
        tokenTypeERC20
      )
    )
      .to.be.revertedWithCustomError(
        fixture.nativeERCBridgePoolBaseErrors,
        "LogicVersionDoesNotExist"
      )
      .withArgs(bridgePoolType, tokenTypeERC20);
  });

  it("attempts to get nonexistent version pool with version ERC721", async () => {
    await fixture.bridgePoolFactory
      .connect(asFactory)
      .deployPoolLogic(
        bridgePoolType,
        bridgePoolTokenType,
        bridgePoolVersion,
        bridgePoolDeployCode,
        bridgePoolValue
      );
    await expect(
      fixture.bridgePoolFactory.getLatestPoolLogicVersion(
        bridgePoolChainId,
        tokenTypeERC721
      )
    )
      .to.be.revertedWithCustomError(
        fixture.nativeERCBridgePoolBaseErrors,
        "LogicVersionDoesNotExist"
      )
      .withArgs(bridgePoolType, tokenTypeERC721);
  });
});
