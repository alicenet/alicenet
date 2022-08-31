import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber, BigNumberish, Contract } from "ethers";
import { ethers } from "hardhat";
import {
  callFunctionAndGetReturnValues,
  getTokenIdFromTx,
  mineBlocks,
} from "../setup";

export interface Position {
  Shares: bigint;
  FreeAfter: bigint;
  WithdrawFreeAfter: bigint;
  AccumulatorEth: bigint;
  AccumulatorToken: bigint;
}

export interface User {
  Address: string;
  TokenBalance: bigint;
  NFTBalance: bigint;
  Idx: number;
  Position: Position;
}

export interface AccumulatorState {
  Accumulator: bigint;
  Slush: bigint;
}

export interface StakingContract {
  Address: string;
  TokenBalance: bigint;
  EthBalance: bigint;
  TotalShares: bigint;
  ReserveTokens: bigint;
  ReserveEth: bigint;
  AccumulatorToken: AccumulatorState;
  AccumulatorEth: AccumulatorState;
}

export interface StakingState {
  BaseStaking: StakingContract;
  Users: User[];
}

export const getCurrentState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  _users: SignerWithAddress[],
  _tokensID: number[]
): Promise<StakingState> => {
  if (_users.length !== _tokensID.length) {
    throw new Error(
      "The _users array should have the same length as _tokensID array." +
        " In case a user doesn't have a minted NFT position send 0 in that position!"
    );
  }
  const [accumulatorEth, slushEth] = await stakingContract.getEthAccumulator();
  const [accumulatorToken, slushToken] =
    await stakingContract.getTokenAccumulator();
  const stakingState: StakingState = {
    BaseStaking: {
      Address: stakingContract.address,
      TokenBalance: (
        await tokenContract.balanceOf(stakingContract.address)
      ).toBigInt(),
      EthBalance: (
        await ethers.provider.getBalance(stakingContract.address)
      ).toBigInt(),
      TotalShares: (await stakingContract.getTotalShares()).toBigInt(),
      ReserveTokens: (await stakingContract.getTotalReserveAToken()).toBigInt(),
      ReserveEth: (await stakingContract.getTotalReserveEth()).toBigInt(),
      AccumulatorToken: {
        Accumulator: accumulatorToken.toBigInt(),
        Slush: slushToken.toBigInt(),
      },
      AccumulatorEth: {
        Accumulator: accumulatorEth.toBigInt(),
        Slush: slushEth.toBigInt(),
      },
    },
    Users: [],
  };

  for (let i = 0; i < _users.length; i++) {
    stakingState.Users.push({
      Address: _users[i].address,
      TokenBalance: (
        await tokenContract.balanceOf(_users[i].address)
      ).toBigInt(),
      NFTBalance: (
        await stakingContract.balanceOf(_users[i].address)
      ).toBigInt(),
      Idx: i,
      Position:
        _tokensID[i] === 0
          ? newPosition(0n, 0n, 0n, 0n, 0n)
          : await getPosition(stakingContract, _tokensID[i]),
    });
  }

  return stakingState;
};

export const newPosition = (
  shares: bigint,
  freeAfter: bigint,
  withdrawFreeAfter: bigint,
  accumulatorEth: bigint,
  accumulatorToken: bigint
): Position => {
  return {
    Shares: shares,
    FreeAfter: freeAfter,
    WithdrawFreeAfter: withdrawFreeAfter,
    AccumulatorEth: accumulatorEth,
    AccumulatorToken: accumulatorToken,
  };
};

export const getPosition = async (
  contract: Contract,
  tokenID: BigNumberish
): Promise<Position> => {
  const [
    _shares,
    _freeAfter,
    _withdrawFreeAfter,
    _accumulatorEth,
    _accumulatorToken,
  ] = await contract.getPosition(tokenID);
  return {
    Shares: _shares.toBigInt(),
    FreeAfter: _freeAfter.toBigInt(),
    WithdrawFreeAfter: _withdrawFreeAfter.toBigInt(),
    AccumulatorEth: _accumulatorEth.toBigInt(),
    AccumulatorToken: _accumulatorToken.toBigInt(),
  };
};

