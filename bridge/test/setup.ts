import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/dist/src/signer-with-address";
import {
  BigNumber,
  BigNumberish,
  BytesLike,
  Contract,
  ContractTransaction,
  Signer,
  Wallet,
} from "ethers";
import { isHexString } from "ethers/lib/utils";
import { ethers, network } from "hardhat";
import {
  AToken,
  ATokenBurnerMock,
  ATokenMinterMock,
  ETHDKG,
  MadByte,
  MadnetFactory,
  MadToken,
  Snapshots,
  SnapshotsMock,
  StakeNFT,
  ValidatorNFT,
  ValidatorPool,
  ValidatorPoolMock,
} from "../typechain-types";
import { ValidatorRawData } from "./ethdkg/setup";

export const PLACEHOLDER_ADDRESS = "0x0000000000000000000000000000000000000000";

export { assert, expect } from "./chai-setup";

export interface Snapshot {
  BClaims: string;
  GroupSignature: string;
  height: BigNumberish;
  validatorIndex: number;
}

export interface Fixture {
  madToken: MadToken;
  madByte: MadByte;
  stakeNFT: StakeNFT;
  validatorNFT: ValidatorNFT;
  validatorPool: ValidatorPool | ValidatorPoolMock;
  snapshots: Snapshots | SnapshotsMock;
  ethdkg: ETHDKG;
  factory: MadnetFactory;
  namedSigners: SignerWithAddress[];
  [key: string]: any;
}

/**
 * Shuffles array in place. ES6 version
 * https://stackoverflow.com/questions/6274339/how-can-i-shuffle-an-array/6274381#6274381
 * @param {Array} a items An array containing the items.
 */
export function shuffle(a: ValidatorRawData[]) {
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [a[i], a[j]] = [a[j], a[i]];
  }
  return a;
}

export const mineBlocks = async (nBlocks: bigint) => {
  if (nBlocks > BigInt(0)) {
    await network.provider.send("hardhat_mine", [
      ethers.utils.hexValue(nBlocks),
    ]);
  }
};

export const getBlockByNumber = async () => {
  return await network.provider.send("eth_getBlockByNumber", [
    "pending",
    false,
  ]);
};

export const getPendingTransactions = async () => {
  return await network.provider.send("eth_pendingTransactions");
};

export const getValidatorEthAccount = async (
  validator: ValidatorRawData | string
): Promise<Signer> => {
  if (typeof validator === "string") {
    return ethers.getSigner(validator);
  } else {
    const balance = await ethers.provider.getBalance(validator.address);
    if (balance.eq(0)) {
      await (
        await ethers.getSigner("0x546F99F244b7B58B855330AE0E2BC1b30b41302F")
      ).sendTransaction({
        to: validator.address,
        value: ethers.utils.parseEther("10"),
      });
    }
    if (typeof validator.privateKey !== "undefined") {
      return new Wallet(validator.privateKey, ethers.provider);
    }
    return ethers.getSigner(validator.address);
  }
};

async function getContractAddressFromDeployedStaticEvent(
  tx: ContractTransaction
): Promise<string> {
  const eventSignature = "event DeployedStatic(address contractAddr)";
  const eventName = "DeployedStatic";
  return await getContractAddressFromEventLog(tx, eventSignature, eventName);
}

async function getContractAddressFromDeployedProxyEvent(
  tx: ContractTransaction
): Promise<string> {
  const eventSignature = "event DeployedProxy(address contractAddr)";
  const eventName = "DeployedProxy";
  return await getContractAddressFromEventLog(tx, eventSignature, eventName);
}

async function getContractAddressFromDeployedRawEvent(
  tx: ContractTransaction
): Promise<string> {
  const eventSignature = "event DeployedRaw(address contractAddr)";
  const eventName = "DeployedRaw";
  return await getContractAddressFromEventLog(tx, eventSignature, eventName);
}

async function getContractAddressFromEventLog(
  tx: ContractTransaction,
  eventSignature: string,
  eventName: string
): Promise<string> {
  const receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  const intrface = new ethers.utils.Interface([eventSignature]);
  let result = "";
  for (const log of receipt.logs) {
    const topics = log.topics;
    const data = log.data;
    const topicHash = intrface.getEventTopic(intrface.getEvent(eventName));
    if (!isHexString(topics[0], 32) || topics[0].toLowerCase() !== topicHash) {
      continue;
    }
    result = intrface.decodeEventLog(eventName, data, topics).contractAddr;
  }
  if (result === "") {
    throw new Error(
      "Couldn't parse logs in the transaction!\nReceipt:\n" + receipt
    );
  }
  return result;
}

function getBytes32Salt(contractName: string) {
  return ethers.utils.formatBytes32String(contractName);
}

