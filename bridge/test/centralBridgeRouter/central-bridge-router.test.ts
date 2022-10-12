import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber, Contract } from "ethers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { CentralBridgeRouter } from "../../typechain-types";
import { BridgeRouterV1Mock } from "../../typechain-types/test/contract-mocks/centralBridgeRouter/BridgeRouterV1Mock";
import { expect } from "../chai-setup";
import {
  deployCreate2WithFactory,
  deployFactoryAndBaseTokens,
  getImpersonatedSigner,
  preFixtureSetup,
} from "../setup";
let user: SignerWithAddress;
let asBToken: SignerWithAddress;
let asFactory: SignerWithAddress;
let centralBridgeRouter: CentralBridgeRouter;
let centralBridgeRouterErrors: Contract;
let bridgeRouter: BridgeRouterV1Mock;
// const fee = 1;
const number = 1;
const nonce = 0;
const poolVersion = 1;
const chainID = 1337;
const destinationAccountType = 1;
const ercType = 0;
const ercContract = ethers.constants.AddressZero;
const destinationAccount = ethers.constants.AddressZero;
let depositData: any;
let encodedDepositData: any;

describe("Testing Central Bridge Router", async () => {
  async function deployFixture() {
    await preFixtureSetup();
    const [admin, user] = await ethers.getSigners();
    const baseTokenFixture = await deployFactoryAndBaseTokens(admin);
    // CentralBridgeRouter
    centralBridgeRouter = (await deployCreate2WithFactory(
      baseTokenFixture.factory,
      "CentralBridgeRouter",
      []
    )) as CentralBridgeRouter;
    centralBridgeRouterErrors = await (
      await (
        await ethers.getContractFactory("CentralBridgeRouterErrors")
      ).deploy()
    ).deployed();
    bridgeRouter = (await deployCreate2WithFactory(
      baseTokenFixture.factory,
      "BridgeRouterV1Mock",
      [1]
    )) as BridgeRouterV1Mock;
    asBToken = await getImpersonatedSigner(baseTokenFixture.bToken.address);
    asFactory = await getImpersonatedSigner(baseTokenFixture.factory.address);

    depositData = [
      ercContract,
      destinationAccountType,
      destinationAccount,
      ercType,
      BigNumber.from(number),
      BigNumber.from(chainID),
      poolVersion,
    ];
    encodedDepositData = defaultAbiCoder.encode(
      ["address", "uint8", "address", "uint8", "uint256", "uint256", "uint16"],
      depositData
    );
    return { user };
  }

  beforeEach(async function () {
    ({ user } = await loadFixture(deployFixture));
  });

  describe("Testing Access Control", async () => {
    it("Should fail to add a router without impersonating factory", async () => {
      await expect(
        centralBridgeRouter.addRouter(bridgeRouter.address)
      ).to.be.revertedWithCustomError(centralBridgeRouter, "OnlyFactory");
    });

    it("Should fail to disable a router without impersonating factory", async () => {
      await expect(
        centralBridgeRouter.disableRouter(0)
      ).to.be.revertedWithCustomError(centralBridgeRouter, "OnlyFactory");
    });
  });

  describe("Testing Logic", async () => {
    it("Should be no router versions on initilization", async () => {
      expect(await centralBridgeRouter.getRouterCount()).to.be.eq(0);
    });

    it("Should add a router", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
    });

    it("Should increment router count when adding a router", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      expect(await centralBridgeRouter.getRouterCount()).to.be.equal(1);
    });

    it("Should get the router address", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      expect(await centralBridgeRouter.getRouterAddress(1)).to.be.equal(
        bridgeRouter.address
      );
    });

    it("Should not get the router address with an incorrect version", async () => {
      await expect(
        centralBridgeRouter.getRouterAddress(0)
      ).to.be.revertedWithCustomError(
        centralBridgeRouterErrors,
        "InvalidPoolVersion"
      );
    });

    it("Should disable router", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      await centralBridgeRouter.connect(asFactory).disableRouter(1);
    });

    it("Should not disable a router with an incorrect version", async () => {
      await expect(
        centralBridgeRouter.connect(asFactory).disableRouter(0)
      ).to.be.revertedWithCustomError(
        centralBridgeRouterErrors,
        "InvalidPoolVersion"
      );
    });

    it("Should not route call to a disabled router", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      await centralBridgeRouter.connect(asFactory).disableRouter(1);
      await expect(
        centralBridgeRouter
          .connect(asBToken)
          .forwardDeposit(user.address, poolVersion, encodedDepositData)
      ).to.be.revertedWithCustomError(
        centralBridgeRouterErrors,
        "DisabledPoolVersion"
      );
    });

    it("Should not route call to an incorrect version", async () => {
      await expect(
        centralBridgeRouter
          .connect(asBToken)
          .forwardDeposit(user.address, poolVersion, encodedDepositData)
      ).to.be.revertedWithCustomError(
        centralBridgeRouterErrors,
        "InvalidPoolVersion"
      );
    });

    it("Should route call and log deposit event with correspondent nonce", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      await expect(
        centralBridgeRouter
          .connect(asBToken)
          .forwardDeposit(user.address, poolVersion, encodedDepositData)
      )
        .to.emit(centralBridgeRouter, "DepositedERCToken")
        .withArgs(
          depositData[0],
          depositData[1],
          depositData[2],
          depositData[3],
          depositData[4],
          depositData[5],
          depositData[6],
          BigNumber.from(nonce)
        );
    });

    it("Should route call with two topics", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      depositData[4] = BigNumber.from(2); // BridgePool mock uses number element of deposit data as a parameter for generated topics array size
      encodedDepositData = defaultAbiCoder.encode(
        [
          "address",
          "uint8",
          "address",
          "uint8",
          "uint256",
          "uint256",
          "uint16",
        ],
        depositData
      );
      await expect(
        centralBridgeRouter
          .connect(asBToken)
          .forwardDeposit(user.address, poolVersion, encodedDepositData)
      )
        .to.emit(centralBridgeRouter, "DepositedERCToken")
        .withArgs(
          depositData[0],
          depositData[1],
          depositData[2],
          depositData[3],
          depositData[4],
          depositData[5],
          depositData[6],
          BigNumber.from(nonce)
        );
    });

    it("Should route call with three topics", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      depositData[4] = BigNumber.from(3); // BridgePool mock uses number element of deposit data as a parameter for generated topics array size
      encodedDepositData = defaultAbiCoder.encode(
        [
          "address",
          "uint8",
          "address",
          "uint8",
          "uint256",
          "uint256",
          "uint16",
        ],
        depositData
      );
      await expect(
        centralBridgeRouter
          .connect(asBToken)
          .forwardDeposit(user.address, poolVersion, encodedDepositData)
      )
        .to.emit(centralBridgeRouter, "DepositedERCToken")
        .withArgs(
          depositData[0],
          depositData[1],
          depositData[2],
          depositData[3],
          depositData[4],
          depositData[5],
          depositData[6],
          BigNumber.from(nonce)
        );
    });

    it("Should route call with four topics", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      depositData[4] = BigNumber.from(4); // BridgePool mock uses number element of deposit data as a parameter for generated topics array size
      encodedDepositData = defaultAbiCoder.encode(
        [
          "address",
          "uint8",
          "address",
          "uint8",
          "uint256",
          "uint256",
          "uint16",
        ],
        depositData
      );
      await expect(
        centralBridgeRouter
          .connect(asBToken)
          .forwardDeposit(user.address, poolVersion, encodedDepositData)
      )
        .to.emit(centralBridgeRouter, "DepositedERCToken")
        .withArgs(
          depositData[0],
          depositData[1],
          depositData[2],
          depositData[3],
          depositData[4],
          depositData[5],
          depositData[6],
          BigNumber.from(nonce)
        );
    });

    it("Should not route call with more than four topics", async () => {
      await centralBridgeRouter
        .connect(asFactory)
        .addRouter(bridgeRouter.address);
      depositData[4] = BigNumber.from(5); // BridgePool mock uses number element of deposit data as a parameter for generated topics array size to force an error
      encodedDepositData = defaultAbiCoder.encode(
        [
          "address",
          "uint8",
          "address",
          "uint8",
          "uint256",
          "uint256",
          "uint16",
        ],
        depositData
      );
      await expect(
        centralBridgeRouter
          .connect(asBToken)
          .forwardDeposit(user.address, poolVersion, encodedDepositData)
      ).to.be.revertedWithCustomError(
        centralBridgeRouter,
        "MissingEventSignature"
      );
    });
  });
});