export const collectEth = async (
  contract: Contract,
  user: SignerWithAddress,
  tokenID: number
): Promise<BigNumber> => {
  const [collectedEth] = await callFunctionAndGetReturnValues(
    contract,
    "collectEth",
    user,
    [tokenID]
  );
  return collectedEth;
};

export const collectToken = async (
  contract: Contract,
  user: SignerWithAddress,
  tokenID: number
): Promise<BigNumber> => {
  const [collectedEth] = await callFunctionAndGetReturnValues(
    contract,
    "collectToken",
    user,
    [tokenID]
  );
  return collectedEth;
};

export const collectAllProfits = async (
  contract: Contract,
  user: SignerWithAddress,
  tokenID: number
): Promise<[BigNumber, BigNumber]> => {
  const [collectedProfits] = await callFunctionAndGetReturnValues(
    contract,
    "collectAllProfits",
    user,
    [tokenID]
  );

  return [collectedProfits.payoutToken, collectedProfits.payoutEth];
};

export const collectAllProfitsTo = async (
  contract: Contract,
  user: SignerWithAddress,
  destinationUser: SignerWithAddress,
  tokenID: number
): Promise<[BigNumber, BigNumber]> => {
  const [collectedProfits] = await callFunctionAndGetReturnValues(
    contract,
    "collectAllProfitsTo",
    user,
    [destinationUser.address, tokenID]
  );

  return [collectedProfits.payoutToken, collectedProfits.payoutEth];
};

export const collectEthTo = async (
  contract: Contract,
  user: SignerWithAddress,
  destinationUser: SignerWithAddress,
  tokenID: number
): Promise<BigNumber> => {
  const [collectedEth] = await callFunctionAndGetReturnValues(
    contract,
    "collectEthTo",
    user,
    [destinationUser.address, tokenID]
  );
  return collectedEth;
};

export const collectTokenTo = async (
  contract: Contract,
  user: SignerWithAddress,
  destinationUser: SignerWithAddress,
  tokenID: number
): Promise<BigNumber> => {
  const [collectedEth] = await callFunctionAndGetReturnValues(
    contract,
    "collectTokenTo",
    user,
    [destinationUser.address, tokenID]
  );
  return collectedEth;
};

export const burnPosition = async (
  contract: Contract,
  user: SignerWithAddress,
  tokenID: number
): Promise<[BigNumber, BigNumber]> => {
  const [[payoutEth, payoutToken]] = await callFunctionAndGetReturnValues(
    contract,
    "burn",
    user,
    [tokenID]
  );
  return [payoutEth, payoutToken];
};

export const mintPosition = async (
  contract: Contract,
  user: SignerWithAddress,
  amountToMint: bigint,
  expectedAccumulatorEth: bigint = 0n,
  expectedAccumulatorToken: bigint = 0n
): Promise<[number, Position]> => {
  const tx = await contract.connect(user).mint(amountToMint);
  await mineBlocks(1n);
  const tokenID = await getTokenIdFromTx(tx);
  const blockNumber = BigInt((await tx.wait()).blockNumber);
  const expectedPosition = newPosition(
    amountToMint,
    blockNumber + 1n,
    blockNumber + 1n,
    expectedAccumulatorEth,
    expectedAccumulatorToken
  );
  await assertPositions(contract, tokenID, expectedPosition, user.address);

  return [tokenID, expectedPosition];
};

