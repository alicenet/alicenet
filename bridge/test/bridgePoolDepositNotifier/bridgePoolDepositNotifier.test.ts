import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  callFunctionAndGetReturnValues,
  factoryCallAny,
  Fixture,
  getContractAddressFromEventLog,
  getFixture,
} from "../setup";

let fixture: Fixture;
let firstOwner: SignerWithAddress;
let user: SignerWithAddress;
const networkId = 1337;
const ercContractAddress = ethers.constants.AddressZero;
const bridgePoolTag = ethers.utils.toUtf8Bytes("ERC");
const nonce = 1;
const ercAmount = 10000;
let bridgePool: any;
const bTokenFeeInETH = 10;
const totalErc20Amount = BigNumber.from(20000).toBigInt();
const erc20Amount = BigNumber.from(100).toBigInt();
const bTokenAmount = BigNumber.from(100).toBigInt();

describe("Testing BridgePoolDepositNotifier", async () => {
  beforeEach(async () => {
    fixture = await getFixture(true, true, false);
    [firstOwner, user] = await ethers.getSigners();
    await fixture.factory.setDelegator(fixture.bridgePoolFactory.address);
    const ethIn = ethers.utils.parseEther(bTokenFeeInETH.toString());
    const deployNewPoolTransaction =
      await fixture.bridgePoolFactory.deployNewPool(
        fixture.aToken.address,
        fixture.bToken.address,
        1
      );
    const eventSignature = "event BridgePoolCreated(address contractAddr)";
    const eventName = "BridgePoolCreated";
    const bridgePoolAddress = await getContractAddressFromEventLog(
      deployNewPoolTransaction,
      eventSignature,
      eventName
    );
    // Final bridgePool address
    bridgePool = (await ethers.getContractFactory("BridgePoolV1")).attach(
      bridgePoolAddress
    );
    // Mint and approve some ERC20 tokens to deposit
    await factoryCallAny(fixture.factory, fixture.aTokenMinter, "mint", [
      user.address,
      totalErc20Amount,
    ]);
    await fixture.aToken
      .connect(user)
      .approve(bridgePool.address, totalErc20Amount);
    // Mint and approve some bTokens to deposit (and burn)
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      firstOwner,
      [user.address, 0],
      ethIn
    );
    await fixture.bToken
      .connect(user)
      .approve(bridgePool.address, BigNumber.from(bTokenAmount));
  });

  it("Should emit an event if called from a BridgePool", async () => {
    const nonce = 1;
    await expect(
      bridgePool
        .connect(user)
        .deposit(1, user.address, erc20Amount, bTokenAmount)
    )
      .to.emit(fixture.bridgePoolDepositNotifier, "Deposited")
      .withArgs(
        BigNumber.from(nonce),
        fixture.aToken.address,
        user.address,
        BigNumber.from(erc20Amount),
        BigNumber.from(networkId)
      );
  });

  it("Should not emit an event if not called from a BridgePool", async () => {
    const salt = ethers.utils.keccak256("0x00");
    const reason = ethers.utils.parseBytes32String(
      await fixture.immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_BRIDGEPOOL()
    );
    await expect(
      fixture.bridgePoolDepositNotifier.doEmit(
        salt,
        ercContractAddress,
        ercAmount,
        user.address
      )
    ).to.be.rejectedWith(reason);
  });
});
