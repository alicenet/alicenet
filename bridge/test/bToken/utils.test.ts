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

  it("Should not allow initialize more than once", async () => {
    await expect(
      fixture.factory.callAny(
        fixture.bToken.address,
        0,
        fixture.bToken.interface.encodeFunctionData("initialize")
      )
    ).to.revertedWith("Initializable: contract is already initialized");
  });

  it("Only factory should be allowed to call initialize", async () => {
    const bToken = await (await ethers.getContractFactory("BToken")).deploy();
    const [, user] = await ethers.getSigners();
    await expect(bToken.connect(user).initialize()).to.revertedWithCustomError(
      bToken,
      "OnlyFactory"
    );
  });

  it("Should calculate correct bounding curves", async () => {
    let eth = 10000;
    let bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(bTokens).to.be.equal(BigInt("944264568799449143863271"));
    eth = 20000;
    bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(bTokens).to.be.equal(BigInt("1020000000000000000000000"));
    eth = 25000;
    bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(
      ethers.utils.parseEther(eth.toString())
    );
    expect(bTokens).to.be.equal(BigInt("1026649288252135716968558"));
  });
});
