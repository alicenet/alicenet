import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { deployUpgradeableWithFactory, Fixture, getFixture } from "../setup";
import { getState, showState } from "./setup";
import {DutchAuction
} from "../../typechain-types";

let startingPrice = 100
let discountRate = 1
let buyPrice = 50

describe("Testing Dutch Auction", async () => {
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture(false,false,false);
    showState("Initial", await getState(fixture));
    const dutchAuction = (await deployUpgradeableWithFactory(
      fixture.factory,
      "DutchAuction",
      "DutchAuction",
      undefined,
      [],
    )) as DutchAuction
  });

  it("Should start a Dutch Auction", async () => {
  });

});
