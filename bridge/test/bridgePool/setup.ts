import { defaultAbiCoder } from "ethers/lib/utils"; 
import { ethers } from "hardhat";
import hre from "hardhat";
import { Fixture } from "../setup";
import { BigNumber } from "ethers";

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

export const getMetamorphicContractAddress = (contractName :string, factoryAddress:string): string => {
  const metamorphicContractBytecodeHash_ = "0x1c0bf703a3415cada9785e89e9d70314c3111ae7d8e04f33bb42eb1d264088be";
  const salt = ethers.utils.formatBytes32String(contractName);
  return "0x"+ethers.utils.keccak256(
    ethers.utils.solidityPack(
      ["bytes1", "address","bytes32", "bytes32"],
      [
        "0xff",
        factoryAddress,
        salt,
        metamorphicContractBytecodeHash_
      ]
    )
  ).slice(-40);
};

export const getSimulatedBridgeRouter = async (factoryAddress: string): Promise<any> => {
  const  bridgeRouterAddress = await getMetamorphicContractAddress("BridgeRouter",factoryAddress)
  const [admin] = await ethers.getSigners();
  await admin.sendTransaction({to: bridgeRouterAddress , value: ethers.utils.parseEther("1")})
await hre.network.provider.request({
  method: "hardhat_impersonateAccount",
  params: [bridgeRouterAddress],
});
return ethers.provider.getSigner(
  bridgeRouterAddress
);
}

export const tokenTypes = [
  {
    it: "ERC20",
    options: {
      ercContractName: "erc20Mock",
      tokenType: 1,
      bridgeImpl: "localERC20BridgePoolV1",
      quantity: valueOrId,
      errorReason: "ERC20: insufficient allowance",
    },
  },
]


export interface state {
  Balances: {
    bToken: {
      address: string;
      admin: bigint;
      user: bigint;
      bridgePool: bigint;
      totalSupply: bigint;
    };
    eth: {
      address: string;
      // We leave user balance as number to round values and avoid consumed gas comparison
      admin: bigint;
      user: bigint;
      bridgePool: bigint;
      aToken: bigint;
      bToken: bigint;
    };
    ERC20: {
      address: string;
      admin: bigint;
      user: bigint;
      bridgePool: bigint;
    };
  };
}

export async function getState(fixture: Fixture, bridgePoolAddress: string) {
  const [admin, user] = await ethers.getSigners();
  const state: state = {
    Balances: {
      bToken: {
        address: fixture.bToken.address.slice(-4),
        admin: (await fixture.bToken.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.bToken.balanceOf(user.address)).toBigInt(),
        bridgePool: (
          await fixture.bToken.balanceOf(bridgePoolAddress)
        ).toBigInt(),
        totalSupply: (await fixture.bToken.totalSupply()).toBigInt(),
      },
      eth: {
        address: "0000",
        admin: formatBigInt(await ethers.provider.getBalance(admin.address)),
        user: formatBigInt(await ethers.provider.getBalance(user.address)),
        bridgePool: (
          await ethers.provider.getBalance(bridgePoolAddress)
        ).toBigInt(),
        aToken: (
          await ethers.provider.getBalance(fixture.aToken.address)
        ).toBigInt(),
        bToken: (
          await ethers.provider.getBalance(fixture.bToken.address)
        ).toBigInt(),
      },
      ERC20: {
        address: fixture.erc20Mock.address.slice(-4),
        admin: (await fixture.erc20Mock.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.erc20Mock.balanceOf(user.address)).toBigInt(),
        bridgePool: (
          await fixture.erc20Mock.balanceOf(bridgePoolAddress)
        ).toBigInt(),
      },
    },
  };
  return state;
}

export function showState(title: string, state: state) {
  if (process.env.npm_config_detailed === "true") {
    // execute "npm run --detailed=true test ..." to see this output
    console.log(title, state);
  }
}

export function format(number: BigNumber) {
  return parseFloat((+ethers.utils.formatEther(number)).toFixed(0));
}

export function formatBigInt(number: BigNumber) {
  return BigInt(parseFloat((+ethers.utils.formatEther(number)).toFixed(0)));
}

