import { expect } from "chai";
import { BigNumber } from "ethers";
import { contract, ethers, network } from "hardhat";
import { ATokenMinter, MockStakingNFT } from "../../typechain-types";
import { getMetamorphicAddress } from "../factory/Setup";
import {
  BaseTokensFixture,
  deployUpgradeableWithFactory,
  factoryCallAny,
  getBaseTokensFixture,
  mineBlocks,
} from "../setup";

contract("StakingNFT", async () => {
  let stakingNFT: MockStakingNFT;
  let fixture: BaseTokensFixture;
  let aTokenMinter: ATokenMinter;
  beforeEach(async () => {
    fixture = await getBaseTokensFixture();
    const hre = await require("hardhat"); // eslint-disable-line
    if (hre.__SOLIDITY_COVERAGE_RUNNING !== true) {
      await network.provider.send("evm_setBlockGasLimit", [
        "0x6000000000000000000000",
      ]);
    }
    stakingNFT = (await deployUpgradeableWithFactory(
      fixture.factory,
      "MockStakingNFT"
    )) as MockStakingNFT;
    aTokenMinter = (await deployUpgradeableWithFactory(
      fixture.factory,
      "ATokenMinter"
    )) as ATokenMinter;
  });

  describe("skimExcessEth", async () => {
    it("skimExcessEth Test 1", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const receipt = await factoryCallAny(
        fixture.factory,
        stakingNFT,
        "skimExcessEth",
        [to]
      );
      expect(receipt.status).to.eq(1);
    });

    it("skimExcessEth Test 2: fail, not factory", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const tx = stakingNFT.skimExcessEthMock(to);
      // Error Code for not calling as factory
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "OnlyFactory")
        .withArgs(to, fixture.factory.address);
    });
  });

  describe("skimExcessToken", async () => {
    it("skimExcessToken Test 1", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const receipt = await factoryCallAny(
        fixture.factory,
        stakingNFT,
        "skimExcessToken",
        [to]
      );
      expect(receipt.status).to.eq(1);
    });

    it("skimExcessToken Test 2: fail, not factory", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const tx = stakingNFT.skimExcessTokenMock(to);
      // Error Code for not calling as factory
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "OnlyFactory")
        .withArgs(to, fixture.factory.address);
    });
  });

  describe("burn", async () => {
    it("burn Test 1: pass, one staker", async () => {
      // Check standard values; nothing should be present
      // Eth accumulator
      let expEthAccumulator = BigNumber.from(0);
      let expEthSlush = BigNumber.from(0);
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      let expTokenAccumulator = BigNumber.from(0);
      let expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      let retTokenAccumulator = retValue[0];
      let retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      let expReserveEth = BigNumber.from(0);
      let retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      let expReserveToken = BigNumber.from(0);
      let retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      let expShares = BigNumber.from(0);
      let retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("25519000000000000000000");
      const tx = await fixture.aToken.approve(stakingNFT.address, amount);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
      const tx2 = await stakingNFT.mintNFTMock(to, amount);
      // We use receipt:any *only* because of events which may not be present
      const receipt2: any = await tx2.wait();
      expect(receipt2.status).to.eq(1);
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt2.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = BigNumber.from(0);
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = BigNumber.from(0);
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // deposit Eth to be distributed
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx3 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const depositedEthScaled = depositedEth.mul(scaleFactor);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);

      // mine blocks
      const duration = BigNumber.from(100);
      await mineBlocks(duration.toBigInt());

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = depositedEthScaled;
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = depositedEth;
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Burn staking position (static call)
      const retValue4 = await stakingNFT.callStatic.burnMock(tokenID);
      const payoutEth = retValue4[0];
      const payoutToken = retValue4[1];
      expect(payoutEth).to.eq(depositedEth);
      expect(payoutToken).to.eq(amount);

      // Check standard values; nothing should change
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = depositedEthScaled;
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = depositedEth;
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Burn staking position
      const tx4 = await stakingNFT.burnMock(tokenID);
      const receipt4 = await tx4.wait();
      expect(receipt4.status).to.eq(1);

      // Check standard values
      // Eth reserves
      expReserveEth = BigNumber.from(0);
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = BigNumber.from(0);
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = BigNumber.from(0);
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth accumulator
      expEthAccumulator = depositedEthScaled.div(amount);
      expEthSlush = BigNumber.from(0); // Because deposited whole number of Wei, only staker
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
    });

    it("burn Test 2: fail from tokenID not existing", async () => {
      const tokenID = BigNumber.from("0");
      const tx = stakingNFT.burnMock(tokenID);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });
  });

  describe("burnTo", async () => {
    it("burnTo Test 1: pass, one staker", async () => {
      // Check standard values; nothing should be present
      // Eth accumulator
      let expEthAccumulator = BigNumber.from(0);
      let expEthSlush = BigNumber.from(0);
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      let expTokenAccumulator = BigNumber.from(0);
      let expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      let retTokenAccumulator = retValue[0];
      let retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      let expReserveEth = BigNumber.from(0);
      let retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      let expReserveToken = BigNumber.from(0);
      let retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      let expShares = BigNumber.from(0);
      let retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("25519000000000000000000");
      const tx = await fixture.aToken.approve(stakingNFT.address, amount);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
      const tx2 = await stakingNFT.mintNFTMock(to, amount);
      // We use receipt:any *only* because of events which may not be present
      const receipt2: any = await tx2.wait();
      expect(receipt2.status).to.eq(1);
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt2.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = BigNumber.from(0);
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = BigNumber.from(0);
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // deposit Eth to be distributed
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx3 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const depositedEthScaled = depositedEth.mul(scaleFactor);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);

      // mine blocks
      const duration = BigNumber.from(100);
      await mineBlocks(duration.toBigInt());

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = depositedEthScaled;
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = depositedEth;
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Burn staking position (static call)
      const from = signers[0].address;
      const retValue4 = await stakingNFT.callStatic.burnNFTMock(
        from,
        to,
        tokenID
      );
      const payoutEth = retValue4[0];
      const payoutToken = retValue4[1];
      expect(payoutEth).to.eq(depositedEth);
      expect(payoutToken).to.eq(amount);

      // Check standard values; nothing should change
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = depositedEthScaled;
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = depositedEth;
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Burn staking position
      const tx4 = await stakingNFT.burnNFTMock(from, to, tokenID);
      const receipt4 = await tx4.wait();
      expect(receipt4.status).to.eq(1);

      // Check standard values
      // Eth reserves
      expReserveEth = BigNumber.from(0);
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = BigNumber.from(0);
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = BigNumber.from(0);
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth accumulator
      expEthAccumulator = depositedEthScaled.div(amount);
      expEthSlush = BigNumber.from(0); // Because deposited whole number of Wei, only staker
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
    });

    it("burnTo Test 2: fail from tokenID not existing", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const tokenID = BigNumber.from("0");
      const tx = stakingNFT.burnToMock(to, tokenID);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });
  });

  describe("burnNFT", async () => {
    it("burnNFT Test 1: pass, one staker", async () => {
      // Check standard values; nothing should be present
      // Eth accumulator
      let expEthAccumulator = BigNumber.from(0);
      let expEthSlush = BigNumber.from(0);
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      let expTokenAccumulator = BigNumber.from(0);
      let expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      let retTokenAccumulator = retValue[0];
      let retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      let expReserveEth = BigNumber.from(0);
      let retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      let expReserveToken = BigNumber.from(0);
      let retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      let expShares = BigNumber.from(0);
      let retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("25519000000000000000000");
      const tx = await fixture.aToken.approve(stakingNFT.address, amount);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
      const tx2 = await stakingNFT.mintNFTMock(to, amount);
      // We use receipt:any *only* because of events which may not be present
      const receipt2: any = await tx2.wait();
      expect(receipt2.status).to.eq(1);
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt2.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = BigNumber.from(0);
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = BigNumber.from(0);
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // deposit Eth to be distributed
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx3 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const depositedEthScaled = depositedEth.mul(scaleFactor);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);

      // mine blocks
      const duration = BigNumber.from(100);
      await mineBlocks(duration.toBigInt());

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = depositedEthScaled;
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = depositedEth;
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Burn staking position (static call)
      const from = signers[0].address;
      const retValue4 = await stakingNFT.callStatic.burnNFTMock(
        from,
        to,
        tokenID
      );
      const payoutEth = retValue4[0];
      const payoutToken = retValue4[1];
      expect(payoutEth).to.eq(depositedEth);
      expect(payoutToken).to.eq(amount);

      // Check standard values; nothing should change
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = depositedEthScaled;
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = depositedEth;
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = amount;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = amount;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Burn staking position
      const tx4 = await stakingNFT.burnNFTMock(from, to, tokenID);
      const receipt4 = await tx4.wait();
      expect(receipt4.status).to.eq(1);

      // Check standard values
      // Eth reserves
      expReserveEth = BigNumber.from(0);
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = BigNumber.from(0);
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = BigNumber.from(0);
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth accumulator
      expEthAccumulator = depositedEthScaled.div(amount);
      expEthSlush = BigNumber.from(0); // Because deposited whole number of Wei, only staker
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
    });

    it("burnNFT Test 2: pass, two stakers", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Check standard values; nothing should be present
      // Eth accumulator
      let expEthAccumulator = BigNumber.from(0);
      let expEthSlush = BigNumber.from(0);
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      let expTokenAccumulator = BigNumber.from(0);
      let expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      let retTokenAccumulator = retValue[0];
      let retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      let expReserveEth = BigNumber.from(0);
      let retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      let expReserveToken = BigNumber.from(0);
      let retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      let expShares = BigNumber.from(0);
      let retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Alice stakes
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const aliceShares = BigNumber.from("25519000000000000000000");
      const tx1 = await fixture.aToken.approve(stakingNFT.address, aliceShares);
      const receipt1 = await tx1.wait();
      expect(receipt1.status).to.eq(1);
      const tx2 = await stakingNFT.mintNFTMock(aliceAddress, aliceShares);
      // We use receipt:any *only* because of events which may not be present
      const receipt2: any = await tx2.wait();
      expect(receipt2.status).to.eq(1);
      const expAliceTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt2.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expAliceTokenID);

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = BigNumber.from(0);
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = BigNumber.from(0);
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = aliceShares;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = aliceShares;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // deposit Eth to be distributed
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx3 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const depositedEthScaled = depositedEth.mul(scaleFactor);

      // mine blocks
      const duration = BigNumber.from(100);
      await mineBlocks(duration.toBigInt());

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = BigNumber.from(0);
      expEthSlush = depositedEthScaled;
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = depositedEth;
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = aliceShares;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = aliceShares;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Bob stakes
      const bobAddress = signers[1].address;
      const bobShares = BigNumber.from("257000000000000000000");
      const tx4 = await fixture.aToken.approve(stakingNFT.address, bobShares);
      const receipt4 = await tx4.wait();
      expect(receipt4.status).to.eq(1);
      const tx5 = await stakingNFT.mintNFTMock(bobAddress, bobShares);
      // We use receipt:any *only* because of events which may not be present
      const receipt5: any = await tx5.wait();
      expect(receipt5.status).to.eq(1);
      const expBobTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt5.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expBobTokenID);

      // Check standard values: Accumulators and Reserves
      // Eth accumulator
      expEthAccumulator = depositedEthScaled.div(aliceShares);
      expEthSlush = depositedEthScaled.sub(expEthAccumulator.mul(aliceShares));
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth reserves
      expReserveEth = depositedEth;
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = aliceShares.add(bobShares);
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = aliceShares.add(bobShares);
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);

      // Burn staking position (static call)
      const retValue6 = await stakingNFT.callStatic.burnNFTMock(
        aliceAddress,
        aliceAddress,
        aliceTokenID
      );
      const payoutEth = retValue6[0];
      const payoutToken = retValue6[1];
      const retValue7 = await stakingNFT.getPositionMock(aliceTokenID);
      const aliceAccumEth = retValue7[3];
      const retValue8 = await stakingNFT.getEthAccumulator();
      const currentAccumEth = retValue8[0];
      const alicePayoutEth = aliceShares
        .mul(currentAccumEth.sub(aliceAccumEth))
        .div(scaleFactor);
      const alicePayoutToken = aliceShares;
      expect(payoutEth).to.eq(alicePayoutEth);
      expect(payoutToken).to.eq(alicePayoutToken);

      // Burn Alice staking position
      const tx9 = await stakingNFT.burnNFTMock(
        aliceAddress,
        aliceAddress,
        aliceTokenID
      );
      const receipt9 = await tx9.wait();
      expect(receipt9.status).to.eq(1);

      // Check standard values
      // Eth reserves
      expReserveEth = depositedEth.sub(alicePayoutEth);
      retReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retReserveEth).to.eq(expReserveEth);
      // Token reserves
      expReserveToken = bobShares;
      retReserveToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retReserveToken).to.eq(expReserveToken);
      // Total shares (staked Tokens)
      expShares = bobShares;
      retShares = await stakingNFT.getTotalSharesMock();
      expect(retShares).to.eq(expShares);
      // Token accumulator
      expTokenAccumulator = BigNumber.from(0);
      expTokenSlush = BigNumber.from(0);
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
      // Eth accumulator
      expEthAccumulator = depositedEthScaled.div(aliceShares);
      expEthSlush = depositedEthScaled.sub(expEthAccumulator.mul(aliceShares));
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
    });

    it("burnNFT Test 3: fail from tokenID not existing", async () => {
      const signers = await ethers.getSigners();
      const from = signers[0].address;
      const to = signers[0].address;
      const tokenID = BigNumber.from("0");
      const tx = stakingNFT.burnNFTMock(from, to, tokenID);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });

    it("burnNFT Test 4: fail from invalid owner", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let tx = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await tx.wait();
      tx = await stakingNFT.mintNFTMock(to, amount);
      receipt = await tx.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Ensure from is different signer so different address
      const from = signers[1].address;
      const tx2 = stakingNFT.burnNFTMock(from, to, tokenID);
      // Error Code for Caller not token owner
      await expect(tx2)
        .to.be.revertedWithCustomError(stakingNFT, "CallerNotTokenOwner")
        .withArgs(to);
    });

    it("burnNFT Test 5: fail from freeAfter", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let tx = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await tx.wait();
      tx = await stakingNFT.mintNFTMock(to, amount);
      receipt = await tx.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      const expFreeAfter = BigNumber.from(blockNumber).add(1);

      // get position info and confirm values
      const retValue = await stakingNFT.getPositionMock(tokenID);
      const freeAfter = retValue[1];
      expect(freeAfter).to.eq(expFreeAfter);

      // Lock position
      const duration = BigNumber.from("100");
      tx = await stakingNFT.lockPositionLowMock(tokenID, duration);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);

      // Need will need to lock position longer
      const from = signers[0].address;
      const tx2 = stakingNFT.burnNFTMock(from, to, tokenID);
      // Error Code 606 for position not ready to be burned
      await expect(tx2).to.be.revertedWithCustomError(
        stakingNFT,
        "FreeAfterTimeNotReached"
      );
    });

    it("burnNFT Test 6: burn in same block as mint", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let tx = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await tx.wait();
      tx = await stakingNFT.mintNFTMock(to, amount);
      receipt = await tx.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Need will need to lock position longer
      const from = signers[0].address;
      const tx2 = stakingNFT.burnNFTMock(from, to, tokenID);
      // Error Code 606 for position not ready to be burned
      await expect(tx2).to.be.revertedWithCustomError(
        stakingNFT,
        "FreeAfterTimeNotReached"
      );
    });
  });

  describe("mint", async () => {
    it("mint Test 1: one staker, zero duration", async () => {
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintMock(amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      const expFreeAfter = BigNumber.from(blockNumber).add(1);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      const retValue = await stakingNFT.getPositionMock(tokenID);
      const shares = retValue[0];
      const freeAfter = retValue[1];
      const withdrawFreeAfter = retValue[2];
      const accumulatorEth = retValue[3];
      const accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("mint Test 2: fail, circuit breaker tripped", async () => {
      await stakingNFT.tripCBLowMock();
      const amount = BigNumber.from("0");
      const tx = stakingNFT.mintMock(amount);
      // Error Code for tripped circuit breaker
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerOpened"
      );
    });

    it("mint Test 3: fail, amount == 0", async () => {
      const amount = BigNumber.from("0");
      const tx = stakingNFT.mintMock(amount);
      // Error Code for Invalid Staking Amount 0
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "MintAmountZero"
      );
    });
  });

  describe("mintTo", async () => {
    it("mintTo Test 1: one staker, zero duration", async () => {
      const duration = BigNumber.from("0");
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintToMock(to, amount, duration);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      const expFreeAfter = BigNumber.from(blockNumber).add(1);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      const retValue = await stakingNFT.getPositionMock(tokenID);
      const shares = retValue[0];
      const freeAfter = retValue[1];
      const withdrawFreeAfter = retValue[2];
      const accumulatorEth = retValue[3];
      const accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("mintTo Test 2: one staker, long duration", async () => {
      const duration = await stakingNFT.getMaxMintLockMock();
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintToMock(to, amount, duration);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      const expFreeAfter = BigNumber.from(blockNumber).add(duration);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      const retValue = await stakingNFT.getPositionMock(tokenID);
      const shares = retValue[0];
      const freeAfter = retValue[1];
      const withdrawFreeAfter = retValue[2];
      const accumulatorEth = retValue[3];
      const accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("mintTo Test 3: fail, circuit breaker tripped", async () => {
      await stakingNFT.tripCBLowMock();
      const duration = BigNumber.from("1");
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("0");
      const tx = stakingNFT.mintToMock(to, amount, duration);
      // Error Code for circuit breaker tripped
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerOpened"
      );
    });

    it("mintTo Test 4: fail, amount == 0", async () => {
      const duration = BigNumber.from("0");
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("0");
      const tx = stakingNFT.mintToMock(to, amount, duration);
      // Error Code 609 for Invalid Staking Amount 0
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "MintAmountZero"
      );
    });

    it("mintTo Test 5: fail, duration > max", async () => {
      const maxMintLock = await stakingNFT.getMaxMintLockMock();
      const duration = await maxMintLock.add(1);
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      const tx = stakingNFT.mintToMock(to, amount, duration);
      // Error Code 602 for lock duration larger than max
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "LockDurationGreaterThanMintLock"
      );
    });
  });

  describe("mintNFT", async () => {
    it("mintNFT Test 1: one staker", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);
    });

    it("mintNFT Test 2: multiple stakers", async () => {
      const signers = await ethers.getSigners();
      let to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      let expTokenID = BigNumber.from("1");
      let retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);

      to = signers[1].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("2");
      retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);

      to = signers[2].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("3");
      retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);

      to = signers[3].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("4");
      retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);
    });

    it("mintNFT Test 3: multiple stakers and reserve checks", async () => {
      let expTokenReserve = BigNumber.from("0");
      let retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);
      const signers = await ethers.getSigners();
      let to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      let expTokenID = BigNumber.from("1");
      let retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);
      expTokenReserve = expTokenReserve.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      to = signers[1].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("2");
      retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);
      expTokenReserve = expTokenReserve.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      to = signers[2].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("3");
      retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);
      expTokenReserve = expTokenReserve.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      to = signers[3].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("4");
      retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);
      expTokenReserve = expTokenReserve.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);
    });

    it("mintNFT Test 4: multiple stakers and slushSkim", async () => {
      // Deposit Eth
      const magic = BigNumber.from("42");
      await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits("1", "ether"),
      });
      let expEthAccumulator = BigNumber.from("0");
      let expEthSlush = BigNumber.from("1000000000000000000000000000000000000");
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // Add staker
      const signers = await ethers.getSigners();
      let to = signers[0].address;
      const amount = BigNumber.from("1000000000000000001");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      let expTokenID = BigNumber.from("1");
      let retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);
      let expTokenReserve = amount;
      let retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Check Eth accumulator; there should be no changes
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // Add another staker
      to = signers[1].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("2");
      retTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(retTokenID).to.eq(expTokenID);
      expTokenReserve = expTokenReserve.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Check Eth accumulator; it should have changed
      expEthAccumulator = expEthSlush.div(amount);
      expEthSlush = expEthSlush.sub(expEthAccumulator.mul(amount));
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
    });

    it("mintNFT Test 5: fail (amount == 0)", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("0");
      const tx = stakingNFT.mintNFTMock(to, amount);
      // Error Code 609 for Invalid Staking Amount 0
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "MintAmountZero"
      );
    });

    it("mintNFT Test 6: fail (amount >= 2**224)", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      // 2**224
      const amount = BigNumber.from(
        "26959946667150639794667015087019630673637144422540572481103610249216"
      );
      const tx = stakingNFT.mintNFTMock(to, amount);
      // Error Code 609 for Invalid Staking Amount larger than maximum allowed
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "MintAmountExceedsMaximumSupply"
      );
    });

    it("mintNFT gas Test", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      let receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);
    });
  });

  describe("collectEth", async () => {
    it("collectEth Test 1: one staker", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Deposit Eth
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx3 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const depositedEthScaled = depositedEth.mul(scaleFactor);
      const expEthAccumulator = BigNumber.from("0");
      const expEthSlush = depositedEthScaled;
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // Add staker
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);
      const expTokenReserve = amount;
      const retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Check Eth accumulator; there should be no changes
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // mine blocks
      const duration = BigNumber.from(2);
      await mineBlocks(duration.toBigInt());

      // collectEth
      const payoutEth = await stakingNFT.callStatic.collectEthMock(tokenID);
      expect(payoutEth).to.eq(depositedEth);
      const tx = await stakingNFT.collectEthMock(tokenID);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });

    it("collectEth Test 2: two stakers", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Add Alice
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, amount);
      receipt = await txResponse.wait();
      let expTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expTokenID);
      let expTokenReserve = amount;
      let retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Add Bob
      const bobAddress = signers[1].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expTokenID);
      expTokenReserve = amount.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Deposit Eth
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx3 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const depositedEthScaled = depositedEth.mul(scaleFactor);
      const expEthAccumulator = BigNumber.from("0");
      const expEthSlush = depositedEthScaled;
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // Check Eth accumulator; there should be no changes
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // mine blocks
      const duration = BigNumber.from(2);
      await mineBlocks(duration.toBigInt());

      // collectEth
      const alicePayoutEth = await stakingNFT.callStatic.collectEthMock(
        aliceTokenID
      );
      expect(alicePayoutEth).to.eq(depositedEth.div(2));
      const tx = await stakingNFT.collectEthMock(aliceTokenID);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });

    it("collectEth Test 3: fail because tokenID does not exist", async () => {
      const tokenID = BigNumber.from(0);
      const tx = stakingNFT.collectEthMock(tokenID);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });

    it("collectEth Test 4: fail because not owner", async () => {
      // Add staker
      const signers = await ethers.getSigners();
      const to = signers[1].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // collectEth
      const tx = stakingNFT.collectEthMock(tokenID);
      // Error Code for Caller not token owner
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "CallerNotTokenOwner")
        .withArgs(signers[0].address);
    });

    it("collectEth Test 5: fail because not free to withdraw", async () => {
      // Add staker
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // collectEth
      const tx = stakingNFT.collectEthMock(tokenID);
      // Error Code 603 for freeAfter not reached
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "LockDurationWithdrawTimeNotReached"
      );
    });
  });

  describe("collectEthTo", async () => {
    it("collectEthTo Test 1: one staker", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Deposit Eth
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx3 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const depositedEthScaled = depositedEth.mul(scaleFactor);
      const expEthAccumulator = BigNumber.from("0");
      const expEthSlush = depositedEthScaled;
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // Add staker
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);
      const expTokenReserve = amount;
      const retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Check Eth accumulator; there should be no changes
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // mine blocks
      const duration = BigNumber.from(2);
      await mineBlocks(duration.toBigInt());

      // collectEthTo
      const payoutEth = await stakingNFT.callStatic.collectEthToMock(
        to,
        tokenID
      );
      expect(payoutEth).to.eq(depositedEth);
      const tx = await stakingNFT.collectEthToMock(to, tokenID);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });

    it("collectEthTo Test 2: two stakers", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Add Alice
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, amount);
      receipt = await txResponse.wait();
      let expTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expTokenID);
      let expTokenReserve = amount;
      let retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Add Bob
      const bobAddress = signers[1].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expTokenID);
      expTokenReserve = amount.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Deposit Eth
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx3 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const depositedEthScaled = depositedEth.mul(scaleFactor);
      const expEthAccumulator = BigNumber.from("0");
      const expEthSlush = depositedEthScaled;
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // Check Eth accumulator; there should be no changes
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);

      // mine blocks
      const duration = BigNumber.from(2);
      await mineBlocks(duration.toBigInt());

      // collectEthTo
      const alicePayoutEth = await stakingNFT.callStatic.collectEthToMock(
        aliceAddress,
        aliceTokenID
      );
      expect(alicePayoutEth).to.eq(depositedEth.div(2));
      const tx = await stakingNFT.collectEthToMock(aliceAddress, aliceTokenID);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });

    it("collectEthTo Test 3: fail because tokenID does not exist", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const tokenID = BigNumber.from(0);
      const tx = stakingNFT.collectEthToMock(to, tokenID);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });

    it("collectEthTo Test 4: fail because not owner", async () => {
      // Add staker
      const signers = await ethers.getSigners();
      const from = signers[0].address;
      const to = signers[1].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // collectEthTo
      const tx = stakingNFT.collectEthToMock(from, tokenID);
      // Error Code for Caller not token owner
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "CallerNotTokenOwner")
        .withArgs(from);
    });

    it("collectEthTo Test 5: fail because not free to withdraw", async () => {
      // Add staker
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // collectEthTo
      const tx = stakingNFT.collectEthToMock(to, tokenID);
      // Error Code 603 for freeAfter not reached
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "LockDurationWithdrawTimeNotReached"
      );
    });
  });

  describe("collectToken", async () => {
    it("collectToken Test 1: one staker", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Deposit Token
      const magic = BigNumber.from("42");
      const tokenString = "1234567890";
      const tokenScale = "1000000000000000000"; // 10**18
      const depositedToken = BigNumber.from(tokenString).mul(
        BigNumber.from(tokenScale)
      );
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      await factoryCallAny(fixture.factory, aTokenMinter, "mint", [
        to,
        depositedToken,
      ]);
      await fixture.aToken.approve(stakingNFT.address, depositedToken);
      const tx3 = await stakingNFT.depositTokenMock(magic, depositedToken);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedTokenScaled = depositedToken.mul(scaleFactor);
      const expTokenAccumulator = BigNumber.from("0");
      const expTokenSlush = depositedTokenScaled;
      let retValue = await stakingNFT.getTokenAccumulator();
      let retTokenAccumulator = retValue[0];
      let retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // Add staker
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);
      const expTokenReserve = amount.add(depositedToken);
      const retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Check Eth accumulator; there should be no changes
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // mine blocks
      const duration = BigNumber.from(2);
      await mineBlocks(duration.toBigInt());

      // collectToken
      const payoutToken = await stakingNFT.callStatic.collectTokenMock(tokenID);
      expect(payoutToken).to.eq(depositedToken);
      const tx = await stakingNFT.collectTokenMock(tokenID);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });

    it("collectToken Test 2: two stakers", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Add Alice
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, amount);
      receipt = await txResponse.wait();
      let expTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expTokenID);
      let expTokenReserve = amount;
      let retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Add Bob
      const bobAddress = signers[1].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expTokenID);
      expTokenReserve = amount.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Deposit Token
      const magic = BigNumber.from("42");
      const tokenString = "1234567890";
      const tokenScale = "1000000000000000000"; // 10**18
      const depositedToken = BigNumber.from(tokenString).mul(
        BigNumber.from(tokenScale)
      );
      const to = signers[0].address;
      await factoryCallAny(fixture.factory, aTokenMinter, "mint", [
        to,
        depositedToken,
      ]);
      await fixture.aToken.approve(stakingNFT.address, depositedToken);
      const tx3 = await stakingNFT.depositTokenMock(magic, depositedToken);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedTokenScaled = depositedToken.mul(scaleFactor);
      const expTokenAccumulator = BigNumber.from("0");
      const expTokenSlush = depositedTokenScaled;
      let retValue = await stakingNFT.getTokenAccumulator();
      let retTokenAccumulator = retValue[0];
      let retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // Check Eth accumulator; there should be no changes
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // mine blocks
      const duration = BigNumber.from(2);
      await mineBlocks(duration.toBigInt());

      // collectEth
      const alicePayoutToken = await stakingNFT.callStatic.collectTokenMock(
        aliceTokenID
      );
      expect(alicePayoutToken).to.eq(depositedToken.div(2));
      const tx = await stakingNFT.collectTokenMock(aliceTokenID);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });

    it("collectToken Test 3: fail because tokenID does not exist", async () => {
      const tokenID = BigNumber.from(0);
      const tx = stakingNFT.collectTokenMock(tokenID);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });

    it("collectToken Test 4: fail because not owner", async () => {
      // Add staker
      const signers = await ethers.getSigners();
      const to = signers[1].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // collectToken
      const tx = stakingNFT.collectTokenMock(tokenID);
      // Error Code for Caller not token owner
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "CallerNotTokenOwner")
        .withArgs(signers[0].address);
    });

    it("collectToken Test 5: fail because not free to withdraw", async () => {
      // Add staker
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // collectToken
      const tx = stakingNFT.collectTokenMock(tokenID);
      // Error Code for freeAfter not reached
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "LockDurationWithdrawTimeNotReached"
      );
    });
  });

  describe("collectTokenTo", async () => {
    it("collectTokenTo Test 1: one staker", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Deposit Token
      const magic = BigNumber.from("42");
      const tokenString = "1234567890";
      const tokenScale = "1000000000000000000"; // 10**18
      const depositedToken = BigNumber.from(tokenString).mul(
        BigNumber.from(tokenScale)
      );
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      await factoryCallAny(fixture.factory, aTokenMinter, "mint", [
        to,
        depositedToken,
      ]);
      await fixture.aToken.approve(stakingNFT.address, depositedToken);
      const tx3 = await stakingNFT.depositTokenMock(magic, depositedToken);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedTokenScaled = depositedToken.mul(scaleFactor);
      const expTokenAccumulator = BigNumber.from("0");
      const expTokenSlush = depositedTokenScaled;
      let retValue = await stakingNFT.getTokenAccumulator();
      let retTokenAccumulator = retValue[0];
      let retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // Add staker
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);
      const expTokenReserve = amount.add(depositedToken);
      const retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Check Eth accumulator; there should be no changes
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // mine blocks
      const duration = BigNumber.from(2);
      await mineBlocks(duration.toBigInt());

      // collectTokenTo
      const payoutToken = await stakingNFT.callStatic.collectTokenToMock(
        to,
        tokenID
      );
      expect(payoutToken).to.eq(depositedToken);
      const tx = await stakingNFT.collectTokenToMock(to, tokenID);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });

    it("collectTokenTo Test 2: two stakers", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();

      // Add Alice
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, amount);
      receipt = await txResponse.wait();
      let expTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expTokenID);
      let expTokenReserve = amount;
      let retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Add Bob
      const bobAddress = signers[1].address;
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, amount);
      receipt = await txResponse.wait();
      expTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expTokenID);
      expTokenReserve = amount.add(amount);
      retTokenReserve = await stakingNFT.getTotalReserveATokenMock();
      expect(retTokenReserve).to.eq(expTokenReserve);

      // Deposit Token
      const magic = BigNumber.from("42");
      const tokenString = "1234567890";
      const tokenScale = "1000000000000000000"; // 10**18
      const depositedToken = BigNumber.from(tokenString).mul(
        BigNumber.from(tokenScale)
      );
      const to = signers[0].address;
      await factoryCallAny(fixture.factory, aTokenMinter, "mint", [
        to,
        depositedToken,
      ]);
      await fixture.aToken.approve(stakingNFT.address, depositedToken);
      const tx3 = await stakingNFT.depositTokenMock(magic, depositedToken);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedTokenScaled = depositedToken.mul(scaleFactor);
      const expTokenAccumulator = BigNumber.from("0");
      const expTokenSlush = depositedTokenScaled;
      let retValue = await stakingNFT.getTokenAccumulator();
      let retTokenAccumulator = retValue[0];
      let retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // Check Token accumulator; there should be no changes
      retValue = await stakingNFT.getTokenAccumulator();
      retTokenAccumulator = retValue[0];
      retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // mine blocks
      const duration = BigNumber.from(2);
      await mineBlocks(duration.toBigInt());

      // collectTokenTo
      const alicePayoutToken = await stakingNFT.callStatic.collectTokenToMock(
        aliceAddress,
        aliceTokenID
      );
      expect(alicePayoutToken).to.eq(depositedToken.div(2));
      const tx = await stakingNFT.collectTokenToMock(
        aliceAddress,
        aliceTokenID
      );
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });

    it("collectTokenTo Test 3: fail because tokenID does not exist", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const tokenID = BigNumber.from(0);
      const tx = stakingNFT.collectTokenToMock(to, tokenID);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });

    it("collectTokenTo Test 4: fail because not owner", async () => {
      // Add staker
      const signers = await ethers.getSigners();
      const from = signers[0].address;
      const to = signers[1].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // collectTokenTo
      const tx = stakingNFT.collectTokenToMock(from, tokenID);
      // Error Code for Caller not token owner
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "CallerNotTokenOwner")
        .withArgs(from);
    });

    it("collectTokenTo Test 5: fail because not free to withdraw", async () => {
      // Add staker
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // collectTokenTo
      const tx = stakingNFT.collectTokenToMock(to, tokenID);
      // Error Code for freeAfter not reached
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "LockDurationWithdrawTimeNotReached"
      );
    });
  });

  describe("lockPosition", async () => {
    it("lockPosition Test 1: one staker, small lock", async () => {
      // Setup ability to perform call as Governance
      const hre = await require("hardhat"); // eslint-disable-line
      const govAddress = getMetamorphicAddress(
        fixture.factory.address,
        "0x476f7665726e616e636500000000000000000000000000000000000000000000"
      );
      await hre.network.provider.request({
        method: "hardhat_impersonateAccount",
        params: [govAddress],
      });
      const govSigner = await hre.ethers.getSigner(govAddress);
      // Get Signers
      const signers = await ethers.getSigners();
      // Send Eth to allow for call
      let txResponse: any = await signers[0].sendTransaction({
        to: govSigner.address,
        value: hre.ethers.utils.parseEther("1", "ether"),
        maxFeePerGas: 1,
        gasLimit: 30000000,
      });
      let receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // Stake position
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      let expFreeAfter = BigNumber.from(blockNumber).add(1);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      let retValue = await stakingNFT.getPositionMock(tokenID);
      let shares = retValue[0];
      let freeAfter = retValue[1];
      let withdrawFreeAfter = retValue[2];
      let accumulatorEth = retValue[3];
      let accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);

      // Lock position
      const duration = BigNumber.from("100");
      const tx = await stakingNFT
        .connect(govSigner)
        .lockPositionMock(to, tokenID, duration);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);

      // specifiy updated expected values
      expFreeAfter = BigNumber.from(blockNumber).add(1).add(duration);

      // Confirm updated values
      retValue = await stakingNFT.getPositionMock(tokenID);
      shares = retValue[0];
      freeAfter = retValue[1];
      withdrawFreeAfter = retValue[2];
      accumulatorEth = retValue[3];
      accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("lockPosition Test 2: one staker, long lock", async () => {
      // Setup ability to perform call as Governance
      const hre = await require("hardhat"); // eslint-disable-line
      const govAddress = getMetamorphicAddress(
        fixture.factory.address,
        "0x476f7665726e616e636500000000000000000000000000000000000000000000"
      );
      await hre.network.provider.request({
        method: "hardhat_impersonateAccount",
        params: [govAddress],
      });
      const govSigner = await hre.ethers.getSigner(govAddress);
      // Get Signers
      const signers = await ethers.getSigners();
      // Send Eth to allow for call
      let txResponse: any = await signers[0].sendTransaction({
        to: govSigner.address,
        value: hre.ethers.utils.parseEther("1", "ether"),
        maxFeePerGas: 1,
        gasLimit: 30000000,
      });
      let receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // Stake position
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      let expFreeAfter = BigNumber.from(blockNumber).add(1);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      let retValue = await stakingNFT.getPositionMock(tokenID);
      let shares = retValue[0];
      let freeAfter = retValue[1];
      let withdrawFreeAfter = retValue[2];
      let accumulatorEth = retValue[3];
      let accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);

      // Lock position
      const maxGovLock = await stakingNFT.getMaxGovernanceLock();
      const duration = maxGovLock;
      const tx = await stakingNFT
        .connect(govSigner)
        .lockPositionMock(to, tokenID, duration);
      receipt = await tx.wait();
      expect(receipt.status).to.eq(1);

      // specifiy updated expected values
      expFreeAfter = BigNumber.from(blockNumber).add(1).add(duration);

      // Confirm updated values
      retValue = await stakingNFT.getPositionMock(tokenID);
      shares = retValue[0];
      freeAfter = retValue[1];
      withdrawFreeAfter = retValue[2];
      accumulatorEth = retValue[3];
      accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("lockPosition Test 3: fail, circuit breaker tripped", async () => {
      const signers = await ethers.getSigners();
      const caller = signers[0].address;
      const tokenID = BigNumber.from("1");
      const duration = BigNumber.from("1");
      await stakingNFT.tripCBLowMock();
      const tx = stakingNFT.lockPositionMock(caller, tokenID, duration);
      // Error Code because failed circuit breaker is open
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerOpened"
      );
    });

    it("lockPosition Test 4: fail, not governance", async () => {
      const signers = await ethers.getSigners();
      const caller = signers[0].address;
      const tokenID = BigNumber.from("1");
      const duration = BigNumber.from("1");
      const tx = stakingNFT.lockPositionMock(caller, tokenID, duration);
      // Error Code because failed onlyGovernance
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "OnlyGovernance"
      );
    });

    it("lockPosition Test 5: fail, does not exist", async () => {
      // Setup ability to perform call as Governance
      const hre = await require("hardhat"); // eslint-disable-line
      const govAddress = getMetamorphicAddress(
        fixture.factory.address,
        "0x476f7665726e616e636500000000000000000000000000000000000000000000"
      );
      await hre.network.provider.request({
        method: "hardhat_impersonateAccount",
        params: [govAddress],
      });
      const govSigner = await hre.ethers.getSigner(govAddress);
      // Get Signers
      const signers = await ethers.getSigners();
      // Send Eth to allow for call
      const txResponse = await signers[0].sendTransaction({
        to: govSigner.address,
        value: hre.ethers.utils.parseEther("1", "ether"),
        maxFeePerGas: 1,
        gasLimit: 30000000,
      });
      const receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      const caller = signers[0].address;
      const tokenID = BigNumber.from("1");
      const duration = BigNumber.from("1");
      const tx = stakingNFT
        .connect(govSigner)
        .lockPositionMock(caller, tokenID, duration);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });

    it("lockPosition Test 6: fail, incorrect caller", async () => {
      // Setup ability to perform call as Governance
      const hre = await require("hardhat"); // eslint-disable-line
      const govAddress = getMetamorphicAddress(
        fixture.factory.address,
        "0x476f7665726e616e636500000000000000000000000000000000000000000000"
      );
      await hre.network.provider.request({
        method: "hardhat_impersonateAccount",
        params: [govAddress],
      });
      const govSigner = await hre.ethers.getSigner(govAddress);
      // Get Signers
      const signers = await ethers.getSigners();
      // Send Eth to allow for call
      let txResponse: any = await signers[0].sendTransaction({
        to: govSigner.address,
        value: hre.ethers.utils.parseEther("1", "ether"),
        maxFeePerGas: 1,
        gasLimit: 30000000,
      });
      let receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // Stake position
      const aliceAddress = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expTokenID);

      const bobAddress = signers[1].address;
      const duration = BigNumber.from("1");
      const tx = stakingNFT
        .connect(govSigner)
        .lockPositionMock(bobAddress, aliceTokenID, duration);
      // Error Code for Caller not token owner
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "CallerNotTokenOwner")
        .withArgs(govAddress);
    });

    it("lockPosition Test 7: fail, above max lock period", async () => {
      // Setup ability to perform call as Governance
      const hre = await require("hardhat"); // eslint-disable-line
      const govAddress = getMetamorphicAddress(
        fixture.factory.address,
        "0x476f7665726e616e636500000000000000000000000000000000000000000000"
      );
      await hre.network.provider.request({
        method: "hardhat_impersonateAccount",
        params: [govAddress],
      });
      const govSigner = await hre.ethers.getSigner(govAddress);
      // Get Signers
      const signers = await ethers.getSigners();
      // Send Eth to allow for call
      let txResponse: any = await signers[0].sendTransaction({
        to: govSigner.address,
        value: hre.ethers.utils.parseEther("1", "ether"),
        maxFeePerGas: 1,
        gasLimit: 30000000,
      });
      let receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // Stake position
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      const maxGovLock = await stakingNFT.getMaxGovernanceLock();
      const duration = maxGovLock.add(1);
      const tx = stakingNFT
        .connect(govSigner)
        .lockPositionMock(to, tokenID, duration);
      // Error Code for lock duration greater than max governance
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "LockDurationGreaterThanGovernanceLock"
      );
    });
  });

  describe("lockOwnPosition", async () => {
    it("lockOwnPosition Test 1: one staker, small lock", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      let expFreeAfter = BigNumber.from(blockNumber).add(1);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      let retValue = await stakingNFT.getPositionMock(tokenID);
      let shares = retValue[0];
      let freeAfter = retValue[1];
      let withdrawFreeAfter = retValue[2];
      let accumulatorEth = retValue[3];
      let accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);

      // Lock position
      const duration = BigNumber.from("100");
      txResponse = await stakingNFT.lockOwnPositionMock(tokenID, duration);
      receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // specifiy updated expected values
      expFreeAfter = BigNumber.from(blockNumber).add(1).add(duration);

      // Confirm updated values
      retValue = await stakingNFT.getPositionMock(tokenID);
      shares = retValue[0];
      freeAfter = retValue[1];
      withdrawFreeAfter = retValue[2];
      accumulatorEth = retValue[3];
      accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("lockOwnPosition Test 2: one staker, long lock", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      let expFreeAfter = BigNumber.from(blockNumber).add(1);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      let retValue = await stakingNFT.getPositionMock(tokenID);
      let shares = retValue[0];
      let freeAfter = retValue[1];
      let withdrawFreeAfter = retValue[2];
      let accumulatorEth = retValue[3];
      let accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);

      // Lock position
      const maxGovLock = await stakingNFT.getMaxGovernanceLock();
      const duration = maxGovLock;
      txResponse = await stakingNFT.lockOwnPositionMock(tokenID, duration);
      receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // specifiy updated expected values
      expFreeAfter = BigNumber.from(blockNumber).add(1).add(duration);

      // Confirm updated values
      retValue = await stakingNFT.getPositionMock(tokenID);
      shares = retValue[0];
      freeAfter = retValue[1];
      withdrawFreeAfter = retValue[2];
      accumulatorEth = retValue[3];
      accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("lockOwnPosition Test 3: fail, circuit breaker tripped", async () => {
      const tokenID = BigNumber.from("1");
      const duration = BigNumber.from("1");
      await stakingNFT.tripCBLowMock();
      const tx = stakingNFT.lockOwnPositionMock(tokenID, duration);
      // Error Code for circuit breaker tripped
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerOpened"
      );
    });

    it("lockOwnPosition Test 4: fail, does not exist", async () => {
      const tokenID = BigNumber.from("1");
      const duration = BigNumber.from("1");
      const tx = stakingNFT.lockOwnPositionMock(tokenID, duration);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });

    it("lockOwnPosition Test 5: fail, wrong owner", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const bobAddress = signers[1].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expTokenID);

      const duration = BigNumber.from("1");
      const tx = stakingNFT.lockOwnPositionMock(bobTokenID, duration);
      // Error Code for Caller not token owner
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "CallerNotTokenOwner")
        .withArgs(signers[0].address);
    });

    it("lockOwnPosition Test 6: fail, above max lock period", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      const maxGovLock = await stakingNFT.getMaxGovernanceLock();
      const duration = maxGovLock.add(1);
      const tx = stakingNFT.lockOwnPositionMock(tokenID, duration);
      // Error Code for lock duration greater than max governance
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "LockDurationGreaterThanGovernanceLock"
      );
    });
  });

  describe("lockPositionLow", async () => {
    it("lockPositionLow Test 1: one staker", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      let expFreeAfter = BigNumber.from(blockNumber).add(1);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      let retValue = await stakingNFT.getPositionMock(tokenID);
      let shares = retValue[0];
      let freeAfter = retValue[1];
      let withdrawFreeAfter = retValue[2];
      let accumulatorEth = retValue[3];
      let accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);

      // Lock position
      const duration = BigNumber.from("100");
      txResponse = await stakingNFT.lockPositionLowMock(tokenID, duration);
      receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // specifiy updated expected values
      expFreeAfter = BigNumber.from(blockNumber).add(1).add(duration);

      // Confirm updated values
      retValue = await stakingNFT.getPositionMock(tokenID);
      shares = retValue[0];
      freeAfter = retValue[1];
      withdrawFreeAfter = retValue[2];
      accumulatorEth = retValue[3];
      accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("lockPositionLow Test 2: fail", async () => {
      const tokenID = BigNumber.from(0);
      const duration = BigNumber.from(1);
      const tx = stakingNFT.lockPositionLowMock(tokenID, duration);
      // Error Code for Invalid NFT TokenID
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "InvalidTokenId"
      );
    });
  });

  describe("lockWithdraw", async () => {
    it("lockWithdraw Test 1: one staker, short lock", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      const expFreeAfter = BigNumber.from(blockNumber).add(1);
      let expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      let retValue = await stakingNFT.getPositionMock(tokenID);
      let shares = retValue[0];
      let freeAfter = retValue[1];
      let withdrawFreeAfter = retValue[2];
      let accumulatorEth = retValue[3];
      let accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);

      // Lock position
      const duration = BigNumber.from("100");
      txResponse = await stakingNFT.lockWithdrawMock(tokenID, duration);
      receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // specifiy updated expected values
      expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1).add(duration);

      // Confirm updated values
      retValue = await stakingNFT.getPositionMock(tokenID);
      shares = retValue[0];
      freeAfter = retValue[1];
      withdrawFreeAfter = retValue[2];
      accumulatorEth = retValue[3];
      accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("lockWithdraw Test 2: one staker, long lock", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      const expFreeAfter = BigNumber.from(blockNumber).add(1);
      let expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      let retValue = await stakingNFT.getPositionMock(tokenID);
      let shares = retValue[0];
      let freeAfter = retValue[1];
      let withdrawFreeAfter = retValue[2];
      let accumulatorEth = retValue[3];
      let accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);

      // Lock position
      const maxGovLock = await stakingNFT.getMaxGovernanceLock();
      const duration = maxGovLock;
      txResponse = await stakingNFT.lockWithdrawMock(tokenID, duration);
      receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // specifiy updated expected values
      expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1).add(duration);

      // Confirm updated values
      retValue = await stakingNFT.getPositionMock(tokenID);
      shares = retValue[0];
      freeAfter = retValue[1];
      withdrawFreeAfter = retValue[2];
      accumulatorEth = retValue[3];
      accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("lockWithdraw Test 3: fail, circuit breaker tripped", async () => {
      const tokenID = BigNumber.from("1");
      const duration = BigNumber.from("1");
      await stakingNFT.tripCBLowMock();
      const tx = stakingNFT.lockWithdrawMock(tokenID, duration);
      // Error Code for circuit breaker tripped
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerOpened"
      );
    });

    it("lockWithdraw Test 4: fail, does not exist", async () => {
      const tokenID = BigNumber.from("1");
      const duration = BigNumber.from("1");
      const tx = stakingNFT.lockWithdrawMock(tokenID, duration);
      // Error Code "ERC721: invalid token ID" for NFT token does not exist
      await expect(tx).to.be.revertedWith("ERC721: invalid token ID");
    });

    it("lockWithdraw Test 5: fail, wrong owner", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const bobAddress = signers[1].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expTokenID);

      const duration = BigNumber.from("1");
      const tx = stakingNFT.lockWithdrawMock(bobTokenID, duration);
      // Error Code for Caller not token owner
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "CallerNotTokenOwner")
        .withArgs(signers[0].address);
    });

    it("lockWithdraw Test 6: fail, above max lock period", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      const maxGovLock = await stakingNFT.getMaxGovernanceLock();
      const duration = maxGovLock.add(1);
      const tx = stakingNFT.lockWithdrawMock(tokenID, duration);
      // Error Code for lock duration greater than max governance
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "LockDurationGreaterThanGovernanceLock"
      );
    });
  });

  describe("lockWithdrawLow", async () => {
    it("lockWithdrawLow Test 1: one staker", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      const expFreeAfter = BigNumber.from(blockNumber).add(1);
      let expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info and confirm values
      let retValue = await stakingNFT.getPositionMock(tokenID);
      let shares = retValue[0];
      let freeAfter = retValue[1];
      let withdrawFreeAfter = retValue[2];
      let accumulatorEth = retValue[3];
      let accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);

      // Lock position
      const duration = BigNumber.from("100");
      txResponse = await stakingNFT.lockWithdrawLowMock(tokenID, duration);
      receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);

      // specifiy updated expected values
      expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1).add(duration);

      // Confirm updated values
      retValue = await stakingNFT.getPositionMock(tokenID);
      shares = retValue[0];
      freeAfter = retValue[1];
      withdrawFreeAfter = retValue[2];
      accumulatorEth = retValue[3];
      accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });
  });

  describe("getEthAccumulator", async () => {
    it("getEthAccumulator Test 1:", async () => {
      const expEthAccumulator = BigNumber.from("0");
      const expEthSlush = BigNumber.from("0");
      const retValue = await stakingNFT.getEthAccumulator();
      const retEthAccumulator = retValue[0];
      const retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
    });

    it("getEthAccumulator Test 2:", async () => {
      let expEthAccumulator = BigNumber.from("0");
      let expEthSlush = BigNumber.from("0");
      let retValue = await stakingNFT.getEthAccumulator();
      let retEthAccumulator = retValue[0];
      let retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      const magic = BigNumber.from("42");
      await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits("1", "ether"),
      });
      let expTotalReserveEth = BigNumber.from("1000000000000000000");
      let retTotalReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retTotalReserveEth).to.eq(expTotalReserveEth);
      expEthAccumulator = BigNumber.from("0");
      expEthSlush = BigNumber.from("1000000000000000000000000000000000000"); // 1 Eth * scale factor
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
      await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits("2", "ether"),
      });
      expTotalReserveEth = BigNumber.from("3000000000000000000");
      retTotalReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retTotalReserveEth).to.eq(expTotalReserveEth);
      expEthAccumulator = BigNumber.from("0");
      expEthSlush = BigNumber.from("3000000000000000000000000000000000000"); // 3 Eth * scale factor
      retValue = await stakingNFT.getEthAccumulator();
      retEthAccumulator = retValue[0];
      retEthSlush = retValue[1];
      expect(retEthAccumulator).to.eq(expEthAccumulator);
      expect(retEthSlush).to.eq(expEthSlush);
    });
  });

  describe("getTokenAccumulator", async () => {
    it("getTokenAccumulator Test 1:", async () => {
      const expTokenAccumulator = BigNumber.from("0");
      const expTokenSlush = BigNumber.from("0");
      const retValue = await stakingNFT.getTokenAccumulator();
      const retTokenAccumulator = retValue[0];
      const retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);
    });
  });

  describe("depositEth", async () => {
    it("depositEth Test 1:", async () => {
      let expTotalReserveEth = BigNumber.from("0");
      let retTotalReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retTotalReserveEth).to.eq(expTotalReserveEth);
      const magic = BigNumber.from("42");
      await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits("1", "ether"),
      });
      expTotalReserveEth = BigNumber.from("1000000000000000000");
      retTotalReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retTotalReserveEth).to.eq(expTotalReserveEth);
      await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits("2", "ether"),
      });
      expTotalReserveEth = BigNumber.from("3000000000000000000");
      retTotalReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retTotalReserveEth).to.eq(expTotalReserveEth);
    });

    it("depositEth Test 2: fail, circuit breaker tripped", async () => {
      await stakingNFT.tripCBLowMock();
      const magic = BigNumber.from("42");
      const tx = stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits("1", "ether"),
      });
      // Error Code for circuit breaker tripped
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerOpened"
      );
    });

    it("depositEth Test 3: fail, invalid magic number", async () => {
      const magic = BigNumber.from("0");
      const tx = stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits("1", "ether"),
      });
      // Error Code for bad magic
      await expect(tx).to.be.revertedWithCustomError(stakingNFT, "BadMagic");
    });
  });

  describe("depositToken", async () => {
    it("depositToken Test 1:", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      await factoryCallAny(fixture.factory, aTokenMinter, "mint", [to, amount]);
      await fixture.aToken.approve(stakingNFT.address, amount);
      let expTotalReserveAToken = BigNumber.from("0");
      let retTotalReserveAToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retTotalReserveAToken).to.eq(expTotalReserveAToken);
      const magic = BigNumber.from("42");
      await stakingNFT.depositTokenMock(magic, amount);
      expTotalReserveAToken = amount;
      retTotalReserveAToken = await stakingNFT.getTotalReserveATokenMock();
      expect(retTotalReserveAToken).to.eq(expTotalReserveAToken);
    });

    it("depositToken Test 2: fail, circuit breaker tripped", async () => {
      await stakingNFT.tripCBLowMock();
      const magic = BigNumber.from("42");
      const amount = BigNumber.from("1000000000000000000");
      const tx = stakingNFT.depositTokenMock(magic, amount);
      // Error Code for tripped circuit breaker
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerOpened"
      );
    });

    it("depositToken Test 3: fail, invalid magic number", async () => {
      const magic = BigNumber.from("0");
      const amount = BigNumber.from("1000000000000000000");
      const tx = stakingNFT.depositTokenMock(magic, amount);
      // Error Code for bad magic
      await expect(tx).to.be.revertedWithCustomError(stakingNFT, "BadMagic");
    });
  });

  describe("getTotal", async () => {
    it("getTotalShares Test:", async () => {
      const expTotalShares = BigNumber.from("0");
      const retTotalShares = await stakingNFT.getTotalSharesMock();
      expect(retTotalShares).to.eq(expTotalShares);
    });
    it("getTotalReserveEth Test:", async () => {
      const expTotalReserveEth = BigNumber.from("0");
      const retTotalReserveEth = await stakingNFT.getTotalReserveEthMock();
      expect(retTotalReserveEth).to.eq(expTotalReserveEth);
    });
    it("getTotalReserveAToken Test:", async () => {
      const expTotalReserveAToken = BigNumber.from("0");
      const retTotalReserveAToken =
        await stakingNFT.getTotalReserveATokenMock();
      expect(retTotalReserveAToken).to.eq(expTotalReserveAToken);
    });
  });

  describe("estimateCollection", async () => {
    it("estimateEthCollection Test 1; no Eth", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      const expEstimatedEth = BigNumber.from(0);
      const estimatedEth = await stakingNFT.estimateEthCollectionMock(tokenID);
      await expect(estimatedEth).to.eq(expEstimatedEth);
    });

    it("estimateEthCollection Test 2; Eth", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // deposit Eth to be distributed
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx2 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const receipt2 = await tx2.wait();
      expect(receipt2.status).to.eq(1);

      const expEstimatedEth = depositedEth;
      const estimatedEth = await stakingNFT.estimateEthCollectionMock(tokenID);
      await expect(estimatedEth).to.eq(expEstimatedEth);
    });

    it("estimateEthCollection Test 3; Eth, 2 stakers, all Alice", async () => {
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const aliceShares = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(
        stakingNFT.address,
        aliceShares
      );
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, aliceShares);
      receipt = await txResponse.wait();
      const expAliceTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expAliceTokenID);

      // deposit Eth to be distributed
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx2 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const receipt2 = await tx2.wait();
      expect(receipt2.status).to.eq(1);

      const bobAddress = signers[1].address;
      const bobShares = BigNumber.from("1000000000000000000");
      txResponse = await fixture.aToken.approve(stakingNFT.address, bobShares);
      // We use receipt:any *only* because of events which may not be present
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, bobShares);
      receipt = await txResponse.wait();
      const expBobTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expBobTokenID);

      // Check
      const expEstimatedEthAlice = depositedEth;
      const estimatedEthAlice = await stakingNFT.estimateEthCollectionMock(
        aliceTokenID
      );
      await expect(estimatedEthAlice).to.eq(expEstimatedEthAlice);
      const expEstimatedEthBob = BigNumber.from(0);
      const estimatedEthBob = await stakingNFT.estimateEthCollectionMock(
        bobTokenID
      );
      await expect(estimatedEthBob).to.eq(expEstimatedEthBob);
    });

    it("estimateEthCollection Test 4; Eth, 2 stakers, split", async () => {
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const aliceShares = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(
        stakingNFT.address,
        aliceShares
      );
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, aliceShares);
      receipt = await txResponse.wait();
      const expAliceTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expAliceTokenID);

      const bobAddress = signers[1].address;
      const bobShares = BigNumber.from("1000000000000000000");
      txResponse = await fixture.aToken.approve(stakingNFT.address, bobShares);
      // We use receipt:any *only* because of events which may not be present
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, bobShares);
      receipt = await txResponse.wait();
      const expBobTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expBobTokenID);

      // deposit Eth to be distributed
      const magic = BigNumber.from("42");
      const ethString = "1234567890";
      const etherScale = "1000000000000000000"; // 10**18
      const tx2 = await stakingNFT.depositEthMock(magic, {
        value: ethers.utils.parseUnits(ethString, "ether"),
      });
      const depositedEth = BigNumber.from(ethString).mul(
        BigNumber.from(etherScale)
      );
      const receipt2 = await tx2.wait();
      expect(receipt2.status).to.eq(1);

      // Check
      const expEstimatedEthAlice = depositedEth.div(2);
      const estimatedEthAlice = await stakingNFT.estimateEthCollectionMock(
        aliceTokenID
      );
      await expect(estimatedEthAlice).to.eq(expEstimatedEthAlice);
      const expEstimatedEthBob = depositedEth.div(2);
      const estimatedEthBob = await stakingNFT.estimateEthCollectionMock(
        bobTokenID
      );
      await expect(estimatedEthBob).to.eq(expEstimatedEthBob);
    });

    it("estimateEthCollection Test 5: fail", async () => {
      const tokenID = BigNumber.from("0");
      const tx = stakingNFT.estimateEthCollectionMock(tokenID);
      // InvalidTokenID
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "InvalidTokenId"
      );
    });

    it("estimateTokenCollection Test 1: no Token", async () => {
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      const expEstimatedToken = BigNumber.from(0);
      const estimatedToken = await stakingNFT.estimateTokenCollectionMock(
        tokenID
      );
      await expect(estimatedToken).to.eq(expEstimatedToken);
    });

    it("estimateTokenCollection Test 2: Token", async () => {
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      // Deposit Token
      const magic = BigNumber.from("42");
      const tokenString = "1234567890";
      const tokenScale = "1000000000000000000"; // 10**18
      const depositedToken = BigNumber.from(tokenString).mul(
        BigNumber.from(tokenScale)
      );
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      await factoryCallAny(fixture.factory, aTokenMinter, "mint", [
        to,
        depositedToken,
      ]);
      await fixture.aToken.approve(stakingNFT.address, depositedToken);
      const tx3 = await stakingNFT.depositTokenMock(magic, depositedToken);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedTokenScaled = depositedToken.mul(scaleFactor);
      const expTokenAccumulator = BigNumber.from("0");
      const expTokenSlush = depositedTokenScaled;
      const retValue = await stakingNFT.getTokenAccumulator();
      const retTokenAccumulator = retValue[0];
      const retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      const expEstimatedToken = depositedToken;
      const estimatedToken = await stakingNFT.estimateTokenCollectionMock(
        tokenID
      );
      await expect(estimatedToken).to.eq(expEstimatedToken);
    });

    it("estimateTokenCollection Test 3; Token, 2 stakers, all Alice", async () => {
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const aliceShares = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(
        stakingNFT.address,
        aliceShares
      );
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, aliceShares);
      receipt = await txResponse.wait();
      const expAliceTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expAliceTokenID);

      // Deposit Token
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const magic = BigNumber.from("42");
      const tokenString = "1234567890";
      const tokenScale = "1000000000000000000"; // 10**18
      const depositedToken = BigNumber.from(tokenString).mul(
        BigNumber.from(tokenScale)
      );
      const to = signers[0].address;
      await factoryCallAny(fixture.factory, aTokenMinter, "mint", [
        to,
        depositedToken,
      ]);
      await fixture.aToken.approve(stakingNFT.address, depositedToken);
      const tx3 = await stakingNFT.depositTokenMock(magic, depositedToken);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedTokenScaled = depositedToken.mul(scaleFactor);
      const expTokenAccumulator = BigNumber.from("0");
      const expTokenSlush = depositedTokenScaled;
      const retValue = await stakingNFT.getTokenAccumulator();
      const retTokenAccumulator = retValue[0];
      const retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // Add Bob
      const bobAddress = signers[1].address;
      const bobShares = BigNumber.from("1000000000000000000");
      txResponse = await fixture.aToken.approve(stakingNFT.address, bobShares);
      // We use receipt:any *only* because of events which may not be present
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, bobShares);
      receipt = await txResponse.wait();
      const expBobTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expBobTokenID);

      // Check
      const expEstimatedTokenAlice = depositedToken;
      const estimatedTokenAlice = await stakingNFT.estimateTokenCollectionMock(
        aliceTokenID
      );
      await expect(estimatedTokenAlice).to.eq(expEstimatedTokenAlice);
      const expEstimatedTokenBob = BigNumber.from(0);
      const estimatedTokenBob = await stakingNFT.estimateTokenCollectionMock(
        bobTokenID
      );
      await expect(estimatedTokenBob).to.eq(expEstimatedTokenBob);
    });

    it("estimateTokenCollection Test 4; Token, 2 stakers, split", async () => {
      const signers = await ethers.getSigners();
      const aliceAddress = signers[0].address;
      const aliceShares = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(
        stakingNFT.address,
        aliceShares
      );
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(aliceAddress, aliceShares);
      receipt = await txResponse.wait();
      const expAliceTokenID = BigNumber.from("1");
      const aliceTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(aliceTokenID).to.eq(expAliceTokenID);

      const bobAddress = signers[1].address;
      const bobShares = BigNumber.from("1000000000000000000");
      txResponse = await fixture.aToken.approve(stakingNFT.address, bobShares);
      receipt = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(bobAddress, bobShares);
      receipt = await txResponse.wait();
      const expBobTokenID = BigNumber.from("2");
      const bobTokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(bobTokenID).to.eq(expBobTokenID);

      // deposit Token to be distributed
      const scaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const magic = BigNumber.from("42");
      const tokenString = "1234567890";
      const tokenScale = "1000000000000000000"; // 10**18
      const depositedToken = BigNumber.from(tokenString).mul(
        BigNumber.from(tokenScale)
      );
      const to = signers[0].address;
      await factoryCallAny(fixture.factory, aTokenMinter, "mint", [
        to,
        depositedToken,
      ]);
      await fixture.aToken.approve(stakingNFT.address, depositedToken);
      const tx3 = await stakingNFT.depositTokenMock(magic, depositedToken);
      const receipt3 = await tx3.wait();
      expect(receipt3.status).to.eq(1);
      const depositedTokenScaled = depositedToken.mul(scaleFactor);
      const expTokenAccumulator = BigNumber.from("0");
      const expTokenSlush = depositedTokenScaled;
      const retValue = await stakingNFT.getTokenAccumulator();
      const retTokenAccumulator = retValue[0];
      const retTokenSlush = retValue[1];
      expect(retTokenAccumulator).to.eq(expTokenAccumulator);
      expect(retTokenSlush).to.eq(expTokenSlush);

      // Check
      const expEstimatedTokenAlice = depositedToken.div(2);
      const estimatedTokenAlice = await stakingNFT.estimateTokenCollectionMock(
        aliceTokenID
      );
      await expect(estimatedTokenAlice).to.eq(expEstimatedTokenAlice);
      const expEstimatedTokenBob = depositedToken.div(2);
      const estimatedTokenBob = await stakingNFT.estimateTokenCollectionMock(
        bobTokenID
      );
      await expect(estimatedTokenBob).to.eq(expEstimatedTokenBob);
    });

    it("estimateTokenCollection Test 5: fail", async () => {
      const tokenID = BigNumber.from(0);
      const tx = stakingNFT.estimateTokenCollectionMock(tokenID);
      // Error Code for Invalid NFT TokenID
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "InvalidTokenId"
      );
    });
  });

  describe("estimateExcess", async () => {
    it("estimateExcessEth Test:", async () => {
      const expExcessEth = BigNumber.from("0");
      const retExcessEth = await stakingNFT.estimateExcessEthMock();
      expect(retExcessEth).to.eq(expExcessEth);
    });
    it("estimateExcessToken Test:", async () => {
      const expExcessToken = BigNumber.from("0");
      const retExcessToken = await stakingNFT.estimateExcessTokenMock();
      expect(retExcessToken).to.eq(expExcessToken);
    });
  });

  describe("getPosition", async () => {
    it("getPosition Test 1", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      const to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      const blockNumber = await ethers.provider.getBlockNumber();
      const expFreeAfter = BigNumber.from(blockNumber).add(1);
      const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      const expAccumulatorEth = BigNumber.from(0);
      const expAccumulatorToken = BigNumber.from(0);

      // get position info
      const retValue = await stakingNFT.getPositionMock(tokenID);
      const shares = retValue[0];
      const freeAfter = retValue[1];
      const withdrawFreeAfter = retValue[2];
      const accumulatorEth = retValue[3];
      const accumulatorToken = retValue[4];
      expect(shares).to.eq(amount);
      expect(freeAfter).to.eq(expFreeAfter);
      expect(withdrawFreeAfter).to.eq(expWithdrawFreeAfter);
      expect(accumulatorEth).to.eq(expAccumulatorEth);
      expect(accumulatorToken).to.eq(expAccumulatorToken);
    });

    it("getPosition Test 2: Fail", async () => {
      const tokenID = BigNumber.from("0");
      const tx = stakingNFT.getPositionMock(tokenID);
      // Error Code for Invalid NFT TokenID
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "InvalidTokenId"
      );
    });
  });

  describe("tokenURI", async () => {
    // TODO: fix test
    // Something is wrong with this test
    /*
    it("tokenURI Test 1", async () => {
      // Stake position
      const signers = await ethers.getSigners();
      let to = signers[0].address;
      const amount = BigNumber.from("1000000000000000000");
      let txResponse = await fixture.aToken.approve(stakingNFT.address, amount);
      // We use receipt:any *only* because of events which may not be present
      let receipt: any = await txResponse.wait();
      txResponse = await stakingNFT.mintNFTMock(to, amount);
      receipt = await txResponse.wait();
      const expTokenID = BigNumber.from("1");
      const tokenID = BigNumber.from(receipt.events[2].args.tokenId);
      expect(tokenID).to.eq(expTokenID);

      // Get block number
      //const blockNumber = await ethers.provider.getBlockNumber();
      //const expFreeAfter = BigNumber.from(blockNumber).add(1);
      //const expWithdrawFreeAfter = BigNumber.from(blockNumber).add(1);
      //const expAccumulatorEth = BigNumber.from(0);
      //const expAccumulatorToken = BigNumber.from(0);

      // get position info
      const retValue = await stakingNFT.tokenURIMock(tokenID);
    });
    */

    it("tokenURI Test 2: Fail", async () => {
      const tokenID = BigNumber.from("0");
      const tx = stakingNFT.tokenURIMock(tokenID);
      // Error Code for Invalid NFT TokenID
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "InvalidTokenId"
      );
    });
  });

  describe("circuitBreaker", async () => {
    it("circuitBreaker Test 1 (low)", async () => {
      let expCircuitBreaker = false;
      let retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
      await stakingNFT.tripCBLowMock();
      expCircuitBreaker = true;
      retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
      await stakingNFT.resetCBLowMock();
      expCircuitBreaker = false;
      retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
    });

    it("circuitBreaker Test 2", async () => {
      let expCircuitBreaker = false;
      let retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
      await factoryCallAny(fixture.factory, stakingNFT, "tripCBMock");
      expCircuitBreaker = true;
      retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
      await stakingNFT.resetCBLowMock();
      expCircuitBreaker = false;
      retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
    });

    it("circuitBreaker Test 3: fail, trip CB when open (low)", async () => {
      let expCircuitBreaker = false;
      let retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
      await stakingNFT.tripCBLowMock();
      expCircuitBreaker = true;
      retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
      const tx = stakingNFT.tripCBLowMock();
      // Error Code for circuit breaker is open
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerOpened"
      );
    });

    it("circuitBreaker Test 4: fail, reset CB when closed (low)", async () => {
      const expCircuitBreaker = false;
      const retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
      const tx = stakingNFT.resetCBLowMock();
      // Error Code for circuit breaker is closed
      await expect(tx).to.be.revertedWithCustomError(
        stakingNFT,
        "CircuitBreakerClosed"
      );
    });

    it("circuitBreaker Test 5: fail, trip CB when open (not factory)", async () => {
      const signers = await ethers.getSigners();
      const expCircuitBreaker = false;
      const retCircuitBreaker = await stakingNFT.circuitBreakerStateMock();
      expect(retCircuitBreaker).to.eq(expCircuitBreaker);
      const tx = stakingNFT.tripCBMock();
      // Error Code for not calling as factory
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "OnlyFactory")
        .withArgs(signers[0].address, fixture.factory.address);
    });
  });

  describe("getAccumulatorScaleFactor", async () => {
    it("getAccumulatorScaleFactor Test", async () => {
      const expAccumulatorScaleFactor = BigNumber.from("1000000000000000000");
      const retAccumulatorScaleFactor =
        await stakingNFT.getAccumulatorScaleFactor();
      expect(retAccumulatorScaleFactor).to.eq(expAccumulatorScaleFactor);
    });
  });

  describe("getMaxMintLock", async () => {
    it("getMaxMintLock Test", async () => {
      const expMaxMintLock = BigNumber.from("1051200");
      const retMaxMintLock = await stakingNFT.getMaxMintLock();
      expect(retMaxMintLock).to.eq(expMaxMintLock);
    });
  });

  describe("counter", async () => {
    it("counter Test", async () => {
      let expCounter = BigNumber.from("0");
      let retCounter = await stakingNFT.getCountMock();
      expect(retCounter).to.eq(expCounter);
      await stakingNFT.incrementMock();
      await stakingNFT.incrementMock();
      retCounter = await stakingNFT.getCountMock();
      expCounter = BigNumber.from("2");
      expect(retCounter).to.eq(expCounter);
      await stakingNFT.incrementMock();
      await stakingNFT.incrementMock();
      await stakingNFT.incrementMock();
      await stakingNFT.incrementMock();
      retCounter = await stakingNFT.getCountMock();
      expCounter = BigNumber.from("6");
      expect(retCounter).to.eq(expCounter);
    });
  });

  describe("collect", async () => {
    it("collectPure Test 1; normal (all zero)", async () => {
      // Test Setup
      const shares = BigNumber.from("1000");
      const stateAccum = BigNumber.from("0");
      const stateSlush = BigNumber.from("0");
      const state = { accumulator: stateAccum, slush: stateSlush };
      const positionShares = BigNumber.from("1000");
      const positionFreeAfter = BigNumber.from("1");
      const positionWithdrawFreeAfter = BigNumber.from("1");
      const positionAccumulatorEth = BigNumber.from("0");
      const positionAccumulatorToken = BigNumber.from("0");
      const position = {
        shares: positionShares,
        freeAfter: positionFreeAfter,
        withdrawFreeAfter: positionWithdrawFreeAfter,
        accumulatorEth: positionAccumulatorEth,
        accumulatorToken: positionAccumulatorToken,
      };
      const positionAccumulatorValue = BigNumber.from("0");

      // Expected Values
      const expPayout = BigNumber.from("0");
      const expUpdatedPositionAccumulatorValue = positionAccumulatorValue;
      const expStateAccumulator = BigNumber.from("0");
      const expStateSlush = BigNumber.from("0");
      const expPositionShares = positionShares;

      // Run collect
      const returned = await stakingNFT.collectPure(
        shares,
        state,
        position,
        positionAccumulatorValue
      );
      const retUpdatedState = returned[0];
      const retUpdatedPosition = returned[1];
      const retUpdatedPositionAccumulatorValue = returned[2];
      const retPayout = returned[3];

      // Verify returned values
      expect(retPayout).to.eq(expPayout);
      expect(retUpdatedPositionAccumulatorValue).to.eq(
        expUpdatedPositionAccumulatorValue
      );
      expect(retUpdatedState.accumulator).to.eq(expStateAccumulator);
      expect(retUpdatedState.slush).to.eq(expStateSlush);
      expect(retUpdatedPosition.shares).to.eq(expPositionShares);
      expect(retUpdatedPositionAccumulatorValue).to.eq(
        expUpdatedPositionAccumulatorValue
      );
      expect(retPayout).to.eq(expPayout);
    });

    it("collectPure Test 2; normal (state nonzero, no slush, no wraparound)", async () => {
      // Test Setup
      const shares = BigNumber.from("1");
      const stateAccum = BigNumber.from("10000000000000000000");
      const stateSlush = BigNumber.from("0");
      const state = { accumulator: stateAccum, slush: stateSlush };
      const positionShares = BigNumber.from("1");
      const positionFreeAfter = BigNumber.from("1");
      const positionWithdrawFreeAfter = BigNumber.from("1");
      const positionAccumulatorEth = BigNumber.from("0");
      const positionAccumulatorToken = BigNumber.from("0");
      const position = {
        shares: positionShares,
        freeAfter: positionFreeAfter,
        withdrawFreeAfter: positionWithdrawFreeAfter,
        accumulatorEth: positionAccumulatorEth,
        accumulatorToken: positionAccumulatorToken,
      };
      const positionAccumulatorValue = BigNumber.from("0");

      const accumScaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      // Compute Expected Values
      let expStateSlush = stateSlush;
      const accumDelta = stateAccum.sub(positionAccumulatorValue);
      let tmp = accumDelta.mul(positionShares);
      if (shares === positionShares) {
        tmp = tmp.add(stateSlush);
        expStateSlush = BigNumber.from("0");
      }
      const expPayout = tmp.div(accumScaleFactor);
      const payoutRem = tmp.sub(expPayout.mul(accumScaleFactor));
      const expPositionAccumulatorValue = stateAccum;
      const expStateAccumulator = stateAccum;
      expStateSlush = expStateSlush.add(payoutRem);
      const expPositionShares = positionShares;

      // Run collect
      const returned = await stakingNFT.collectPure(
        shares,
        state,
        position,
        positionAccumulatorValue
      );
      const retState = returned[0];
      const retPosition = returned[1];
      const retPositionAccumulatorValue = returned[2];
      const retPayout = returned[3];

      // Verify returned values
      expect(retPositionAccumulatorValue).to.eq(expPositionAccumulatorValue);
      expect(retState.accumulator).to.eq(expStateAccumulator);
      expect(retState.slush).to.eq(expStateSlush);
      expect(retPosition.shares).to.eq(expPositionShares);
      expect(retPayout).to.eq(expPayout);
    });

    it("collectPure Test 3; normal (state nonzero, slush, no wraparound)", async () => {
      // Test Setup
      const shares = BigNumber.from("10000");
      const stateAccum = BigNumber.from("10000000000000000000");
      const stateSlush = BigNumber.from("123456789");
      const state = { accumulator: stateAccum, slush: stateSlush };
      const positionShares = BigNumber.from("1234");
      const positionFreeAfter = BigNumber.from("1");
      const positionWithdrawFreeAfter = BigNumber.from("1");
      const positionAccumulatorEth = BigNumber.from("0");
      const positionAccumulatorToken = BigNumber.from("0");
      const position = {
        shares: positionShares,
        freeAfter: positionFreeAfter,
        withdrawFreeAfter: positionWithdrawFreeAfter,
        accumulatorEth: positionAccumulatorEth,
        accumulatorToken: positionAccumulatorToken,
      };
      const positionAccumulatorValue = BigNumber.from("0");

      const accumScaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const stateSlushAfterSkim = stateSlush.sub(
        stateSlush.div(shares).mul(shares)
      );
      const stateAccumAfterSkim = stateAccum.add(stateSlush.div(shares));
      // Compute Expected Values
      let expStateSlush = stateSlushAfterSkim;
      const accumDelta = stateAccumAfterSkim.sub(positionAccumulatorValue);
      let tmp = accumDelta.mul(positionShares);
      if (shares === positionShares) {
        tmp = tmp.add(stateSlush);
        expStateSlush = BigNumber.from("0");
      }
      const expPayout = tmp.div(accumScaleFactor);
      const payoutRem = tmp.sub(expPayout.mul(accumScaleFactor));
      const expPositionAccumulatorValue = stateAccumAfterSkim;
      const expStateAccumulator = stateAccumAfterSkim;
      expStateSlush = expStateSlush.add(payoutRem);
      const expPositionShares = positionShares;

      // Run collect
      const returned = await stakingNFT.collectPure(
        shares,
        state,
        position,
        positionAccumulatorValue
      );
      const retState = returned[0];
      const retPosition = returned[1];
      const retPositionAccumulatorValue = returned[2];
      const retPayout = returned[3];

      // Verify returned values
      expect(retPositionAccumulatorValue).to.eq(expPositionAccumulatorValue);
      expect(retState.accumulator).to.eq(expStateAccumulator);
      expect(retState.slush).to.eq(expStateSlush);
      expect(retPosition.shares).to.eq(expPositionShares);
      expect(retPayout).to.eq(expPayout);
    });

    it("collectPure Test 4; normal (state nonzero, slush, wraparound)", async () => {
      // Test Setup
      const twoPower168 = BigNumber.from(
        "374144419156711147060143317175368453031918731001856"
      );
      const shares = BigNumber.from("1000000");
      const stateAccum = BigNumber.from("1234567890");
      const stateSlush = BigNumber.from("987654321");
      const state = { accumulator: stateAccum, slush: stateSlush };
      const positionShares = BigNumber.from("13579");
      const positionFreeAfter = BigNumber.from("1");
      const positionWithdrawFreeAfter = BigNumber.from("1");
      const positionAccumulatorEth = BigNumber.from("0");
      const positionAccumulatorToken = BigNumber.from("0");
      const position = {
        shares: positionShares,
        freeAfter: positionFreeAfter,
        withdrawFreeAfter: positionWithdrawFreeAfter,
        accumulatorEth: positionAccumulatorEth,
        accumulatorToken: positionAccumulatorToken,
      };
      const positionAccumulatorValue = BigNumber.from(
        "374144419156711147059143317175368453031918731001856"
      ); // 2**168 - 10**30

      const accumScaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const stateSlushAfterSkim = stateSlush.sub(
        stateSlush.div(shares).mul(shares)
      );
      const stateAccumAfterSkim = stateAccum.add(stateSlush.div(shares));
      // Compute Expected Values
      let expStateSlush = stateSlushAfterSkim;
      const accumDelta = stateAccumAfterSkim.add(
        twoPower168.sub(positionAccumulatorValue)
      );
      let tmp = accumDelta.mul(positionShares);
      if (shares === positionShares) {
        tmp = tmp.add(stateSlush);
        expStateSlush = BigNumber.from("0");
      }
      const expPayout = tmp.div(accumScaleFactor);
      const payoutRem = tmp.sub(expPayout.mul(accumScaleFactor));
      const expPositionAccumulatorValue = stateAccumAfterSkim;
      const expStateAccumulator = stateAccumAfterSkim;
      expStateSlush = expStateSlush.add(payoutRem);
      const expPositionShares = positionShares;

      // Run collect
      const returned = await stakingNFT.collectPure(
        shares,
        state,
        position,
        positionAccumulatorValue
      );
      const retState = returned[0];
      const retPosition = returned[1];
      const retPositionAccumulatorValue = returned[2];
      const retPayout = returned[3];

      // Verify returned values
      expect(retPositionAccumulatorValue).to.eq(expPositionAccumulatorValue);
      expect(retState.accumulator).to.eq(expStateAccumulator);
      expect(retState.slush).to.eq(expStateSlush);
      expect(retPosition.shares).to.eq(expPositionShares);
      expect(retPayout).to.eq(expPayout);
    });

    it("collect gas Test", async () => {
      // Test Setup
      const shares = BigNumber.from("10000");
      const stateAccum = BigNumber.from("10000000000000000000");
      const stateSlush = BigNumber.from("123456789");
      const state = { accumulator: stateAccum, slush: stateSlush };
      const positionShares = BigNumber.from("1234");
      const positionFreeAfter = BigNumber.from("1");
      const positionWithdrawFreeAfter = BigNumber.from("1");
      const positionAccumulatorEth = BigNumber.from("0");
      const positionAccumulatorToken = BigNumber.from("0");
      const position = {
        shares: positionShares,
        freeAfter: positionFreeAfter,
        withdrawFreeAfter: positionWithdrawFreeAfter,
        accumulatorEth: positionAccumulatorEth,
        accumulatorToken: positionAccumulatorToken,
      };
      const positionAccumulatorValue = BigNumber.from("1");
      // Run collect
      const txResponse = await stakingNFT.collectMock(
        shares,
        state,
        position,
        positionAccumulatorValue
      );
      const receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);
    });
  });

  describe("slushSkim", async () => {
    it("slushSkimPure Test 1; normal", async () => {
      const shares = BigNumber.from("1000000");
      const accumulator = BigNumber.from("1234567890");
      const slush = BigNumber.from("5678901234");
      const deltaAccum = slush.div(shares);
      const slushExp = slush.sub(deltaAccum.mul(shares));
      const accumulatorExp = accumulator.add(deltaAccum);
      const returned = await stakingNFT.slushSkimPure(
        shares,
        accumulator,
        slush
      );
      const accumulatorRet = returned[0];
      const slushRet = returned[1];
      expect(slushRet.toNumber()).to.eq(slushExp);
      expect(accumulatorRet.toNumber()).to.eq(accumulatorExp);
    });

    it("slushSkimPure Test 2; wrap around", async () => {
      const shares = BigNumber.from("1000");
      // 2**168 - 1
      const accumulator = BigNumber.from(
        "374144419156711147060143317175368453031918731001855"
      );
      const slush = BigNumber.from("1001");
      const slushExp = BigNumber.from("1");
      const accumulatorExp = BigNumber.from("0");
      const returned = await stakingNFT.slushSkimPure(
        shares,
        accumulator,
        slush
      );
      const accumulatorRet = returned[0];
      const slushRet = returned[1];
      expect(slushRet).to.eq(slushExp);
      expect(accumulatorRet).to.eq(accumulatorExp);
    });
    // Need to know how to work with BigInt

    // Curly braces on the fly
    it("slushSkim gas Test", async () => {
      // Base Tx gas cost: 21K gas
      // Thus, subtract 21K off of all stated gas costs.
      const shares = BigNumber.from("1111111111111111");
      const accumulator = BigNumber.from("123456789012345");
      const slush = BigNumber.from("567890123456789");
      const txResponse = await stakingNFT.slushSkimMock(
        shares,
        accumulator,
        slush
      );
      const receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);
    });
  });

  describe("deposit", async () => {
    it("depositPure Test 1; normal (deposit == 0)", async () => {
      const delta = BigNumber.from("0");
      const stateAccum = BigNumber.from("5678901234");
      const stateSlush = BigNumber.from("1234");
      const state = { accumulator: stateAccum, slush: stateSlush };
      const accumScaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const stateSlushExp = stateSlush.add(delta.mul(accumScaleFactor));
      const stateExp = { accumulator: stateAccum, slush: stateSlushExp };
      const returned = await stakingNFT.depositPure(delta, state);
      const stateRet = returned;
      expect(stateRet.accumulator).to.eq(stateExp.accumulator);
      expect(stateRet.slush).to.eq(stateExp.slush);
    });

    it("depositPure Test 2; normal (deposit > 0)", async () => {
      const delta = BigNumber.from("1234567890");
      const stateAccum = BigNumber.from("5678901234");
      const stateSlush = BigNumber.from("1234");
      const state = { accumulator: stateAccum, slush: stateSlush };
      const accumScaleFactor = await stakingNFT.getAccumulatorScaleFactor();
      const stateSlushExp = stateSlush.add(delta.mul(accumScaleFactor));
      const stateExp = { accumulator: stateAccum, slush: stateSlushExp };
      const returned = await stakingNFT.depositPure(delta, state);
      const stateRet = returned;
      expect(stateRet.accumulator).to.eq(stateExp.accumulator);
      expect(stateRet.slush).to.eq(stateExp.slush);
    });

    it("depositPure Test 3; fail: (failed require [slush overflow])", async () => {
      // Set initial values
      const delta = BigNumber.from("0");
      const stateAccumInitial = BigNumber.from("0");
      // 2**167
      const stateSlushInitial = BigNumber.from(
        "187072209578355573530071658587684226515959365500928"
      );
      const state = {
        accumulator: stateAccumInitial,
        slush: stateSlushInitial,
      };
      // Call deposit
      const tx = stakingNFT.depositPure(delta, state);
      // Error Code for Slush Too Large; must have slush < 2**167
      await expect(tx)
        .to.be.revertedWithCustomError(stakingNFT, "SlushTooLarge")
        .withArgs(stateSlushInitial);
    });

    it("deposit gas Test", async () => {
      const delta = BigNumber.from("1234567890");
      const stateAccum = BigNumber.from("5678901234");
      const stateSlush = BigNumber.from("1234");
      const state = { accumulator: stateAccum, slush: stateSlush };
      const txResponse = await stakingNFT.depositMock(delta, state);
      const receipt = await txResponse.wait();
      expect(receipt.status).to.eq(1);
    });
  });
});
