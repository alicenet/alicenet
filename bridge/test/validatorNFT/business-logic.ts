import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumberish } from "ethers";
import { ethers } from "hardhat";
import { ValidatorPoolMock } from "../../typechain-types";
import { expect } from "../chai-setup";
import { Fixture, getFixture, mineBlocks } from "../setup";

describe("ValidatorNFT: Tests ValidatorNFT Business Logic methods", async () => {
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
    await fixture.madToken.approve(validatorPool.address, amount);
  });

  it("Should mint a token and sender should be the payer and owner", async function () {
    const madBalanceBefore = await fixture.madToken.balanceOf(
      adminSigner.address
    );
    const nftBalanceBefore = await fixture.validatorNFT.balanceOf(
      validatorPool.address
    );
    const rcpt = await (await validatorPool.mintValidatorNFT()).wait();
    expect(rcpt.status).to.be.equal(1);
    expect(await fixture.validatorNFT.ownerOf(1)).to.be.eq(
      validatorPool.address
    );
    expect(
      await fixture.validatorNFT // NFT +1
        .balanceOf(validatorPool.address)
    ).to.equal(nftBalanceBefore.add(1));
    expect(
      await fixture.madToken // MAD -= amount
        .balanceOf(adminSigner.address)
    ).to.equal(madBalanceBefore.sub(amount));
  });

  it("Should burn a token and sender should receive funds", async function () {
    let tx = await validatorPool.mintValidatorNFT();
    const rcpt = await tx.wait();
    expect(rcpt.status).to.be.equal(1);
    await mineBlocks(1n);
    const nftBalanceBefore = await fixture.validatorNFT.balanceOf(
      validatorPool.address
    );
    const madBalanceBefore = await fixture.madToken.balanceOf(
      validatorPool.address
    );
    tx = await validatorPool.burnValidatorNFT(1);
    expect((await tx.wait()).status).to.be.equal(1);
    expect(await fixture.madToken.balanceOf(validatorPool.address)).to.be.eq(
      amount
    );
    expect(
      await fixture.validatorNFT // NFT -1
        .balanceOf(validatorPool.address)
    ).to.equal(nftBalanceBefore.sub(1));
    expect(
      await fixture.madToken // MAD +=amount
        .balanceOf(validatorPool.address)
    ).to.equal(madBalanceBefore.add(amount));
  });

  it("Should mint a token to an address and send staking funds from sender address", async function () {
    const madBalanceBefore = await fixture.madToken.balanceOf(
      adminSigner.address
    );
    const nftBalanceBefore = await fixture.validatorNFT.balanceOf(
      notAdminSigner.address
    );
    const rcpt = await (
      await validatorPool.mintToValidatorNFT(notAdminSigner.address)
    ).wait();
    expect(rcpt.status).to.be.equal(1);
    expect(
      await fixture.validatorNFT // NFT +1
        .balanceOf(notAdminSigner.address)
    ).to.equal(nftBalanceBefore.add(1));
    expect(
      await fixture.madToken // MAD -= amount
        .balanceOf(adminSigner.address)
    ).to.equal(madBalanceBefore.sub(amount));
  });

  it("Should burn a token from an address and return staking funds", async function () {
    let tx = await validatorPool.mintValidatorNFT();
    const rcpt = await tx.wait();
    expect(rcpt.status).to.be.equal(1);
    await mineBlocks(1n);
    const nftBalanceBefore = await fixture.validatorNFT.balanceOf(
      validatorPool.address
    );
    const madBalanceBefore = await fixture.madToken.balanceOf(
      notAdminSigner.address
    );
    tx = await validatorPool.burnToValidatorNFT(1, notAdminSigner.address);
    expect((await tx.wait()).status).to.be.equal(1);
    expect(
      await fixture.validatorNFT // NFT -1
        .balanceOf(validatorPool.address)
    ).to.equal(nftBalanceBefore.sub(1));
    expect(
      await fixture.madToken // MAD +=amount
        .balanceOf(notAdminSigner.address)
    ).to.equal(madBalanceBefore.add(amount));
  });
});
