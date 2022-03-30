import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { factoryCallAnyFixture, Fixture, getFixture } from "../setup";
import { getState, showState } from "./setup";

describe("Testing BToken Settings", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin, user] = signers;
    showState("Initial", await getState(fixture));
    await factoryCallAnyFixture(fixture, "bToken", "setAdmin", [admin.address]);
  });

  it("Should fail to set split not being an admin", async () => {
    await expect(
      fixture.bToken.connect(user).setSplits(300, 300, 300, 100)
    ).to.be.revertedWith("Must be admin");
  });

  it("Should fail to set splits greater than one unit", async () => {
    await expect(
      fixture.bToken.connect(admin).setSplits(333, 333, 333, 2)
    ).to.be.revertedWith(
      "BToken: All the split values must sum to _PERCENTAGE_SCALE!"
    );
  });

  it("Should fail to set all splits to 0", async () => {
    await expect(
      fixture.bToken.connect(admin).setSplits(0, 0, 0, 0)
    ).to.be.revertedWith(
      "BToken: All the split values must sum to _PERCENTAGE_SCALE!"
    );
  });

  it("Should set some splits to 0", async () => {
    await fixture.bToken.connect(admin).setSplits(0, 0, 1000, 0);
  });

  it("Should correctly set the splits", async () => {
    await fixture.bToken.connect(admin).setSplits(300, 300, 300, 100);
  });
});
