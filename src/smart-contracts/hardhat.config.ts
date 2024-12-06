import os from "os";
import * as dotenv from "dotenv";
import { HardhatUserConfig } from "hardhat/config";
import "hardhat-contract-sizer";
import "@nomicfoundation/hardhat-verify";
import "@typechain/hardhat";
import "@nomicfoundation/hardhat-ethers";
import "@nomicfoundation/hardhat-chai-matchers";

dotenv.config();

const config: HardhatUserConfig = {
  networks: {
    dev: {
      url: "http://127.0.0.1:8545",
    },
    sepolia: {
      url: process.env.ALCHEMY_PROJECT_ID
        ? `https://eth-sepolia.g.alchemy.com/v2/${process.env.ALCHEMY_PROJECT_ID}`
        : "http://127.0.0.1:8545",
      gas: "auto",
      gasMultiplier: 2,
      gasPrice: "auto",
      accounts: [
        process.env.SEPOLIA_PK
          ? process.env.SEPOLIA_PK
          : "0x0000000000000000000000000000000000000000000000000000000000000000",
      ],
    },
    mainnet: {
      url: process.env.ALCHEMY_PROJECT_ID
        ? `https://eth-mainnet.g.alchemy.com/v2/${process.env.ALCHEMY_PROJECT_ID}`
        : "http://127.0.0.1:8545",
      gas: "auto",
      gasMultiplier: 2,
      gasPrice: "auto",
      accounts: [
        process.env.MAINNET_PK ||
          "0x0000000000000000000000000000000000000000000000000000000000000000",
      ],
    },
    alicenetTestnet: {
      url: "https://api.staging.alice.net",
      gas: "auto",
      gasMultiplier: 2,
      gasPrice: "auto",
      accounts: [
        process.env.ALICENET_PK ||
          "0x0000000000000000000000000000000000000000000000000000000000000001",
      ],
    },
    alicenet: {
      url: "https://api.prod.alice.net",
      gas: "auto",
      gasMultiplier: 2,
      gasPrice: "auto",
      accounts: [
        process.env.ALICENET_PK ||
          "0x0000000000000000000000000000000000000000000000000000000000000001",
      ],
    },
    hardhat: {
      chainId: 11155111,
      allowUnlimitedContractSize: false,
      mining: {
        auto: true,
        interval: 2200,
      },
    },
  },
  etherscan: {
    apiKey: {
      sepolia:
        process.env.ETHERSCAN_API_KEY || "0000000000000000000000000000000000",
      mainnet:
        process.env.ETHERSCAN_API_KEY || "0000000000000000000000000000000000",
      alicenet: "0000000000000000000000000000000000",
      alicenetTestnet: "0000000000000000000000000000000000",
    },
    customChains: [
      {
        network: "alicenetTestnet",
        chainId: 10042,
        urls: {
          apiURL: "https://block-explorer.staging.alice.net/api",
          browserURL: "https://block-explorer.staging.alice.net",
        },
      },
      {
        network: "alicenet",
        chainId: 1042,
        urls: {
          apiURL: "https://block-explorer.prod.alice.net/api",
          browserURL: "https://block-explorer.prod.alice.net",
        },
      },
    ],
  },
  solidity: {
    compilers: [
      {
        version: "0.8.26",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
        },
      },
      {
        version: "0.8.20",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
        },
      },
    ],
  },
  paths: {
    artifacts: "./artifacts",
    cache: "./cache",
    sources: "./contracts",
    tests: "./test",
  },
  contractSizer: {
    alphaSort: true,
    disambiguatePaths: false,
    runOnCompile: true,
    strict: true,
  },
  mocha: {
    timeout: 2200000,
    jobs: os.cpus().length / 2 > 1 ? os.cpus().length / 2 : 1,
  },
};

export default config;
