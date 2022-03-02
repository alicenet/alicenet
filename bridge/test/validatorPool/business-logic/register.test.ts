import {
  Fixture,
  getValidatorEthAccount,
  getFixture,
  factoryCallAny,
} from "../../setup";
import { completeETHDKGRound } from "../../ethdkg/setup";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  validatorsSnapshots,
} from "../../snapshots/assets/4-validators-snapshots-1";
import { BigNumber, ContractTransaction, Signer } from "ethers";
import { commitSnapshots, createValidators, getCurrentState, showState, stakeValidators } from "../setup";

describe("ValidatorPool: Registration logic", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let maxNumValidators = 4; //default number
  let stakeAmount = 20000;
  let stakeAmountMadWei = ethers.utils.parseUnits(stakeAmount.toString(), 18);
  let validators: string[];
  let stakingTokenIds: BigNumber[];
  let dummyAddress = "0x000000000000000000000000000000000000dEaD";

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
  });

  it("Should not allow registering validators if the STAKENFT position doesn’t have enough MADTokens staked", async function () {
    let rcpt = await factoryCallAny(
      fixture,
      "validatorPool",
      "setStakeAmount",
      [stakeAmountMadWei.add(1)]
    ); //Add 1 to max amount allowed
    expect(rcpt.status).to.equal(1);
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith(
      "ValidatorStakeNFT: Error, the Stake position doesn't have enough funds!"
    );
  });

  it("Should not allow registering more validators that the current number of free spots in the pool", async function () {
    await factoryCallAny(fixture, "validatorPool", "setMaxNumValidators", [
      maxNumValidators - 1,
    ]); // Set maxNumValidators to 1 under default number of validators
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: There are not enough free spots for all new validators!"
    );
  });

  it("Should not allow registering validators if the size of the input data is not correct", async function () {
    validators.length = maxNumValidators;
    stakingTokenIds.length = maxNumValidators - 1;
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: Both input array should have same length!"
    );
  });

  it('Should not allow registering validators if "Madnet consensus is running" or ETHDKG round is running', async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    // Complete ETHDKG Round
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: Error Madnet Consensus should be halted!"
    );
  });

  it("Should not allow registering validators if the STAKENFT position was not given permissions for the ValidatorPool contract burn it", async function () {
    for (const tokenID of stakingTokenIds) {
      //Disallow validatorPool to withdraw validator's NFT from StakeNFT
      await factoryCallAny(fixture, "stakeNFT", "approve", [
        dummyAddress,
        tokenID,
      ]);
    }
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("ERC721: transfer caller is not owner nor approved");
  });

  it("Should not allow registering an address that is already a validator", async function () {
    //Clone validators array
    let _validatorsSnapshots = validatorsSnapshots.slice();
    //Repeat the first validator
    _validatorsSnapshots[1] = _validatorsSnapshots[0];
    let newValidators = await createValidators(fixture, _validatorsSnapshots);
    //Approve first validator for twice the amount
    await fixture.madToken
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .approve(fixture.stakeNFT.address, stakeAmountMadWei.mul(2));
    await stakeValidators(fixture, newValidators);
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        newValidators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: Address is already a validator or it is in the exiting line!"
    );
  });

  it("Should not allow registering an address that is in the exiting queue", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    showState("after registering", await getCurrentState(fixture, validators));
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    showState("after un-registering", await getCurrentState(fixture, validators));
    let newValidators= await createValidators(fixture, validatorsSnapshots);
    let newTokensIds = await stakeValidators(fixture, newValidators);
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        newValidators,
        newTokensIds,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: Address is already a validator or it is in the exiting line!"
    );
  });

  it("Should successfully register validators if all conditions are met", async function () {
    let expectedState = await getCurrentState(fixture, validators);
    //Expect that NFTs are transferred from each validator to ValidatorPool sd ValidatorNFTs
    validators.map((_, index) => {
      expectedState.Factory.StakeNFT--;
      expectedState.ValidatorPool.ValNFT++;
      expectedState.validators[index].Acc = true
      expectedState.validators[index].Reg = true
    });
    //Expect that all validators funds are transferred from StakeNFT to ValidatorNFT
    expectedState.StakeNFT.MAD -= stakeAmount * validators.length;
    expectedState.ValidatorNFT.MAD += stakeAmount * validators.length;
    // Register validators
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let currentState = await getCurrentState(fixture, validators);
    await showState("after registering", currentState);
    await showState("Expected state after registering", expectedState);
    await showState("Current state after registering", currentState);

    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Set and get the validator's location after registering", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let currentState = await getCurrentState(fixture, validators);
    await showState("after registering", currentState);
    await fixture.validatorPool
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .setLocation("1.1.1.1");
    expect(
      await fixture.validatorPool.getLocation(await validators[0])
    ).to.be.equal("1.1.1.1");
  });

  it("Should not allow non-validator to set an IP location", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let currentState = await getCurrentState(fixture, validators);
    await showState("after registering", currentState);
    // Original Validators has all the validators not only the maximun to register
    // Position 5 is not a register validator
    await expect(
      fixture.validatorPool.connect(adminSigner).setLocation("1.1.1.1")
    ).to.be.revertedWith("ValidatorPool: Only validators allowed!");
  });

  it("Should not allow users to register reusing previous stakeNFT that cannot be burned", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await commitSnapshots(fixture, 4);
    stakingTokenIds = [];
    for (const validator of validatorsSnapshots) {
      let claimTx = (await fixture.validatorPool
        .connect(await getValidatorEthAccount(validator))
        .claimExitingNFTPosition()) as ContractTransaction;
      let receipt = await ethers.provider.getTransactionReceipt(claimTx.hash);
      //When a token is claimed it gets burned an minted so it gets a new tokenId so we need to update array for re-registration
      const tokenId = BigNumber.from(receipt.logs[0].topics[3]);
      stakingTokenIds.push(tokenId);
      // Transfer to factory for re-registering
      await fixture.stakeNFT
        .connect(await getValidatorEthAccount(validator))
        .transferFrom(validator.address, fixture.factory.address, tokenId);
      // Approve new tokensIds
      await factoryCallAny(fixture, "stakeNFT", "approve", [
        fixture.validatorPool.address,
        tokenId,
      ]);
    }
    await showState(
      "After claiming:",
      await getCurrentState(fixture, validators)
    );
    // After claiming, position is locked for a period of 172800 Madnet epochs
    // To perform this test with no revert, POSITION_LOCK_PERIOD can be set to 3
    // in ValidatorPool and then take 4 snapshots
    // await commitSnapshots(fixture, 4)
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("StakeNFT: The position is not ready to be burned!");
  });
});
