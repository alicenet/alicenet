import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import { BaseTokensFixture, getBaseTokensFixture } from "../../setup";
import { getPosition, newPosition } from "../setup";

describe("StakeNFT: Basics", async () => {
  let fixture: BaseTokensFixture;
  let adminSigner: SignerWithAddress;
  let blockNumber: bigint;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner] = await ethers.getSigners();
    await fixture.madToken.approve(fixture.stakeNFT.address, 1000);
    const tx = await fixture.stakeNFT.connect(adminSigner).mint(1000);
    blockNumber = BigInt(tx.blockNumber as number);
  });
  it("Check ERC721 name and symbol", async function () {
    expect(await fixture.stakeNFT.name()).to.be.equals("MNSNFT");
    expect(await fixture.stakeNFT.symbol()).to.be.equals("MNS");
  });

  it("Should be able to get information about a valid position", async function () {
    const expectedPosition = newPosition(
      1000n,
      blockNumber + 1n,
      blockNumber + 1n,
      0n,
      0n
    );
    expect(await getPosition(fixture.stakeNFT, 1)).to.be.deep.equals(
      expectedPosition
    );
  });

  it("Should not be able to get a position that doesn't exist", async function () {
    await expect(fixture.stakeNFT.getPosition(2)).to.be.rejectedWith(
      "StakeNFT: Token ID doesn't exist!"
    );
  });
});
