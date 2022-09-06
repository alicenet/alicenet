import { BigNumber } from "ethers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { Fixture } from "../setup";

const ERC1155_FIXED_TOKEN_ID = 1;
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
    ERC721: {
      address: string;
      admin: bigint;
      user: bigint;
      bridgePool: bigint;
    };
    ERC1155: {
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
      ERC721: {
        address: fixture.erc721Mock.address.slice(-4),
        admin: (await fixture.erc721Mock.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.erc721Mock.balanceOf(user.address)).toBigInt(),
        bridgePool: (
          await fixture.erc721Mock.balanceOf(bridgePoolAddress)
        ).toBigInt(),
      },
      ERC1155: {
        address: fixture.erc1155Mock.address.slice(-4),
        admin: (
          await fixture.erc1155Mock.balanceOf(
            admin.address,
            ERC1155_FIXED_TOKEN_ID
          )
        ).toBigInt(),
        user: (
          await fixture.erc1155Mock.balanceOf(
            user.address,
            ERC1155_FIXED_TOKEN_ID
          )
        ).toBigInt(),
        bridgePool: (
          await fixture.erc1155Mock.balanceOf(
            bridgePoolAddress,
            ERC1155_FIXED_TOKEN_ID
          )
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

export function getMockBlockClaimsForStateRoot(stateRoot: string) {
  return defaultAbiCoder.encode(
    ["uint32", "uint32", "uint32", "bytes32", "bytes32", "bytes32", "bytes32"],
    [
      0,
      0,
      0,
      "0x0000000000000000000000000000000000000000000000000000000000000000",
      "0x0000000000000000000000000000000000000000000000000000000000000000",
      stateRoot,
      "0x0000000000000000000000000000000000000000000000000000000000000000",
    ]
  );
}
export const valueOrId = 100; // value if ERC20 , tokenId if ERC721
export const maxEth = 1;
export const maxTokens = 1010; // has to be > bTokenFee => 10
export const valueSent = ethers.utils.parseEther("1.0");

export const tokenTypes = [
  {
    it: "ERC20",
    options: {
      ercContractName: "erc20Mock",
      poolType: 1,
      bridgeImpl: "localERC20BridgePoolV1",
      quantity: valueOrId,
      errorReason: "ERC20: insufficient allowance",
    },
  },
/*   {
    it: "ERC721",
    options: {
      ercContractName: "erc721Mock",
      poolType: 2,
      bridgeImpl: "LocalERC721BridgePoolV1",
      quantity: 1,
      errorReason: "ERC721: invalid token ID",
    },
  },
  {
    it: "ERC1155",
    options: {
      ercContractName: "erc1155Mock",
      poolType: 3,
      bridgeImpl: "LocalERC1155BridgePoolV1",
      quantity: 1,
      errorReason: "ERC1155: caller is not token owner nor approved",
    },
  },
 */];

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
