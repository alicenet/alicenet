import { anyValue } from "@nomicfoundation/hardhat-chai-matchers/withArgs";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";
import { PromiseOrValue } from "../../typechain-types/common";
import {
  deployUpgradeableWithFactory,
  Fixture,
  getFixture,
  getImpersonatedSigner,
  mineBlocks,
} from "../setup";

let dutchAuction: DutchAuction;
let asFactory: SignerWithAddress;
let admin: SignerWithAddress;
let fixture: Fixture;
const INITIAL_PRICE_ETH = ethers.utils.parseEther("1000000");
const DECAY = 16;
const SCALE_FACTOR = 10;
const gasPrice = 100000000000; // 100 GWei

const VALIDATORS: PromiseOrValue<string>[] = [];
const EXPECTED_PRICE_INITIAL = ethers.utils.parseEther(
  "1000000.0000000000000000000"
);
const EXPECTED_PRICE_BLOCK_ZERO = ethers.utils.parseEther(
  "100000.864000000000000000"
);
const EXPECTED_PRICE_BLOCK_ONE = ethers.utils.parseEther(
  "86207.773793103448275862"
);
const EXPECTED_PRICE_BLOCK_TWO = ethers.utils.parseEther(
  "75758.463030303030303030"
);
const EXPECTED_PRICE_BLOCK_THREE = ethers.utils.parseEther(
  "67568.462702702702702702"
);
const EXPECTED_PRICE_BLOCK_FOUR = ethers.utils.parseEther(
  "60976.511219512195121951"
);
const EXPECTED_FINAL_PRICE = ethers.utils.parseEther("0.960000000000000000");

const dailyExpectedPriceFirstMonth = [
  "109.349230435725124647",
  "55.184001735169721288",
  "37.115869549497432930",
  "28.079353473992515051",
  "22.656659579084400086",
  "19.041203486059378729",
  "16.458574749697777502",
  "14.521515636442539803",
  "13.014862212792632061",
  "11.809506780948247802",
  "10.823285266210324896",
  "10.001419142510985334",
  "9.305983408169056402",
  "8.709887936512857076",
  "8.193266112115732368",
  "7.741217636607760432",
  "7.342347940414342426",
  "6.987794427901481633",
  "6.670560206494055301",
  "6.385047686214940595",
  "6.126725429613631901",
  "5.891885856324163304",
  "5.677466152147864401",
  "5.480913948841288642",
  "5.300085239355930732",
  "5.133165850116431438",
  "4.978610363202353300",
  "4.835094126126684698",
  "4.701475190254196068"
];

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
    await expect(
      dutchAuction.startAuction(INITIAL_PRICE_ETH, { gasPrice: gasPrice })
    ).to.be.revertedWithCustomError(dutchAuction, "OnlyFactory");
  });

  it("Should obtain bid price at first auction block", async () => {
    await dutchAuction
      .connect(asFactory)
      .startAuction(INITIAL_PRICE_ETH, { gasPrice: gasPrice });
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("Should obtain prices through five blocks according to dutch auction curve", async () => {
    await dutchAuction
      .connect(asFactory)
      .startAuction(INITIAL_PRICE_ETH, { gasPrice: gasPrice });
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
    await dutchAuction
      .connect(asFactory)
      .startAuction(INITIAL_PRICE_ETH, { gasPrice: gasPrice });
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_ONE);
    await expect(
      dutchAuction
        .connect(asFactory)
        .startAuction(INITIAL_PRICE_ETH, { gasPrice: gasPrice })
    )
      .to.emit(dutchAuction, "AuctionStarted")
      .withArgs(2, anyValue, INITIAL_PRICE_ETH, EXPECTED_FINAL_PRICE);
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("should get correct prices for the first 28 days", async () => {
    await dutchAuction
      .connect(asFactory)
      .startAuction(INITIAL_PRICE_ETH, { gasPrice: gasPrice });
    for (let day = 0; day <= 28; day++) {
      await mineBlocks(5760n); // approximately blocks per day at 15 seconds per block
      console.log("|", day+1, "|", await dutchAuction.getPrice(),"|");
      expect(await dutchAuction.getPrice()).to.be.equal(
        ethers.utils.parseEther(dailyExpectedPriceFirstMonth[day])
      );    }
  });

  it("Should bid for current price and end auction", async () => {
    await dutchAuction
      .connect(asFactory)
      .startAuction(INITIAL_PRICE_ETH, { gasPrice: gasPrice });
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    // expect next block bid price since hardhat is mining for each tx
    await expect(dutchAuction.bid())
      .to.emit(dutchAuction, "BidPlaced")
      .withArgs(1, admin.address, EXPECTED_PRICE_BLOCK_ONE);
  });

  it("Should not start an auction with start price lower than final price", async () => {
    await expect(
      dutchAuction.connect(asFactory).startAuction(0, { gasPrice: gasPrice })
    ).to.be.revertedWithCustomError(
      dutchAuction,
      "StartPriceLowerThanFinalPrice"
    );
  });
});
