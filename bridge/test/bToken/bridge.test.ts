import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getState, showState, state } from "./setup";

describe("Testing BToken bridge methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let eths: BigNumber;
  let fixture: Fixture;
  const eth = 40;
  let ethForMinting: BigNumber;
  let bTokens: BigNumber;
  const minBTokens = 0;
  const marketSpread = 4;
  let ethsFromBurning: BigNumber;
  let depositCallData: any;
  let encodedDepositCallData: string;
  const valueOrId = 100;
  const tokenType = 1; // ERC20
  const chainId = 1337;
  const poolVersion = 1;
  const maxTokens = 100;
  const bTokenFee = 1000; // Fee that's returned by BridgeRouterMok
  const minEthFeeForDeposit = 8; // For an actual BridgeRouterMok fee of 1000 btokens

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin, user] = signers;
    ethForMinting = ethers.utils.parseEther(eth.toString());
    [bTokens] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      user,
      [minBTokens],
      ethForMinting
    );
    ethsFromBurning = await fixture.bToken.getLatestEthFromBTokensBurn(bTokens);
    depositCallData = {
      ERCContract: ethers.constants.AddressZero,
      tokenType: tokenType,
      number: valueOrId,
      chainID: chainId,
      poolVersion: poolVersion,
    };
    encodedDepositCallData = ethers.utils.defaultAbiCoder.encode(
      [
        "tuple(address ERCContract, uint8 tokenType, uint256 number, uint256 chainID, uint16 poolVersion)",
      ],
      [depositCallData]
    );
    ethsFromBurning = await fixture.bToken.getLatestEthFromBTokensBurn(
      bTokenFee
    );
    showState("Initial", await getState(fixture));
  });

  it("Should deposit tokens into the bridge and destroy the correspondant btoken fee if no eth fee sent", async () => {
    expectedState = await getState(fixture);
    expectedState.Balances.bToken.user -= BigInt(bTokenFee);
    expectedState.Balances.bToken.totalSupply -= BigInt(bTokenFee);
    expectedState.Balances.bToken.poolBalance -= ethsFromBurning.toBigInt();
    await fixture.bToken
      .connect(user)
      .depositTokensOnBridges(maxTokens, poolVersion, encodedDepositCallData);
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should not deposit btokens into an inexistent bridge router version", async () => {
    expectedState = await getState(fixture);
    const poolVersion = 2; // inexistent version
    depositCallData = {
      ERCContract: ethers.constants.AddressZero,
      tokenType: tokenType,
      number: valueOrId,
      chainID: chainId,
      poolVersion: poolVersion,
    };
    encodedDepositCallData = ethers.utils.defaultAbiCoder.encode(
      [
        "tuple(address ERCContract, uint8 tokenType, uint256 number, uint256 chainID, uint16 poolVersion)",
      ],
      [depositCallData]
    );
    await expect(
      fixture.bToken
        .connect(user)
        .depositTokensOnBridges(maxTokens, poolVersion, encodedDepositCallData)
    ).to.be.revertedWith("InexistentRouterContract");
  });

  it("Should not deposit tokens into the bridge when insufficient eth fee is sent", async () => {
    await expect(
      fixture.bToken
        .connect(user)
        .depositTokensOnBridges(
          maxTokens,
          poolVersion,
          encodedDepositCallData,
          { value: minEthFeeForDeposit - 1 }
        )
    ).to.be.revertedWith("InsufficientFee");
  });

  it("Should deposit tokens into the bridge and receive a refund when greater than sufficient eth fee is sent", async () => {
    expectedState = await getState(fixture);
    const refund = 1;
    const ethFeeForDeposit = minEthFeeForDeposit + refund
    expectedState.Balances.eth.user -= ethFeeForDeposit;
    expectedState.Balances.eth.user += refund;
    expectedState.Balances.eth.bToken += BigInt(minEthFeeForDeposit);
    await fixture.bToken
      .connect(user)
      .depositTokensOnBridges(maxTokens, poolVersion, encodedDepositCallData, { value: ethFeeForDeposit });
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
 });
});
