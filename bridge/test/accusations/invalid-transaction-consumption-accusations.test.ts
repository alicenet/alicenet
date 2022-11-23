import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { assert, expect } from "chai";
import { ethers } from "hardhat";
import {
  AccusationsLibraryMock,
  InvalidTxConsumptionAccusation,
} from "../../typechain-types";
import { signedData } from "../math/assets/4-validators-1000-snapshots";
import { Fixture, getFixture } from "../setup";
import { validSnapshot1024 } from "../snapshots/assets/4-validators-snapshots-1";
import {
  addValidators,
  deployLibrary,
  getAccusationDataForNonExistentUTXOWithInvalidSigGroup,
  getInvalidAccusationDataWithSpendingValidDeposit,
  getValidAccusationDataForNonExistentUTXO,
  getValidAccusationDataForNonExistentUTXOChainId2,
  getValidAccusationDataForNonExistentUTXOWithInvalidHeight,
} from "./accusations-test-helpers";

describe("InvalidTxConsumptionAccusation: Tests InvalidTxConsumptionAccusation methods", async () => {
  let fixture: Fixture;
  let accusationMock: AccusationsLibraryMock;
  let accusation: InvalidTxConsumptionAccusation;

  async function deployFixture() {
    accusationMock = await deployLibrary();
    return getFixture(true, true, true);
  }

  async function setMPKAsValid(bClaimsSig: string) {
    const mockETHDKG = await ethers.getContractAt(
      "ETHDKGMock",
      fixture.ethdkg.address
    );
    const [pk] = await accusationMock.recoverGroupSignature(bClaimsSig);

    const sigHash = ethers.utils.solidityKeccak256(
      ["uint256", "uint256", "uint256", "uint256"],
      pk
    );

    await mockETHDKG.setValidMasterPublicKey(sigHash);
  }

  beforeEach(async function () {
    fixture = await loadFixture(deployFixture);
    accusation = fixture.invalidTxConsumptionAccusation;
    const snapshotsMock = await ethers.getContractAt(
      "SnapshotsMock",
      fixture.snapshots.address
    );

    await snapshotsMock.setEpochLength(1024);
    await snapshotsMock.snapshotWithValidData(
      validSnapshot1024.GroupSignature,
      validSnapshot1024.BClaims
    );
  });

  describe("accuseInvalidTransactionConsumption:", async () => {
    const signerAccount0 = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";
    it("should successfully accuse non existent utxo", async function () {
      await addValidators(fixture.validatorPool, [signerAccount0]);
      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXO();

      await setMPKAsValid(bClaimsSigGroup);

      const signer = await accusation.accuseInvalidTransactionConsumption(
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs
      );

      assert.equal(signer, signerAccount0);
    });

    it("should successfully accuse non existent utxo when height is in valid range", async function () {
      await addValidators(fixture.validatorPool, [signerAccount0]);
      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXO();

      const snapshotsMock = await ethers.getContractAt(
        "SnapshotsMock",
        fixture.snapshots.address
      );

      await snapshotsMock.snapshotWithValidData(
        signedData[1].GroupSignature,
        signedData[1].BClaims
      );

      await setMPKAsValid(bClaimsSigGroup);

      await accusation.accuseInvalidTransactionConsumption(
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs
      );
    });

    it("reverts when accusation is older than 2 epochs", async function () {
      await addValidators(fixture.validatorPool, [signerAccount0]);
      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXO();

      const snapshotsMock = await ethers.getContractAt(
        "SnapshotsMock",
        fixture.snapshots.address
      );

      await snapshotsMock.snapshotWithValidData(
        signedData[2].GroupSignature,
        signedData[2].BClaims
      );

      await setMPKAsValid(bClaimsSigGroup);

      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(accusation, "ExpiredAccusation");
    });

    it("accusations works when in height range", async function () {
      await addValidators(fixture.validatorPool, [signerAccount0]);
      const { pClaims, pClaimsSig, txInPreImage, proofs } =
        getValidAccusationDataForNonExistentUTXO();

      const snapshotsMock = await ethers.getContractAt(
        "SnapshotsMock",
        fixture.snapshots.address
      );

      // doing a snapshot at 3072
      await snapshotsMock.snapshotWithValidData(
        signedData[2].GroupSignature,
        signedData[2].BClaims
      );

      await setMPKAsValid(signedData[2].GroupSignature);

      // height in the range of 2 epochs older
      const newPClaimsHeight1025 =
        pClaims.substring(0, 58) + "0104" + pClaims.substring(62);

      // height in the future of the current snapshot and older than latestSnapshotHeight + 1024
      const newPClaimsHeight3073 =
        pClaims.substring(0, 58) + "010c" + pClaims.substring(62);

      // since we change the data without changing the signature, we expect to pass the height check
      // and fail down the line
      await expect(
        accusation.accuseInvalidTransactionConsumption(
          newPClaimsHeight1025,
          pClaimsSig,
          signedData[0].BClaims,
          signedData[0].GroupSignature,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(accusation, "SignerNotValidValidator");

      // since we change the data without changing the signature, we expect to pass the height check
      // and fail down the line
      await expect(
        accusation.accuseInvalidTransactionConsumption(
          newPClaimsHeight3073,
          pClaimsSig,
          signedData[2].BClaims,
          signedData[2].GroupSignature,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(accusation, "SignerNotValidValidator");
    });

    it("reverts when accusation is newer than last snapshot + 1 epoch", async function () {
      await addValidators(fixture.validatorPool, [signerAccount0]);
      const { pClaims, pClaimsSig, txInPreImage, proofs } =
        getValidAccusationDataForNonExistentUTXO();

      const snapshotsMock = await ethers.getContractAt(
        "SnapshotsMock",
        fixture.snapshots.address
      );

      // doing a snapshot at 3072
      await snapshotsMock.snapshotWithValidData(
        signedData[2].GroupSignature,
        signedData[2].BClaims
      );

      await setMPKAsValid(signedData[2].GroupSignature);

      const newPClaimsHeight4097 =
        pClaims.substring(0, 58) + "0110" + pClaims.substring(62);

      await expect(
        accusation.accuseInvalidTransactionConsumption(
          newPClaimsHeight4097,
          pClaimsSig,
          signedData[3].BClaims,
          signedData[3].GroupSignature,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(accusation, "ExpiredAccusation");
    });

    it("reverts with InvalidAccusation (ConsumptionOfValidDeposit)", async function () {
      await addValidators(fixture.validatorPool, [signerAccount0]);

      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getInvalidAccusationDataWithSpendingValidDeposit();

      await setMPKAsValid(bClaimsSigGroup);

      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(
        accusation,
        "InvalidNonInclusionMerkleProof"
      );
    });

    it("reverts when validator is not valid", async function () {
      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXO();

      await setMPKAsValid(bClaimsSigGroup);

      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      )
        .to.be.revertedWithCustomError(accusation, "SignerNotValidValidator")
        .withArgs(signerAccount0);
    });

    it("reverts when chain id is not valid in BClaims and PClaims", async function () {
      const address2 = "0x03e0AcB2Bf2B41D7E102Cd44937f6c5c6F1d5353";
      await addValidators(fixture.validatorPool, [signerAccount0, address2]);
      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXO();

      const snapshotsMock = await ethers.getContractAt(
        "SnapshotsMock",
        fixture.snapshots.address
      );

      await snapshotsMock.snapshot(
        validSnapshot1024.GroupSignature,
        validSnapshot1024.BClaims
      );

      await setMPKAsValid(bClaimsSigGroup);

      const expectedBClaimsChainId = 1;
      const expectedpClaimsChainId = 1;
      const expectedActualChainId = 0;
      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      )
        .to.be.revertedWithCustomError(accusation, "ChainIdDoesNotMatch")
        .withArgs(
          expectedBClaimsChainId,
          expectedpClaimsChainId,
          expectedActualChainId
        );
    });

    it("reverts when chain id is not valid", async function () {
      const address2 = "0x03e0AcB2Bf2B41D7E102Cd44937f6c5c6F1d5353";
      await addValidators(fixture.validatorPool, [signerAccount0, address2]);
      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXOChainId2();

      await setMPKAsValid(bClaimsSigGroup);

      const expectedBClaimsChainId = 1337;
      const expectedpClaimsChainId = 1;
      const expectedActualChainId = 1;
      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      )
        .to.be.revertedWithCustomError(accusation, "ChainIdDoesNotMatch")
        .withArgs(
          expectedBClaimsChainId,
          expectedpClaimsChainId,
          expectedActualChainId
        );
    });

    it("reverts when height is not valid", async function () {
      const address2 = "0x03e0AcB2Bf2B41D7E102Cd44937f6c5c6F1d5353";
      await addValidators(fixture.validatorPool, [signerAccount0, address2]);

      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXOWithInvalidHeight();

      await setMPKAsValid(bClaimsSigGroup);

      const expectedBClaimsHeight = 42;
      const expectedPClaimsHeight = 2;

      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      )
        .to.be.revertedWithCustomError(accusation, "HeightDeltaShouldBeOne")
        .withArgs(expectedBClaimsHeight, expectedPClaimsHeight);
    });

    it("reverts when sig group is not valid", async function () {
      const signerAccount0 = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      await addValidators(fixture.validatorPool, [signerAccount0]);

      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getAccusationDataForNonExistentUTXOWithInvalidSigGroup();

      await setMPKAsValid(bClaimsSigGroup);

      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(accusation, "EllipticCurvePairingFailed");
    });

    it("reverts when block is not signed with a valid master public key", async function () {
      const signerAccount0 = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      await addValidators(fixture.validatorPool, [signerAccount0]);

      const {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getAccusationDataForNonExistentUTXOWithInvalidSigGroup();

      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(accusation, "InvalidMasterPublicKey");
    });

    it("reverts when sig group is signed with a different key", async function () {
      const signerAccount0 = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      await addValidators(fixture.validatorPool, [signerAccount0]);

      let {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXO();

      bClaimsSigGroup =
        "0x258aa89365a642358d92db67a13cb25d73e6eedf0d25100d8d91566882fac54b1ccedfb0425434b54999a88cd7d993e05411955955c0cfec9dd33066605bd4a60f6bbfbab37349aaa762c23281b5749932c514f3b8723cf9bb05f9841a7f2d0e0f75e42fd6c8e9f0edadac3dcfb7416c2d4b2470f4210f2afa93138615b1deb10cdc89f164e81cc49e06c4a7e1dcdcf7c0108e8cc9bb1032f9df6d4e834f1bb318accba7ae3f4b28bd9ba81695ba475f70d40a14b12ca3ef9764f2a6d9bfc53a";
      await setMPKAsValid(bClaimsSigGroup);

      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(
        accusation,
        "SignatureVerificationFailed"
      );
    });

    it("reverts when BClaims is invalid without transactions", async function () {
      const signerAccount0 = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      await addValidators(fixture.validatorPool, [signerAccount0]);

      let {
        pClaims,
        pClaimsSig,
        bClaims,
        bClaimsSigGroup,
        txInPreImage,
        proofs,
      } = getValidAccusationDataForNonExistentUTXO();

      // inject an invalid pClaims that doesn't have transactions
      pClaims =
        "0x0000000000000200" + // struct definition capn proto https://capnproto.org/encoding.html
        "0400000001000400" + // BClaims struct definition
        "5400000000000200" + // RCert struct definition
        "01000000" + // chainId NOTE: BClaim starts here
        "02000000" + // height
        "0d00000002010000" + // list(uint8) definition for prevBlock
        "1900000002010000" + // list(uint8) definition for txRoot
        "2500000002010000" + // list(uint8) definition for stateRoot
        "3100000002010000" + // list(uint8) definition for headerRoot
        "41b1a0649752af1b28b3dc29a1556eee781e4a4c3a1f7f53f90fa834de098c4d" + // prevBlock
        "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470" + // txRoot
        "b58904fe94d4dca4102566c56402dfa153037d18263b3f6d5574fd9e622e5627" + // stateRoot
        "3e9768bd0513722b012b99bccc3f9ccbff35302f7ec7d75439178e5a80b45800" + // headerRoot
        "0400000002000100" + // RClaims struct definition NOTE:RCert starts here
        "1d00000002060000" + // list(uint8) definition for sigGroup
        "01000000" + // chainID
        "02000000" + // Height
        "01000000" + // round
        "00000000" + // zeros pads for the round (capnproto operates using 8 bytes word)
        "0100000002010000" + // list(uint8) definition for prevBlock
        "41b1a0649752af1b28b3dc29a1556eee781e4a4c3a1f7f53f90fa834de098c4d" + // prevBlock
        "258aa89365a642358d92db67a13cb25d73e6eedf0d25100d8d91566882fac54b" +
        "1ccedfb0425434b54999a88cd7d993e05411955955c0cfec9dd33066605bd4a6" +
        "0f6bbfbab37349aaa762c23281b5749932c514f3b8723cf9bb05f9841a7f2d0e" +
        "0f75e42fd6c8e9f0edadac3dcfb7416c2d4b2470f4210f2afa93138615b1deb1" +
        "06f5308b02f59062b735d0021ba93b1b9c09f3e168384b96b1eccfed65935714" +
        "2a7bd3532dc054cb5be81e9d559128229d61a00474b983a3569f538eb03d07ce";

      await setMPKAsValid(bClaimsSigGroup);
      await expect(
        accusation.accuseInvalidTransactionConsumption(
          pClaims,
          pClaimsSig,
          bClaims,
          bClaimsSigGroup,
          txInPreImage,
          proofs
        )
      ).to.be.revertedWithCustomError(
        accusation,
        "NoTransactionInAccusedProposal"
      );
    });
  });
});
