import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { StakingSVGMock } from "../../typechain-types";
import { expect } from "../chai-setup";
import { getStakingSVG } from "../setup";

describe("StakingSVG: Tests StakingSVG library methods", async () => {
  let stakingNFTSVG: StakingSVGMock;

  async function deployFixture() {
    const StakingSVGFactory = await ethers.getContractFactory("StakingSVGMock");
    const stakingNFTSVG = await StakingSVGFactory.deploy();
    await stakingNFTSVG.deployed();
    return { stakingNFTSVG };
  }

  beforeEach(async function () {
    ({ stakingNFTSVG } = await loadFixture(deployFixture));
  });

  it("Should return correct token uri", async function () {
    const inputData = {
      shares: "shares value",
      freeAfter: "freeAfter value",
      withdrawFreeAfter: "withdrawFreeAfter value",
      accumulatorEth: "accumulatorEth value",
      accumulatorToken: "accumulatorToken value",
    };
    const expectedSvgData = getStakingSVG(
      inputData.shares,
      inputData.freeAfter,
      inputData.withdrawFreeAfter
    );
    const generatedSVG = await stakingNFTSVG.generateSVG(inputData);
    await expect(generatedSVG).to.be.equal(expectedSvgData);
  });
});