export const estimateAndCollectTokens = async (
  contract: Contract,
  tokenID: number,
  user: SignerWithAddress,
  expectedCollectAmount: bigint
) => {
  expect(
    (await contract.estimateTokenCollection(tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmount,
    "Token collection don't match expected value!"
  );

  expect((await collectToken(contract, user, tokenID)).toBigInt()).to.be.equals(
    expectedCollectAmount,
    "Collected amount to match expected value!"
  );
};

export const estimateAndCollectEth = async (
  contract: Contract,
  tokenID: number,
  user: SignerWithAddress,
  expectedCollectAmount: bigint
) => {
  expect(
    (await contract.estimateEthCollection(tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmount,
    "Token collection don't match expected value!"
  );

  expect((await collectEth(contract, user, tokenID)).toBigInt()).to.be.equals(
    expectedCollectAmount,
    "Collected amount to match expected value!"
  );
};

export const estimateAndCollectAllProfits = async (
  contract: Contract,
  tokenID: number,
  user: SignerWithAddress,
  expectedCollectAmountTokens: bigint,
  expectedCollectAmountEth: bigint
) => {
  expect(
    (await contract.estimateTokenCollection(tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmountTokens,
    "Token collection does not match expected value!"
  );

  expect(
    (await contract.estimateEthCollection(tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmountEth,
    "Eth collection does not match expected value!"
  );

  const [payoutToken, payoutEth] = await collectAllProfits(
    contract,
    user,
    tokenID
  );
  expect(payoutToken.toBigInt()).to.be.equals(
    expectedCollectAmountTokens,
    "Collected token amount to match expected value!"
  );
  expect(payoutEth.toBigInt()).to.be.equals(
    expectedCollectAmountTokens,
    "Collected eth amount to match expected value!"
  );
};

export const estimateAndCollectTokensTo = async (
  contract: Contract,
  tokenID: number,
  user: SignerWithAddress,
  destinationUser: SignerWithAddress,
  expectedCollectAmount: bigint
) => {
  expect(
    (await contract.estimateTokenCollection(tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmount,
    "Token collection don't match expected value!"
  );

  expect(
    (await collectTokenTo(contract, user, destinationUser, tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmount,
    "Collected amount to match expected value!"
  );
};

export const estimateAndCollectEthTo = async (
  contract: Contract,
  tokenID: number,
  user: SignerWithAddress,
  destinationUser: SignerWithAddress,
  expectedCollectAmount: bigint
) => {
  expect(
    (await contract.estimateEthCollection(tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmount,
    "Token collection don't match expected value!"
  );

  expect(
    (await collectEthTo(contract, user, destinationUser, tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmount,
    "Collected amount to match expected value!"
  );
};

export const estimateAndCollectAllProfitsTo = async (
  contract: Contract,
  tokenID: number,
  user: SignerWithAddress,
  destinationUser: SignerWithAddress,
  expectedCollectAmountTokens: bigint,
  expectedCollectAmountEth: bigint
) => {
  expect(
    (await contract.estimateTokenCollection(tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmountTokens,
    "Token collection does not match expected value!"
  );

  expect(
    (await contract.estimateEthCollection(tokenID)).toBigInt()
  ).to.be.equals(
    expectedCollectAmountEth,
    "Eth collection does not match expected value!"
  );

  const [payoutToken, payoutEth] = await collectAllProfitsTo(
    contract,
    user,
    destinationUser,
    tokenID
  );
  expect(payoutToken.toBigInt()).to.be.equals(
    expectedCollectAmountTokens,
    "Collected token amount to match expected value!"
  );
  expect(payoutEth.toBigInt()).to.be.equals(
    expectedCollectAmountTokens,
    "Collected eth amount to match expected value!"
  );
};

export const mintPositionCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  amount: bigint,
  userIdx: number,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string
) => {
  if (stakingState.BaseStaking.TotalShares > BigInt(0)) {
    // Update eth accum
    const deltaAccumEth =
      stakingState.BaseStaking.AccumulatorEth.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Slush -=
      deltaAccumEth * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Accumulator += deltaAccumEth;
    // Update token accum
    const deltaAccumToken =
      stakingState.BaseStaking.AccumulatorToken.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Slush -=
      deltaAccumToken * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Accumulator += deltaAccumToken;
  }
  const [tokenID, expectedPosition] = await mintPosition(
    stakingContract,
    users[userIdx],
    amount,
    stakingState.BaseStaking.AccumulatorEth.Accumulator,
    stakingState.BaseStaking.AccumulatorToken.Accumulator
  );

  tokensID[userIdx] = tokenID;
  stakingState.BaseStaking.TokenBalance += amount;
  stakingState.BaseStaking.ReserveTokens += amount;
  stakingState.BaseStaking.TotalShares += amount;
  stakingState.Users[userIdx].TokenBalance -= amount;
  stakingState.Users[userIdx].NFTBalance++;
  stakingState.Users[userIdx].Position = expectedPosition;

  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const depositTokensCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  amountDeposited: bigint,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string
) => {
  const scaleFactor = (
    await stakingContract.getAccumulatorScaleFactor()
  ).toBigInt();
  await stakingContract.depositToken(42, amountDeposited);
  const scaledAmount = amountDeposited * scaleFactor;
  stakingState.BaseStaking.TokenBalance += amountDeposited;
  stakingState.BaseStaking.ReserveTokens += amountDeposited;
  let accumulatorValue = stakingState.BaseStaking.AccumulatorToken.Accumulator;
  const overflowValue = 2n ** 168n;
  // simulating accumulator overflow which happens at 2n ** 168
  if (accumulatorValue >= overflowValue) {
    accumulatorValue = accumulatorValue - overflowValue;
  }
  stakingState.BaseStaking.AccumulatorToken.Accumulator = accumulatorValue;
  stakingState.BaseStaking.AccumulatorToken.Slush += scaledAmount;
  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const depositEthCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  amountDeposited: bigint,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string
) => {
  const scaleFactor = (
    await stakingContract.getAccumulatorScaleFactor()
  ).toBigInt();
  await stakingContract.depositEth(42, {
    value: amountDeposited,
  });
  const scaledAmount = amountDeposited * scaleFactor;
  stakingState.BaseStaking.EthBalance += amountDeposited;
  stakingState.BaseStaking.ReserveEth += amountDeposited;
  let accumulatorValue = stakingState.BaseStaking.AccumulatorEth.Accumulator;
  const overflowValue = 2n ** 168n;
  // simulating accumulator overflow which happens at 2n ** 168
  if (accumulatorValue >= overflowValue) {
    accumulatorValue = accumulatorValue - overflowValue;
  }
  stakingState.BaseStaking.AccumulatorEth.Accumulator = accumulatorValue;
  stakingState.BaseStaking.AccumulatorEth.Slush += scaledAmount;
  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const collectTokensCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  expectedAmountCollected: bigint,
  userIdx: number,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string,
  expectedSlush?: bigint
) => {
  if (stakingState.BaseStaking.TotalShares > BigInt(0)) {
    const overflowValue = 2n ** 168n;
    // Update token accum
    const deltaAccumToken =
      stakingState.BaseStaking.AccumulatorToken.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Slush -=
      deltaAccumToken * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Accumulator += deltaAccumToken;
    if (
      stakingState.BaseStaking.AccumulatorToken.Accumulator >= overflowValue
    ) {
      stakingState.BaseStaking.AccumulatorToken.Accumulator -= overflowValue;
    }
  }
  await estimateAndCollectTokens(
    stakingContract,
    tokensID[userIdx],
    users[userIdx],
    expectedAmountCollected
  );
  stakingState.Users[userIdx].TokenBalance += expectedAmountCollected;
  stakingState.Users[userIdx].Position.AccumulatorToken =
    stakingState.BaseStaking.AccumulatorToken.Accumulator;
  stakingState.BaseStaking.TokenBalance -= expectedAmountCollected;
  stakingState.BaseStaking.ReserveTokens -= expectedAmountCollected;
  if (expectedSlush !== undefined) {
    stakingState.BaseStaking.AccumulatorToken.Slush = expectedSlush;
  }
  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const collectEthCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  expectedAmountCollected: bigint,
  userIdx: number,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string,
  expectedSlush?: bigint,
  expectedSlushToken?: bigint
) => {
  if (stakingState.BaseStaking.TotalShares > BigInt(0)) {
    const overflowValue = 2n ** 168n;
    // Update eth accum
    const deltaAccumEth =
      stakingState.BaseStaking.AccumulatorEth.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Slush -=
      deltaAccumEth * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Accumulator += deltaAccumEth;
    if (stakingState.BaseStaking.AccumulatorEth.Accumulator >= overflowValue) {
      stakingState.BaseStaking.AccumulatorEth.Accumulator -= overflowValue;
    }
  }
  await estimateAndCollectEth(
    stakingContract,
    tokensID[userIdx],
    users[userIdx],
    expectedAmountCollected
  );
  stakingState.Users[userIdx].Position.AccumulatorEth =
    stakingState.BaseStaking.AccumulatorEth.Accumulator;
  stakingState.BaseStaking.EthBalance -= expectedAmountCollected;
  stakingState.BaseStaking.ReserveEth -= expectedAmountCollected;
  if (expectedSlush !== undefined) {
    stakingState.BaseStaking.AccumulatorEth.Slush = expectedSlush;
  }
  if (expectedSlushToken !== undefined) {
    stakingState.BaseStaking.AccumulatorToken.Slush = expectedSlushToken;
  }
  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const collectAllProfitsCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  expectedEthAmountCollected: bigint,
  expectedTokenAmountCollected: bigint,
  userIdx: number,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string,
  expectedSlush?: bigint,
  expectedSlushToken?: bigint
) => {
  if (stakingState.BaseStaking.TotalShares > BigInt(0)) {
    const overflowValue = 2n ** 168n;
    // Update eth accum
    const deltaAccumEth =
      stakingState.BaseStaking.AccumulatorEth.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Slush -=
      deltaAccumEth * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Accumulator += deltaAccumEth;
    if (stakingState.BaseStaking.AccumulatorEth.Accumulator >= overflowValue) {
      stakingState.BaseStaking.AccumulatorEth.Accumulator -= overflowValue;
    }
    const deltaAccumToken =
      stakingState.BaseStaking.AccumulatorToken.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Slush -=
      deltaAccumToken * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Accumulator += deltaAccumToken;
    if (
      stakingState.BaseStaking.AccumulatorToken.Accumulator >= overflowValue
    ) {
      stakingState.BaseStaking.AccumulatorToken.Accumulator -= overflowValue;
    }
  }
  await estimateAndCollectAllProfits(
    stakingContract,
    tokensID[userIdx],
    users[userIdx],
    expectedTokenAmountCollected,
    expectedEthAmountCollected
  );

  stakingState.Users[userIdx].TokenBalance += expectedTokenAmountCollected;
  stakingState.Users[userIdx].Position.AccumulatorToken =
    stakingState.BaseStaking.AccumulatorToken.Accumulator;
  stakingState.BaseStaking.TokenBalance -= expectedTokenAmountCollected;
  stakingState.BaseStaking.ReserveTokens -= expectedTokenAmountCollected;

  stakingState.Users[userIdx].Position.AccumulatorEth =
    stakingState.BaseStaking.AccumulatorEth.Accumulator;
  stakingState.BaseStaking.EthBalance -= expectedEthAmountCollected;
  stakingState.BaseStaking.ReserveEth -= expectedEthAmountCollected;
  if (expectedSlush !== undefined) {
    stakingState.BaseStaking.AccumulatorEth.Slush = expectedSlush;
  }
  if (expectedSlushToken !== undefined) {
    stakingState.BaseStaking.AccumulatorToken.Slush = expectedSlushToken;
  }
  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const collectTokensToCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  expectedAmountCollected: bigint,
  userIdx: number,
  destinationUserIdx: number,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string,
  expectedSlush?: bigint
) => {
  if (stakingState.BaseStaking.TotalShares > BigInt(0)) {
    const overflowValue = 2n ** 168n;
    // Update token accum
    const deltaAccumToken =
      stakingState.BaseStaking.AccumulatorToken.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Slush -=
      deltaAccumToken * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Accumulator += deltaAccumToken;
    if (
      stakingState.BaseStaking.AccumulatorToken.Accumulator >= overflowValue
    ) {
      stakingState.BaseStaking.AccumulatorToken.Accumulator -= overflowValue;
    }
  }
  await estimateAndCollectTokensTo(
    stakingContract,
    tokensID[userIdx],
    users[userIdx],
    users[destinationUserIdx],
    expectedAmountCollected
  );
  stakingState.Users[destinationUserIdx].TokenBalance +=
    expectedAmountCollected;
  stakingState.Users[userIdx].Position.AccumulatorToken =
    stakingState.BaseStaking.AccumulatorToken.Accumulator;
  stakingState.BaseStaking.TokenBalance -= expectedAmountCollected;
  stakingState.BaseStaking.ReserveTokens -= expectedAmountCollected;
  if (expectedSlush !== undefined) {
    stakingState.BaseStaking.AccumulatorToken.Slush = expectedSlush;
  }
  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const collectEthToCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  expectedAmountCollected: bigint,
  userIdx: number,
  destinationUserIdx: number,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string,
  expectedSlush?: bigint,
  expectedSlushToken?: bigint
) => {
  if (stakingState.BaseStaking.TotalShares > BigInt(0)) {
    const overflowValue = 2n ** 168n;
    // Update eth accum
    const deltaAccumEth =
      stakingState.BaseStaking.AccumulatorEth.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Slush -=
      deltaAccumEth * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Accumulator += deltaAccumEth;
    if (stakingState.BaseStaking.AccumulatorEth.Accumulator >= overflowValue) {
      stakingState.BaseStaking.AccumulatorEth.Accumulator -= overflowValue;
    }
  }
  await estimateAndCollectEthTo(
    stakingContract,
    tokensID[userIdx],
    users[userIdx],
    users[destinationUserIdx],
    expectedAmountCollected
  );
  stakingState.Users[userIdx].Position.AccumulatorEth =
    stakingState.BaseStaking.AccumulatorEth.Accumulator;
  stakingState.BaseStaking.EthBalance -= expectedAmountCollected;
  stakingState.BaseStaking.ReserveEth -= expectedAmountCollected;
  if (expectedSlush !== undefined) {
    stakingState.BaseStaking.AccumulatorEth.Slush = expectedSlush;
  }
  if (expectedSlushToken !== undefined) {
    stakingState.BaseStaking.AccumulatorToken.Slush = expectedSlushToken;
  }
  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const collectAllProfitsToCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  expectedEthAmountCollected: bigint,
  expectedTokenAmountCollected: bigint,
  userIdx: number,
  destinationUserIdx: number,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string,
  expectedSlush?: bigint,
  expectedSlushToken?: bigint
) => {
  if (stakingState.BaseStaking.TotalShares > BigInt(0)) {
    const overflowValue = 2n ** 168n;
    // Update eth accum
    const deltaAccumEth =
      stakingState.BaseStaking.AccumulatorEth.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Slush -=
      deltaAccumEth * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Accumulator += deltaAccumEth;
    if (stakingState.BaseStaking.AccumulatorEth.Accumulator >= overflowValue) {
      stakingState.BaseStaking.AccumulatorEth.Accumulator -= overflowValue;
    }

    const deltaAccumToken =
      stakingState.BaseStaking.AccumulatorToken.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Slush -=
      deltaAccumToken * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Accumulator += deltaAccumToken;
    if (
      stakingState.BaseStaking.AccumulatorToken.Accumulator >= overflowValue
    ) {
      stakingState.BaseStaking.AccumulatorToken.Accumulator -= overflowValue;
    }
  }

  await estimateAndCollectAllProfitsTo(
    stakingContract,
    tokensID[userIdx],
    users[userIdx],
    users[destinationUserIdx],
    expectedTokenAmountCollected,
    expectedEthAmountCollected
  );
  stakingState.Users[userIdx].Position.AccumulatorEth =
    stakingState.BaseStaking.AccumulatorEth.Accumulator;
  stakingState.BaseStaking.EthBalance -= expectedEthAmountCollected;
  stakingState.BaseStaking.ReserveEth -= expectedEthAmountCollected;
  if (expectedSlush !== undefined) {
    stakingState.BaseStaking.AccumulatorEth.Slush = expectedSlush;
  }
  if (expectedSlushToken !== undefined) {
    stakingState.BaseStaking.AccumulatorToken.Slush = expectedSlushToken;
  }
  stakingState.Users[destinationUserIdx].TokenBalance +=
    expectedTokenAmountCollected;
  stakingState.Users[userIdx].Position.AccumulatorToken =
    stakingState.BaseStaking.AccumulatorToken.Accumulator;
  stakingState.BaseStaking.TokenBalance -= expectedTokenAmountCollected;
  stakingState.BaseStaking.ReserveTokens -= expectedTokenAmountCollected;

  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const burnPositionCheckAndUpdateState = async (
  stakingContract: Contract,
  tokenContract: Contract,
  expectedShares: bigint,
  expectedPayoutEth: bigint,
  expectedPayoutToken: bigint,
  userIdx: number,
  users: SignerWithAddress[],
  tokensID: number[],
  stakingState: StakingState,
  errorMessage: string,
  expectedSlushEth?: bigint,
  expectedSlushToken?: bigint
) => {
  if (stakingState.BaseStaking.TotalShares > BigInt(0)) {
    const overflowValue = 2n ** 168n;
    // Update eth accum
    const deltaAccumEth =
      stakingState.BaseStaking.AccumulatorEth.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Slush -=
      deltaAccumEth * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorEth.Accumulator += deltaAccumEth;
    if (stakingState.BaseStaking.AccumulatorEth.Accumulator >= overflowValue) {
      stakingState.BaseStaking.AccumulatorEth.Accumulator -= overflowValue;
    }
    // Update token accum
    const deltaAccumToken =
      stakingState.BaseStaking.AccumulatorToken.Slush /
      stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Slush -=
      deltaAccumToken * stakingState.BaseStaking.TotalShares;
    stakingState.BaseStaking.AccumulatorToken.Accumulator += deltaAccumToken;
    if (
      stakingState.BaseStaking.AccumulatorToken.Accumulator >= overflowValue
    ) {
      stakingState.BaseStaking.AccumulatorToken.Accumulator -= overflowValue;
    }
  }
  const [payoutEth, payoutToken] = await burnPosition(
    stakingContract,
    users[userIdx],
    tokensID[userIdx]
  );
  expect(payoutEth.toBigInt()).to.be.equals(
    expectedPayoutEth,
    "Payout Eth don't match expected amount!"
  );
  expect(payoutToken.toBigInt()).to.be.equals(
    expectedPayoutToken,
    "Payout token don't match expected amount!"
  );
  tokensID[userIdx] = 0;
  stakingState.Users[userIdx].NFTBalance--;
  stakingState.Users[userIdx].TokenBalance += expectedPayoutToken;
  stakingState.Users[userIdx].Position = newPosition(0n, 0n, 0n, 0n, 0n);
  stakingState.BaseStaking.TokenBalance -= expectedPayoutToken;
  stakingState.BaseStaking.ReserveTokens -= expectedPayoutToken;
  stakingState.BaseStaking.EthBalance -= expectedPayoutEth;
  stakingState.BaseStaking.ReserveEth -= expectedPayoutEth;
  stakingState.BaseStaking.TotalShares -= expectedShares;
  if (expectedSlushEth !== undefined) {
    stakingState.BaseStaking.AccumulatorEth.Slush = expectedSlushEth;
  }
  if (expectedSlushToken !== undefined) {
    stakingState.BaseStaking.AccumulatorToken.Slush = expectedSlushToken;
  }
  expect(
    await getCurrentState(stakingContract, tokenContract, users, tokensID)
  ).to.be.deep.equals(stakingState, errorMessage);
};

export const assertAccumulatorAndSlushEth = async (
  contract: Contract,
  accumulatorExpectedAmount: bigint,
  slushExpectedAmount: bigint
) => {
  const [accumulator, slush] = await contract.getEthAccumulator();
  expect(accumulator.toBigInt()).to.be.equals(
    accumulatorExpectedAmount,
    "Accumulator didn't match the expected amount!"
  );
  expect(slush.toBigInt()).to.be.equals(
    slushExpectedAmount,
    "Slush didn't match the expected amount!"
  );
};

export const assertAccumulatorAndSlushToken = async (
  contract: Contract,
  accumulatorExpectedAmount: bigint,
  slushExpectedAmount: bigint
) => {
  const [accumulator, slush] = await contract.getTokenAccumulator();
  expect(accumulator.toBigInt()).to.be.equals(
    accumulatorExpectedAmount,
    "Accumulator didn't match the expected amount!"
  );
  expect(slush.toBigInt()).to.be.equals(
    slushExpectedAmount,
    "Slush didn't match the expected amount!"
  );
};

export const assertTotalReserveAndZeroExcess = async (
  contract: Contract,
  reserveAmountToken: bigint,
  reserveAmountEth: bigint
) => {
  expect((await contract.getTotalReserveAToken()).toBigInt()).to.be.equals(
    reserveAmountToken,
    "Total reserve tokens don't match expected value!"
  );
  expect((await contract.getTotalReserveEth()).toBigInt()).to.be.equals(
    reserveAmountEth,
    "Total reserve eth don't match expected value!"
  );
  expect((await contract.estimateExcessToken()).toBigInt()).to.be.equals(
    BigInt(0),
    "Excess token is not 0!"
  );
  expect((await contract.estimateExcessEth()).toBigInt()).to.be.equals(
    BigInt(0),
    "Excess eth is not 0!"
  );
};

export const assertPositions = async (
  contract: Contract,
  tokenID: number,
  expectedPosition: Position,
  ownerAddress?: string,
  expectedBalance?: bigint,
  reserveAmountToken?: bigint,
  reserveAmountEth?: bigint
) => {
  expect(await getPosition(contract, tokenID)).to.be.deep.equals(
    expectedPosition,
    "Actual position didn't match expected position!"
  );
  if (ownerAddress !== undefined) {
    expect((await contract.ownerOf(tokenID)).toLowerCase()).to.be.equals(
      ownerAddress.toLowerCase(),
      "Owner address didn't match expected address!"
    );
    if (expectedBalance !== undefined) {
      expect((await contract.balanceOf(ownerAddress)).toBigInt()).to.be.equals(
        expectedBalance,
        "Balance didn't match the expected amount!"
      );
    }
  }
  if (reserveAmountToken !== undefined && reserveAmountEth !== undefined) {
    await assertTotalReserveAndZeroExcess(
      contract,
      reserveAmountToken,
      reserveAmountEth
    );
  }
};

export const assertERC20Balance = async (
  contract: Contract,
  ownerAddress: string,
  expectedAmount: bigint
) => {
  expect((await contract.balanceOf(ownerAddress)).toBigInt()).to.be.equals(
    expectedAmount,
    "ERC20 Balance didn't match the expected amount!"
  );
};
