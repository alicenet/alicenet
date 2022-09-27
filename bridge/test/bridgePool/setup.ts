import { BytesLike, defaultAbiCoder } from "ethers/lib/utils";
import hre, { ethers } from "hardhat";

// The following values can be obtained from accusation_builder_test.go execution
export const txRoot =
  "0xb272723f931f12cc652801de2f17ec7bdcb9d07f278d9141627c0624b355860c";
export const stateRoot =
  "0x0d66a8a0babec3d38b67b5239c1683f15a57e087f3825fac3d70fd6a243ed30b";
export const headerRoot =
  "0xd2a29ef9245ea33f1b47508646a26334c276d2a1ebd671c8439f48d8a2f235fb";

export const bClaims =
  "0x0000000002000400010000000200000001000000000000000d00000002010000190000000201000025000000020100003100000002010000fb9caafb31a4d4cada31e5d20ef0952832d8f37a70aa40c7035a693199e964afb272723f931f12cc652801de2f17ec7bdcb9d07f278d9141627c0624b355860c0d66a8a0babec3d38b67b5239c1683f15a57e087f3825fac3d70fd6a243ed30bd2a29ef9245ea33f1b47508646a26334c276d2a1ebd671c8439f48d8a2f235fb";

export const bClaimsSigGroup =
  "0xf99d19d6173cf8e0c99dde2fb9d50e7f15559eba9f86a4def23cb600bdf233fd537d9c67d29b0d80cc8054bf80769663f66fd1086a0adac2bbff0c810bc3325b01";

export const proofs: [BytesLike, BytesLike, BytesLike, BytesLike] = [
  // proofOfInclusionStateRootCapnProto
  "0x010005cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000af469f3b9864a5132323df8bdd9cbd59ea728cd7525b65252133a5a02f1566ee00010003a8793650a7050ac58cf53ea792426b97212251673788bf0b4045d0bb5bdc3843aafb9eb5ced6edc2826e734abad6235c8cf638c812247fd38f04e7080d431933b9c6d6f24756341fde3e8055dd3a83743a94dddc122ab3f32a3db0c4749ff57bad",

  // proofOfInclusionTXRootCapnProto
  "0x010000b4aec67f3220a8bcdee78d4aaec6ea419171e3db9c27c65d70cc85d60e07a3f70000000000000000000000000000000000000000000000000000000000000000111c8c4c333349644418902917e1a334a6f270b8b585661a91165298792437ed0001000000",

  // proofOfInclusionTXHashCapnProto
  "0x010002cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000db3b45f6122fb536c80ace7701d8ade36fb508adf5296f12edfb1dfdbb242e0d00010002c042a165d6c097eb5ac77b72581c298f93375322fc57a03283891c2acc2ce66f8cb3dc4df7060209ffb1efa66f693ed888c15034398312231b51894f86343c2421",

  // proofOfInclusionHeaderRootCapnProto
  "0x010003d97c1b45464c01481c8df20932ce3292c99e7d5e5df07f6e1ca639dedb05b42b0000000000000000000000000000000000000000000000000000000000000000323bfb4d198651454355911116e41425aa3668eec380c958f89bc3fe76c88ab500010003e0ad098365a75c471199ee17c3e53b26ab5e404ad6697dc3033c686d84479207c40de2b3c2e453811cc0f64df31de8864ca32742fee909b443660d9f20eb7f280e64abf1c51eedeef9d757bfce9edc3af361b304bdeedd321e343921c04f3b2005",
];
export const txInPreImage =
  "0x0000000001000100010000000000000001000000020100007b802d223569d7b75cec992b1b028b0c2092d950d992b11187f11ee568c469bd";

export const wrongProofs: [BytesLike, BytesLike, BytesLike, BytesLike] = [
  // proofOfInclusionStateRootCapnProto
  "0x010005cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000af469f3b9864a5132323df8bdd9cbd59ea728cd7525b65252133a5a02f1566ee00010003a8793650a7050ac58cf53ea792426b97212251673788bf0b4045d0bb5bdc3843aafb9eb5ced6edc2826e734abad6235c8cf638c812247fd38f04e7080d431933b9c6d6f24756341fde3e8055dd3a83743a94dddc122ab3f32a3db0c4749ff57fff",

  // proofOfInclusionTXRootCapnProto
  "0x010000b4aec67f3220a8bcdee78d4aaec6ea419171e3db9c27c65d70cc85d60e07a3f70000000000000000000000000000000000000000000000000000000000000000111c8c4c333349644418902917e1a334a6f270b8b585661a91165298792437ed0001000fff",

  // proofOfInclusionTXHashCapnProto
  "0x010002cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000db3b45f6122fb536c80ace7701d8ade36fb508adf5296f12edfb1dfdbb242e0d00010002c042a165d6c097eb5ac77b72581c298f93375322fc57a03283891c2acc2ce66f8cb3dc4df7060209ffb1efa66f693ed888c15034398312231b51894f86343c2fff",

  // proofOfInclusionHeaderRootCapnProto
  "0x010003d97c1b45464c01481c8df20932ce3292c99e7d5e5df07f6e1ca639dedb05b42b0000000000000000000000000000000000000000000000000000000000000000323bfb4d198651454355911116e41425aa3668eec380c958f89bc3fe76c88ab500010003e0ad098365a75c471199ee17c3e53b26ab5e404ad6697dc3033c686d84479207c40de2b3c2e453811cc0f64df31de8864ca32742fee909b443660d9f20eb7f280e64abf1c51eedeef9d757bfce9edc3af361b304bdeedd321e343921c04f3b2fff",
];

export function getMockBlockClaimsForSnapshot() {
  return defaultAbiCoder.encode(
    ["uint32", "uint32", "uint32", "bytes32", "bytes32", "bytes32", "bytes32"],
    [1, 0, 0, ethers.constants.HashZero, txRoot, stateRoot, headerRoot]
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
