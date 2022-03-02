import { expect, assert } from "../chai-setup";
import { ethers } from "hardhat";
import {
  BigNumber,
  BigNumberish,
  ContractTransaction,
} from "ethers";
import {
  ETHDKG,
  ValidatorPool,
  ValidatorPoolMock,
} from "../../typechain-types";
import { getFixture, getValidatorEthAccount, mineBlocks } from "../setup";
import { factoryCallAny } from "../setup";

export const PLACEHOLDER_ADDRESS = "0x0000000000000000000000000000000000000000";

export { expect, assert } from "../chai-setup";

export enum Phase {
  RegistrationOpen,
  ShareDistribution,
  DisputeShareDistribution,
  KeyShareSubmission,
  MPKSubmission,
  GPKJSubmission,
  DisputeGPKJSubmission,
  Completion,
}

export interface ValidatorRawData {
  privateKey?: string;
  address: string;
  madNetPublicKey: [BigNumberish, BigNumberish];
  encryptedShares: BigNumberish[];
  commitments: [BigNumberish, BigNumberish][];
  keyShareG1: [BigNumberish, BigNumberish];
  keyShareG1CorrectnessProof: [BigNumberish, BigNumberish];
  keyShareG2: [BigNumberish, BigNumberish, BigNumberish, BigNumberish];
  mpk: [BigNumberish, BigNumberish, BigNumberish, BigNumberish];
  gpkj: [BigNumberish, BigNumberish, BigNumberish, BigNumberish];
  sharedKey?: [BigNumberish, BigNumberish];
  sharedKeyProof?: [BigNumberish, BigNumberish];
  encryptedSharesHash?: BigNumberish[];
  groupCommitments?: [BigNumberish, BigNumberish][][];
}

// Event asserts

export const assertRegistrationComplete = async (tx: ContractTransaction) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event RegistrationComplete(uint256 blockNumber)",
  ]);
  let data = receipt.logs[1].data;
  let topics = receipt.logs[1].topics;
  let event = intrface.decodeEventLog("RegistrationComplete", data, topics);
  expect(event.blockNumber).to.equal(receipt.blockNumber);
};

export const assertEventSharesDistributed = async (
  tx: ContractTransaction,
  account: string,
  index: BigNumberish,
  nonce: BigNumberish,
  encryptedShares: BigNumberish[],
  commitments: [BigNumberish, BigNumberish][]
) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event SharesDistributed(address account, uint256 index, uint256 nonce, uint256[] encryptedShares, uint256[2][] commitments)",
  ]);
  let data = receipt.logs[0].data;
  let topics = receipt.logs[0].topics;
  let event = intrface.decodeEventLog("SharesDistributed", data, topics);
  expect(event.account).to.equal(account);
  expect(event.index).to.equal(index);
  expect(event.nonce).to.equal(nonce);
  assertEqEncryptedShares(event.encryptedShares, encryptedShares);
  assertEqCommitments(event.commitments, commitments);
};

export const assertEventShareDistributionComplete = async (
  tx: ContractTransaction
) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event ShareDistributionComplete(uint256 blockNumber)",
  ]);
  let data = receipt.logs[1].data;
  let topics = receipt.logs[1].topics;
  let event = intrface.decodeEventLog(
    "ShareDistributionComplete",
    data,
    topics
  );
  expect(event.blockNumber).to.equal(receipt.blockNumber);
};

export const assertEventKeyShareSubmitted = async (
  tx: ContractTransaction,
  account: string,
  index: BigNumberish,
  nonce: BigNumberish,
  keyShareG1: [BigNumberish, BigNumberish],
  keyShareG1CorrectnessProof: [BigNumberish, BigNumberish],
  keyShareG2: [BigNumberish, BigNumberish, BigNumberish, BigNumberish]
) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event KeyShareSubmitted(address account, uint256 index, uint256 nonce, uint256[2] keyShareG1, uint256[2] keyShareG1CorrectnessProof, uint256[4] keyShareG2)",
  ]);
  let data = receipt.logs[0].data;
  let topics = receipt.logs[0].topics;
  let event = intrface.decodeEventLog("KeyShareSubmitted", data, topics);
  expect(event.account).to.equal(account);
  expect(event.index).to.equal(index);
  expect(event.nonce).to.equal(nonce);
  assertEqKeyShareG1(event.keyShareG1, keyShareG1);
  assertEqKeyShareG1(
    event.keyShareG1CorrectnessProof,
    keyShareG1CorrectnessProof
  );
  assertEqKeyShareG2(event.keyShareG2, keyShareG2);
};

