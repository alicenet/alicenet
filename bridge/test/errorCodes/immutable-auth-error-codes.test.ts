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
  it("IMMUTEABLEAUTH_ONLY_LIQUIDITYPROVIDERSTAKING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_LIQUIDITYPROVIDERSTAKING,
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
  it("IMMUTEABLEAUTH_ONLY_PUBLICSTAKING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_PUBLICSTAKING,
      2007
    );
  });
  it("IMMUTEABLEAUTH_ONLY_SNAPSHOTS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_SNAPSHOTS,
      2008
    );
  });
  it("IMMUTEABLEAUTH_ONLY_STAKINGPOSITIONDESCRIPTOR returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_STAKINGPOSITIONDESCRIPTOR,
      2009
    );
  });
  it("IMMUTEABLEAUTH_ONLY_VALIDATORPOOL returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_VALIDATORPOOL,
      2010
    );
  });
  it("IMMUTEABLEAUTH_ONLY_VALIDATORSTAKING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_VALIDATORSTAKING,
      2011
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ATOKENBURNER returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ATOKENBURNER,
      2012
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ATOKENMINTER returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ATOKENMINTER,
      2013
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ETHDKGACCUSATIONS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ETHDKGACCUSATIONS,
      2014
    );
  });
  it("IMMUTEABLEAUTH_ONLY_ETHDKGPHASES returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_ETHDKGPHASES,
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
