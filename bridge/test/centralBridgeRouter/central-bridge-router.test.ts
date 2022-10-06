import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  factoryCallAnyFixture,
  getTopicData,
  preFixtureSetup,
} from "../setup";
interface CentralBridgeRouterFixture extends BaseTokensFixture {}
let fixture: CentralBridgeRouterFixture;
let user: SignerWithAddress;
const fee = 1;
const depositNumber = 1;
const nonce = 0;
const depositEventInterface = new ethers.utils.Interface([
  "event DepositedERCToken( address ercContract, uint8 destinationAccountType, address destinationAccount, uint8 ercType, uint256 number, uint256 chainID, uint16 poolVersion, uint256 nonce)",
]);
const depositEventTopic =
  depositEventInterface.getEventTopic("DepositedERCToken");

describe("Testing Central Bridge Router", async () => {
  async function deployFixture() {
    await preFixtureSetup();
    const [admin, user] = await ethers.getSigners();
    const baseTokenFixture = await deployFactoryAndBaseTokens(admin);
    const centralBridgeRouterErrors = await (
      await (
        await ethers.getContractFactory("CentralBridgeRouterErrors")
      ).deploy()
    ).deployed();
    const bridgeRouter = await (
      await (await ethers.getContractFactory("BridgeRouterV1Mock")).deploy(fee)
    ).deployed();
    const fixture: CentralBridgeRouterFixture = {
      ...baseTokenFixture,
      bridgeRouter,
      centralBridgeRouterErrors,
    };
    return { fixture, user };
  }

  beforeEach(async function () {
    ({ fixture, user } = await loadFixture(deployFixture));
  });

  describe("Testing Access Control", async () => {
    it("Should fail to add a router without impersonating factory", async () => {
      await expect(
        fixture.centralBridgeRouter.addRouter(fixture.bridgeRouter.address)
      ).to.be.revertedWithCustomError(
        fixture.centralBridgeRouter,
        "OnlyFactory"
      );
    });

    it("Should fail to disable a router without impersonating factory", async () => {
      await expect(
        fixture.centralBridgeRouter.disableRouter(0)
      ).to.be.revertedWithCustomError(
        fixture.centralBridgeRouter,
        "OnlyFactory"
      );
    });
  });

  describe("Testing Logic", async () => {
    it("Should add a router", async () => {
      await factoryCallAnyFixture(fixture, "centralBridgeRouter", "addRouter", [
        fixture.bridgeRouter.address,
      ]);
    });

    it("Should be no router versions on initilization", async () => {
      expect(await fixture.centralBridgeRouter.getRouterCount()).to.be.eq(0);
    });

    it("Should increment router count when adding a router", async () => {
      await factoryCallAnyFixture(fixture, "centralBridgeRouter", "addRouter", [
        fixture.bridgeRouter.address,
      ]);
      expect(await fixture.centralBridgeRouter.getRouterCount()).to.be.equal(1);
    });

    it("Should get the router address", async () => {
      await factoryCallAnyFixture(fixture, "centralBridgeRouter", "addRouter", [
        fixture.bridgeRouter.address,
      ]);
      expect(await fixture.centralBridgeRouter.getRouterAddress(1)).to.be.equal(
        fixture.bridgeRouter.address
      );
    });

    it("Should not get the router address with an incorrect version", async () => {
      await expect(
        fixture.centralBridgeRouter.getRouterAddress(0)
      ).to.be.revertedWithCustomError(
        fixture.centralBridgeRouterErrors,
        "InvalidPoolVersion"
      );
    });

    it("Should disable router", async () => {
      await factoryCallAnyFixture(fixture, "centralBridgeRouter", "addRouter", [
        fixture.bridgeRouter.address,
      ]);
      await factoryCallAnyFixture(
        fixture,
        "centralBridgeRouter",
        "disableRouter",
        [depositNumber]
      );
    });

    it("Should not disable a router with an incorrect version", async () => {
      await expect(
        factoryCallAnyFixture(fixture, "centralBridgeRouter", "disableRouter", [
          0,
        ])
      ).to.be.revertedWithCustomError(
        fixture.centralBridgeRouterErrors,
        "InvalidPoolVersion"
      );
    });

    it("Should not route call to a disable router", async () => {
      await factoryCallAnyFixture(fixture, "centralBridgeRouter", "addRouter", [
        fixture.bridgeRouter.address,
      ]);
      await factoryCallAnyFixture(
        fixture,
        "centralBridgeRouter",
        "disableRouter",
        [depositNumber]
      );
      await expect(
        fixture.bToken
          .connect(user)
          .depositTokensOnBridges(
            1,
            defaultAbiCoder.encode(["uint256"], [depositNumber])
          )
      ).to.be.revertedWithCustomError(
        fixture.centralBridgeRouterErrors,
        "DisabledPoolVersion"
      );
    });

    it("Should not route call to an incorrect version", async () => {
      await expect(
        fixture.bToken
          .connect(user)
          .depositTokensOnBridges(
            1,
            defaultAbiCoder.encode(["uint256"], [depositNumber])
          )
      ).to.be.revertedWithCustomError(
        fixture.centralBridgeRouterErrors,
        "InvalidPoolVersion"
      );
    });

    it("Should route call and log deposit", async () => {
      await factoryCallAnyFixture(fixture, "centralBridgeRouter", "addRouter", [
        fixture.bridgeRouter.address,
      ]);
      const tx = await fixture.bToken
        .connect(user)
        .depositTokensOnBridges(
          1,
          defaultAbiCoder.encode(["uint256"], [depositNumber]),
          {
            value: 1,
          }
        );
      expect(await getTopicData(tx, depositEventTopic)).to.be.equal(
        defaultAbiCoder.encode(["uint256", "uint256"], [depositNumber, nonce])
      );
    });

    it("Should not route call and log deposit whith in", async () => {
      await factoryCallAnyFixture(fixture, "centralBridgeRouter", "addRouter", [
        fixture.bridgeRouter.address,
      ]);
      await expect(
        fixture.bToken.connect(user).depositTokensOnBridges(
          1,
          defaultAbiCoder.encode(["uint256"], [5]), // mock uses depositNumber as size for the topics array
          {
            value: 1,
          }
        )
      ).to.be.revertedWithCustomError(
        fixture.centralBridgeRouterErrors,
        "MissingEventSignature"
      );
    });
  });
});
