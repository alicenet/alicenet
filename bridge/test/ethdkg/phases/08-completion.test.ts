import { getValidatorEthAccount } from "../../setup";
import { validators4 } from "../assets/4-validators-successful-case";
import {
  assertETHDKGPhase,
  assertEventValidatorSetCompleted,
  endCurrentPhase,
  expect,
  getInfoForIncorrectPhaseCustomError,
  Phase,
  PLACEHOLDER_ADDRESS,
  startAtGPKJ,
  submitValidatorsGPKJ,
} from "../setup";

describe("ETHDKG: ETHDKG Completion", () => {
  it("should not allow completion until after the DisputeGPKj phase", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    const txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();
    const [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeGPKJSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();

    await assertEventValidatorSetCompleted(
      tx,
      validators4.length,
      expectedNonce,
      0,
      0,
      0,
      validators4[0].mpk
    );
  });

  it("should not allow non-validators to complete ETHDKG", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );
    const txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();
    const [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeGPKJSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    const validatorAddress = "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac";
    // non-validator tries to complete ethdkg
    await expect(
      ethdkg.connect(await getValidatorEthAccount(validatorAddress)).complete()
    )
      .to.be.revertedWithCustomError(ethdkg, "OnlyValidatorsAllowed")
      .withArgs(validatorAddress);
  });

  it("should not allow double completion of ETHDKG", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    let txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();
    let [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeGPKJSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();

    await assertEventValidatorSetCompleted(
      tx,
      validators4.length,
      expectedNonce,
      0,
      0,
      0,
      validators4[0].mpk
    );

    // try completing again
    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeGPKJSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
  });

  it("should not allow validators to participate in previous phases", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    let txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();
    let [
      ethDKGPhases,
      ETHDKGAccusations,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeGPKJSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();

    await assertEventValidatorSetCompleted(
      tx,
      validators4.length,
      expectedNonce,
      0,
      0,
      0,
      validators4[0].mpk
    );

    // try participating in previous phases
    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .register(validators4[0].aliceNetPublicKey);
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.RegistrationOpen,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .accuseParticipantNotRegistered([]);
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.RegistrationOpen,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .distributeShares(
        validators4[0].encryptedShares,
        validators4[0].commitments
      );
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    // distribute shares before the time
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.ShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .accuseParticipantDidNotDistributeShares([]);
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .accuseParticipantDistributedBadShares(
        PLACEHOLDER_ADDRESS,
        [],
        [
          [0, 0],
          [0, 0],
        ],
        [0, 0],
        [0, 0]
      );
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);

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

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .submitKeyShare(
        validators4[0].keyShareG1,
        validators4[0].keyShareG1CorrectnessProof,
        validators4[0].keyShareG2
      );
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.KeyShareSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .accuseParticipantDidNotSubmitKeyShares([]);
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.KeyShareSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .submitMasterPublicKey([0, 0, 0, 0]);
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.MPKSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .submitGPKJ([0, 0, 0, 0]);
    [
      ,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.GPKJSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .accuseParticipantDidNotSubmitGPKJ([]);
    [
      ,
      ETHDKGAccusations,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.GPKJSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .accuseParticipantSubmittedBadGPKJ(
        [],
        [],
        [[[0, 0]]],
        PLACEHOLDER_ADDRESS
      );
    [
      ,
      ETHDKGAccusations,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeGPKJSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.GPKJSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();
    [
      ,
      ETHDKGAccusations,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.DisputeGPKJSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
  });
});