export const deployStaticWithFactory = async (
  factory: MadnetFactory,
  contractName: string,
  initCallData?: any[],
  constructorArgs?: any[]
): Promise<Contract> => {
  const _Contract = await ethers.getContractFactory(contractName);
  let contractTx;
  if (constructorArgs !== undefined) {
    contractTx = await factory.deployTemplate(
      _Contract.getDeployTransaction(...constructorArgs).data as BytesLike
    );
  } else {
    contractTx = await factory.deployTemplate(
      _Contract.getDeployTransaction().data as BytesLike
    );
  }
  let receipt = await ethers.provider.getTransactionReceipt(contractTx.hash);
  if (receipt.gasUsed.gt(10_000_000)) {
    throw new Error(
      `Contract deployment size:${receipt.gasUsed} is greater than 10 million`
    );
  }

  let initCallDataBin;
  try {
    initCallDataBin = _Contract.interface.encodeFunctionData(
      "initialize",
      initCallData
    );
  } catch (error) {
    console.log(
      `Warning couldnt get init call data for contract: ${contractName}`
    );
    console.log(error);
    initCallDataBin = "0x";
  }
  const tx = await factory.deployStatic(
    getBytes32Salt(contractName),
    initCallDataBin
  );
  receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  if (receipt.gasUsed.gt(10_000_000)) {
    throw new Error(
      `Contract deployment size:${receipt.gasUsed} is greater than 10 million`
    );
  }

  return _Contract.attach(await getContractAddressFromDeployedStaticEvent(tx));
};

async function deployUpgradeableWithFactory(
  factory: MadnetFactory,
  contractName: string,
  salt?: string,
  initCallData?: any[],
  constructorArgs?: any[]
): Promise<Contract> {
  const _Contract = await ethers.getContractFactory(contractName);
  let deployCode: BytesLike;
  let contractTx;
  if (constructorArgs !== undefined) {
    contractTx = await factory.deployTemplate(
      (deployCode = _Contract.getDeployTransaction(...constructorArgs)
        .data as BytesLike)
    );
  } else {
    contractTx = await factory.deployTemplate(
      (deployCode = _Contract.getDeployTransaction().data as BytesLike)
    );
  }
  let receipt = await ethers.provider.getTransactionReceipt(contractTx.hash);
  if (receipt.gasUsed.gt(10_000_000)) {
    throw new Error(
      `Contract deployment size:${receipt.gasUsed} is greater than 10 million`
    );
  }
  const transaction = await factory.deployCreate(deployCode);
  receipt = await ethers.provider.getTransactionReceipt(transaction.hash);
  if (receipt.gasUsed.gt(10_000_000)) {
    throw new Error(
      `Contract deployment size:${receipt.gasUsed} is greater than 10 million`
    );
  }
  const logicAddr = await getContractAddressFromDeployedRawEvent(transaction);
  let saltBytes;
  if (salt === undefined) {
    saltBytes = getBytes32Salt(contractName);
  } else {
    saltBytes = getBytes32Salt(salt);
  }

  const transaction2 = await factory.deployProxy(saltBytes);
  receipt = await ethers.provider.getTransactionReceipt(transaction2.hash);
  if (receipt.gasUsed.gt(10_000_000)) {
    throw new Error(
      `Contract deployment size:${receipt.gasUsed} is greater than 10 million`
    );
  }

  let initCallDataBin = "0x";
  if (initCallData !== undefined) {
    try {
      initCallDataBin = _Contract.interface.encodeFunctionData(
        "initialize",
        initCallData
      );
    } catch (error) {
      console.warn(
        `Error deploying contract ${contractName} couldn't get initialize arguments: ${error}`
      );
    }
  }
  await factory.upgradeProxy(saltBytes, logicAddr, initCallDataBin);
  return _Contract.attach(
    await getContractAddressFromDeployedProxyEvent(transaction2)
  );
}

