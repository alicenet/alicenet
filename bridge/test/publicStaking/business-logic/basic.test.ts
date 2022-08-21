import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import { BaseTokensFixture, getBaseTokensFixture } from "../../setup";
import { getPosition, newPosition } from "../setup";

describe("PublicStaking: Basics", async () => {
  let fixture: BaseTokensFixture;
  let adminSigner: SignerWithAddress;
  let blockNumber: bigint;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner] = await ethers.getSigners();
    await fixture.aToken.approve(fixture.publicStaking.address, 1000);
    const tx = await fixture.publicStaking.connect(adminSigner).mint(1000);
    blockNumber = BigInt(tx.blockNumber as number);
  });

  it("Should not allow initialize more than once", async () => {
    await expect(
      fixture.factory.callAny(
        fixture.publicStaking.address,
        0,
        fixture.publicStaking.interface.encodeFunctionData("initialize")
      )
    ).to.revertedWith("Initializable: contract is already initialized");
  });

  it("Only factory should be allowed to call initialize", async () => {
    const publicStaking = await (
      await ethers.getContractFactory("PublicStaking")
    ).deploy();
    const [, user] = await ethers.getSigners();
    await expect(
      publicStaking.connect(user).initialize()
    ).to.revertedWithCustomError(publicStaking, "OnlyFactory");
  });

  it("Check ERC721 name and symbol", async function () {
    expect(await fixture.publicStaking.name()).to.be.equals("APSNFT");
    expect(await fixture.publicStaking.symbol()).to.be.equals("APS");
  });

  it("Should be able to get information about a valid position", async function () {
    const expectedPosition = newPosition(
      1000n,
      blockNumber + 1n,
      blockNumber + 1n,
      0n,
      0n
    );
    expect(await getPosition(fixture.publicStaking, 1)).to.be.deep.equals(
      expectedPosition
    );
  });

  it("Should not be able to get a position that doesn't exist", async function () {
    await expect(fixture.publicStaking.getPosition(2))
      .to.be.revertedWithCustomError(fixture.publicStaking, "InvalidTokenId")
      .withArgs(2);
  });
});
