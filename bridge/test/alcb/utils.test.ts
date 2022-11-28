import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";
import { getState, showState } from "./setup";

describe("Testing ALCB Utils methods", async () => {
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await loadFixture(getFixture);
    showState("Initial", await getState(fixture));
  });

  it("Should calculate correct bounding curves", async () => {
    let eth = 10000;
    let alcbs = await fixture.alcb.getLatestMintedALCBsFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(alcbs).to.be.equal(BigInt("944264568799449143863271"));
    eth = 20000;
    alcbs = await fixture.alcb.getLatestMintedALCBsFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(alcbs).to.be.equal(BigInt("1020000000000000000000000"));
    eth = 25000;
    alcbs = await fixture.alcb.getLatestMintedALCBsFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(alcbs).to.be.equal(BigInt("1026649288252135716968558"));
  });
});
