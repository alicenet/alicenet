import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BytesLike, ContractFactory } from "ethers";
import { ethers } from "hardhat";
import { MOCK_INITIALIZABLE } from "../../scripts/lib/constants";
import { BridgePoolFactory } from "../../typechain-types/contracts/BridgePoolFactory";
import { getEventVar } from "../factory/Setup";
import { getImpersonatedSigner } from "../lockup/setup";
import {
  BaseTokensFixture,
  deployUpgradeableWithFactory,
  Fixture,
  getFixture,
  preFixtureSetup,
} from "../setup";
// const nativeBridgePool = 0;
const poolTypeExternal = 1;
const tokenTypeERC20 = 0;
const bridgePoolExternalChainId = 9999;
let bridgePoolImplFactory: ContractFactory;
let bridgePoolFactory: Contract;
let fixture: Fixture;

interface BridgePoolFactoryFixture extends BaseTokensFixture {
  bridgePoolFactory: BridgePoolFactory;
}
let baseTokenFixture: BaseTokensFixture;
const bridgePoolVersion = 1;
let asFactory: SignerWithAddress;
const bridgePoolType = 0; // Native
const bridgePoolTokenType = 0; // ERC20
const bridgePoolNativeChainId = 1337;
const bridgePoolValue = 0;

describe("Testing BridgePool Factory - Arbitrary Deployments", async () => {
  async function deployFixture() {
    await preFixtureSetup();
    const [admin] = await ethers.getSigners();
    fixture = await getFixture(true, true, false);
    bridgePoolFactory = (await deployUpgradeableWithFactory(
      fixture.factory,
      "BridgePoolFactory",
      "BridgePoolFactory"
    )) as BridgePoolFactory;
    asFactory = await getImpersonatedSigner(fixture.factory.address);
    return { fixture };
  }

  beforeEach(async function () {
    ({ fixture } = await loadFixture(deployFixture));
  });

  it("should deploy arbitrary pool logic and then deploy a arbitrary pool and initialize", async () => {
    const mockInitBase = await ethers.getContractFactory(MOCK_INITIALIZABLE);
    const mockInitDeployTx = mockInitBase.getDeployTransaction();
    const initCallData = mockInitBase.interface.encodeFunctionData(
      "initialize",
      [fixture.legacyToken.address]
    );
    bridgePoolImplFactory = await ethers.getContractFactory(
      "NativeERC20BridgePoolV1"
    );
    await bridgePoolFactory
      .connect(asFactory)
      .deployPoolLogic(
        bridgePoolType,
        bridgePoolTokenType,
        bridgePoolVersion,
        mockInitDeployTx.data as BytesLike,
        bridgePoolValue
      );
    const txResponse = await bridgePoolFactory
      .connect(asFactory)
      .deployNewArbitraryPool(
        bridgePoolType,
        bridgePoolTokenType,
        fixture.legacyToken.address,
        bridgePoolVersion,
        bridgePoolExternalChainId,
        initCallData
      );
    const poolAddress = await getEventVar(
      txResponse,
      "BridgePoolCreated",
      "poolAddress"
    );
    const mockInit = await ethers.getContractAt(
      MOCK_INITIALIZABLE,
      poolAddress
    );
    expect(await mockInit.getImut()).to.equal(fixture.legacyToken.address);
  });
});