export const assertEventKeyShareSubmissionComplete = async (
  tx: ContractTransaction
) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event KeyShareSubmissionComplete(uint256 blockNumber)",
  ]);
  let data = receipt.logs[1].data;
  let topics = receipt.logs[1].topics;
  let event = intrface.decodeEventLog(
    "KeyShareSubmissionComplete",
    data,
    topics
  );
  expect(event.blockNumber).to.equal(receipt.blockNumber);
};

export const assertEventMPKSet = async (
  tx: ContractTransaction,
  nonce: BigNumberish,
  mpk: [BigNumberish, BigNumberish, BigNumberish, BigNumberish]
) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event MPKSet(uint256 blockNumber, uint256 nonce, uint256[4] mpk)",
  ]);
  let data = receipt.logs[0].data;
  let topics = receipt.logs[0].topics;
  let event = intrface.decodeEventLog("MPKSet", data, topics);
  expect(event.blockNumber).to.equal(receipt.blockNumber);
  expect(event.nonce).to.equal(nonce);
  expect(event.mpk[0]).to.equal(mpk[0]);
  expect(event.mpk[1]).to.equal(mpk[1]);
  expect(event.mpk[2]).to.equal(mpk[2]);
  expect(event.mpk[3]).to.equal(mpk[3]);
};

// submit GPKj phase
export const assertEventValidatorMemberAdded = async (
  tx: ContractTransaction,
  account: string,
  index: BigNumberish,
  nonce: BigNumberish,
  epoch: BigNumberish,
  gpkj: [BigNumberish, BigNumberish, BigNumberish, BigNumberish]
) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event ValidatorMemberAdded(address account, uint256 index, uint256 nonce, uint256 epoch, uint256 share0, uint256 share1, uint256 share2, uint256 share3)",
  ]);
  let data = receipt.logs[0].data;
  let topics = receipt.logs[0].topics;
  let event = intrface.decodeEventLog("ValidatorMemberAdded", data, topics);
  expect(event.account).to.equal(account);
  expect(event.index).to.equal(index);
  expect(event.nonce).to.equal(nonce);
  expect(event.epoch).to.equal(epoch);
  expect(event.share0).to.equal(gpkj[0]);
  expect(event.share1).to.equal(gpkj[1]);
  expect(event.share2).to.equal(gpkj[2]);
  expect(event.share3).to.equal(gpkj[3]);
};

export const assertEventGPKJSubmissionComplete = async (
  tx: ContractTransaction
) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event GPKJSubmissionComplete(uint256 blockNumber)",
  ]);
  let data = receipt.logs[1].data;
  let topics = receipt.logs[1].topics;
  let event = intrface.decodeEventLog("GPKJSubmissionComplete", data, topics);
  expect(event.blockNumber).to.equal(receipt.blockNumber);
};

// COMPLETE PHASE

export const assertEventValidatorSetCompleted = async (
  tx: ContractTransaction,
  validatorCount: BigNumberish,
  nonce: BigNumberish,
  epoch: BigNumberish,
  ethHeight: BigNumberish,
  madHeight: BigNumberish,
  mpk: [BigNumberish, BigNumberish, BigNumberish, BigNumberish]
) => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event ValidatorSetCompleted(uint256 validatorCount, uint256 nonce, uint256 epoch, uint256 ethHeight, uint256 madHeight, uint256 groupKey0, uint256 groupKey1, uint256 groupKey2, uint256 groupKey3)",
  ]);
  let data = receipt.logs[0].data;
  let topics = receipt.logs[0].topics;
  let event = intrface.decodeEventLog("ValidatorSetCompleted", data, topics);
  expect(event.validatorCount).to.equal(validatorCount);
  expect(event.nonce).to.equal(nonce);
  expect(event.epoch).to.equal(epoch);
  expect(event.ethHeight).to.equal(ethHeight);
  expect(event.madHeight).to.equal(madHeight);
  expect(event.groupKey0).to.equal(mpk[0]);
  expect(event.groupKey1).to.equal(mpk[1]);
  expect(event.groupKey2).to.equal(mpk[2]);
  expect(event.groupKey3).to.equal(mpk[3]);
};

