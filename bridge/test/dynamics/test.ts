import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { assertErrorMessage, assertError } from "../chai-helpers";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";
import { Dynamics } from "../../typechain-types";
import { assert } from "console";
import { ConfigurationStructOutput } from "../../typechain-types/contracts/Dynamics.sol/Dynamics";

describe("Testing Dynamics methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let dynamics: Dynamics;
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture(false, true, false);
    const signers = await ethers.getSigners();
    [admin, user] = signers;
    dynamics = fixture.dynamics;
  });

  it("Should get initial Configuration", async () => {
    const configuration = await dynamics.getConfiguration();
    expect(configuration[0]).to.be.equal(BigNumber.from(2))
    expect(configuration[1]).to.be.equal(BigNumber.from(336))
  });


});
