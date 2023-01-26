import { anyValue } from "@nomicfoundation/hardhat-chai-matchers/withArgs";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";
import { PromiseOrValue } from "../../typechain-types/common";
import { getImpersonatedSigner } from "../lockup/setup";
import { deployUpgradeableWithFactory, Fixture, getFixture, mineBlocks } from "../setup";

let dutchAuction: DutchAuction;
let asFactory: SignerWithAddress;
let admin: SignerWithAddress;
let fixture: Fixture;
const INITIAL_PRICE_ETH = ethers.utils.parseEther("1000000")
const DECAY = 16;
const SCALE_FACTOR = 100;
const gasPrice = 1000000000 // 1 GWei

const VALIDATORS: PromiseOrValue<string>[] = [];
const EXPECTED_PRICE_INITIAL = ethers.utils.parseEther(
  "1000000.0000000000000000000"
);
const EXPECTED_PRICE_BLOCK_ZERO = ethers.utils.parseEther(
  "10000.009504000000000000"
);
const EXPECTED_PRICE_BLOCK_ONE = ethers.utils.parseEther(
  "9984.035063258785942492"
);
const EXPECTED_PRICE_BLOCK_TWO = ethers.utils.parseEther(
  "9968.111577671451355661"
);
const EXPECTED_PRICE_BLOCK_THREE = ethers.utils.parseEther(
  "9952.238803821656050955"
);
const EXPECTED_PRICE_BLOCK_FOUR = ethers.utils.parseEther(
  "9936.416499841017488076"
);
const EXPECTED_FINAL_PRICE = ethers.utils.parseEther("0.009600000000000000");

const dailyExpectedPriceFirstMonth = [
"10000.009504000000000000",
"978.866285982772122161",
"514.624662988884314532",
"349.074103769896676905",
"264.112703317135009507",
"212.414015972812234494",
"177.642112150063947704",
"152.653388985224081084",
"133.828247682260999892",
"119.136635928714381015",
"107.351805925289823958",
"97.688742611549581933",
"89.621757717399096709",
"82.785574306337328653",
"76.918477622592752107",
"71.828042286699224360",
"67.369625219593684323",
"63.432401156831904206",
"59.930025099467906620",
"56.794226720574206150",
"53.970316080293546298",
"51.413966821565160176",
"49.088872370332561152",
"46.965010690808008714",
"45.017340899434702768",
"43.224811339671564390",
"41.569595611264421318",
"40.036497691248519004",
"38.612484036935239801",
"37.286312134315450451",
]
async function deployFixture() {
  fixture = await getFixture(true, false, false);
  [admin] = await ethers.getSigners();
  asFactory = await getImpersonatedSigner(fixture.factory.address);
  // Simulate 4 registered validators
  VALIDATORS.push("0x0000000000000000000000000000000000000000");
  VALIDATORS.push("0x0000000000000000000000000000000000000001");
  VALIDATORS.push("0x0000000000000000000000000000000000000002");
  VALIDATORS.push("0x0000000000000000000000000000000000000003");
  await fixture.validatorPool.registerValidators(VALIDATORS, []);
  dutchAuction = (await deployUpgradeableWithFactory(
    fixture.factory,
    "DutchAuction",
    "DutchAuction",
    undefined,
    [DECAY, SCALE_FACTOR]
  )) as DutchAuction;
  return { fixture, dutchAuction };
}

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    await loadFixture(deployFixture);
  });

  it("Should fail to start auction if not factory", async () => {
    await expect(dutchAuction.startAuction(INITIAL_PRICE_ETH,{gasPrice: gasPrice})).to.be.revertedWithCustomError(
      dutchAuction,
      "OnlyFactory"
    );
  });

  // From here on we skip these tests on coverage ([ @skip-on-coverage ]) since they are gasPrice based and coverage distorts gas consumption
  it("Should obtain bid price at first auction block", async () => {
    await dutchAuction.connect(asFactory).startAuction(INITIAL_PRICE_ETH,{gasPrice: gasPrice});
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("Should obtain prices through five blocks according to dutch auction curve", async () => {
    await dutchAuction.connect(asFactory).startAuction(INITIAL_PRICE_ETH,{gasPrice: gasPrice});
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_ONE);
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_TWO);
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_THREE
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_FOUR
    );
  });

  it("Should re-start the auction", async () => {
    await dutchAuction.connect(asFactory).startAuction(INITIAL_PRICE_ETH,{gasPrice: gasPrice});
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_ONE);
    await expect(dutchAuction.connect(asFactory).startAuction(INITIAL_PRICE_ETH,{gasPrice: gasPrice}))
      .to.emit(dutchAuction, "AuctionStarted")
      .withArgs(2, anyValue, INITIAL_PRICE_ETH, EXPECTED_FINAL_PRICE);
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("should get right prices for the first 30 days", async () => {
    await dutchAuction.connect(asFactory).startAuction(INITIAL_PRICE_ETH,{gasPrice: gasPrice});
  for (let day = 0; day <= 28; day++) {
       expect(await dutchAuction.getPrice()).to.be.equal(
        ethers.utils.parseEther(dailyExpectedPriceFirstMonth[day])
      ); 
      await mineBlocks(5760n); // blocks per day at 15 seconds per block
    }
  });

  it("Should bid for current price and end auction", async () => {
    await dutchAuction.connect(asFactory).startAuction(INITIAL_PRICE_ETH,{gasPrice: gasPrice});
  expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    // expect next block bid price since hardhat is mining for each tx
    await expect(dutchAuction.bid())
      .to.emit(dutchAuction, "BidPlaced")
      .withArgs(1, admin.address, EXPECTED_PRICE_BLOCK_ONE);
  });

  it("Should not start an auction with start price lower than final price", async () => {
    await expect(dutchAuction.connect(asFactory).startAuction(0, {gasPrice: gasPrice})).to.be.revertedWithCustomError(
      dutchAuction,
      "StartPriceLowerThanFinalPrice"
    );

  });
});