// Special asserts
export const assertEqEncryptedShares = (
  actualEncryptedShares: BigNumberish[],
  expectedEncryptedShares: BigNumberish[]
) => {
  actualEncryptedShares.forEach((element, i) => {
    assert(BigNumber.from(element).eq(expectedEncryptedShares[i]), "Incorrect encrypted shares!");
  });
};

export const assertEqCommitments = (
  actualCommitments: [BigNumberish, BigNumberish][],
  expectedCommitments: [BigNumberish, BigNumberish][]
) => {
  actualCommitments.forEach((element, i) => {
    assert(BigNumber.from(element[0]).eq(expectedCommitments[i][0]), "Incorrect commitments[0]");
    assert(BigNumber.from(element[1]).eq(expectedCommitments[i][1]), "Incorrect commitments[1]");
  });
};

export const assertEqKeyShareG1 = (
  actualKeySharesG1: [BigNumberish, BigNumberish],
  expectedKeySharesG1: [BigNumberish, BigNumberish]
) => {
  assert(BigNumber.from(actualKeySharesG1[0]).eq(expectedKeySharesG1[0]), "Incorrect encryptedKeys[0]");
  assert(BigNumber.from(actualKeySharesG1[1]).eq(expectedKeySharesG1[1]), "Incorrect encryptedKeys[1]");
};

export const assertEqKeyShareG2 = (
  actualKeySharesG2: [BigNumberish, BigNumberish, BigNumberish, BigNumberish],
  expectedKeySharesG2: [BigNumberish, BigNumberish, BigNumberish, BigNumberish]
) => {
  assert(BigNumber.from(actualKeySharesG2[0]).eq(expectedKeySharesG2[0]), "Incorrect actualKeySharesG2[0]");
  assert(BigNumber.from(actualKeySharesG2[1]).eq(expectedKeySharesG2[1]), "Incorrect actualKeySharesG2[1]");
  assert(BigNumber.from(actualKeySharesG2[2]).eq(expectedKeySharesG2[2]), "Incorrect actualKeySharesG2[2]");
  assert(BigNumber.from(actualKeySharesG2[3]).eq(expectedKeySharesG2[3]), "Incorrect actualKeySharesG2[3]");
};

export const assertETHDKGPhase = async (
  ethdkg: ETHDKG,
  expectedPhase: Phase
) => {
  let actualPhase = await ethdkg.getETHDKGPhase();
  assert(actualPhase === expectedPhase, "Incorrect Phase");
};

// Aux functions
export const addValidators = async (
  validatorPool: ValidatorPoolMock | ValidatorPool,
  validators: ValidatorRawData[]
) => {
  if ((<ValidatorPoolMock>validatorPool).isMock) {
    for (let validator of validators) {
      expect(await validatorPool.isValidator(validator.address)).to.equal(
        false
      );
      await validatorPool.registerValidators([validator.address], [0]);
      expect(await validatorPool.isValidator(validator.address)).to.equal(true);
    }
  }
};

export const endCurrentPhase = async (ethdkg: ETHDKG) => {
  // advance enough blocks to timeout a phase
  let phaseStart = await ethdkg.getPhaseStartBlock();
  let phaseLength = await ethdkg.getPhaseLength();
  let bn = await ethers.provider.getBlockNumber();
  let endBlock = phaseStart.add(phaseLength);
  let blocksToMine = endBlock.sub(bn);
  await mineBlocks(blocksToMine.toNumber());
};

