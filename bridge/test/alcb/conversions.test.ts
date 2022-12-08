import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
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
  const alcbs = await fixture.alcb.getLatestMintedTokensFromEth(eth);

  const convertedEth = await fixture.alcb.getLatestEthToMintTokens(alcbs);
  // due to rounding errors during integer math, we need to discard errors in the
  // last 2 decimal places
  expect(convertedEth.div(100)).to.be.equal(eth.div(100));

  const [alcbsMinted] = await callFunctionAndGetReturnValues(
    fixture.alcb,
    "mintTo",
    admin,
    [admin.address, 0],
    eth
  );
  expect(alcbs).to.be.equal(alcbsMinted);

  // check the mint values in the past
  const poolBalanceBefore = (await fixture.alcb.getPoolBalance()).sub(
    eth.div(marketSpread)
  );
  expect(
    await fixture.alcb.getMintedTokensFromEth(poolBalanceBefore, eth)
  ).to.be.equal(alcbs);
  const totalSupplyBefore = (await fixture.alcb.totalSupply()).sub(alcbs);
  expect(
    (await fixture.alcb.getEthToMintTokens(totalSupplyBefore, alcbs)).div(100)
  ).to.be.equal(eth.div(100));
};

const checkConversionAndBurnALCB = async (
  fixture: Fixture | BaseTokensFixture,
  alcbs: number,
  admin: SignerWithAddress
) => {
  await checkConversionAndBurnALCBWei(
    fixture,
    ethers.utils.parseEther(alcbs.toString()),
    admin
  );
};

const checkConversionAndBurnALCBWei = async (
  fixture: Fixture | BaseTokensFixture,
  alcbs: BigNumber,
  admin: SignerWithAddress
) => {
  const eth = await fixture.alcb.getLatestEthFromTokensBurn(alcbs);
  const [burnedEth] = await callFunctionAndGetReturnValues(
    fixture.alcb,
    "burn",
    admin,
    [alcbs, 0],
    BigNumber.from(0)
  );
  expect(eth).to.be.equal(burnedEth);
  // check the mint value in the past
  const poolBalanceBefore = (await fixture.alcb.getPoolBalance()).add(eth);
  const totalSupplyBefore = (await fixture.alcb.totalSupply()).add(alcbs);
  expect(
    await fixture.alcb.getEthFromTokensBurn(
      poolBalanceBefore,
      totalSupplyBefore,
      alcbs
    )
  ).to.be.equal(eth);
};

describe("Testing ALCB conversion functions", async () => {
  let fixture: Fixture;
  let admin: SignerWithAddress;

  beforeEach(async function () {
    fixture = await loadFixture(getFixture);
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
    await checkConversionAndBurnALCB(fixture, 1, admin);
    await checkConversionAndBurnALCB(fixture, 9, admin);
    await checkConversionAndBurnALCB(fixture, 11, admin);
    await checkConversionAndBurnALCB(fixture, 75, admin);
    await checkConversionAndBurnALCB(fixture, 143, admin);
    await checkConversionAndBurnALCB(fixture, 555, admin);
    await checkConversionAndBurnALCB(fixture, 1986, admin);
    await checkConversionAndBurnALCB(fixture, 2200, admin);
    await checkConversionAndBurnALCB(fixture, 10001, admin);
  });
});
