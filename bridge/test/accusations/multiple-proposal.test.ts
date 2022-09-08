import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { Fixture, getFixture } from "../setup";
//import { getState, init, state } from "./setup";

describe("Testing Multiple Proposal Accusation", async () => {
  let user: SignerWithAddress;
  let admin: SignerWithAddress;
  const amount = 1000;
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture();
    [admin, user] = await ethers.getSigners();
  });

  describe("Testing invalid accusation", async () => {
    it("Should revert on invalid accusation", async function () {
      // await expect(
      //   fixture.accusations.accuseMultipleProposal("aaa", pclaims0, sig1, pclaim1)
      // ).to.be.revertedWith("2013");
    });
  });
});
