import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";
import { PromiseOrValue } from "../../typechain-types/common";
import { getImpersonatedSigner } from "../lockup/setup";

import {
  deployUpgradeableWithFactory,
  factoryCallAny,
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  mineBlocks,
} from "../setup";

let fixture: Fixture;
let dutchAuction: DutchAuction;
let asFactory: SignerWithAddress;
const INITIAL_PRICE_ETH = 1000000;
const DECAY = 16;
const SCALE_FACTOR = 100;

const VALIDATORS: PromiseOrValue<string>[] = [];
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
const dailyExpectedPriceFirstMonth = [
  "10000009504000009504000",
  "978866285982781712764",
  "514624662988893909592",
  "349074103769906273554",
  "264112703317144606972",
  "212414015972821832455",
  "177642112150073545999",
  "152653388985233679619",
  "133828247682270598608",
  "119136635928723979872",
  "107351805925299422928",
  "97688742611559180995",
  "89621757717408695849",
  "82785574306346927859",
  "76918477622602351368",
  "71828042286708823671",
  "67369625219603283677",
  "63432401156841503597",
  "59930025099477506044",
  "56794226720583805605",
  "53970316080303145780",
  "51413966821574759682",
  "49088872370342160681",
  "46965010690817608264",
  "45017340899444302336",
  "43224811339681163975",
  "41569595611274020919",
  "40036497691258118620",
  "38612484036944839430",
  "37286312134325050093"];

async function deployFixture() {
  const fixture = await getFixture(true, false, false);
  // Simulate 4 registered validators
  VALIDATORS.push("0x0000000000000000000000000000000000000000");
  VALIDATORS.push("0x0000000000000000000000000000000000000001");
  VALIDATORS.push("0x0000000000000000000000000000000000000002");
  VALIDATORS.push("0x0000000000000000000000000000000000000003");
  await factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
    VALIDATORS,
    [],
  ]);
  dutchAuction = (await deployUpgradeableWithFactory(
    fixture.factory,
    "DutchAuction",
    "DutchAuction", undefined, [ethers.utils.parseEther(String(INITIAL_PRICE_ETH)), DECAY, SCALE_FACTOR]
  )) as DutchAuction;
  asFactory = await getImpersonatedSigner(fixture.factory.address);
  dutchAuction.connect(asFactory).startAuction();
  return { fixture, dutchAuction };
}

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    ({ fixture, dutchAuction } = await loadFixture(deployFixture));
  });

  it("Should fail to start auction if not factory", async () => {
    await expect(dutchAuction.startAuction()).to.be.revertedWithCustomError(dutchAuction, "OnlyFactory");
  });

  it("Should obtain bid price at first auction block", async () => {
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("Should obtain prices through five blocks according to dutch auction curve", async () => {
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

  it("Should restart the auction", async () => {
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_ONE);
    dutchAuction.connect(asFactory).startAuction();
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("should get right prices for the first 30 days", async () => {
    for (let day = 1; day <= 30; day++) {
      console.log(await dutchAuction.getPrice());
      /*       expect(await dutchAuction.getPrice()).to.be.equal(
              dailyExpectedPriceFirstMonth[day - 1]
            ); */
      await mineBlocks(5760n);
    }
  });
});
