import { ethers } from "hardhat";
import { StakingDescriptorMock } from "../../typechain-types";
import { expect } from "../chai-setup";

describe("stakingDescriptor: Tests stakingDescriptor methods", async () => {
  let stakingDescriptor: StakingDescriptorMock;

  beforeEach(async function () {
    const stakingDescriptorFactory = await ethers.getContractFactory(
      "StakingDescriptorMock"
    );
    stakingDescriptor = await stakingDescriptorFactory.deploy();
    await stakingDescriptor.deployed();
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
    const expectedTokenUriData =
      'data:application/json;utf8,{"name":"AliceNet Staked Token For Position #123", "description":"This NFT represents a staked position on AliceNet. The owner of this NFT can modify or redeem the position.", "attributes": [{"trait_type": "Shares", "value": "456"},{"trait_type": "Free After", "value": "789"},{"trait_type": "Withdraw Free After", "value": "987"},{"trait_type": "Accumulator Eth", "value": "123456789"},{"trait_type": "Accumulator Token", "value": "123456789"},{"trait_type": "Token ID", "value": "123"}], "image_data": "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNTAwIiBoZWlnaHQ9IjUwMCIgdmlld0JveD0iMCAwIDUwMCA1MDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9J2h0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsnPjx0ZXh0IHg9JzEwJyB5PScyMCc+U2hhcmVzOiA0NTY8L3RleHQ+PHRleHQgeD0nMTAnIHk9JzQwJz5GcmVlIGFmdGVyOiA3ODk8L3RleHQ+PHRleHQgeD0nMTAnIHk9JzYwJz5XaXRoZHJhdyBGcmVlIEFmdGVyOiA5ODc8L3RleHQ+PHRleHQgeD0nMTAnIHk9JzgwJz5BY2N1bXVsYXRvciAoRVRIKTogMTIzNDU2Nzg5PC90ZXh0Pjx0ZXh0IHg9JzEwJyB5PScxMDAnPkFjY3VtdWxhdG9yIChUb2tlbik6IDEyMzQ1Njc4OTwvdGV4dD48L3N2Zz4="}';
    const tokenUri = await stakingDescriptor.constructTokenURI(inputData);

    await expect(tokenUri).to.be.equal(expectedTokenUriData);
  });
});
