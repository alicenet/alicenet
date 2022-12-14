import { anyValue } from "@nomicfoundation/hardhat-chai-matchers/withArgs";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";
import { PromiseOrValue } from "../../typechain-types/common";
import { getImpersonatedSigner } from "../lockup/setup";
import { deployUpgradeableWithFactory, getFixture, mineBlocks } from "../setup";

let dutchAuction: DutchAuction;
let asFactory: SignerWithAddress;
let admin: SignerWithAddress;
const INITIAL_PRICE_ETH = 1000000;
const DECAY = 16;
const SCALE_FACTOR = 100;

const VALIDATORS: PromiseOrValue<string>[] = [];
const EXPECTED_PRICE_INITIAL = ethers.utils.parseEther(
  "1000000.0000000000000000000"
);
const EXPECTED_PRICE_BLOCK_ZERO = ethers.utils.parseEther(
  "10000.009504000009504000"
);
const EXPECTED_PRICE_BLOCK_ONE = ethers.utils.parseEther(
  "9984.035063258795446645"
);
const EXPECTED_PRICE_BLOCK_TWO = ethers.utils.parseEther(
  "9968.111577671460859968"
);
const EXPECTED_PRICE_BLOCK_THREE = ethers.utils.parseEther(
  "9952.238803821665555414"
);
const EXPECTED_PRICE_BLOCK_FOUR = ethers.utils.parseEther(
  "9936.416499841026992686"
);
const EXPECTED_FINAL_PRICE = ethers.utils.parseEther("0.009600000009600000");
const dailyExpectedPriceFirstMonth = [
  ethers.utils.parseEther("10000.009504000009504000"),
  ethers.utils.parseEther("978.866285982781712764"),
  ethers.utils.parseEther("514.624662988893909592"),
  ethers.utils.parseEther("349.074103769906273554"),
  ethers.utils.parseEther("264.112703317144606972"),
  ethers.utils.parseEther("212.414015972821832455"),
  ethers.utils.parseEther("177.642112150073545999"),
  ethers.utils.parseEther("152.653388985233679619"),
  ethers.utils.parseEther("133.828247682270598608"),
  ethers.utils.parseEther("119.136635928723979872"),
  ethers.utils.parseEther("107.351805925299422928"),
  ethers.utils.parseEther("97.688742611559180995"),
  ethers.utils.parseEther("89.621757717408695849"),
  ethers.utils.parseEther("82.785574306346927859"),
  ethers.utils.parseEther("76.918477622602351368"),
  ethers.utils.parseEther("71.828042286708823671"),
  ethers.utils.parseEther("67.369625219603283677"),
  ethers.utils.parseEther("63.432401156841503597"),
  ethers.utils.parseEther("59.930025099477506044"),
  ethers.utils.parseEther("56.794226720583805605"),
  ethers.utils.parseEther("53.970316080303145780"),
  ethers.utils.parseEther("51.413966821574759682"),
  ethers.utils.parseEther("49.088872370342160681"),
  ethers.utils.parseEther("46.965010690817608264"),
  ethers.utils.parseEther("45.017340899444302336"),
  ethers.utils.parseEther("43.224811339681163975"),
  ethers.utils.parseEther("41.569595611274020919"),
  ethers.utils.parseEther("40.036497691258118620"),
  ethers.utils.parseEther("38.612484036944839430"),
  ethers.utils.parseEther("37.286312134325050093"),
];

async function deployFixture() {
  const fixture = await getFixture(true, false, false);
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
    [ethers.utils.parseEther(String(INITIAL_PRICE_ETH)), DECAY, SCALE_FACTOR]
  )) as DutchAuction;
  await dutchAuction.connect(asFactory).startAuction();
  return { fixture, dutchAuction };
}

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    await loadFixture(deployFixture);
  });

  it("Should fail to start auction if not factory", async () => {
    await expect(dutchAuction.startAuction()).to.be.revertedWithCustomError(
      dutchAuction,
      "OnlyFactory"
    );
  });

  // From here on we skip these tests on coverage ([ @skip-on-coverage ]) since they are gasPrice based and coverage distorts gas consumption
  it("Should obtain bid price at first auction block [ @skip-on-coverage ]", async () => {
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("Should obtain prices through five blocks according to dutch auction curve [ @skip-on-coverage ]", async () => {
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

  it("Should re-start the auction [ @skip-on-coverage ]", async () => {
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_ONE);
    await expect(dutchAuction.connect(asFactory).startAuction())
      .to.emit(dutchAuction, "AuctionStarted")
      .withArgs(2, anyValue, EXPECTED_PRICE_INITIAL, EXPECTED_FINAL_PRICE);
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("should get right prices for the first 30 days [ @skip-on-coverage ]", async () => {
    for (let day = 1; day <= 30; day++) {
      expect(await dutchAuction.getPrice()).to.be.equal(
        dailyExpectedPriceFirstMonth[day - 1]
      );
      await mineBlocks(5760n); // blocks per day
    }
  });

  it("Should bid for current price and end auction [ @skip-on-coverage ]", async () => {
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    // expect next block bid price sincehardhat is automining for each tx
    await expect(dutchAuction.bid())
      .to.emit(dutchAuction, "AuctionEnded")
      .withArgs(1, admin.address, EXPECTED_PRICE_BLOCK_ONE);
  });
});
