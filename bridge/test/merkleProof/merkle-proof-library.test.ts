import { expect } from "chai";
import { MockMerkleProofLibrary } from "../../typechain-types";
import { deployLibrary } from "./setup";

const ZERO_BYTES32 =
  "0x0000000000000000000000000000000000000000000000000000000000000000";
const VALID_AUDIT_PATH =
  "0x066c7a6ef776fbae26f10eabcc5f0eb72b0f527c4cad8c4037940a28c2fe31596974f1d60877fdff3125a5c1adb630afe3aa899820a0531cea8ee6a85eb1150925fc463686d6201f9c3973a9ebeabe4375d5a76d935bbec6cbb9d18bffe67216";
const VALID_MERKLE_ROOT =
  "0x51f16d008cec2409af8104a0bca9facec585e02c12d2fa5221707672410dc692";
const VALID_PROOF_VALUE =
  "0x0391f56ce9575815216c9c0fcffa1d50767adb008c1491b7da2dbc323b8c1fb5";
const VALID_HEIGHT_FOR_VALID_PROOF = 0x4;
const VALID_HEIGHT_RANGE_MAX = 0x100; // 256
const VALID_HEIGHT_RANGE_MIN = 0x00;
const INVALID_HEIGHT_ABOVE_RANGE_MAX = 0x101; // 257
const VALID_BITSET = "0xb0";
const VALID_KEY =
  "0x80ab269d23d84721a53f9f3accb024a1947bcf5e4910a152f38d55d7d644c995"; // utxoID
const VALID_KEYHASH =
  "0xa53ec428ed37200bcb4944a99107b738c1a58ef76287b130583095c58b0f45e4";

