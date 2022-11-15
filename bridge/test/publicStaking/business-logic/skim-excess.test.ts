import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import {
  AliceNetFactory,
  AToken,
  PublicStaking,
} from "../../../typechain-types";
import {
  createUsers,
  deployAliceNetFactory,
  deployUpgradeableWithFactory,
  factoryCallAny,
  getMetamorphicAddress,
  mineBlocks,
  posFixtureSetup,
  preFixtureSetup,
} from "../../setup";
import {
  burnPositionCheckAndUpdateState,
  collectEthCheckAndUpdateState,
  collectTokensCheckAndUpdateState,
  depositEthCheckAndUpdateState,
  depositTokensCheckAndUpdateState,
  getCurrentState,
  mintPositionCheckAndUpdateState,
} from "../setup";

describe("PublicStaking: Skim excess of tokens", async () => {
  let stakingContract: PublicStaking;
  let aToken: AToken;
  let factory: AliceNetFactory;
  const numberUsers: number = 3;
  let etherExcess: bigint;
  let tokenExcess: bigint;

  async function deployFixture() {
    const [adminSigner] = await ethers.getSigners();
    await preFixtureSetup();

    const etherExcess = ethers.utils.parseEther("100").toBigInt();

    const legacyToken = await (
      await ethers.getContractFactory("LegacyToken")
    ).deploy();

    const factory = await deployAliceNetFactory(legacyToken.address);

    // AToken
    const aToken = await ethers.getContractAt(
      "AToken",
      await factory.lookup(ethers.utils.formatBytes32String("AToken"))
    );

    const publicStakingAddress = getMetamorphicAddress(
      factory.address,
      "PublicStaking"
    );
    // transferring ether before contract deployment to get eth excess
    await adminSigner.sendTransaction({
      to: publicStakingAddress,
      value: etherExcess,
    });
    const stakingContract = (await deployUpgradeableWithFactory(
      factory,
      "PublicStaking",
      "PublicStaking"
    )) as PublicStaking;
    await posFixtureSetup(factory, aToken);
    const tokenExcess = ethers.utils.parseUnits("100", 18).toBigInt();
    await aToken.approve(
      stakingContract.address,
      ethers.utils.parseUnits("1000000", 18)
    );
    await aToken.transfer(publicStakingAddress, tokenExcess);
    return {
      stakingContract,
      aToken,
      factory,
      etherExcess,
      tokenExcess,
    };
  }

  beforeEach(async function () {
    ({ stakingContract, aToken, factory, etherExcess, tokenExcess } =
      await loadFixture(deployFixture));
  });

  it("Skim excess of token and ether", async function () {
    expect(
      (await ethers.provider.getBalance(stakingContract.address)).toBigInt()
    ).to.equals(etherExcess, "Excess Eth amount doesn't match");

    expect(
      (await aToken.balanceOf(stakingContract.address)).toBigInt()
    ).to.equals(tokenExcess, "Excess token amount doesn't match");

    const [userWithoutEth] = await createUsers(1, true);

    await factoryCallAny(factory, stakingContract, "skimExcessEth", [
      userWithoutEth.address,
    ]);

    expect(
      (await ethers.provider.getBalance(stakingContract.address)).toBigInt()
    ).to.equals(0n, "Excess Eth amount doesn't match after skim");

    expect(
      (await ethers.provider.getBalance(userWithoutEth.address)).toBigInt()
    ).to.equals(etherExcess, "Excess Eth amount doesn't match after skim User");

    await factoryCallAny(factory, stakingContract, "skimExcessToken", [
      userWithoutEth.address,
    ]);

    expect(
      (await aToken.balanceOf(stakingContract.address)).toBigInt()
    ).to.equals(0n, "Excess token amount doesn't match after skim");

    expect(
      (await aToken.balanceOf(userWithoutEth.address)).toBigInt()
    ).to.equals(
      tokenExcess,
      "Excess token amount doesn't match after skim User"
    );
  });

  it("Mint, deposit, collect and burn excess of token and ether", async function () {
    expect(
      (await ethers.provider.getBalance(stakingContract.address)).toBigInt()
    ).to.equals(etherExcess, "Excess Eth amount doesn't match");

    expect(
      (await aToken.balanceOf(stakingContract.address)).toBigInt()
    ).to.equals(tokenExcess, "Excess token amount doesn't match");

    const sharesPerUser = ethers.utils.parseUnits("10", 18).toBigInt();
    const totalShares = sharesPerUser * BigInt(numberUsers);
    const users = await createUsers(numberUsers);
    const tokensID: number[] = [];
    for (let i = 0; i < users.length; i++) {
      await aToken.transfer(await users[i].getAddress(), sharesPerUser);
      await aToken
        .connect(users[i])
        .approve(stakingContract.address, sharesPerUser);
      tokensID.push(0);
    }

    const expectedState = await getCurrentState(
      stakingContract,
      aToken,
      users,
      tokensID
    );

    for (let i = 0; i < numberUsers; i++) {
      await mintPositionCheckAndUpdateState(
        stakingContract,
        aToken,
        sharesPerUser,
        i,
        users,
        tokensID,
        expectedState,
        "After mint 1-" + i
      );
    }
    await mineBlocks(2n);

    for (let i = 0; i < 2; i++) {
      // deposit and collect only with 1 user
      const amountDeposited = ethers.utils.parseUnits("1000", 0).toBigInt();
      await depositTokensCheckAndUpdateState(
        stakingContract,
        aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1 Tokens " + i
      );
      await depositEthCheckAndUpdateState(
        stakingContract,
        aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1 Eth " + i
      );
      const expectedCollectedAmount = [330n, 330n, 330n];
      for (let j = 0; j < numberUsers; j++) {
        await collectTokensCheckAndUpdateState(
          stakingContract,
          aToken,
          expectedCollectedAmount[i],
          j,
          users,
          tokensID,
          expectedState,
          "After collect 1 Token " + j
        );

        await collectEthCheckAndUpdateState(
          stakingContract,
          aToken,
          expectedCollectedAmount[i],
          j,
          users,
          tokensID,
          expectedState,
          "After collect 1 Eth " + j
        );
      }
    }

    const expectedSlush = 20n;
    expect(
      (await ethers.provider.getBalance(stakingContract.address)).toBigInt()
    ).to.equals(
      expectedSlush + etherExcess,
      "Excess Eth amount doesn't match before skim"
    );

    expect(
      (await aToken.balanceOf(stakingContract.address)).toBigInt()
    ).to.equals(
      expectedSlush + totalShares + tokenExcess,
      "Excess token amount doesn't match after skim"
    );

    // deposit and collect only with 1 user
    const amountDeposited = ethers.utils.parseUnits("900", 0).toBigInt();
    await depositTokensCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2 Tokens"
    );

    await depositEthCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2 Eth"
    );
    const expectedPayoutAmountEth = [300n, 310n, 310n];
    // last two users burning split the slush of 20
    const expectedPayoutAmountToken = [
      sharesPerUser + 300n,
      sharesPerUser + 310n,
      sharesPerUser + 310n,
    ];

    const expectedSlushes = [20000000000000000000n, 0n, 0n];
    for (let j = 0; j < numberUsers; j++) {
      await burnPositionCheckAndUpdateState(
        stakingContract,
        aToken,
        sharesPerUser,
        expectedPayoutAmountEth[j],
        expectedPayoutAmountToken[j],
        j,
        users,
        tokensID,
        expectedState,
        "After burn " + j,
        expectedSlushes[j],
        expectedSlushes[j]
      );
    }

    const [userWithoutEth] = await createUsers(1, true);

    await factoryCallAny(factory, stakingContract, "skimExcessEth", [
      userWithoutEth.address,
    ]);

    expect(
      (await ethers.provider.getBalance(stakingContract.address)).toBigInt()
    ).to.equals(0n, "Excess Eth amount doesn't match after skim");

    expect(
      (await ethers.provider.getBalance(userWithoutEth.address)).toBigInt()
    ).to.equals(etherExcess, "Excess Eth amount doesn't match after skim User");

    await factoryCallAny(factory, stakingContract, "skimExcessToken", [
      userWithoutEth.address,
    ]);

    expect(
      (await aToken.balanceOf(stakingContract.address)).toBigInt()
    ).to.equals(0n, "Excess token amount doesn't match after skim");

    expect(
      (await aToken.balanceOf(userWithoutEth.address)).toBigInt()
    ).to.equals(
      tokenExcess,
      "Excess token amount doesn't match after skim User"
    );
  });
});
