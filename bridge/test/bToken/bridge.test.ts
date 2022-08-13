import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
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

describe("Testing BToken bridge methods", async () => {
  let user: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 40;
  let ethForMinting: BigNumber;
  let bTokens: BigNumber;
  const minBTokens = 0;
  let ethsFromBurning: BigNumber;
  let depositCallData: any;
  let encodedDepositCallData: string;
  const valueOrId = 100;
  const _tokenType = 1; // ERC20
  const chainId = 1337;
  const _poolVersion = 1;
  const bTokenFee = 1000; // Fee that's returned by BridgeRouterMok
  const minEthFeeForDeposit = 8; // Curve value for the BridgeRouterMok returned fee
  const refund = 1;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [, user] = signers;
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
      tokenType: _tokenType,
      number: valueOrId,
      chainID: chainId,
      poolVersion: _poolVersion,
    };
    encodedDepositCallData = ethers.utils.defaultAbiCoder.encode(
      [
        "tuple(address ERCContract, uint8 tokenType, uint256 number, uint256 chainID, uint16 poolVersion)",
      ],
      [depositCallData]
    );
    await deployUpgradeableWithFactory(
      fixture.factory,
      "BridgeRouterMock",
      getBridgeRouterSalt(1)
    );
    showState("Initial", await getState(fixture));
  });

  it("Should deposit tokens into the bridge and destroy the correspondent BToken fee if no eth fee is sent", async () => {
    expectedState = await getState(fixture);
    const tx = await fixture.bToken
      .connect(user)
      .depositTokensOnBridges(_poolVersion, encodedDepositCallData);
    ethsFromBurning = await fixture.bToken.getLatestEthFromBTokensBurn(
      bTokenFee
    );
    expectedState.Balances.bToken.user -= BigInt(bTokenFee);
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.bToken.totalSupply -= BigInt(bTokenFee);
    expectedState.Balances.bToken.poolBalance -= ethsFromBurning.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should not deposit BTokens into an inexistent bridge router version", async () => {
    expectedState = await getState(fixture);
    const _poolVersion = 2; // inexistent version
    depositCallData = {
      ERCContract: ethers.constants.AddressZero,
      tokenType: _tokenType,
      number: valueOrId,
      chainID: chainId,
      poolVersion: _poolVersion,
    };
    encodedDepositCallData = ethers.utils.defaultAbiCoder.encode(
      [
        "tuple(address ERCContract, uint8 tokenType, uint256 number, uint256 chainID, uint16 poolVersion)",
      ],
      [depositCallData]
    );
    const expectedErrAddress = await fixture.factory.lookup(
      getBridgeRouterSalt(2)
    );
    await expect(
      fixture.bToken
        .connect(user)
        .depositTokensOnBridges(_poolVersion, encodedDepositCallData)
    )
      .to.be.revertedWithCustomError(fixture.bToken, "InexistentRouterContract")
      .withArgs(expectedErrAddress);
  });

  it("Should not deposit tokens into the bridge when insufficient eth fee is sent", async () => {
    await expect(
      fixture.bToken
        .connect(user)
        .depositTokensOnBridges(_poolVersion, encodedDepositCallData, {
          value: minEthFeeForDeposit - 1,
        })
    )
      .to.be.revertedWithCustomError(fixture.bToken, "InsufficientFee")
      .withArgs(minEthFeeForDeposit - 1, minEthFeeForDeposit);
  });

  it("Should deposit tokens into the bridge and receive a refund when greater than sufficient eth fee is sent", async () => {
    expectedState = await getState(fixture);
    const ethFeeForDeposit = minEthFeeForDeposit + refund;
    const tx = await fixture.bToken
      .connect(user)
      .depositTokensOnBridges(_poolVersion, encodedDepositCallData, {
        value: ethFeeForDeposit,
      });
    expectedState.Balances.eth.user -= BigInt(ethFeeForDeposit);
    expectedState.Balances.eth.user += BigInt(refund);
    expectedState.Balances.eth.bToken += BigInt(minEthFeeForDeposit);
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });
});
