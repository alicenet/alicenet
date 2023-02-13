import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumberish } from "ethers";
import { ethers } from "hardhat";
import { ValidatorPoolMock } from "../../typechain-types";
import { expect } from "../chai-setup";
import { Fixture, getFixture, getStakingSVG, mineBlocks } from "../setup";

describe("ValidatorStaking: Tests ValidatorStaking Business Logic methods", async () => {
  let fixture: Fixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;
  let amount: BigNumberish;
  let validatorPool: ValidatorPoolMock;

  async function deployFixture() {
    const fixture = await getFixture(true, true);
    const [admin, notAdmin] = fixture.namedSigners;
    const adminSigner = await ethers.getSigner(admin.address);
    const notAdminSigner = await ethers.getSigner(notAdmin.address);
    const validatorPool = fixture.validatorPool as ValidatorPoolMock;
    const amount = await validatorPool.getStakeAmount();
    await fixture.alca.approve(validatorPool.address, amount);
    return {
      fixture,
      adminSigner,
      notAdminSigner,
      validatorPool,
      amount,
    };
  }

  beforeEach(async function () {
    ({ fixture, adminSigner, notAdminSigner, validatorPool, amount } =
      await loadFixture(deployFixture));
  });

  it("Should mint a token and sender should be the payer and owner", async function () {
    const tokenBalanceBefore = await fixture.alca.balanceOf(
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
      await fixture.alca // ATK -= amount
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
    const tokenBalanceBefore = await fixture.alca.balanceOf(
      validatorPool.address
    );
    tx = await validatorPool.burnValidatorStaking(1);
    expect((await tx.wait()).status).to.be.equal(1);
    expect(await fixture.alca.balanceOf(validatorPool.address)).to.be.eq(
      amount
    );
    expect(
      await fixture.validatorStaking // NFT -1
        .balanceOf(validatorPool.address)
    ).to.equal(nftBalanceBefore.sub(1));
    expect(
      await fixture.alca // ATK +=amount
        .balanceOf(validatorPool.address)
    ).to.equal(tokenBalanceBefore.add(amount));
  });

  it("Should mint a token to an address and send staking funds from sender address", async function () {
    const tokenBalanceBefore = await fixture.alca.balanceOf(
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
      await fixture.alca // ATK -= amount
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
    const tokenBalanceBefore = await fixture.alca.balanceOf(
      notAdminSigner.address
    );
    tx = await validatorPool.burnToValidatorStaking(1, notAdminSigner.address);
    expect((await tx.wait()).status).to.be.equal(1);
    expect(
      await fixture.validatorStaking // NFT -1
        .balanceOf(validatorPool.address)
    ).to.equal(nftBalanceBefore.sub(1));
    expect(
      await fixture.alca // ATK +=amount
        .balanceOf(notAdminSigner.address)
    ).to.equal(tokenBalanceBefore.add(amount));
  });

  it("Should return correct token uri", async function () {
    const tx = await validatorPool.mintValidatorStaking();
    await tx.wait();

    const tokenId = 1;
    const positionData = await fixture.validatorStaking.getPosition(tokenId);

    const svg = getStakingSVG(
      positionData.shares.toString(),
      positionData.freeAfter.toString(),
      positionData.withdrawFreeAfter.toString()
    );

    const tokenUriJson =
      `{"name":"AliceNet Staked Token For Position #${tokenId.toString()}",` +
      ` "description":"This NFT represents a staked position on AliceNet. The owner of this NFT can modify or redeem the position.",` +
      ` "attributes": [` +
      `{"trait_type": "Shares", "value": "${positionData.shares.toString()}"},` +
      `{"trait_type": "Free After", "value": "${positionData.freeAfter.toString()}"},` +
      `{"trait_type": "Withdraw Free After", "value": "${positionData.withdrawFreeAfter.toString()}"},` +
      `{"trait_type": "Accumulator Eth", "value": "${positionData.accumulatorEth.toString()}"},` +
      `{"trait_type": "Accumulator Token", "value": "${positionData.accumulatorToken.toString()}"},` +
      `{"trait_type": "Token ID", "value": "${tokenId.toString()}"}` +
      `], "image_data": "data:image/svg+xml;base64,${Buffer.from(svg).toString(
        "base64"
      )}"}`;

    const expectedTokenUriData = `data:application/json;utf8,${tokenUriJson}`;

    const tokenUri = await fixture.validatorStaking.tokenURI(tokenId);

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
