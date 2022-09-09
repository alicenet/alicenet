import { defaultAbiCoder } from "ethers/lib/utils"; 
import { ethers } from "hardhat";

export const valueOrId = 100;
// The following merkle proof and stateRoot values can be obtained from accusation_builder_test.go execution
export const merkleProof =
  "0x010005cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000af469f3b9864a5132323df8bdd9cbd59ea728cd7525b65252133a5a02f1566ee00010003a8793650a7050ac58cf53ea792426b97212251673788bf0b4045d0bb5bdc3843aafb9eb5ced6edc2826e734abad6235c8cf638c812247fd38f04e7080d431933b9c6d6f24756341fde3e8055dd3a83743a94dddc122ab3f32a3db0c4749ff57bad"; // capnproto
export const wrongMerkleProof =
  "0x010005cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000af469f3b9864a5132323df8bdd9cbd59ea728cd7525b65252133a5a02f1566ee00010003a8793650a7050ac58cf53ea792426b97212251673788bf0b4045d0bb5bdc3843aafb9eb5ced6edc2826e734abad6235c8cf638c812247fd38f04e7080d431933b9c6d6f24756341fde3e8055dd3a83743a94dddc122ab3f32a3db0c4749ff57fff";
export const stateRoot =
  "0x0d66a8a0babec3d38b67b5239c1683f15a57e087f3825fac3d70fd6a243ed30b"; // stateRoot
// Mock a merkle proof for a burned UTXO on alicenet

export function getMockBlockClaimsForStateRoot(stateRoot: string) {
  return defaultAbiCoder.encode(
    ["uint32", "uint32", "uint32", "bytes32", "bytes32", "bytes32", "bytes32"],
    [
      0,
      0,
      0,
      ethers.constants.HashZero,
      ethers.constants.HashZero,
      stateRoot,
      ethers.constants.HashZero,
    ]
  );
}

export function getEncodedBurnedUTXO(userAddress: string, valueOrId: number) {
  return defaultAbiCoder.encode(
    [
      "tuple(uint256 chainId, address owner, uint256 value, uint256 fee, bytes32 txHash)",
    ],
    [
      {
        chainId: 0,
        owner: userAddress,
        value: valueOrId,
        fee: 1,
        txHash: ethers.constants.HashZero,
      },
    ]
  );
}
