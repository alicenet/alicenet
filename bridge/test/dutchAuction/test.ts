import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";

import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";
import { deployStaticWithFactory, Fixture, getFixture } from "../setup";
import { getState, showState } from "./setup";

let startingPrice = ethers.utils.parseEther("10.0");
let discountRate = BigNumber.from(1);
let admin: SignerWithAddress;
let user: SignerWithAddress;
let fixture: Fixture;
let dutchAuction: DutchAuction;

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    await network.provider.send("evm_setAutomine", [true]);
    fixture = await getFixture(false, false, false);
    [admin, user] = await ethers.getSigners();
    dutchAuction = (await deployStaticWithFactory(
      fixture.factory,
      "DutchAuction",
      "DutchAuction",
      [startingPrice, discountRate]
    )) as DutchAuction;
    showState("Initial", await getState(fixture));
  });

  it("Should get lower bid prices through time", async () => {
    const initialPrice = await dutchAuction.getPrice();
    await network.provider.send("evm_mine"); // Mine one block
    await network.provider.send("evm_mine"); // Mine another
    const currentPrice = await dutchAuction.getPrice();
    expect(currentPrice).to.be.equal(initialPrice.sub(discountRate.mul(2)));
  });

  it("Should to start over dutch auction", async () => {
    const initialPrice = await dutchAuction.getPrice();
    await network.provider.send("evm_mine"); // Mine one block
    await network.provider.send("evm_mine"); // Mine another
    await dutchAuction.resetAuction();
    const currentPrice = await dutchAuction.getPrice();
    expect(currentPrice).to.be.equal(initialPrice);
  });
});
