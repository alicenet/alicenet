import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { StakingSVGMock } from "../../typechain-types";
import { expect } from "../chai-setup";

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
    const expectedSvgData = `<svg width="500" height="500" viewBox="0 0 500 500" xmlns="http://www.w3.org/2000/svg" xmlns:xlink='http://www.w3.org/1999/xlink'><text x='10' y='20'>Shares: ${inputData.shares}</text><text x='10' y='40'>Free after: ${inputData.freeAfter}</text><text x='10' y='60'>Withdraw Free After: ${inputData.withdrawFreeAfter}</text><text x='10' y='80'>Accumulator (ETH): ${inputData.accumulatorEth}</text><text x='10' y='100'>Accumulator (Token): ${inputData.accumulatorToken}</text></svg>`;
    const generatedSVG = await stakingNFTSVG.generateSVG(inputData);

    await expect(generatedSVG).to.be.equal(expectedSvgData);
  });
});