describe("Testing Merkle Proof Library", async () => {
  let MerkleProofLibrary: MockMerkleProofLibrary;

  beforeEach(async function () {
    MerkleProofLibrary = await deployLibrary();
  });

  describe("verifyNonInclusion:", async () => {
    it("does not revert with non inclusion default hash MerkleProof", async () => {
      const auditPath =
        "0xe602c66f5176c6d2a33d6eb3addf38c937e0e32457e58148883578cb4655b826";
      const root =
        "0x529312e0c69f0cc47d27630461d884d9537ebfa51d57c1ecf7f38f77c801373f";
      const key = VALID_KEY;
      const bitset = "0x80";
      const height = 0x1;
      const included = false;
      const proofKeyZero = ZERO_BYTES32;
      const proofValueZero = ZERO_BYTES32;

      await expect(
        MerkleProofLibrary.verifyNonInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey: proofKeyZero,
            proofValue: proofValueZero,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).not.to.be.reverted;
    });

    it("does not revert with non inclusion leaf node MerkleProof", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const key =
        "0x80ab269d23d84721a53f9f3accb024a1947bcf5e4910a152f38d55d7d644c996"; // utxoID
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;
      const included = false;
      const proofKey = VALID_KEY;
      const proofValue = VALID_PROOF_VALUE;

      await expect(
        MerkleProofLibrary.verifyNonInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey,
            proofValue,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).not.to.be.reverted;
    });

    it("reverts with non inclusion proof value 0 and proof key different from 0", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const key = VALID_KEY;
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;
      const included = false;
      const proofKey = VALID_PROOF_VALUE;
      const proofValueZero = ZERO_BYTES32;

      await expect(
        MerkleProofLibrary.verifyNonInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey,
            proofValue: proofValueZero,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).to.be.revertedWithCustomError(
        MerkleProofLibrary,
        "InvalidNonInclusionMerkleProof"
      );
    });

    it("reverts with valid proof with key not in the key path", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const key = VALID_KEY;
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;
      const included = true;
      // Proof key is inside the trie but is not in the path of our Key
      const proofKey =
        "0x7a6315f5d19bf3f3bed9ef4e6002ebf76d4d05a7f7e84547e20b40fde2c34411";
      const proofValue =
        "0xe10fbdbaa5b72d510af6dd5ebd08da8b2fbd2b06d4787ce15a6eaf518c2d97fc";

      await expect(
        MerkleProofLibrary.verifyNonInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey,
            proofValue,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).to.be.revertedWithCustomError(
        MerkleProofLibrary,
        "ProvidedLeafNotFoundInKeyPath"
      );
    });

    it("reverts with invalid proof key", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const key = VALID_KEY;
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;
      const included = true;
      // Proof key is inside the trie but is not in the path of our Key
      const proofKey =
        "0x7a6315f5d19bf3f3bed9ef4e6002ebf76d4d05a7f7e84547e20b40fde2c34416";
      const proofValue =
        "0xe10fbdbaa5b72d510af6dd5ebd08da8b2fbd2b06d4787ce15a6eaf518c2d97fc";

      await expect(
        MerkleProofLibrary.verifyNonInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey,
            proofValue,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).to.be.revertedWithCustomError(
        MerkleProofLibrary,
        "ProvidedLeafNotFoundInKeyPath"
      );
    });

    it("reverts with non inclusion of an included key merkle proof", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const key = VALID_KEY;
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;
      const included = true;
      const proofKeyZero = ZERO_BYTES32;
      const proofValue = VALID_PROOF_VALUE;

      await expect(
        MerkleProofLibrary.verifyNonInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey: proofKeyZero,
            proofValue,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).to.be.revertedWithCustomError(
        MerkleProofLibrary,
        "InvalidNonInclusionMerkleProof"
      );
    });
  });

  describe("verifyInclusion:", async () => {
    it("Does not revert if proof is included", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const key = VALID_KEY;
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;
      const included = true;
      const proofKeyZero = ZERO_BYTES32;
      const proofValue = VALID_PROOF_VALUE;

      await expect(
        MerkleProofLibrary.verifyInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey: proofKeyZero,
            proofValue,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).not.to.be.reverted;
    });

    it("reverts with invalid inclusion Merkle Proof", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const key =
        "0x80ab269d23d84721a53f9f3accb024a1947bcf5e4910a152f38d55d7d644c996"; // utxoID
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;
      const included = false;
      const proofKeyZero = ZERO_BYTES32;
      const proofValue = VALID_PROOF_VALUE;

      await expect(
        MerkleProofLibrary.verifyInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey: proofKeyZero,
            proofValue,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).to.be.revertedWithCustomError(
        MerkleProofLibrary,
        "ProofDoesNotMatchTrieRoot"
      );
    });

    it("reverts with inclusion without proof value", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const key = VALID_KEY;
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;
      const included = false;
      const proofKeyZero = ZERO_BYTES32;
      const proofValueZero = ZERO_BYTES32;

      await expect(
        MerkleProofLibrary.verifyInclusion(
          {
            included,
            keyHeight: height,
            key,
            proofKey: proofKeyZero,
            proofValue: proofValueZero,
            bitmap: bitset,
            auditPath,
          },
          root
        )
      ).to.be.revertedWithCustomError(MerkleProofLibrary, "InclusionZero");
    });
  });
  describe("checkProof:", async () => {
    it("returns correct leaf hash when data is valid", async () => {
      const key = VALID_KEY;
      const proofValue = VALID_PROOF_VALUE;
      const height = 0x4;

      const expectedKeyHash =
        "0xa53ec428ed37200bcb4944a99107b738c1a58ef76287b130583095c58b0f45e4";

      const actualKeyHash = await MerkleProofLibrary.computeLeafHash(
        key,
        proofValue,
        height
      );
      await expect(actualKeyHash).to.be.equal(expectedKeyHash);
    });

    it("reverts when height is out of range", async () => {
      const key = VALID_KEY;
      const proofValue = VALID_PROOF_VALUE;
      const height = INVALID_HEIGHT_ABOVE_RANGE_MAX;

      await expect(MerkleProofLibrary.computeLeafHash(key, proofValue, height))
        .to.be.revertedWithCustomError(MerkleProofLibrary, "InvalidProofHeight")
        .withArgs(height);
    });

    it("succeeds when height is zero and returns correct leaf hash", async () => {
      const key = VALID_KEY;
      const proofValue = VALID_PROOF_VALUE;
      const height = VALID_HEIGHT_RANGE_MIN;

      const expectedKeyHash =
        "0x435d8718a62e73fac8e2b7f99d89161dc10e713284ab59edd6f22c9858ab1617";

      const actualKeyHash = await MerkleProofLibrary.computeLeafHash(
        key,
        proofValue,
        height
      );
      await expect(actualKeyHash).to.be.equal(expectedKeyHash);
    });
  });

  describe("checkProof:", async () => {
    it("returns true when key is valid inside MerkleTree", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const keyHash = VALID_KEYHASH;
      const key = VALID_KEY;
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;

      const valid = await MerkleProofLibrary.checkProof(
        auditPath,
        root,
        keyHash,
        key,
        bitset,
        height
      );
      await expect(valid).to.be.true;
    });

    it("returns false when height is greater than 256", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const keyHash = VALID_KEYHASH;
      const key = VALID_KEY;
      const bitset = VALID_BITSET;
      const height = INVALID_HEIGHT_ABOVE_RANGE_MAX;

      await expect(
        MerkleProofLibrary.checkProof(
          auditPath,
          root,
          keyHash,
          key,
          bitset,
          height
        )
      )
        .to.be.revertedWithCustomError(MerkleProofLibrary, "InvalidProofHeight")
        .withArgs(height);
    });

    it("returns true when proof is valid inside MerkleTree with height 0", async () => {
      const auditPath = "0x";
      const root =
        "0xa368ad44a1ae7ea22d0857c32da7e679fb29c4357e4de838394faf0bd57e1493";
      const keyHash =
        "0xa368ad44a1ae7ea22d0857c32da7e679fb29c4357e4de838394faf0bd57e1493";
      const key =
        "0x6e286c26d8715685894787919c58c9aeee7dff73f88bf476dab1d282d535e5f2"; // utxoID
      const bitset = "0x00";
      const height = VALID_HEIGHT_RANGE_MIN;

      const valid = await MerkleProofLibrary.checkProof(
        auditPath,
        root,
        keyHash,
        key,
        bitset,
        height
      );
      await expect(valid).to.be.true;
    });

    it("returns true when proof is valid inside MerkleTree with height 256", async () => {
      const auditPath =
        "0x166eacf747876e3a8eb60fdd36cfac41e9ae27098cf605d90a019a895e98af17" +
        "1ddb3652b5ecb0881027ca462000d3614fd7735855db004c598622855e9190c4" +
        "9bd43690efe794fff0a9790fc9c3f694e251cd15dbf06a5217005cffc23943c2" +
        "54ec9d782b8991251f91af85390450ad21fc2befaf8f473367a30d579c36e221" +
        "480543d6cc63a90d7c121e5219495ebc0091dd7355763da9d97931997f52260b" +
        "1c925246c3b9255ebd67fc7daf4769f556b5eaefbd62d31daf7f6c43da4ffe2c";
      const root =
        "0x0d66a8a0babec3d38b67b5239c1683f15a57e087f3825fac3d70fd6a243ed30b";
      const key =
        "0x0000000000000000000000000000000000000000000000000000000000000030";
      const keyHash =
        "0x5179fc581e28dfb3f4f7202cc76cf896b86c982c2a70e7b607009ce1a9e86395";
      const bitset =
        "0xe00000000000000000000000000000000000000000000000000000000000002300";
      const height = VALID_HEIGHT_RANGE_MAX;

      const valid = await MerkleProofLibrary.checkProof(
        auditPath,
        root,
        keyHash,
        key,
        bitset,
        height
      );
      await expect(valid).to.be.true;
    });

    it("returns false when key is not valid inside MerkleTree", async () => {
      const auditPath = VALID_AUDIT_PATH;
      const root = VALID_MERKLE_ROOT;
      const keyHash =
        "0xd2056ce8d2aca5dfebd7a2ee14e01d4beda84ba1c5968cd7f0717de28d89988c";
      const key =
        "0x80ab269d23d84721a53f9f3accb024a1947bcf5e4910a152f38d55d7d644c996"; // utxoID
      const bitset = VALID_BITSET;
      const height = VALID_HEIGHT_FOR_VALID_PROOF;

      const valid = await MerkleProofLibrary.checkProof(
        auditPath,
        root,
        keyHash,
        key,
        bitset,
        height
      );
      await expect(valid).to.be.false;
    });

    it("returns false when key is not valid inside MerkleTree 2", async () => {
      const auditPath =
        "0xae68be4a30b6e4158f672298af814ce905ec1d486c0705f3a964859610eaef45";
      const root =
        "0x6cfa7a7076fd4b09b8dd9bb1d0737254f200c794bc46394b01b0b43f9de67090";
      const keyHash =
        "0x9105f42f6c6392bfcd9ca1cf5c4556e2455554e5467348569d357a6e7874ca82";
      const key =
        "0x80ab269d23d84721a53f9f3accb024a1947bcf5e4910a152f38d55d7d644c996"; // utxoID
      const bitset = "0x80";
      const height = 0x1;

      const valid = await MerkleProofLibrary.checkProof(
        auditPath,
        root,
        keyHash,
        key,
        bitset,
        height
      );
      await expect(valid).to.be.false;
    });
  });
});