export const endCurrentAccusationPhase = async (ethdkg: ETHDKG) => {
  // advance enough blocks to timeout a phase
  let phaseStart = await ethdkg.getPhaseStartBlock();
  let phaseLength = await ethdkg.getPhaseLength();
  let bn = await ethers.provider.getBlockNumber();
  let endBlock = phaseStart.add(phaseLength.mul(2));
  let blocksToMine = endBlock.sub(bn);
  await mineBlocks(blocksToMine.toNumber());
};

export const waitNextPhaseStartDelay = async (ethdkg: ETHDKG) => {
  // advance enough blocks to timeout a phase
  let phaseStart = await ethdkg.getPhaseStartBlock();
  let bn = await ethers.provider.getBlockNumber();
  let blocksToMine = phaseStart.sub(bn).add(1);
  await mineBlocks(blocksToMine.toNumber());
};

export const initializeETHDKG = async (
  ethdkg: ETHDKG,
  validatorPool: ValidatorPoolMock | ValidatorPool
) => {
  let nonce = await ethdkg.getNonce();
  await expect(
    validatorPool.initializeETHDKG()
  ).to.emit(
    ethdkg,
    "RegistrationOpened"
  );
  expect(await ethdkg.getNonce()).to.eq(nonce.add(1));
  await assertETHDKGPhase(ethdkg, Phase.RegistrationOpen);
};

export const registerValidators = async (
  ethdkg: ETHDKG,
  validatorPool: ValidatorPoolMock | ValidatorPool,
  validators: ValidatorRawData[],
  expectedNonce: number
) => {
  //validators = shuffle(validators);
  for (let validator of validators) {
    let numParticipantsBefore = await ethdkg.getNumParticipants();
    let tx = ethdkg
      .connect(await getValidatorEthAccount(validator))
      .register(validator.madNetPublicKey)
    let receipt = await tx;
    let participant = await ethdkg.getParticipantInternalState(
      validator.address
    );
    expect(tx)
      .to.emit(ethdkg, "AddressRegistered")
      .withArgs(
        ethers.utils.getAddress(validator.address),
        participant.index,
        expectedNonce,
        validator.madNetPublicKey
      );
    let numValidators = await validatorPool.getValidatorsCount();
    let numParticipants = await ethdkg.getNumParticipants();
    // if all validators in the Pool participated in this round
    if (numParticipantsBefore.add(1).eq(numValidators)) {
      await assertRegistrationComplete(receipt);
      expect(await ethdkg.getNumParticipants()).to.eq(0);
      await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);
    } else {
      expect(numParticipants).to.eq(numParticipantsBefore.add(1));
    }
  }
};

export const distributeValidatorsShares = async (
  ethdkg: ETHDKG,
  validatorPool: ValidatorPoolMock | ValidatorPool,
  validators: ValidatorRawData[],
  expectedNonce: number
) => {
  //validators = shuffle(validators);
  for (let validator of validators) {
    let numParticipantsBefore = await ethdkg.getNumParticipants();
    let tx = await ethdkg
      .connect(await getValidatorEthAccount(validator))
      .distributeShares(validator.encryptedShares, validator.commitments);
    let participant = await ethdkg.getParticipantInternalState(
      validator.address
    );
    await assertEventSharesDistributed(
      tx,
      ethers.utils.getAddress(validator.address),
      participant.index,
      expectedNonce,
      validator.encryptedShares,
      validator.commitments
    );
    // if all validators in the Pool participated in this round
    let numValidators = await validatorPool.getValidatorsCount();
    let numParticipants = await ethdkg.getNumParticipants();
    if (numParticipantsBefore.add(1).eq(numValidators)) {
      await assertEventShareDistributionComplete(tx);
      expect(await ethdkg.getNumParticipants()).to.eq(0);
      await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
    } else {
      expect(numParticipants).to.eq(numParticipantsBefore.add(1));
    }
  }
};

