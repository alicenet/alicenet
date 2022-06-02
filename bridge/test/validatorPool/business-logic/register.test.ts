import { BigNumber, ContractTransaction, Signer } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import { completeETHDKGRound } from "../../ethdkg/setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
  mineBlocks,
} from "../../setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import {
  burnStakeTo,
  commitSnapshots,
  createValidators,
  getCurrentState,
  showState,
  stakeValidators,
} from "../setup";

describe("ValidatorPool: Registration logic", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  const maxNumValidators = 4; // default number
  let stakeAmount: bigint;
  let validators: string[];
  let stakingTokenIds: BigNumber[];
  const dummyAddress = "0x000000000000000000000000000000000000dEaD";

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
    stakeAmount = (await fixture.validatorPool.getStakeAmount()).toBigInt();
  });

  it("Should not allow registering validators if the PublicStaking position doesn’t have enough Tokens staked", async function () {
    const rcpt = await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "setStakeAmount",
      [stakeAmount + BigInt(1)]
    ); // Add 1 to max amount allowed
    expect(rcpt.status).to.equal(1);
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("821");
  });

  it("Should not allow registering more validators that the current number of free spots in the pool", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "setMaxNumValidators",
      [maxNumValidators - 1]
    ); // Set maxNumValidators to 1 under default number of validators
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("805");
  });

  it("Should not allow registering validators if the size of the input state is not correct", async function () {
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators.slice(0, 3),
        stakingTokenIds,
      ])
    ).to.be.revertedWith("806");
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds.slice(0, 3),
      ])
    ).to.be.revertedWith("806");
  });

  it('Should not allow registering validators if "AliceNet consensus is running" or ETHDKG round is running', async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    // Complete ETHDKG Round
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("802");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("801");
  });

  it("Should not allow registering validators if the PublicStaking position was not given permissions for the ValidatorPool contract burn it", async function () {
    for (const tokenID of stakingTokenIds) {
      // Disallow validatorPool to withdraw validator's NFT from PublicStaking
      await factoryCallAnyFixture(fixture, "publicStaking", "approve", [
        dummyAddress,
        tokenID,
      ]);
    }
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("ERC721: transfer caller is not owner nor approved");
  });

  it("Should not allow registering an address that is already a validator", async function () {
    // Clone validators array
    const _validatorsSnapshots = validatorsSnapshots.slice();
    // Repeat the first validator
    _validatorsSnapshots[1] = _validatorsSnapshots[0];
    const newValidators = await createValidators(fixture, _validatorsSnapshots);
    // Approve first validator for twice the amount
    await fixture.aToken
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .approve(fixture.publicStaking.address, stakeAmount * BigInt(2));
    await stakeValidators(fixture, newValidators);
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        newValidators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("816");
  });

  it("Should not allow registering an address that is in the exiting queue", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );
    const newValidators = await createValidators(fixture, validatorsSnapshots);
    const newTokensIds = await stakeValidators(fixture, newValidators);
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        newValidators,
        newTokensIds,
      ])
    ).to.be.revertedWith("816");
  });

  it("Should successfully register validators if all conditions are met", async function () {
    const expectedState = await getCurrentState(fixture, validators);
    // Expect that NFTs are transferred from each validator from ValidatorPool to ValidatorStaking
    for (let index = 0; index < validators.length; index++) {
      expectedState.Factory.PublicStaking--;
      expectedState.ValidatorPool.ValNFT++;
      expectedState.validators[index].Acc = true;
      expectedState.validators[index].Reg = true;
    }
    // Expect that all validators funds are transferred from PublicStaking to ValidatorStaking
    expectedState.PublicStaking.ATK -= stakeAmount * BigInt(validators.length);
    expectedState.ValidatorStaking.ATK +=
      stakeAmount * BigInt(validators.length);
    // Register validators
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const currentState = await getCurrentState(fixture, validators);
    await showState("after registering", currentState);
    await showState("Expected state after registering", expectedState);
    await showState("Current state after registering", currentState);

    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Should be able to register validators even if the contract has excess of Tokens and eth", async function () {
    // Mint a publicStaking and burn it to the ValidatorPool contract. Besides a contract self destructing
    // itself, this is a method to send eth accidentally to the validatorPool contract
    const etherAmount = ethers.utils.parseEther("1");
    const aTokenAmount = ethers.utils.parseEther("2");
    await burnStakeTo(fixture, etherAmount, aTokenAmount, adminSigner);

    const expectedState = await getCurrentState(fixture, validators);
    // Expect that NFTs are transferred from each validator from ValidatorPool to ValidatorStaking
    for (let index = 0; index < validators.length; index++) {
      expectedState.Factory.PublicStaking--;
      expectedState.ValidatorPool.ValNFT++;
      expectedState.validators[index].Acc = true;
      expectedState.validators[index].Reg = true;
    }
    // Expect that all validators funds are transferred from PublicStaking to ValidatorStaking
    expectedState.PublicStaking.ATK -= stakeAmount * BigInt(validators.length);
    expectedState.PublicStaking.ETH = BigInt(0);
    expectedState.ValidatorStaking.ATK +=
      stakeAmount * BigInt(validators.length);
    // Register validators
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const currentState = await getCurrentState(fixture, validators);
    await showState("after registering", currentState);
    await showState("Expected state after registering", expectedState);
    await showState("Current state after registering", currentState);

    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Set and get the validator's location after registering", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const currentState = await getCurrentState(fixture, validators);
    await showState("after registering", currentState);
    await fixture.validatorPool
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .setLocation("1.1.1.1");
    expect(
      await fixture.validatorPool.getLocation(await validators[0])
    ).to.be.equal("1.1.1.1");
  });

  it("Should not allow non-validator to set an IP location", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const currentState = await getCurrentState(fixture, validators);
    await showState("after registering", currentState);
    // Original Validators has all the validators not only the maximum to register
    // Position 5 is not a register validator
    await expect(
      fixture.validatorPool.connect(adminSigner).setLocation("1.1.1.1")
    ).to.be.revertedWith("800");
  });

  it("Should not allow users to register reusing previous publicStaking that cannot be burned", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );
    await commitSnapshots(fixture, 4);
    stakingTokenIds = [];
    for (const validator of validatorsSnapshots) {
      const claimTx = (await fixture.validatorPool
        .connect(await getValidatorEthAccount(validator))
        .claimExitingNFTPosition()) as ContractTransaction;
      const receipt = await ethers.provider.getTransactionReceipt(claimTx.hash);
      // When a token is claimed it gets burned an minted so it gets a new tokenId so we need to update array for re-registration
      const tokenId = BigNumber.from(receipt.logs[0].topics[3]);
      stakingTokenIds.push(tokenId);
      // Transfer to factory for re-registering
      await fixture.publicStaking
        .connect(await getValidatorEthAccount(validator))
        .transferFrom(validator.address, fixture.factory.address, tokenId);
      // Approve new tokensIds
      await factoryCallAnyFixture(fixture, "publicStaking", "approve", [
        fixture.validatorPool.address,
        tokenId,
      ]);
    }
    await showState(
      "After claiming:",
      await getCurrentState(fixture, validators)
    );
    // After claiming, position is locked for a period of 172800 AliceNet epochs
    // To perform this test with no revert, POSITION_LOCK_PERIOD can be set to 3
    // in ValidatorPool and then take 4 snapshots
    // await commitSnapshots(fixture, 4)
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith("606");
  });
});
