import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BytesLike } from "ethers";
import { ethers } from "hardhat";
import { MOCK_INITIALIZABLE } from "../../scripts/lib/constants";
import { BridgePoolFactory } from "../../typechain-types/contracts/BridgePoolFactory";
import { getEventVar } from "../factory/Setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  factoryCallAny,
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
    return { fixture, admin };
  }

  beforeEach(async function () {
    ({ fixture, admin } = await loadFixture(deployFixture));
  });
  it("should deploy arbitrary pool logic and then deploy a arbitrary pool and initialize", async () => {
    const mockInitBase = await ethers.getContractFactory(MOCK_INITIALIZABLE);
    const mockInitDeployTx = mockInitBase.getDeployTransaction();
    const initCallData = mockInitBase.interface.encodeFunctionData(
      "initialize",
      [2]
    );

    await factoryCallAny(
      fixture.factory,
      fixture.bridgePoolFactory,
      "deployPoolLogic",
      [3, 4, bridgePoolVersion, mockInitDeployTx.data as BytesLike]
    );

    const deployNewArbitraryPool =
      fixture.bridgePoolFactory.interface.encodeFunctionData(
        "deployNewArbitraryPool",
        [
          3,
          4,
          fixture.legacyToken.address,
          bridgePoolVersion,
          bridgePoolExternalChainId,
          initCallData,
        ]
      );
    let txResponse = await fixture.factory.callAny(
      fixture.bridgePoolFactory.address,
      0,
      deployNewArbitraryPool
    );
    await txResponse.wait();

    const poolAddress = await getEventVar(
      txResponse,
      "BridgePoolCreated",
      "poolAddress"
    );
    console.log(poolAddress);
    const arbitraryPoolSalt = fixture.bridgePoolFactory.getBridgePoolSalt(
      fixture.legacyToken.address,
      4,
      bridgePoolExternalChainId,
      bridgePoolVersion
    );
    const bridgePoolAddress =
      await fixture.bridgePoolFactory.lookupBridgePoolAddress(
        arbitraryPoolSalt
      );
    const mockInit = await ethers.getContractAt(
      MOCK_INITIALIZABLE,
      bridgePoolAddress
    );

    expect(await mockInit.getImut()).to.equal(2);
  });
});