export const submitValidatorsKeyShares = async (
  ethdkg: ETHDKG,
  validatorPool: ValidatorPoolMock | ValidatorPool,
  validators: ValidatorRawData[],
  expectedNonce: number
) => {
  //validators = shuffle(validators);
  for (let validator of validators) {
    let numParticipantsBefore = await ethdkg.getNumParticipants();
    let tx = await ethdkg
      .connect(await getValidatorEthAccount(validator))
      .submitKeyShare(
        validator.keyShareG1,
        validator.keyShareG1CorrectnessProof,
        validator.keyShareG2
      );
    let participant = await ethdkg.getParticipantInternalState(
      validator.address
    );
    await assertEventKeyShareSubmitted(
      tx,
      ethers.utils.getAddress(validator.address),
      participant.index,
      expectedNonce,
      validator.keyShareG1,
      validator.keyShareG1CorrectnessProof,
      validator.keyShareG2
    );
    //
    // if all validators in the Pool participated in this round
    let numValidators = await validatorPool.getValidatorsCount();
    if (numValidators.eq(1)) {
      await assertETHDKGPhase(ethdkg, Phase.KeyShareSubmission);
    }
    let numParticipants = await ethdkg.getNumParticipants();
    if (numParticipantsBefore.add(1).eq(numValidators)) {
      await assertEventKeyShareSubmissionComplete(tx);
      expect(await ethdkg.getNumParticipants()).to.eq(0);
      await assertETHDKGPhase(ethdkg, Phase.MPKSubmission);
    } else {
      expect(numParticipants).to.eq(numParticipantsBefore.add(1));
    }
  }
};

export const submitMasterPublicKey = async (
  ethdkg: ETHDKG,
  validators: ValidatorRawData[],
  expectedNonce: number
) => {
  // choose an random validator from the list to send the mpk
  var index = Math.floor(Math.random() * validators.length);
  let tx = await ethdkg
    .connect(await getValidatorEthAccount(validators[index]))
    .submitMasterPublicKey(validators[index].mpk);
  await assertEventMPKSet(tx, expectedNonce, validators[index].mpk);
  expect(await ethdkg.getNumParticipants()).to.eq(0);
  await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  // The other validators should fail
  for (let validator of validators) {
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validator))
        .submitMasterPublicKey(validator.mpk)
    ).to.be.revertedWith(
      "ETHDKG: cannot participate on master public key submission phase"
    );
  }
};

export const submitValidatorsGPKJ = async (
  ethdkg: ETHDKG,
  validatorPool: ValidatorPoolMock | ValidatorPool,
  validators: ValidatorRawData[],
  expectedNonce: number,
  expectedEpoch: number
) => {
  //validators = shuffle(validators);
  for (let validator of validators) {
    let numParticipantsBefore = await ethdkg.getNumParticipants();
    let tx = await ethdkg
      .connect(await getValidatorEthAccount(validator))
      .submitGPKJ(validator.gpkj);
    let participant = await ethdkg.getParticipantInternalState(
      validator.address
    );
    await assertEventValidatorMemberAdded(
      tx,
      ethers.utils.getAddress(validator.address),
      participant.index,
      expectedNonce,
      expectedEpoch,
      validator.gpkj
    );
    // if all validators in the Pool participated in this round
    let numValidators = await validatorPool.getValidatorsCount();
    let numParticipants = await ethdkg.getNumParticipants();
    if (numParticipantsBefore.add(1).eq(numValidators)) {
      await assertEventGPKJSubmissionComplete(tx);
      expect(await ethdkg.getNumParticipants()).to.eq(0);
      await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    } else {
      expect(numParticipants).to.eq(numParticipantsBefore.add(1));
    }
  }
};

export const completeETHDKG = async (
  ethdkg: ETHDKG,
  validators: ValidatorRawData[],
  expectedNonce: number,
  expectedEpoch: number,
  expectedMadHeight: number
) => {
  // choose an random validator from the list to send the mpk
  var index = Math.floor(Math.random() * validators.length);
  let tx = await ethdkg
    .connect(await getValidatorEthAccount(validators[index]))
    .complete();
  await assertEventValidatorSetCompleted(
    tx,
    validators.length,
    expectedNonce,
    expectedEpoch,
    0,
    expectedMadHeight,
    validators[index].mpk
  );
  await assertETHDKGPhase(ethdkg, Phase.Completion);
  // The other validators should fail
  for (let validator of validators) {
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validator))
        .submitMasterPublicKey(validator.mpk)
    ).to.be.revertedWith(
      "ETHDKG: cannot participate on master public key submission phase"
    );
  }
};