export const getFixture = async (
  mockValidatorPool?: boolean,
  mockSnapshots?: boolean,
  mockETHDKG?: boolean
): Promise<Fixture> => {
  await network.provider.send("evm_setAutomine", [true]);
  // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
  // being sent as input to the function (the contract bytecode), so we need to increase the block
  // gas limit temporally in order to deploy the template
  await network.provider.send("evm_setBlockGasLimit", ["0x3000000000000000"]);

  const namedSigners = await ethers.getSigners();
  const [admin] = namedSigners;

  const txCount = await ethers.provider.getTransactionCount(admin.address);
  // calculate the factory address for the constructor arg
  const futureFactoryAddress = ethers.utils.getContractAddress({
    from: admin.address,
    nonce: txCount,
  });

  const Factory = await ethers.getContractFactory("MadnetFactory");
  const factory = await Factory.deploy(futureFactoryAddress);
  await factory.deployed();

  // MadToken
  const madToken = (await deployStaticWithFactory(
    factory,
    "MadToken"
  )) as MadToken;

  // MadByte
  const madByte = (await deployStaticWithFactory(
    factory,
    "MadByte"
  )) as MadByte;

  // StakeNFT
  const stakeNFT = (await deployStaticWithFactory(
    factory,
    "StakeNFT"
  )) as StakeNFT;

  // ValidatorNFT
  const validatorNFT = (await deployStaticWithFactory(
    factory,
    "ValidatorNFT"
  )) as ValidatorNFT;

  let validatorPool;
  if (typeof mockValidatorPool !== "undefined" && mockValidatorPool) {
    // ValidatorPoolMock
    validatorPool = (await deployUpgradeableWithFactory(
      factory,
      "ValidatorPoolMock",
      "ValidatorPool"
    )) as ValidatorPoolMock;
  } else {
    // ValidatorPool
    validatorPool = (await deployUpgradeableWithFactory(
      factory,
      "ValidatorPool",
      "ValidatorPool",
      [
        ethers.utils.parseUnits("20000", 18),
        10,
        ethers.utils.parseUnits("3", 18),
      ]
    )) as ValidatorPool;
  }

  // ETHDKG Accusations
  await deployUpgradeableWithFactory(factory, "ETHDKGAccusations");

  // ETHDKG Phases
  await deployUpgradeableWithFactory(factory, "ETHDKGPhases");

  // ETHDKG
  let ethdkg;
  if (typeof mockETHDKG !== "undefined" && mockETHDKG) {
    // ValidatorPoolMock
    ethdkg = (await deployUpgradeableWithFactory(
      factory,
      "ETHDKGMock",
      "ETHDKG",
      [BigNumber.from(40), BigNumber.from(6)]
    )) as ETHDKG;
  } else {
    // ValidatorPool
    ethdkg = (await deployUpgradeableWithFactory(factory, "ETHDKG", "ETHDKG", [
      BigNumber.from(40),
      BigNumber.from(6),
    ])) as ETHDKG;
  }

  let snapshots;
  if (typeof mockSnapshots !== "undefined" && mockSnapshots) {
    // Snapshots Mock
    snapshots = (await deployUpgradeableWithFactory(
      factory,
      "SnapshotsMock",
      "Snapshots",
      [10, 40],
      [1, 1]
    )) as Snapshots;
  } else {
    // Snapshots
    snapshots = (await deployUpgradeableWithFactory(
      factory,
      "Snapshots",
      "Snapshots",
      [10, 40],
      [1, 1024]
    )) as Snapshots;
  }

  const aToken = (await deployStaticWithFactory(factory, "AToken", undefined, [
    madToken.address,
  ])) as AToken;
  const aTokenMinter = (await deployUpgradeableWithFactory(
    factory,
    "ATokenMinterMock",
    "ATokenMinter"
  )) as ATokenMinterMock;
  const aTokenBurner = (await deployUpgradeableWithFactory(
    factory,
    "ATokenBurnerMock",
    "ATokenBurner"
  )) as ATokenBurnerMock;

  // finish workaround, putting the blockgas limit to the previous value 30_000_000
  await network.provider.send("evm_setBlockGasLimit", ["0x1C9C380"]);
  await factory.callAny(
    madToken.address,
    0,
    madToken.interface.encodeFunctionData("transfer", [
      admin.address,
      ethers.utils.parseEther("100000000"),
    ])
  );
  const blockNumber = BigInt(await ethers.provider.getBlockNumber());
  const phaseLength = (await ethdkg.getPhaseLength()).toBigInt();
  if (phaseLength >= blockNumber) {
    await mineBlocks(phaseLength);
  }

  return {
    madToken,
    madByte,
    stakeNFT,
    validatorNFT,
    validatorPool,
    snapshots,
    ethdkg,
    factory,
    namedSigners,
    aToken,
    aTokenMinter,
    aTokenBurner,
  };
};

export async function getTokenIdFromTx(tx: any) {
  const abi = [
    "event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)",
  ];
  const iface = new ethers.utils.Interface(abi);
  const receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  const logs =
    typeof receipt.logs[2] !== "undefined" ? receipt.logs[2] : receipt.logs[0];
  const log = iface.parseLog(logs);
  return log.args[2];
}

export async function factoryCallAny(
  fixture: Fixture,
  contractName: string,
  functionName: string,
  args?: Array<any>
) {
  const factory = fixture.factory;
  const contract = fixture[contractName];
  if (args === undefined) {
    args = [];
  }
  const txResponse = await factory.callAny(
    contract.address,
    0,
    contract.interface.encodeFunctionData(functionName, args)
  );
  const receipt = await txResponse.wait();
  return receipt;
}
