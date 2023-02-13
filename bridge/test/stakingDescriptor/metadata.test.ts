import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { StakingDescriptorMock } from "../../typechain-types";
import { expect } from "../chai-setup";
import { getStakingSVGBase64 } from "../setup";

describe("stakingDescriptor: Tests stakingDescriptor methods", async () => {
  let stakingDescriptor: StakingDescriptorMock;

  async function deployFixture() {
    const stakingDescriptorFactory = await ethers.getContractFactory(
      "StakingDescriptorMock"
    );
    const stakingDescriptor = await stakingDescriptorFactory.deploy();
    await stakingDescriptor.deployed();
    return { stakingDescriptor };
  }

  beforeEach(async function () {
    ({ stakingDescriptor } = await loadFixture(deployFixture));
  });

  it("Should return correct token uri", async function () {
    const inputData = {
      tokenId: 123,
      shares: 456,
      freeAfter: 789,
      withdrawFreeAfter: 987,
      accumulatorEth: 123456789,
      accumulatorToken: 123456789,
    };
    const expectedTokenUriData = `data:application/json;utf8,{"name":"AliceNet Staked Token For Position #123", "description":"This NFT represents a staked position on AliceNet. The owner of this NFT can modify or redeem the position.", "attributes": [{"trait_type": "Shares", "value": "456"},{"trait_type": "Free After", "value": "789"},{"trait_type": "Withdraw Free After", "value": "987"},{"trait_type": "Accumulator Eth", "value": "123456789"},{"trait_type": "Accumulator Token", "value": "123456789"},{"trait_type": "Token ID", "value": "123"}], "image_data": "data:image/svg+xml;base64,${getStakingSVGBase64(
      inputData.shares.toString(),
      inputData.freeAfter.toString(),
      inputData.withdrawFreeAfter.toString()
    )}"}`;
    const tokenUri = await stakingDescriptor.constructTokenURI(inputData);

    await expect(tokenUri).to.be.equal(expectedTokenUriData);
  });
});
