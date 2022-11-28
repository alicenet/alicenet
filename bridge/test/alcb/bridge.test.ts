import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  BaseTokensFixture,
  callFunctionAndGetReturnValues,
  deployUpgradeableWithFactory,
  Fixture,
  getFixture,
} from "../setup";
import {
  getBridgeRouterSalt,
  getEthConsumedAsGas,
  getState,
  showState,
  state,
} from "./setup";

const depositAndCheckRefundEth = async (
  fixture: Fixture | BaseTokensFixture,
  refund: number,
  user: SignerWithAddress,
  _poolVersion: number,
  encodedDepositCallData: string
) => {
  await depositAndCheckRefundWei(
    fixture,
    ethers.utils.parseEther(refund.toString()).toBigInt(),
    user,
    _poolVersion,
    encodedDepositCallData
  );
};

const depositAndCheckRefundWei = async (
  fixture: Fixture | BaseTokensFixture,
  refund: bigint,
  user: SignerWithAddress,
  _poolVersion: number,
  encodedDepositCallData: string
) => {
  const minEthFeeForDeposit = 8n; // Curve value for the BridgeRouterMok returned fee
  const expectedState = await getState(fixture);
  const ethFeeForDeposit = minEthFeeForDeposit + refund;
  const tx = await fixture.alcb
    .connect(user)
    .depositTokensOnBridges(_poolVersion, encodedDepositCallData, {
      value: ethFeeForDeposit,
    });
  expectedState.Balances.eth.alcb += BigInt(minEthFeeForDeposit);
  const balanceBefore = expectedState.Balances.eth.user;
  expectedState.Balances.eth.user = (
    await ethers.provider.getBalance(user.address)
  ).toBigInt();
  expect(await getState(fixture)).to.be.deep.equal(
    expectedState,
    `state check failed for refund ${refund}`
  );
  // ethereum balance has to be compared outside since there's round errors due to
  // the integer math in the bounding curve when computing the eth required to
  // mint a ALCB amount
  expectedState.Balances.eth.user = balanceBefore;
  expectedState.Balances.eth.user -=
    getEthConsumedAsGas(await tx.wait()) + minEthFeeForDeposit;
  let roundingErrorFactor = 1n;
  // if refund is to low, less than 100Wei, divide by 10 to account for rounding errors
  if (refund < 100n) {
    roundingErrorFactor = 10n;
  }

  expect(
    (await ethers.provider.getBalance(user.address)).toBigInt() /
      roundingErrorFactor
  ).to.be.equal(
    expectedState.Balances.eth.user / roundingErrorFactor,
    `balance check failed for refund ${refund}`
  );
};

describe("Testing ALCB bridge methods", async () => {
  let user: SignerWithAddress;
  let admin: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 40;

  const minALCBs = 0;
  let ethsFromBurning: BigNumber;

  let encodedDepositCallData: string;
  const valueOrId = 100;
  const _tokenType = 1; // ERC20
  const chainId = 1337;
  const _poolVersion = 1;
  const alcbFee = 1000; // Fee that's returned by BridgeRouterMok
  const minEthFeeForDeposit = 8; // Curve value for the BridgeRouterMok returned fee

  async function deployFixture() {
    const fixture = await getFixture();
    const signers = await ethers.getSigners();
    const [admin, user] = signers;
    const ethForMinting = ethers.utils.parseEther(eth.toString());
    const [alcbs] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      user,
      [minALCBs],
      ethForMinting
    );
    const ethsFromBurning = await fixture.alcb.getLatestEthFromALCBsBurn(alcbs);
    const depositCallData = {
      ERCContract: ethers.constants.AddressZero,
      tokenType: _tokenType,
      number: valueOrId,
      chainID: chainId,
      poolVersion: _poolVersion,
    };
    const encodedDepositCallData = ethers.utils.defaultAbiCoder.encode(
      [
        "tuple(address ERCContract, uint8 tokenType, uint256 number, uint256 chainID, uint16 poolVersion)",
      ],
      [depositCallData]
    );
    await deployUpgradeableWithFactory(
      fixture.factory,
      "BridgeRouterMock",
      getBridgeRouterSalt(1),
      undefined,
      [1000]
    );
    return {
      fixture,
      user,
      admin,
      ethForMinting,
      alcbs,
      ethsFromBurning,
      depositCallData,
      encodedDepositCallData,
    };
  }

  beforeEach(async function () {
    ({ fixture, user, admin, ethsFromBurning, encodedDepositCallData } =
      await loadFixture(deployFixture));

    showState("Initial", await getState(fixture));
  });

  it("Should deposit tokens into the bridge and destroy the correspondent ALCB fee if no eth fee is sent", async () => {
    expectedState = await getState(fixture);
    ethsFromBurning = await fixture.alcb.getLatestEthFromALCBsBurn(alcbFee);
    const tx = await fixture.alcb
      .connect(user)
      .depositTokensOnBridges(_poolVersion, encodedDepositCallData);
    expectedState.Balances.alcb.user -= BigInt(alcbFee);
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.alcb.totalSupply -= BigInt(alcbFee);
    expectedState.Balances.alcb.poolBalance -= ethsFromBurning.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should not deposit with insufficient alcb", async () => {
    await fixture.alcb.connect(admin).mint(0, { value: 10000000n });
    // burn all tokens before depositing
    await fixture.alcb
      .connect(user)
      .burn((await fixture.alcb.balanceOf(user.address)).sub(alcbFee - 1), 0);
    await expect(
      fixture.alcb
        .connect(user)
        .depositTokensOnBridges(_poolVersion, encodedDepositCallData)
    ).to.be.revertedWith("ERC20: burn amount exceeds balance");
  });

  it("Should not deposit tokens into the bridge when insufficient eth fee is sent", async () => {
    await expect(
      fixture.alcb
        .connect(user)
        .depositTokensOnBridges(_poolVersion, encodedDepositCallData, {
          value: minEthFeeForDeposit - 1,
        })
    )
      .to.be.revertedWithCustomError(fixture.alcb, "InsufficientFee")
      .withArgs(minEthFeeForDeposit - 1, minEthFeeForDeposit);
  });

  it("Should deposit tokens into the bridge and receive a refund when greater than sufficient eth fee is sent", async () => {
    await depositAndCheckRefundWei(
      fixture,
      1n,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundWei(
      fixture,
      1n * 10n ** 9n,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundWei(
      fixture,
      1n * 10n ** 15n + 1n,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundWei(
      fixture,
      1n * 10n ** 17n + 1n,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      1,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      3,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      11,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      77,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      113,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      257,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      776,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      1277,
      user,
      _poolVersion,
      encodedDepositCallData
    );
    await depositAndCheckRefundEth(
      fixture,
      5634,
      user,
      _poolVersion,
      encodedDepositCallData
    );
  });
});
