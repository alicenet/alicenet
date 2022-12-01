import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BytesLike } from "ethers";
import { ethers } from "hardhat";
import { MOCK_INITIALIZABLE } from "../../scripts/lib/constants";
import { BridgePoolFactory } from "../../typechain-types/contracts/BridgePoolFactory";
import { getEventVar } from "../factory/Setup";
import { getImpersonatedSigner } from "../lockup/setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  factoryCallAny,
  preFixtureSetup,
} from "../setup";
// const nativeBridgePool = 0;
const poolTypeExternal = 1;
const tokenTypeERC20 = 0;
const bridgePoolExternalChainId = 9999;
interface BridgePoolFactoryFixture extends BaseTokensFixture {
  bridgePoolFactory: BridgePoolFactory;
}
let fixture: BridgePoolFactoryFixture;
let baseTokenFixture: BaseTokensFixture;
const bridgePoolVersion = 1;
let asFactory: SignerWithAddress;

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
    asFactory = await getImpersonatedSigner(baseTokenFixture.factory.address);
    fixture = {
      ...baseTokenFixture,
      bridgePoolFactory,
    };
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
    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [
        poolTypeExternal,
        tokenTypeERC20,
        bridgePoolVersion,
        mockInitDeployTx.data as BytesLike,
      ]
    );
    const txResponse = await fixture.bridgePoolFactory
      .connect(asFactory)
      .deployNewArbitraryPool(
        poolTypeExternal,
        tokenTypeERC20,
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
