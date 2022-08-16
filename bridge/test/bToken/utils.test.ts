import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";
import { getState, showState } from "./setup";

describe("Testing BToken Utils methods", async () => {
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture();
    showState("Initial", await getState(fixture));
  });

  it("Should calculate correct bounding curves", async () => {
    let eth = 10000;
    let bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(bTokens).to.be.equal(BigInt("936764568799449143863271"));
    eth = 20000;
    bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(bTokens).to.be.equal(BigInt("1005000000000000000000000"));
    eth = 25000;
    bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(bTokens).to.be.equal(BigInt("1007899288252135716968558"));
  });
});
