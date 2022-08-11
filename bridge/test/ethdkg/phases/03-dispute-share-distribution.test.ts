import { BigNumberish } from "ethers";
import { ethers } from "hardhat";
import {
  getReceiptForFailedTransaction,
  getValidatorEthAccount,
  mineBlocks,
} from "../../setup";
import { validators4BadDistributeSharesLast } from "../assets/4-validators-1-bad-distribute-shares-last";
import { validators4BadDistributeShares } from "../assets/4-validators-bad-distribute-shares";
import { validators4 } from "../assets/4-validators-successful-case";
import {
  assertETHDKGPhase,
  completeETHDKG,
  distributeValidatorsShares,
  endCurrentAccusationPhase,
  endCurrentPhase,
  expect,
  Phase,
  PLACEHOLDER_ADDRESS,
  startAtDistributeShares,
  submitMasterPublicKey,
  submitValidatorsGPKJ,
  submitValidatorsKeyShares,
} from "../setup";

describe("ETHDKG: Dispute bad shares", () => {
  it("should not allow accusations before time", async function () {
    const [ethdkg, ,] = await startAtDistributeShares(validators4);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);
    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // try accusing bad shares
    const txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0]))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [[0, 0]],
        [0, 0],
        [0, 0]
      );
    const expectedBlockNumber = (
      await getReceiptForFailedTransaction(txPromise)
    ).blockNumber;
    const expectedCurrentPhase = await ethdkg.getETHDKGPhase();
    const phaseStartBlock = await ethdkg.getPhaseStartBlock();
    const phaseLength = await ethdkg.getPhaseLength();
    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
  });

  it("should not allow accusations unless in DisputeShareDistribution phase, or expired ShareDistribution phase", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);
    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // try accusing bad shares
    let txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0]))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [[0, 0]],
        [0, 0],
        [0, 0]
      );
    let expectedBlockNumber = (await getReceiptForFailedTransaction(txPromise))
      .blockNumber;
    let expectedCurrentPhase = await ethdkg.getETHDKGPhase();
    let phaseStartBlock = await ethdkg.getPhaseStartBlock();
    const phaseLength = await ethdkg.getPhaseLength();
    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

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
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce
    );

    // try accusing bad shares
    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0]))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [[0, 0]],
        [0, 0],
        [0, 0]
      );
    phaseStartBlock = await ethdkg.getPhaseStartBlock();
    expectedBlockNumber = (await getReceiptForFailedTransaction(txPromise))
      .blockNumber;
    expectedCurrentPhase = await ethdkg.getETHDKGPhase();

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    // await endCurrentPhase(ethdkg)
    await assertETHDKGPhase(ethdkg, Phase.MPKSubmission);

    // try accusing bad shares
    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0]))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [[0, 0]],
        [0, 0],
        [0, 0]
      );
    phaseStartBlock = await ethdkg.getPhaseStartBlock();
    expectedBlockNumber = (await getReceiptForFailedTransaction(txPromise))
      .blockNumber;
    expectedCurrentPhase = await ethdkg.getETHDKGPhase();

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    // submit MPK
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());
    await submitMasterPublicKey(ethdkg, validators4, expectedNonce);

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // try accusing bad shares
    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0]))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [[0, 0]],
        [0, 0],
        [0, 0]
      );
    phaseStartBlock = await ethdkg.getPhaseStartBlock();
    expectedBlockNumber = (await getReceiptForFailedTransaction(txPromise))
      .blockNumber;
    expectedCurrentPhase = await ethdkg.getETHDKGPhase();

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    // submit GPKj
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    // try accusing bad shares
    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0]))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [[0, 0]],
        [0, 0],
        [0, 0]
      );
    phaseStartBlock = await ethdkg.getPhaseStartBlock();
    expectedBlockNumber = (await getReceiptForFailedTransaction(txPromise))
      .blockNumber;
    expectedCurrentPhase = await ethdkg.getETHDKGPhase();

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    await endCurrentPhase(ethdkg);

    // try accusing bad shares
    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0]))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [[0, 0]],
        [0, 0],
        [0, 0]
      );
    phaseStartBlock = await ethdkg.getPhaseStartBlock();
    expectedBlockNumber = (await getReceiptForFailedTransaction(txPromise))
      .blockNumber;
    expectedCurrentPhase = await ethdkg.getETHDKGPhase();

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    // complete ethdkg
    await completeETHDKG(ethdkg, validators4, expectedNonce, 0, 0);

    await assertETHDKGPhase(ethdkg, Phase.Completion);

    // try accusing bad shares
    txPromise = ethdkg
      .connect(await ethers.getSigner(validators4[0].address))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [[0, 0]],
        [0, 0],
        [0, 0]
      );
    phaseStartBlock = await ethdkg.getPhaseStartBlock();
    expectedBlockNumber = (await getReceiptForFailedTransaction(txPromise))
      .blockNumber;
    expectedCurrentPhase = await ethdkg.getETHDKGPhase();

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
  });

  it("should not allow accusation of a non-participating validator", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // 3/4 validators will distribute shares, 4th validator will not
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 3),
      expectedNonce
    );

    await endCurrentPhase(ethdkg);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // try accusing the 4th validator of bad shares, when it did not even distribute them
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0]))
        .accuseParticipantDistributedBadShares(
          validators4[3].address,
          [],
          [[0, 0]],
          [0, 0],
          [0, 0]
        )
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `AccusedDidNotDistributeSharesInRound`
      )
      .withArgs(validators4[3].address);
  });

  it("should not allow accusation from a non-participating validator", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // 3/4 validators will distribute shares, 4th validator will not
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 3),
      expectedNonce
    );

    await endCurrentPhase(ethdkg);
    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // validator 4 will try accusing the 1st validator of bad shares, when it did not even distribute them itself
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[3]))
        .accuseParticipantDistributedBadShares(
          validators4[0].address,
          [],
          [[0, 0]],
          [0, 0],
          [0, 0]
        )
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `DisputerDidNotDistributeSharesInRound`
      )
      .withArgs(ethers.utils.getAddress(validators4[3].address));
  });

  it("should not allow accusation with an incorrect index", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    //
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());
    // await endCurrentPhase(ethdkg)
  });

  it("should not allow accusation with incorrect encrypted shares or commitments", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // try accusing the 1st validator of bad shares using incorrect encrypted shares and commitments
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[3]))
        .accuseParticipantDistributedBadShares(
          validators4[0].address,
          [],
          [[0, 0]],
          [0, 0],
          [0, 0]
        )
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `SharesAndCommitmentsMismatch`
      )
      .withArgs(
        "0x054a84b56c856021fcc61c75be9db7be0847b2b851356b213666caf7799bacc2",
        "0x3335c5eb3a24a7ecf92ed63d6ff8617a657471874b65a3c842544037a7ff8c0f"
      );
  });

  it("should not allow double accusation of a validator on two separate calls", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4BadDistributeShares);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4BadDistributeShares,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());

    // try accusing the 1st validator of bad shares using valid encrypted shares and commitments
    await ethdkg
      .connect(await getValidatorEthAccount(validators4BadDistributeShares[3]))
      .accuseParticipantDistributedBadShares(
        validators4BadDistributeShares[0].address,
        validators4BadDistributeShares[0].encryptedShares,
        validators4BadDistributeShares[0].commitments,
        validators4BadDistributeShares[3].sharedKey as [
          BigNumberish,
          BigNumberish
        ],
        validators4BadDistributeShares[3].sharedKeyProof as [
          BigNumberish,
          BigNumberish
        ]
      );

    expect(
      await validatorPool.isValidator(validators4BadDistributeShares[3].address)
    ).to.equal(true);
    expect(
      await validatorPool.isValidator(validators4BadDistributeShares[0].address)
    ).to.equal(false);
    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // try accusing the 1st validator again of bad shares using valid encrypted shares and commitments
    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount(validators4BadDistributeShares[3])
        )
        .accuseParticipantDistributedBadShares(
          validators4BadDistributeShares[0].address,
          validators4BadDistributeShares[0].encryptedShares,
          validators4BadDistributeShares[0].commitments,
          validators4BadDistributeShares[3].sharedKey as [
            BigNumberish,
            BigNumberish
          ],
          validators4BadDistributeShares[3].sharedKeyProof as [
            BigNumberish,
            BigNumberish
          ]
        )
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(
        ethers.utils.getAddress(validators4BadDistributeShares[0].address)
      );
  });

  it("should not allow double accusation of a validator on two separate calls, with a different index order", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4BadDistributeSharesLast);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4BadDistributeSharesLast,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());

    // try accusing the 4th validator of bad shares using valid encrypted shares and commitments
    await ethdkg
      .connect(
        await getValidatorEthAccount(validators4BadDistributeSharesLast[0])
      )
      .accuseParticipantDistributedBadShares(
        validators4BadDistributeSharesLast[3].address,
        validators4BadDistributeSharesLast[3].encryptedShares,
        validators4BadDistributeSharesLast[3].commitments,
        validators4BadDistributeSharesLast[0].sharedKey as [
          BigNumberish,
          BigNumberish
        ],
        validators4BadDistributeSharesLast[0].sharedKeyProof as [
          BigNumberish,
          BigNumberish
        ]
      );

    expect(
      await validatorPool.isValidator(
        validators4BadDistributeSharesLast[0].address
      )
    ).to.equal(true);
    expect(
      await validatorPool.isValidator(
        validators4BadDistributeSharesLast[3].address
      )
    ).to.equal(false);
    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // try accusing the 4th validator again of bad shares using valid encrypted shares and commitments
    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount(validators4BadDistributeSharesLast[0])
        )
        .accuseParticipantDistributedBadShares(
          validators4BadDistributeSharesLast[3].address,
          validators4BadDistributeSharesLast[3].encryptedShares,
          validators4BadDistributeSharesLast[3].commitments,
          validators4BadDistributeSharesLast[0].sharedKey as [
            BigNumberish,
            BigNumberish
          ],
          validators4BadDistributeSharesLast[0].sharedKeyProof as [
            BigNumberish,
            BigNumberish
          ]
        ),
      `AccusedNotValidator("${ethers.utils.getAddress(
        validators4BadDistributeSharesLast[3].address
      )}")`
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(
        ethers.utils.getAddress(validators4BadDistributeSharesLast[3].address)
      );
  });

  it("should not allow proceeding to next phase (KeyShareSubmission) after accusations take place", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4BadDistributeShares);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4BadDistributeShares,
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());

    // try accusing the 1st validator of bad shares using valid encrypted shares and commitments
    await ethdkg
      .connect(await getValidatorEthAccount(validators4BadDistributeShares[3]))
      .accuseParticipantDistributedBadShares(
        validators4BadDistributeShares[0].address,
        validators4BadDistributeShares[0].encryptedShares,
        validators4BadDistributeShares[0].commitments,
        validators4BadDistributeShares[3].sharedKey as [
          BigNumberish,
          BigNumberish
        ],
        validators4BadDistributeShares[3].sharedKeyProof as [
          BigNumberish,
          BigNumberish
        ]
      );

    expect(
      await validatorPool.isValidator(validators4BadDistributeShares[3].address)
    ).to.equal(true);
    expect(
      await validatorPool.isValidator(validators4BadDistributeShares[0].address)
    ).to.equal(false);

    // try moving into the next phase - KeyShareSubmission
    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
    await endCurrentPhase(ethdkg);
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());
    await assertETHDKGPhase(ethdkg, Phase.DisputeShareDistribution);
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );

    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount(validators4BadDistributeShares[3])
        )
        .submitKeyShare(
          validators4BadDistributeShares[3].keyShareG1,
          validators4BadDistributeShares[3].keyShareG1CorrectnessProof,
          validators4BadDistributeShares[3].keyShareG2
        )
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInKeyshareSubmissionPhase`
      )
      .withArgs(Phase.DisputeShareDistribution);
  });

  it("should not allow proceeding to next phase (KeyShareSubmission) after bad shares accusations take place, along with accusations of non-participant validators", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4BadDistributeShares);

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // the 4th validator won't distribute shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4BadDistributeShares.slice(0, 3),
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);
    await endCurrentPhase(ethdkg);
    // await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt())
    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    // accuse the 1st validator of bad shares using valid encrypted shares and commitments
    await ethdkg
      .connect(await getValidatorEthAccount(validators4BadDistributeShares[2]))
      .accuseParticipantDistributedBadShares(
        validators4BadDistributeShares[0].address,
        validators4BadDistributeShares[0].encryptedShares,
        validators4BadDistributeShares[0].commitments,
        validators4BadDistributeShares[2].sharedKey as [
          BigNumberish,
          BigNumberish
        ],
        validators4BadDistributeShares[2].sharedKeyProof as [
          BigNumberish,
          BigNumberish
        ]
      );

    expect(
      await validatorPool.isValidator(validators4BadDistributeShares[2].address)
    ).to.equal(true);
    expect(
      await validatorPool.isValidator(validators4BadDistributeShares[0].address)
    ).to.equal(false);

    // also accuse the 4th validator of not distributing shares
    await ethdkg.accuseParticipantDidNotDistributeShares([
      validators4BadDistributeShares[3].address,
    ]);
    expect(
      await validatorPool.isValidator(validators4BadDistributeShares[3].address)
    ).to.equal(false);
    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // try accusing the 1st validator again of bad shares using valid encrypted shares and commitments
    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount(validators4BadDistributeSharesLast[2])
        )
        .accuseParticipantDistributedBadShares(
          validators4BadDistributeSharesLast[0].address,
          validators4BadDistributeSharesLast[0].encryptedShares,
          validators4BadDistributeSharesLast[0].commitments,
          validators4BadDistributeSharesLast[2].sharedKey as [
            BigNumberish,
            BigNumberish
          ],
          validators4BadDistributeSharesLast[2].sharedKeyProof as [
            BigNumberish,
            BigNumberish
          ]
        )
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(
        ethers.utils.getAddress(validators4BadDistributeSharesLast[0].address)
      );

    // try moving into the next phase - KeyShareSubmission
    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);
    await endCurrentPhase(ethdkg);
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());
    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount(validators4BadDistributeShares[2])
        )
        .submitKeyShare(
          validators4BadDistributeShares[2].keyShareG1,
          validators4BadDistributeShares[2].keyShareG1CorrectnessProof,
          validators4BadDistributeShares[2].keyShareG2
        )
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInKeyshareSubmissionPhase`
      )
      .withArgs(Phase.ShareDistribution);

    // try accusing the 1st validator again of bad shares using valid encrypted shares and commitments
    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount(validators4BadDistributeSharesLast[2])
        )
        .accuseParticipantDistributedBadShares(
          validators4BadDistributeSharesLast[0].address,
          validators4BadDistributeSharesLast[0].encryptedShares,
          validators4BadDistributeSharesLast[0].commitments,
          validators4BadDistributeSharesLast[2].sharedKey as [
            BigNumberish,
            BigNumberish
          ],
          validators4BadDistributeSharesLast[2].sharedKeyProof as [
            BigNumberish,
            BigNumberish
          ]
        )
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(
        ethers.utils.getAddress(validators4BadDistributeSharesLast[0].address)
      );

    await endCurrentAccusationPhase(ethdkg);
    await mineBlocks((await ethdkg.getConfirmationLength()).toBigInt());
    await assertETHDKGPhase(ethdkg, Phase.ShareDistribution);

    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount(validators4BadDistributeShares[2])
        )
        .submitKeyShare(
          validators4BadDistributeShares[2].keyShareG1,
          validators4BadDistributeShares[2].keyShareG1CorrectnessProof,
          validators4BadDistributeShares[2].keyShareG2
        )
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInKeyshareSubmissionPhase`
      )
      .withArgs(Phase.ShareDistribution);

    // try accusing the 1st validator again of bad shares using valid encrypted shares and commitments
    const txPromise = ethdkg
      .connect(
        await getValidatorEthAccount(validators4BadDistributeSharesLast[2])
      )
      .accuseParticipantDistributedBadShares(
        validators4BadDistributeSharesLast[0].address,
        validators4BadDistributeSharesLast[0].encryptedShares,
        validators4BadDistributeSharesLast[0].commitments,
        validators4BadDistributeSharesLast[2].sharedKey as [
          BigNumberish,
          BigNumberish
        ],
        validators4BadDistributeSharesLast[2].sharedKeyProof as [
          BigNumberish,
          BigNumberish
        ]
      );
    const expectedBlockNumber = (
      await getReceiptForFailedTransaction(txPromise)
    ).blockNumber;
    const expectedCurrentPhase = await ethdkg.getETHDKGPhase();
    const phaseStartBlock = await ethdkg.getPhaseStartBlock();
    const phaseLength = await ethdkg.getPhaseLength();
    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
  });
});
