import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { ERC721ReceiverAccount } from "../../../typechain-types";
import { expect } from "../../chai-setup";
import { BaseTokensFixture, getBaseTokensFixture } from "../../setup";

describe("PublicStaking: NFT transfer", async () => {
  let fixture: BaseTokensFixture;
  let adminSigner: SignerWithAddress;
  let notAdminSigner: SignerWithAddress;
  let erc721ReceiverContract: ERC721ReceiverAccount;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
    await fixture.aToken.approve(fixture.publicStaking.address, 2000);
    await fixture.publicStaking.connect(adminSigner).mint(1000);
    erc721ReceiverContract = await (
      await ethers.getContractFactory("ERC721ReceiverAccount")
    ).deploy();
  });

  it("Safe transfer Staking NFT position to user", async function () {
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      adminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    await fixture.publicStaking["safeTransferFrom(address,address,uint256)"](
      adminSigner.address,
      notAdminSigner.address,
      1
    );
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      notAdminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(1n);
  });

  it("Transfer Staking NFT position to user", async function () {
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      adminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    await fixture.publicStaking.transferFrom(
      adminSigner.address,
      notAdminSigner.address,
      1
    );
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      notAdminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(1n);
  });

  it("Approval should be able to safe transfer Staking NFT position", async function () {
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      adminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    await fixture.publicStaking.approve(notAdminSigner.address, 1);
    await fixture.publicStaking
      .connect(notAdminSigner)
      .transferFrom(adminSigner.address, notAdminSigner.address, 1);
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      notAdminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(1n);
  });

  it("Approval for one position should not be able to transfer other Staking NFT position", async function () {
    await fixture.publicStaking.connect(adminSigner).mint(1000);
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      adminSigner.address.toLowerCase()
    );
    expect((await fixture.publicStaking.ownerOf(2)).toLowerCase()).to.be.equals(
      adminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(2n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    await fixture.publicStaking.approve(notAdminSigner.address, 1);

    // transfer valid position
    await fixture.publicStaking
      .connect(notAdminSigner)
      .transferFrom(adminSigner.address, notAdminSigner.address, 1);
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      notAdminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    await expect(
      fixture.publicStaking
        .connect(notAdminSigner)
        .transferFrom(adminSigner.address, notAdminSigner.address, 2)
    ).to.be.revertedWith("ERC721: transfer caller is not owner nor approved");
  });

  it("Approval for all should be able to safe transfer Staking NFT position", async function () {
    await fixture.publicStaking.connect(adminSigner).mint(1000);
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      adminSigner.address.toLowerCase()
    );
    expect((await fixture.publicStaking.ownerOf(2)).toLowerCase()).to.be.equals(
      adminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(2n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    await fixture.publicStaking.approve(notAdminSigner.address, 1);

    // transfer valid position
    await fixture.publicStaking
      .connect(notAdminSigner)
      .transferFrom(adminSigner.address, notAdminSigner.address, 1);
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      notAdminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    await fixture.publicStaking.transferFrom(
      adminSigner.address,
      notAdminSigner.address,
      2
    );

    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    expect(
      (await fixture.publicStaking.balanceOf(notAdminSigner.address)).toBigInt()
    ).to.be.equals(2n);
  });

  it("Safe transfer Staking NFT position to contract that implements ERC721Receiver interface", async function () {
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      adminSigner.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    expect(
      (
        await fixture.publicStaking.balanceOf(erc721ReceiverContract.address)
      ).toBigInt()
    ).to.be.equals(0n);

    await fixture.publicStaking["safeTransferFrom(address,address,uint256)"](
      adminSigner.address,
      erc721ReceiverContract.address,
      1
    );
    expect((await fixture.publicStaking.ownerOf(1)).toLowerCase()).to.be.equals(
      erc721ReceiverContract.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(0n);

    expect(
      (
        await fixture.publicStaking.balanceOf(erc721ReceiverContract.address)
      ).toBigInt()
    ).to.be.equals(1n);
  });

  it("Shouldn't allow safeTransferFrom/transferFrom without being approved or owner", async function () {
    await expect(
      fixture.publicStaking
        .connect(notAdminSigner)
        ["safeTransferFrom(address,address,uint256)"](
          adminSigner.address,
          notAdminSigner.address,
          1
        )
    ).to.be.rejectedWith("ERC721: transfer caller is not owner nor approved");
    await expect(
      fixture.publicStaking
        .connect(notAdminSigner)
        .transferFrom(adminSigner.address, notAdminSigner.address, 1)
    ).to.be.rejectedWith("ERC721: transfer caller is not owner nor approved");
  });

  it("Shouldn't allow safe transfer NFT to a contract that doesn't implement ERC721Receiver interface", async function () {
    await expect(
      fixture.publicStaking["safeTransferFrom(address,address,uint256)"](
        adminSigner.address,
        fixture.bToken.address,
        1
      )
    ).to.be.rejectedWith("ERC721: transfer to non ERC721Receiver implementer");
  });

  describe("After transfer", async () => {
    beforeEach(async function () {
      await fixture.publicStaking["safeTransferFrom(address,address,uint256)"](
        adminSigner.address,
        notAdminSigner.address,
        1
      );
    });
    it("Old owner shouldn't be able to lock or withdrawalLock position", async function () {
      await expect(
        fixture.publicStaking.lockOwnPosition(1, 100)
      ).to.be.rejectedWith(
        "Error, token doesn't exist or doesn't belong to the caller!"
      );

      await expect(
        fixture.publicStaking.lockWithdraw(1, 100)
      ).to.be.rejectedWith(
        "Error, token doesn't exist or doesn't belong to the caller!"
      );
    });

    it("Old owner shouldn't be able to burn or burnTo position", async function () {
      await expect(fixture.publicStaking.burn(1)).to.be.rejectedWith(
        "PublicStaking: User is not the owner of the tokenID!"
      );

      await expect(
        fixture.publicStaking.burnTo(notAdminSigner.address, 1)
      ).to.be.rejectedWith(
        "PublicStaking: User is not the owner of the tokenID!"
      );
    });

    it("Old owner shouldn't be able to collect profits position", async function () {
      await expect(fixture.publicStaking.collectEth(1)).to.be.rejectedWith(
        "PublicStaking: Error sender is not the owner of the tokenID!"
      );

      await expect(
        fixture.publicStaking.collectEthTo(notAdminSigner.address, 1)
      ).to.be.rejectedWith(
        "PublicStaking: Error sender is not the owner of the tokenID!"
      );

      await expect(fixture.publicStaking.collectToken(1)).to.be.rejectedWith(
        "PublicStaking: Error sender is not the owner of the tokenID!"
      );

      await expect(
        fixture.publicStaking.collectTokenTo(notAdminSigner.address, 1)
      ).to.be.rejectedWith(
        "PublicStaking: Error sender is not the owner of the tokenID!"
      );
    });

    it("New owner should be able to lock or withdrawalLock position", async function () {
      await fixture.publicStaking
        .connect(notAdminSigner)
        .lockOwnPosition(1, 100);
      await fixture.publicStaking.connect(notAdminSigner).lockWithdraw(1, 100);
    });

    it("New owner should be able to burn position", async function () {
      await fixture.publicStaking.connect(notAdminSigner).burn(1);
    });

    it("New owner should be able to burnTo position", async function () {
      await fixture.publicStaking
        .connect(notAdminSigner)
        .burnTo(adminSigner.address, 1);
    });

    it("New owner should be able to collect profits position", async function () {
      await fixture.publicStaking.connect(notAdminSigner).collectEth(1);
      await fixture.publicStaking
        .connect(notAdminSigner)
        .collectEthTo(notAdminSigner.address, 1);
      await fixture.publicStaking.connect(notAdminSigner).collectToken(1);
      await fixture.publicStaking
        .connect(notAdminSigner)
        .collectTokenTo(notAdminSigner.address, 1);
    });
  });
});
