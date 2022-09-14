import { defaultAbiCoder } from "ethers/lib/utils";
import hre, { ethers } from "hardhat";

export const tokenId = 1;
export const tokenAmount = 100;
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

export function getEncodedBurnedUTXO(userAddress: string) {
  return defaultAbiCoder.encode(
    [
      "tuple(uint256 chainId, address owner, uint256 tokenId_, uint256 tokenAmount_, uint256 fee, bytes32 txHash)",
    ],
    [
      {
        chainId: 0,
        owner: userAddress,
        tokenId_: tokenId,
        tokenAmount_: tokenAmount,
        fee: 1,
        txHash:
          "0x0000000000000000000000000000000000000000000000000000000000000000",
      },
    ]
  );
}

export const getBridgePoolMetamorphicAddress = (
  factoryAddress: string,
  salt: string
): string => {
  const initCode = "0x6020363636335afa1536363636515af43d36363e3d36f3";
  return ethers.utils.getCreate2Address(
    factoryAddress,
    salt,
    ethers.utils.keccak256(initCode)
  );
};

export const getMetamorphicContractAddress = (
  contractName: string,
  factoryAddress: string
): string => {
  const metamorphicContractBytecodeHash_ =
    "0x1c0bf703a3415cada9785e89e9d70314c3111ae7d8e04f33bb42eb1d264088be";
  const salt = ethers.utils.formatBytes32String(contractName);
  return (
    "0x" +
    ethers.utils
      .keccak256(
        ethers.utils.solidityPack(
          ["bytes1", "address", "bytes32", "bytes32"],
          ["0xff", factoryAddress, salt, metamorphicContractBytecodeHash_]
        )
      )
      .slice(-40)
  );
};

export const getSimulatedBridgeRouter = async (
  factoryAddress: string
): Promise<any> => {
  const bridgeRouterAddress = await getMetamorphicContractAddress(
    "BridgeRouter",
    factoryAddress
  );
  const [admin] = await ethers.getSigners();
  await admin.sendTransaction({
    to: bridgeRouterAddress,
    value: ethers.utils.parseEther("1"),
  });
  await hre.network.provider.request({
    method: "hardhat_impersonateAccount",
    params: [bridgeRouterAddress],
  });
  return ethers.provider.getSigner(bridgeRouterAddress);
};
