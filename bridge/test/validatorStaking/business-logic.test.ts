import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumberish } from "ethers";
import { ethers } from "hardhat";
import { ValidatorPoolMock } from "../../typechain-types";
import { expect } from "../chai-setup";
import { Fixture, getFixture, mineBlocks } from "../setup";

describe("ValidatorStaking: Tests ValidatorStaking Business Logic methods", async () => {
  let fixture: Fixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;
  let amount: BigNumberish;
  let validatorPool: ValidatorPoolMock;

  beforeEach(async function () {
    fixture = await getFixture(true, true);
    const [admin, notAdmin] = fixture.namedSigners;
    adminSigner = await ethers.getSigner(admin.address);
    notAdminSigner = await ethers.getSigner(notAdmin.address);
    validatorPool = fixture.validatorPool as ValidatorPoolMock;
    amount = await validatorPool.getStakeAmount();
    await fixture.aToken.approve(validatorPool.address, amount);
  });

  it("Should mint a token and sender should be the payer and owner", async function () {
    const tokenBalanceBefore = await fixture.aToken.balanceOf(
      adminSigner.address
    );
    const nftBalanceBefore = await fixture.validatorStaking.balanceOf(
      validatorPool.address
    );
    const rcpt = await (await validatorPool.mintValidatorStaking()).wait();
    expect(rcpt.status).to.be.equal(1);
    expect(await fixture.validatorStaking.ownerOf(1)).to.be.eq(
      validatorPool.address
    );
    expect(
      await fixture.validatorStaking // NFT +1
        .balanceOf(validatorPool.address)
    ).to.equal(nftBalanceBefore.add(1));
    expect(
      await fixture.aToken // ATK -= amount
        .balanceOf(adminSigner.address)
    ).to.equal(tokenBalanceBefore.sub(amount));
  });

  it("Should burn a token and sender should receive funds", async function () {
    let tx = await validatorPool.mintValidatorStaking();
    const rcpt = await tx.wait();
    expect(rcpt.status).to.be.equal(1);
    await mineBlocks(1n);
    const nftBalanceBefore = await fixture.validatorStaking.balanceOf(
      validatorPool.address
    );
    const tokenBalanceBefore = await fixture.aToken.balanceOf(
      validatorPool.address
    );
    tx = await validatorPool.burnValidatorStaking(1);
    expect((await tx.wait()).status).to.be.equal(1);
    expect(await fixture.aToken.balanceOf(validatorPool.address)).to.be.eq(
      amount
    );
    expect(
      await fixture.validatorStaking // NFT -1
        .balanceOf(validatorPool.address)
    ).to.equal(nftBalanceBefore.sub(1));
    expect(
      await fixture.aToken // ATK +=amount
        .balanceOf(validatorPool.address)
    ).to.equal(tokenBalanceBefore.add(amount));
  });

  it("Should mint a token to an address and send staking funds from sender address", async function () {
    const tokenBalanceBefore = await fixture.aToken.balanceOf(
      adminSigner.address
    );
    const nftBalanceBefore = await fixture.validatorStaking.balanceOf(
      notAdminSigner.address
    );
    const rcpt = await (
      await validatorPool.mintToValidatorStaking(notAdminSigner.address)
    ).wait();
    expect(rcpt.status).to.be.equal(1);
    expect(
      await fixture.validatorStaking // NFT +1
        .balanceOf(notAdminSigner.address)
    ).to.equal(nftBalanceBefore.add(1));
    expect(
      await fixture.aToken // ATK -= amount
        .balanceOf(adminSigner.address)
    ).to.equal(tokenBalanceBefore.sub(amount));
  });

  it("Should burn a token from an address and return staking funds", async function () {
    let tx = await validatorPool.mintValidatorStaking();
    const rcpt = await tx.wait();
    expect(rcpt.status).to.be.equal(1);
    await mineBlocks(1n);
    const nftBalanceBefore = await fixture.validatorStaking.balanceOf(
      validatorPool.address
    );
    const tokenBalanceBefore = await fixture.aToken.balanceOf(
      notAdminSigner.address
    );
    tx = await validatorPool.burnToValidatorStaking(1, notAdminSigner.address);
    expect((await tx.wait()).status).to.be.equal(1);
    expect(
      await fixture.validatorStaking // NFT -1
        .balanceOf(validatorPool.address)
    ).to.equal(nftBalanceBefore.sub(1));
    expect(
      await fixture.aToken // ATK +=amount
        .balanceOf(notAdminSigner.address)
    ).to.equal(tokenBalanceBefore.add(amount));
  });

  it("Should return correct token uri", async function () {
    const tx = await validatorPool.mintValidatorStaking();
    await tx.wait();

    const tokenId = 1;
    const positionData = await fixture.validatorStaking.getPosition(tokenId);

    const svg =
      `<svg width="500" height="500" viewBox="0 0 500 500" xmlns="http://www.w3.org/2000/svg" xmlns:xlink='http://www.w3.org/1999/xlink'>` +
      `<text x='10' y='20'>Shares: ${positionData.shares.toString()}</text>` +
      `<text x='10' y='40'>Free after: ${positionData.freeAfter.toString()}</text>` +
      `<text x='10' y='60'>Withdraw Free After: ${positionData.withdrawFreeAfter.toString()}</text>` +
      `<text x='10' y='80'>Accumulator (ETH): ${positionData.accumulatorEth.toString()}</text>` +
      `<text x='10' y='100'>Accumulator (Token): ${positionData.accumulatorToken.toString()}</text></svg>`;

    const tokenUriJson =
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

    const expectedTokenUriData = `data:application/json;base64,${btoa(
      tokenUriJson
    )}`;

    const tokenUri = await fixture.validatorStaking.tokenURI(tokenId);

    const parsedJson = JSON.parse(
      atob(tokenUri.replace("state:application/json;base64,", ""))
    );

    await expect(tokenUri).to.be.equal(expectedTokenUriData);
    await expect(
      atob(parsedJson.image.replace("state:image/svg+xml;base64,", ""))
    ).to.be.equal(svg);
  });
});
