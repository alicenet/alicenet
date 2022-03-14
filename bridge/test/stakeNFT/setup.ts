import { expect } from "chai";
import { BigNumberish, Contract } from "ethers";

export interface Position {
  shares: bigint;
  freeAfter: bigint;
  withdrawFreeAfter: bigint;
  accumulatorEth: bigint;
  accumulatorToken: bigint;
}

export const newPosition = (
  shares: bigint,
  freeAfter: bigint,
  withdrawFreeAfter: bigint,
  accumulatorEth: bigint,
  accumulatorToken: bigint
): Position => {
  return {
    shares,
    freeAfter,
    withdrawFreeAfter,
    accumulatorEth,
    accumulatorToken,
  };
};

export const getPosition = async (
  contract: Contract,
  tokenID: BigNumberish
): Promise<Position> => {
  let [
    _shares,
    _freeAfter,
    _withdrawFreeAfter,
    _accumulatorEth,
    _accumulatorToken,
  ] = await contract.getPosition(tokenID);
  return {
    shares: _shares.toBigInt(),
    freeAfter: _freeAfter.toBigInt(),
    withdrawFreeAfter: _withdrawFreeAfter.toBigInt(),
    accumulatorEth: _accumulatorEth.toBigInt(),
    accumulatorToken: _accumulatorToken.toBigInt(),
  };
};

export const assertTotalReserveAndZeroExcess = async (
  contract: Contract,
  reserveAmountToken: bigint,
  reserveAmountEth: bigint
) => {
  expect((await contract.getTotalReserveMadToken()).toBigInt()).to.be.equals(
    reserveAmountToken
  );
  expect((await contract.getTotalReserveEth()).toBigInt()).to.be.equals(
    reserveAmountEth
  );
  expect((await contract.estimateExcessToken()).toBigInt()).to.be.equals(
    BigInt(0)
  );
  expect((await contract.estimateExcessEth()).toBigInt()).to.be.equals(
    BigInt(0)
  );
};
