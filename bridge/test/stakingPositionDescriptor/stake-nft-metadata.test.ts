import { BigNumber, BigNumberish, Signer } from "ethers";
import { ethers } from "hardhat";
import {
  PublicStaking,
  StakingPositionDescriptor,
} from "../../typechain-types";
import { expect } from "../chai-setup";
import {
  Fixture,
  getFixture,
  getTokenIdFromTx,
  getValidatorEthAccount,
} from "../setup";

describe("StakingPositionDescriptor: Tests StakingPositionDescriptor methods", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let publicStaking: PublicStaking;
  let stakingPositionDescriptor: StakingPositionDescriptor;
  const stakeAmount = 20000;
  const stakeAmountATokenWei = ethers.utils.parseUnits(
    stakeAmount.toString(),
    18
  );
  const lockTime = 1;
  let tokenId: BigNumberish;

  beforeEach(async function () {
    fixture = await getFixture(true, true);

    const [admin] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);

    publicStaking = fixture.publicStaking;
    stakingPositionDescriptor = fixture.stakingPositionDescriptor;

    await fixture.aToken.approve(
      fixture.publicStaking.address,
      BigNumber.from(stakeAmountATokenWei)
    );
    const tx = await fixture.publicStaking
      .connect(adminSigner)
      .mintTo(admin.address, stakeAmountATokenWei, lockTime);
    tokenId = await getTokenIdFromTx(tx);
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

      svg =
        `<svg width="500" height="500" viewBox="0 0 500 500" xmlns="http://www.w3.org/2000/svg" xmlns:xlink='http://www.w3.org/1999/xlink'>` +
        `<text x='10' y='20'>Shares: ${positionData.shares.toString()}</text>` +
        `<text x='10' y='40'>Free after: ${positionData.freeAfter.toString()}</text>` +
        `<text x='10' y='60'>Withdraw Free After: ${positionData.withdrawFreeAfter.toString()}</text>` +
        `<text x='10' y='80'>Accumulator (ETH): ${positionData.accumulatorEth.toString()}</text>` +
        `<text x='10' y='100'>Accumulator (Token): ${positionData.accumulatorToken.toString()}</text></svg>`;

      tokenUriJson =
        `{"name":"AliceNet Staked token for position #1", ` +
        `"description":"This NFT represents a staked position on AliceNet.` +
        `\\nThe owner of this NFT can modify or redeem the position.` +
        `\\n Shares: ${positionData.shares.toString()}` +
        `\\nFree After: ${positionData.freeAfter.toString()}` +
        `\\nWithdraw Free After: ${positionData.withdrawFreeAfter.toString()}` +
        `\\nAccumulator Eth: ${positionData.accumulatorEth.toString()}` +
        `\\nAccumulator Token: ${positionData.accumulatorToken.toString()}` +
        `\\nToken ID: ${tokenId.toString()}", ` +
        `"image": "data:image/svg+xml;base64,${btoa(svg)}"}`;

      expectedTokenUriData = `data:application/json;base64,${btoa(
        tokenUriJson
      )}`;
    });

    it("StakingPositionDescriptor should return correct token uri", async function () {
      const tokenUri = await stakingPositionDescriptor.tokenURI(
        publicStaking.address,
        tokenId
      );

      const parsedJson = JSON.parse(
        atob(tokenUri.replace("data:application/json;base64,", ""))
      );

      await expect(tokenUri).to.be.equal(expectedTokenUriData);
      await expect(
        atob(parsedJson.image.replace("data:image/svg+xml;base64,", ""))
      ).to.be.equal(svg);
    });

    it("PublicStaking contract Should return correct token uri", async function () {
      const tokenUri = await publicStaking.tokenURI(tokenId);

      const parsedJson = JSON.parse(
        atob(tokenUri.replace("data:application/json;base64,", ""))
      );

      await expect(tokenUri).to.be.equal(expectedTokenUriData);
      await expect(
        atob(parsedJson.image.replace("data:image/svg+xml;base64,", ""))
      ).to.be.equal(svg);
    });
  });
});
