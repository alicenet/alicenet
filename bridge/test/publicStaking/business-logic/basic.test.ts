import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BytesLike } from "ethers";
import { ethers, network } from "hardhat";
import { expect } from "../../chai-setup";
import {
  BaseTokensFixture,
  getBaseTokensFixture,
  getContractAddressFromDeployedRawEvent,
} from "../../setup";
import { getPosition, newPosition } from "../setup";

describe("PublicStaking: Basics", async () => {
  let fixture: BaseTokensFixture;
  let adminSigner: SignerWithAddress;
  let otherSigner: SignerWithAddress;
  let blockNumber: bigint;

  async function deployFixture() {
    const fixture = await getBaseTokensFixture();
    const [adminSigner, otherSigner] = await ethers.getSigners();
    await fixture.alca.approve(fixture.publicStaking.address, 1000);
    const tx = await fixture.publicStaking.connect(adminSigner).mint(1000);
    const blockNumber = BigInt(tx.blockNumber as number);
    return { fixture, adminSigner, otherSigner, blockNumber };
  }

  beforeEach(async function () {
    ({ fixture, adminSigner, otherSigner, blockNumber } = await loadFixture(
      deployFixture
    ));
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
    const deployData = (
      await ethers.getContractFactory("PublicStaking")
    ).getDeployTransaction().data as BytesLike;
    await network.provider.send("evm_setBlockGasLimit", ["0x3000000000000000"]);
    const publicStakingAddress = await getContractAddressFromDeployedRawEvent(
      await fixture.factory.deployCreate(deployData)
    );
    const publicStaking = await ethers.getContractAt(
      "PublicStaking",
      publicStakingAddress
    );
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

  it("Should return correct owner address for valid position", async function () {
    expect(await fixture.publicStaking.ownerOf(1)).to.be.equal(
      adminSigner.address
    );
  });

  it("Should return correct token id for index of owned token", async function () {
    const index = 0;
    const expectedTokenId = 1;
    expect(
      await fixture.publicStaking.tokenOfOwnerByIndex(
        adminSigner.address,
        index
      )
    ).to.be.equal(expectedTokenId);
  });

  it("Should revert for incorrect token id for index out of owned token bounds", async function () {
    const index = 1;

    await expect(
      fixture.publicStaking.tokenOfOwnerByIndex(adminSigner.address, index)
    ).to.be.revertedWith("ERC721Enumerable: owner index out of bounds");
  });

  it("Should not be able to get a position that doesn't exist", async function () {
    await expect(fixture.publicStaking.getPosition(2))
      .to.be.revertedWithCustomError(fixture.publicStaking, "InvalidTokenId")
      .withArgs(2);
  });

  describe("With multiple tokens minted", async function () {
    beforeEach(async function () {
      await fixture.alca.approve(fixture.publicStaking.address, 3000);
      await fixture.publicStaking.connect(adminSigner).mint(1000);
      await fixture.publicStaking.connect(adminSigner).mint(1000);
      await fixture.publicStaking.connect(adminSigner).mint(1000);

      await fixture.publicStaking
        .connect(adminSigner)
        .transferFrom(adminSigner.address, otherSigner.address, 1);
    });

    it("Should return correct balance for address", async function () {
      const expectedAdminSignerBalance = 3;
      expect(
        await fixture.publicStaking.balanceOf(adminSigner.address)
      ).to.be.equal(expectedAdminSignerBalance);

      const expectedOtherSignerBalance = 1;
      expect(
        await fixture.publicStaking.balanceOf(otherSigner.address)
      ).to.be.equal(expectedOtherSignerBalance);
    });

    it("Should return correct token id by index of token owned by address", async function () {
      const adminBalance = 3;
      const expectedTokenIds = [4, 2, 3];
      for (let i = 0; i < adminBalance; i++) {
        expect(
          await fixture.publicStaking.tokenOfOwnerByIndex(
            adminSigner.address,
            i
          )
        ).to.be.equal(expectedTokenIds[i]);
      }

      const index = 0;
      const expectedTokenId = 1;
      expect(
        await fixture.publicStaking.tokenOfOwnerByIndex(
          otherSigner.address,
          index
        )
      ).to.be.equal(expectedTokenId);
    });
  });
});
