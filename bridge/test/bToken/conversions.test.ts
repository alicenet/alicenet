import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  BaseTokensFixture,
  callFunctionAndGetReturnValues,
  Fixture,
  getFixture,
} from "../setup";

const checkConversionAndMintEth = async (
  fixture: Fixture | BaseTokensFixture,
  eth: number,
  admin: SignerWithAddress
) => {
  await checkConversionAndMintWei(
    fixture,
    ethers.utils.parseEther(eth.toString()),
    admin
  );
};

const checkConversionAndMintWei = async (
  fixture: Fixture | BaseTokensFixture,
  eth: BigNumber,
  admin: SignerWithAddress
) => {
  const marketSpread = 4;
  const bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(eth);

  const convertedEth = await fixture.bToken.getLatestEthToMintBTokens(bTokens);
  // due to rounding errors during integer math, we need to discard errors in the
  // last 2 decimal places
  expect(convertedEth.div(100)).to.be.equal(eth.div(100));

  const [bTokensMinted] = await callFunctionAndGetReturnValues(
    fixture.bToken,
    "mintTo",
    admin,
    [admin.address, 0],
    eth
  );
  expect(bTokens).to.be.equal(bTokensMinted);

  // check the mint values in the past
  const poolBalanceBefore = (await fixture.bToken.getPoolBalance()).sub(
    eth.div(marketSpread)
  );
  expect(
    await fixture.bToken.getMintedBTokensFromEth(poolBalanceBefore, eth)
  ).to.be.equal(bTokens);
  const totalSupplyBefore = (await fixture.bToken.totalSupply()).sub(bTokens);
  expect(
    (await fixture.bToken.getEthToMintBTokens(totalSupplyBefore, bTokens)).div(
      100
    )
  ).to.be.equal(eth.div(100));
};

const checkConversionAndBurnBToken = async (
  fixture: Fixture | BaseTokensFixture,
  bTokens: number,
  admin: SignerWithAddress
) => {
  await checkConversionAndBurnBTokenWei(
    fixture,
    ethers.utils.parseEther(bTokens.toString()),
    admin
  );
};

const checkConversionAndBurnBTokenWei = async (
  fixture: Fixture | BaseTokensFixture,
  bTokens: BigNumber,
  admin: SignerWithAddress
) => {
  const eth = await fixture.bToken.getLatestEthFromBTokensBurn(bTokens);
  const [burnedEth] = await callFunctionAndGetReturnValues(
    fixture.bToken,
    "burn",
    admin,
    [bTokens, 0],
    BigNumber.from(0)
  );
  expect(eth).to.be.equal(burnedEth);
  // check the mint value in the past
  const poolBalanceBefore = (await fixture.bToken.getPoolBalance()).add(eth);
  const totalSupplyBefore = (await fixture.bToken.totalSupply()).add(bTokens);
  expect(
    await fixture.bToken.getEthFromBTokensBurn(
      poolBalanceBefore,
      totalSupplyBefore,
      bTokens
    )
  ).to.be.equal(eth);
};

describe("Testing BToken conversion functions", async () => {
  let fixture: Fixture;
  let admin: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin] = signers;
  });

  it("Check conversion and mint", async () => {
    // mint
    await checkConversionAndMintWei(fixture, BigNumber.from(4), admin);
    await checkConversionAndMintWei(fixture, BigNumber.from(711), admin);
    await checkConversionAndMintWei(
      fixture,
      BigNumber.from(123455678891),
      admin
    );
    await checkConversionAndMintWei(
      fixture,
      BigNumber.from("871238675431823641276"),
      admin
    );
    await checkConversionAndMintEth(fixture, 1, admin);
    await checkConversionAndMintEth(fixture, 100, admin);
    await checkConversionAndMintEth(fixture, 111, admin);
    await checkConversionAndMintEth(fixture, 765, admin);
    await checkConversionAndMintEth(fixture, 1463, admin);
    await checkConversionAndMintEth(fixture, 10000, admin);
    await checkConversionAndMintEth(fixture, 20000, admin);
    await checkConversionAndMintEth(fixture, 25000, admin);
    await checkConversionAndMintEth(fixture, 1000001, admin);
  });

  it("Check conversion and burn", async () => {
    // mint first
    await checkConversionAndMintEth(fixture, 100000000, admin);
    await checkConversionAndMintEth(fixture, 100000000, admin);
    await checkConversionAndMintEth(fixture, 100000000, admin);

    // burn
    await checkConversionAndBurnBToken(fixture, 1, admin);
    await checkConversionAndBurnBToken(fixture, 9, admin);
    await checkConversionAndBurnBToken(fixture, 11, admin);
    await checkConversionAndBurnBToken(fixture, 75, admin);
    await checkConversionAndBurnBToken(fixture, 143, admin);
    await checkConversionAndBurnBToken(fixture, 555, admin);
    await checkConversionAndBurnBToken(fixture, 1986, admin);
    await checkConversionAndBurnBToken(fixture, 2200, admin);
    await checkConversionAndBurnBToken(fixture, 10001, admin);
  });
});
