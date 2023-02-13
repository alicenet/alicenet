import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BigNumber, BigNumberish } from "ethers";
import { ethers } from "hardhat";
import {
  PublicStaking,
  StakingPositionDescriptor,
} from "../../typechain-types";
import { expect } from "../chai-setup";
import {
  Fixture,
  getFixture,
  getStakingSVG,
  getTokenIdFromTx,
  getValidatorEthAccount,
} from "../setup";

describe("StakingPositionDescriptor: Tests StakingPositionDescriptor methods", async () => {
  let fixture: Fixture;
  let publicStaking: PublicStaking;
  let stakingPositionDescriptor: StakingPositionDescriptor;
  const stakeAmount = 20000;
  const stakeAmountALCAWei = ethers.utils.parseUnits(
    stakeAmount.toString(),
    18
  );
  const lockTime = 1;
  let tokenId: BigNumberish;

  async function deployFixture() {
    const fixture = await getFixture(true, true);

    const [admin] = fixture.namedSigners;
    const adminSigner = await getValidatorEthAccount(admin.address);

    const publicStaking = fixture.publicStaking;
    const stakingPositionDescriptor = fixture.stakingPositionDescriptor;

    await fixture.alca.approve(
      fixture.publicStaking.address,
      BigNumber.from(stakeAmountALCAWei)
    );
    const tx = await fixture.publicStaking
      .connect(adminSigner)
      .mintTo(admin.address, stakeAmountALCAWei, lockTime);
    const tokenId = await getTokenIdFromTx(tx);
    return {
      fixture,
      adminSigner,
      publicStaking,
      stakingPositionDescriptor,
      tokenId,
    };
  }

  beforeEach(async function () {
    ({ fixture, publicStaking, stakingPositionDescriptor, tokenId } =
      await loadFixture(deployFixture));
  });

  it("Fails if token at id does not exist", async function () {
    const invalidTokenId = 1234;

    await expect(publicStaking.tokenURI(invalidTokenId))
      .to.be.revertedWithCustomError(fixture.publicStaking, "InvalidTokenId")
      .withArgs(invalidTokenId);
  });

  describe("Given valid token id", async () => {
    let positionData;
    let svg: string;
    let tokenUriJson: string;
    let expectedTokenUriData: string;
    beforeEach(async function () {
      positionData = await publicStaking.getPosition(tokenId);

      svg = getStakingSVG(
        positionData.shares.toString(),
        positionData.freeAfter.toString(),
        positionData.withdrawFreeAfter.toString()
      );

      tokenUriJson =
        `{"name":"AliceNet Staked Token For Position #${tokenId.toString()}",` +
        ` "description":"This NFT represents a staked position on AliceNet. The owner of this NFT can modify or redeem the position.",` +
        ` "attributes": [` +
        `{"trait_type": "Shares", "value": "${positionData.shares.toString()}"},` +
        `{"trait_type": "Free After", "value": "${positionData.freeAfter.toString()}"},` +
        `{"trait_type": "Withdraw Free After", "value": "${positionData.withdrawFreeAfter.toString()}"},` +
        `{"trait_type": "Accumulator Eth", "value": "${positionData.accumulatorEth.toString()}"},` +
        `{"trait_type": "Accumulator Token", "value": "${positionData.accumulatorToken.toString()}"},` +
        `{"trait_type": "Token ID", "value": "${tokenId.toString()}"}` +
        `], "image_data": "data:image/svg+xml;base64,${Buffer.from(
          svg
        ).toString("base64")}"}`;

      expectedTokenUriData = `data:application/json;utf8,${tokenUriJson}`;
    });

    it("StakingPositionDescriptor should return correct token uri", async function () {
      const tokenUri = await stakingPositionDescriptor.tokenURI(
        publicStaking.address,
        tokenId
      );

      const parsedJson = JSON.parse(
        tokenUri.replace("data:application/json;utf8,", "")
      );

      expect(tokenUri).to.be.equal(expectedTokenUriData);
      expect(
        Buffer.from(
          parsedJson.image_data.replace("data:image/svg+xml;base64,", ""),
          "base64"
        ).toString()
      ).to.be.equal(svg);
    });

    it("PublicStaking contract Should return correct token uri", async function () {
      const tokenUri = await publicStaking.tokenURI(tokenId);

      const parsedJson = JSON.parse(
        tokenUri.replace("data:application/json;utf8,", "")
      );

      expect(tokenUri).to.be.equal(expectedTokenUriData);
      expect(
        Buffer.from(
          parsedJson.image_data.replace("data:image/svg+xml;base64,", ""),
          "base64"
        ).toString()
      ).to.be.equal(svg);
    });
  });
});
