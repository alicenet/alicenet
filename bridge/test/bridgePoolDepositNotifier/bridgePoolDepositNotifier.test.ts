import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";

let fixture: Fixture;
let admin: SignerWithAddress;
const networkId = 1337;
const ercContractAddress = ethers.constants.AddressZero;
const bridgePoolTag = ethers.utils.toUtf8Bytes("ERC");
const nonce = 1;
const ercAmount = 10000;

describe("Testing BridgePoolDepositNotifier", async () => {
  beforeEach(async () => {
    fixture = await getFixture(true, true, false);
    [admin] = await ethers.getSigners();
  });

  it("Should emit an event if called with correct salt", async () => {
    const salt = ethers.utils.keccak256(
      ethers.utils.hexConcat([
        ethers.utils.keccak256(ercContractAddress),
        ethers.utils.keccak256(bridgePoolTag),
        ethers.utils.keccak256(
          ethers.utils.hexZeroPad(ethers.utils.hexlify(networkId), 32)
        ),
      ])
    );
    await expect(
      fixture.bridgePoolDepositNotifier.doEmit(
        salt,
        ercContractAddress,
        ercAmount,
        admin.address
      )
    )
      .to.emit(fixture.bridgePoolDepositNotifier, "Deposited")
      .withArgs(
        BigNumber.from(nonce),
        ercContractAddress,
        admin.address,
        BigNumber.from(ercAmount),
        BigNumber.from(networkId)
      );
  });

  it("Should not emit an event if called with incorrect salt", async () => {
    const salt = ethers.utils.keccak256("0x00");
    const reason = ethers.utils.parseBytes32String(
      await fixture.immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_BRIDGEPOOL()
    );
    await expect(
      fixture.bridgePoolDepositNotifier.doEmit(
        salt,
        ercContractAddress,
        ercAmount,
        admin.address
      )
    ).to.be.rejectedWith(reason);
  });
});
