import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing ImmutableAuth error codes", async () => {
  let immutableAuthErrorCodesContract: Contract;

  beforeEach(async function () {
    immutableAuthErrorCodesContract = await deployLibrary(
      "ImmutableAuthErrorCodes"
    );
  });

  it("IMMUTEABLEAUTH_ONLY_FACTORY returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_FACTORY,
      2000
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ATOKEN returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ATOKEN,
      2001
    );
  });
  it("IMMUTEABLEAUTH_ONLY_FOUNDATION returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_FOUNDATION,
      2002
    );
  });
  it("IMMUTEABLEAUTH_ONLY_GOVERNANCE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_GOVERNANCE,
      2003
    );
  });
  it("IMMUTEABLEAUTH_ONLY_LP_STAKING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_LP_STAKING,
      2004
    );
  });
  it("IMMUTEABLEAUTH_ONLY_BTOKEN returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_BTOKEN,
      2005
    );
  });
  it("IMMUTEABLEAUTH_ONLY_MADTOKEN returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_MADTOKEN,
      2006
    );
  });
  it("IMMUTEABLEAUTH_ONLY_PUBLIC_STAKING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_PUBLIC_STAKING,
      2007
    );
  });
  it("IMMUTEABLEAUTH_ONLY_SNAPSHOTS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_SNAPSHOTS,
      2008
    );
  });
  it("IMMUTEABLEAUTH_ONLY_STAKING_POSITION_DESCRIPTOR returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_STAKING_POSITION_DESCRIPTOR,
      2009
    );
  });
  it("IMMUTEABLEAUTH_ONLY_VALIDATOR_POOL returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_VALIDATOR_POOL,
      2010
    );
  });
  it("IMMUTEABLEAUTH_ONLY_VALIDATOR_STAKING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_VALIDATOR_STAKING,
      2011
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ATOKEN_BURNER returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ATOKEN_BURNER,
      2012
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ATOKEN_MINTER returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ATOKEN_MINTER,
      2013
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ETHDKG_ACCUSATIONS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ETHDKG_ACCUSATIONS,
      2014
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ETHDKG_PHASES returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ETHDKG_PHASES,
      2015
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ETHDKG returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ETHDKG,
      2016
    );
  });
});
