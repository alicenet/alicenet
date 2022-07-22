import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";

import { ethers, network } from "hardhat";
import { DutchAuction, ERC721Mock } from "../../typechain-types";
import { deployStaticWithFactory, Fixture, getFixture } from "../setup";
import { getState, showState, state } from "./setup";

let startingPrice =  ethers.utils.parseEther("10.0");;
let discountRate = BigNumber.from(1);
let buyPrice = BigNumber.from(5);
let nftTokenId = BigNumber.from(1);
let duration = BigNumber.from(10); // 10 seconds
let admin: SignerWithAddress;
let user: SignerWithAddress;
let fixture: Fixture;
let dutchAuction: DutchAuction;
let erc721Mock: ERC721Mock;
let expectedState: state;

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    await network.provider.send("evm_setAutomine", [true]);
    fixture = await getFixture(false, false, false);
    [admin, user] = await ethers.getSigners();
    dutchAuction = (await deployStaticWithFactory(
      fixture.factory,
      "DutchAuction",
      "DutchAuction",
      [
        startingPrice,
        discountRate,
        admin.address,
        fixture.erc721Mock.address,
        nftTokenId,
      ]
    )) as DutchAuction;
    fixture.erc721Mock.mint(admin.address, nftTokenId);
    fixture.erc721Mock.approve(dutchAuction.address, nftTokenId);
    showState("Initial", await getState(fixture));
  });

  it("Should get lower bid prices through time", async () => {
    const initialPrice = await dutchAuction.getPrice();
    await network.provider.send("evm_mine"); // Add 1 second to block.timestamp
    const currentPrice = await dutchAuction.getPrice();
    expect(currentPrice).to.be.equal(initialPrice.sub(discountRate.mul(1)));
  });

  it("Should let buy when price and value matches", async () => {
    expectedState = await getState(fixture);
    expectedState.Balances.erc721.admin -= BigNumber.from(1).toBigInt();
    expectedState.Balances.erc721.user += BigNumber.from(1).toBigInt();
    do {
      await network.provider.send("evm_mine"); // Add 1 second to block.timestamp
    } while ((await dutchAuction.getPrice()).gt(buyPrice));
    await dutchAuction.connect(user).buy({ value: buyPrice });
    showState("After Buy", await getState(fixture));
    expect(expectedState).to.be.deep.equal(await getState(fixture));
  });

  it("Should not let buy when auction has expired", async () => {
    do {
      await network.provider.send("evm_mine"); // Add 1 second to block.timestamp
    } while ((await dutchAuction.getRemainingTime()).gt(0));
    await expect(
      dutchAuction.connect(user).buy({ value: buyPrice })
    ).to.be.revertedWith("auction expired");
  });

  it("Should not let buy with value less than price", async () => {
    const lessPrice = buyPrice.sub(5);
    do {
      await network.provider.send("evm_mine"); // Add 1 second to block.timestamp
    } while ((await dutchAuction.getPrice()).gt(buyPrice));
    console.log(lessPrice, buyPrice)
    await expect(
      dutchAuction.connect(user).buy({ value: lessPrice })
    ).to.be.revertedWith("ETH < price");
  });


  it("Should refund value surplus", async () => {
    const refund = BigNumber.from(5)
    const valueSurplus = buyPrice.add(refund);
    do {
      await network.provider.send("evm_mine"); // Add 1 second to block.timestamp
    } while ((await dutchAuction.getPrice()).gt(buyPrice));
  });

});
