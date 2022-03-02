import { validators4 } from "../assets/4-validators-successful-case";
import { validators4BadDistributeShares } from "../assets/4-validators-bad-distribute-shares";
import { validators4BadDistributeSharesLast } from "../assets/4-validators-1-bad-distribute-shares-last";
import { ethers } from "hardhat";
import {
  endCurrentPhase,
  endCurrentAccusationPhase,
  distributeValidatorsShares,
  startAtDistributeShares,
  expect,
  assertETHDKGPhase,
  Phase,
  PLACEHOLDER_ADDRESS,
  submitValidatorsKeyShares,
  submitMasterPublicKey,
  submitValidatorsGPKJ,
  completeETHDKG,
} from "../setup";
import { BigNumberish } from "ethers";
import { getValidatorEthAccount, mineBlocks } from "../../setup";

describe("ETHDKG: Dispute bad shares", () => {

  it("should not allow accusations before time", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // try accusing bad shares
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[0])).accuseParticipantDistributedBadShares(PLACEHOLDER_ADDRESS, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")
  });

  it("should not allow accusations unless in DisputeShareDistribution phase, or expired ShareDistribution phase", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // try accusing bad shares
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[0])).accuseParticipantDistributedBadShares(PLACEHOLDER_ADDRESS, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")

    // distribute shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce
    );
    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);

    // skipping the distribute shares accusation phase
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);

    // submit key shares phase
    await submitValidatorsKeyShares(ethdkg, validatorPool, validators4, expectedNonce)

    // try accusing bad shares
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[0])).accuseParticipantDistributedBadShares(PLACEHOLDER_ADDRESS, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")

    //await endCurrentPhase(ethdkg)
    await assertETHDKGPhase(ethdkg, Phase.MPKSubmission);

    // try accusing bad shares
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[0])).accuseParticipantDistributedBadShares(PLACEHOLDER_ADDRESS, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")

    // submit MPK
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())
    await submitMasterPublicKey(ethdkg, validators4, expectedNonce)

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // try accusing bad shares
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[0])).accuseParticipantDistributedBadShares(PLACEHOLDER_ADDRESS, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")

    // submit GPKj
    await submitValidatorsGPKJ(ethdkg, validatorPool, validators4, expectedNonce, 0)

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission)

    // try accusing bad shares
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[0])).accuseParticipantDistributedBadShares(PLACEHOLDER_ADDRESS, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")

    await endCurrentPhase(ethdkg)

    // try accusing bad shares
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[0])).accuseParticipantDistributedBadShares(PLACEHOLDER_ADDRESS, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")

    // complete ethdkg
    await completeETHDKG(ethdkg, validators4, expectedNonce, 0, 0)

    await assertETHDKGPhase(ethdkg, Phase.Completion)

    // try accusing bad shares
    await expect(ethdkg.connect(await ethers.getSigner(validators4[0].address)).accuseParticipantDistributedBadShares(PLACEHOLDER_ADDRESS, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")
  });


  it("should not allow accusation of a non-participating validator", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // 3/4 validators will distribute shares, 4th validator will not
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 3),
      expectedNonce
    );

    await endCurrentPhase(ethdkg)

    // try accusing the 4th validator of bad shares, when it did not even distribute them
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[0])).accuseParticipantDistributedBadShares(validators4[3].address, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! dishonestParticipant did not distribute shares!")
  });

  it("should not allow accusation from a non-participating validator", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // 3/4 validators will distribute shares, 4th validator will not
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 3),
      expectedNonce
    );

    await endCurrentPhase(ethdkg)

    // validator 4 will try accusing the 1st validator of bad shares, when it did not even distribute them itself
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[3])).accuseParticipantDistributedBadShares(validators4[0].address, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Disputer did not distribute shares!")
  });

  it("should not allow accusation with an incorrect index", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    //
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution)
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())
    //await endCurrentPhase(ethdkg)
  });

  it("should not allow accusation with incorrect encrypted shares or commitments", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution)
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())

    // try accusing the 1st validator of bad shares using incorrect encrypted shares and commitments
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4[3])).accuseParticipantDistributedBadShares(validators4[0].address, [], [[0,0]], [0,0], [0,0]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Submitted commitments and encrypted shares don't match!")
  });

  it("should not allow double accusation of a validator on two separate calls", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4BadDistributeShares
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4BadDistributeShares,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution)
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())

    // try accusing the 1st validator of bad shares using valid encrypted shares and commitments
    await ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeShares[3]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeShares[0].address,
      validators4BadDistributeShares[0].encryptedShares,
      validators4BadDistributeShares[0].commitments,
      validators4BadDistributeShares[3].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeShares[3].sharedKeyProof as [BigNumberish, BigNumberish])

    expect(await validatorPool.isValidator(validators4BadDistributeShares[3].address))
    .to.equal(true)
    expect(await validatorPool.isValidator(validators4BadDistributeShares[0].address))
    .to.equal(false)

    // try accusing the 1st validator again of bad shares using valid encrypted shares and commitments
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeShares[3]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeShares[0].address,
      validators4BadDistributeShares[0].encryptedShares,
      validators4BadDistributeShares[0].commitments,
      validators4BadDistributeShares[3].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeShares[3].sharedKeyProof as [BigNumberish, BigNumberish]))
    .to.be.revertedWith("ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!")

  });

  it("should not allow double accusation of a validator on two separate calls, with a different index order", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4BadDistributeSharesLast
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4BadDistributeSharesLast,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution)
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())

    // try accusing the 4th validator of bad shares using valid encrypted shares and commitments
    await ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeSharesLast[0]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeSharesLast[3].address,
      validators4BadDistributeSharesLast[3].encryptedShares,
      validators4BadDistributeSharesLast[3].commitments,
      validators4BadDistributeSharesLast[0].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeSharesLast[0].sharedKeyProof as [BigNumberish, BigNumberish])

    expect(await validatorPool.isValidator(validators4BadDistributeSharesLast[0].address))
    .to.equal(true)
    expect(await validatorPool.isValidator(validators4BadDistributeSharesLast[3].address))
    .to.equal(false)

    // try accusing the 4th validator again of bad shares using valid encrypted shares and commitments
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeSharesLast[0]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeSharesLast[3].address,
      validators4BadDistributeSharesLast[3].encryptedShares,
      validators4BadDistributeSharesLast[3].commitments,
      validators4BadDistributeSharesLast[0].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeSharesLast[0].sharedKeyProof as [BigNumberish, BigNumberish]))
    .to.be.revertedWith("ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!")

  });

  it("should not allow proceeding to next phase (KeyShareSubmission) after accusations take place", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4BadDistributeShares
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4BadDistributeShares,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution)
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())

    // try accusing the 1st validator of bad shares using valid encrypted shares and commitments
    await ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeShares[3]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeShares[0].address,
      validators4BadDistributeShares[0].encryptedShares,
      validators4BadDistributeShares[0].commitments,
      validators4BadDistributeShares[3].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeShares[3].sharedKeyProof as [BigNumberish, BigNumberish])

    expect(await validatorPool.isValidator(validators4BadDistributeShares[3].address))
    .to.equal(true)
    expect(await validatorPool.isValidator(validators4BadDistributeShares[0].address))
    .to.equal(false)

    // try moving into the next phase - KeyShareSubmission
    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution)
    await endCurrentPhase(ethdkg)
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())
    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution)

    await expect(ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeShares[3]))
    .submitKeyShare(
      validators4BadDistributeShares[3].keyShareG1,
      validators4BadDistributeShares[3].keyShareG1CorrectnessProof,
      validators4BadDistributeShares[3].keyShareG2))
    .to.be.revertedWith("ETHDKG: cannot participate on key share submission phase")

  });

  it("should not allow proceeding to next phase (KeyShareSubmission) after bad shares accusations take place, along with accusations of non-participant validators", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4BadDistributeShares
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // the 4th validator won't distribute shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4BadDistributeShares.slice(0,3),
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution)
    await endCurrentPhase(ethdkg)
    //await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())
    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution)

    // accuse the 1st validator of bad shares using valid encrypted shares and commitments
    await ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeShares[2]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeShares[0].address,
      validators4BadDistributeShares[0].encryptedShares,
      validators4BadDistributeShares[0].commitments,
      validators4BadDistributeShares[2].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeShares[2].sharedKeyProof as [BigNumberish, BigNumberish])

    expect(await validatorPool.isValidator(validators4BadDistributeShares[2].address))
    .to.equal(true)
    expect(await validatorPool.isValidator(validators4BadDistributeShares[0].address))
    .to.equal(false)

    // also accuse the 4th validator of not distributing shares
    await ethdkg.accuseParticipantDidNotDistributeShares([validators4BadDistributeShares[3].address])
    expect(await validatorPool.isValidator(validators4BadDistributeShares[3].address))
    .to.equal(false)

    // try accusing the 1st validator again of bad shares using valid encrypted shares and commitments
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeSharesLast[2]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeSharesLast[0].address,
      validators4BadDistributeSharesLast[0].encryptedShares,
      validators4BadDistributeSharesLast[0].commitments,
      validators4BadDistributeSharesLast[2].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeSharesLast[2].sharedKeyProof as [BigNumberish, BigNumberish]))
    .to.be.revertedWith("ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!")

    // try moving into the next phase - KeyShareSubmission
    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution)
    await endCurrentPhase(ethdkg)
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())
    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution)

    await expect(ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeShares[2]))
    .submitKeyShare(
      validators4BadDistributeShares[2].keyShareG1,
      validators4BadDistributeShares[2].keyShareG1CorrectnessProof,
      validators4BadDistributeShares[2].keyShareG2))
    .to.be.revertedWith("ETHDKG: cannot participate on key share submission phase")

    // try accusing the 1st validator again of bad shares using valid encrypted shares and commitments
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeSharesLast[2]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeSharesLast[0].address,
      validators4BadDistributeSharesLast[0].encryptedShares,
      validators4BadDistributeSharesLast[0].commitments,
      validators4BadDistributeSharesLast[2].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeSharesLast[2].sharedKeyProof as [BigNumberish, BigNumberish]))
    .to.be.revertedWith("ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!")

    await endCurrentAccusationPhase(ethdkg)
    await mineBlocks((await ethdkg.getConfirmationLength()).toNumber())
    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution)

    await expect(ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeShares[2]))
    .submitKeyShare(
      validators4BadDistributeShares[2].keyShareG1,
      validators4BadDistributeShares[2].keyShareG1CorrectnessProof,
      validators4BadDistributeShares[2].keyShareG2))
    .to.be.revertedWith("ETHDKG: cannot participate on key share submission phase")

    // try accusing the 1st validator again of bad shares using valid encrypted shares and commitments
    await expect(ethdkg.connect(await getValidatorEthAccount(validators4BadDistributeSharesLast[2]))
    .accuseParticipantDistributedBadShares(
      validators4BadDistributeSharesLast[0].address,
      validators4BadDistributeSharesLast[0].encryptedShares,
      validators4BadDistributeSharesLast[0].commitments,
      validators4BadDistributeSharesLast[2].sharedKey as [BigNumberish, BigNumberish],
      validators4BadDistributeSharesLast[2].sharedKeyProof as [BigNumberish, BigNumberish]))
    .to.be.revertedWith("ETHDKG: Dispute failed! Contract is not in dispute phase!")

  });

});