export const startAtDistributeShares = async (
  validators: ValidatorRawData[],
  contracts?: {
    ethdkg: ETHDKG;
    validatorPool: ValidatorPoolMock | ValidatorPool;
  }
): Promise<[ETHDKG, ValidatorPoolMock | ValidatorPool, number]> => {
  const { ethdkg, validatorPool } =
    typeof contracts !== "undefined" ? contracts : await getFixture(true);
  // add validators
  if ((<ValidatorPoolMock>validatorPool).isMock) {
    await addValidators(validatorPool, validators);
    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);
  }
  const expectedNonce = (await ethdkg.getNonce()).toNumber();
  // register all validators
  await registerValidators(ethdkg, validatorPool, validators, expectedNonce);
  await waitNextPhaseStartDelay(ethdkg);
  return [ethdkg, validatorPool, expectedNonce];
};

export const startAtSubmitKeyShares = async (
  validators: ValidatorRawData[],
  contracts?: {
    ethdkg: ETHDKG;
    validatorPool: ValidatorPoolMock | ValidatorPool;
  }
): Promise<[ETHDKG, ValidatorPoolMock | ValidatorPool, number]> => {
  let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
    validators,
    contracts
  );
  // distribute shares for all validators
  await distributeValidatorsShares(
    ethdkg,
    validatorPool,
    validators,
    expectedNonce
  );

  // skipping the distribute shares accusation phase
  await endCurrentPhase(ethdkg);
  await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
  return [ethdkg, validatorPool, expectedNonce];
};

export const startAtMPKSubmission = async (
  validators: ValidatorRawData[],
  contracts?: {
    ethdkg: ETHDKG;
    validatorPool: ValidatorPoolMock | ValidatorPool;
  }
): Promise<[ETHDKG, ValidatorPoolMock | ValidatorPool, number]> => {
  let [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
    validators,
    contracts
  );
  // distribute shares for all validators
  await submitValidatorsKeyShares(
    ethdkg,
    validatorPool,
    validators,
    expectedNonce
  );

  await waitNextPhaseStartDelay(ethdkg);
  await assertETHDKGPhase(ethdkg, Phase.MPKSubmission);
  return [ethdkg, validatorPool, expectedNonce];
};

export const startAtGPKJ = async (
  validators: ValidatorRawData[],
  contracts?: {
    ethdkg: ETHDKG;
    validatorPool: ValidatorPoolMock | ValidatorPool;
  }
): Promise<[ETHDKG, ValidatorPoolMock | ValidatorPool, number]> => {
  let [ethdkg, validatorPool, expectedNonce] = await startAtMPKSubmission(
    validators,
    contracts
  );
  // Submit the Master Public key
  await submitMasterPublicKey(ethdkg, validators, expectedNonce);
  await waitNextPhaseStartDelay(ethdkg);

  return [ethdkg, validatorPool, expectedNonce];
};

export const completeETHDKGRound = async (
  validators: ValidatorRawData[],
  contracts?: {
    ethdkg: ETHDKG;
    validatorPool: ValidatorPoolMock | ValidatorPool;
  },
  expectedEpoch?: number
): Promise<
  [ETHDKG, ValidatorPoolMock | ValidatorPool, number, number, number]
> => {
  let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
    validators,
    contracts
  );
  let _expectedEpoch = 0;
  if (typeof(expectedEpoch) !== "undefined") {
    _expectedEpoch = expectedEpoch
  }
  const expectedMadHeight = 0;
  // Submit GPKj for all validators
  await submitValidatorsGPKJ(
    ethdkg,
    validatorPool,
    validators,
    expectedNonce,
    _expectedEpoch
  );

  // skipping the distribute shares accusation phase
  await endCurrentPhase(ethdkg);
  await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

  // Complete ETHDKG
  await completeETHDKG(
    ethdkg,
    validators,
    expectedNonce,
    _expectedEpoch,
    expectedMadHeight
  );
  return [
    ethdkg,
    validatorPool,
    expectedNonce,
    _expectedEpoch,
    expectedMadHeight,
  ];
};

