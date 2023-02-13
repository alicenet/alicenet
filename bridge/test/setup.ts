import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/dist/src/signer-with-address";
import {
  BigNumber,
  BigNumberish,
  Contract,
  ContractTransaction,
  Signer,
  Wallet,
} from "ethers";
import { isHexString } from "ethers/lib/utils";
import { ethers, network } from "hardhat";
import {
  deployCreateAndRegister,
  deployFactory,
} from "../scripts/lib/alicenetFactory";
import { deployUpgradeableProxyTask } from "../scripts/lib/deployment/tasks";
import {
  calculateSalt,
  extractFullContractInfoByContractName,
  populateConstructorArgs,
  populateInitializerArgs,
} from "../scripts/lib/deployment/utils";
import {
  ALCA,
  ALCABurner,
  ALCAMinter,
  ALCB,
  AliceNetFactory,
  Distribution,
  Dynamics,
  ETHDKG,
  Foundation,
  InvalidTxConsumptionAccusation,
  LegacyToken,
  LiquidityProviderStaking,
  MultipleProposalAccusation,
  PublicStaking,
  Snapshots,
  SnapshotsMock,
  StakingPositionDescriptor,
  ValidatorPool,
  ValidatorPoolMock,
  ValidatorStaking,
} from "../typechain-types";
import { ValidatorRawData } from "./ethdkg/setup";

export const PLACEHOLDER_ADDRESS = "0x0000000000000000000000000000000000000000";

export { assert, expect } from "./chai-setup";

export interface SignedBClaims {
  BClaims: string;
  GroupSignature: string;
}

export interface Snapshot {
  BClaims: string;
  GroupSignature: string;
  height: BigNumberish;
  validatorIndex: number;
  GroupSignatureDeserialized?: [
    [string, string, string, string],
    [string, string]
  ];
  BClaimsDeserialized?: [
    number,
    number,
    number,
    string,
    string,
    string,
    string
  ];
}

export interface BaseFixture {
  factory: AliceNetFactory;
  [key: string]: any;
}

export interface BaseTokensFixture extends BaseFixture {
  alca: ALCA;
  alcb: ALCB;
  legacyToken: LegacyToken;
  publicStaking: PublicStaking;
}

export interface Fixture extends BaseTokensFixture {
  alcaMinter: ALCAMinter;
  validatorStaking: ValidatorStaking;
  validatorPool: ValidatorPool | ValidatorPoolMock;
  snapshots: Snapshots | SnapshotsMock;
  ethdkg: ETHDKG;
  stakingPositionDescriptor: StakingPositionDescriptor;
  namedSigners: SignerWithAddress[];
  invalidTxConsumptionAccusation: InvalidTxConsumptionAccusation;
  multipleProposalAccusation: MultipleProposalAccusation;
  distribution: Distribution;
  dynamics: Dynamics;
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
  const hre = await require("hardhat");
  if (hre.__SOLIDITY_COVERAGE_RUNNING === true) {
    await network.provider.send("hardhat_setNextBlockBaseFeePerGas", ["0x1"]);
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
  const hre = await require("hardhat");
  const amount = hre.__SOLIDITY_COVERAGE_RUNNING === true ? "100000" : "10";
  const signers = await await ethers.getSigners();
  if (typeof validator === "string") {
    return ethers.getSigner(validator);
  } else {
    const balance = await ethers.provider.getBalance(validator.address);
    if (balance.eq(0)) {
      await signers[0].sendTransaction({
        to: validator.address,
        value: ethers.utils.parseEther(amount),
      });
    }
    if (typeof validator.privateKey !== "undefined") {
      return new Wallet(validator.privateKey, ethers.provider);
    }
    return ethers.getSigner(validator.address);
  }
};

export const createUsers = async (
  numberOfUsers: number,
  createWithNoFunds: boolean = false
): Promise<SignerWithAddress[]> => {
  const hre: any = await require("hardhat");
  const users: SignerWithAddress[] = [];
  const admin = (await ethers.getSigners())[0];
  for (let i = 0; i < numberOfUsers; i++) {
    const user = new Wallet(Wallet.createRandom(), ethers.provider);
    if (!createWithNoFunds) {
      const balance = await ethers.provider.getBalance(user.address);
      if (balance.eq(0)) {
        const value = hre.__SOLIDITY_COVERAGE_RUNNING ? "1000000" : "1";
        await admin.sendTransaction({
          to: user.address,
          value: ethers.utils.parseEther(value),
        });
      }
    }
    users.push(user as Signer as SignerWithAddress);
  }
  return users;
};

export async function getContractAddressFromDeployedStaticEvent(
  tx: ContractTransaction
): Promise<string> {
  const eventSignature = "event DeployedStatic(address contractAddr)";
  const eventName = "DeployedStatic";
  return await getContractAddressFromEventLog(tx, eventSignature, eventName);
}

export async function getContractAddressFromDeployedProxyEvent(
  tx: ContractTransaction
): Promise<string> {
  const eventSignature = "event DeployedProxy(address contractAddr)";
  const eventName = "DeployedProxy";
  return await getContractAddressFromEventLog(tx, eventSignature, eventName);
}

export async function getContractAddressFromDeployedRawEvent(
  tx: ContractTransaction
): Promise<string> {
  const eventSignature = "event DeployedRaw(address contractAddr)";
  const eventName = "DeployedRaw";
  return await getContractAddressFromEventLog(tx, eventSignature, eventName);
}

export async function getContractAddressFromEventLog(
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

export const deployUpgradeableWithFactory = async (
  factory: AliceNetFactory,
  contractName: string,
  salt?: string,
  initCallData?: any[],
  constructorArgs: any[] = [],
  saltType?: string
): Promise<Contract> => {
  process.env.silencer = "true";
  const hre: any = await require("hardhat");

  const contractData = await extractFullContractInfoByContractName(
    contractName,
    hre.artifacts,
    hre.ethers
  );

  initCallData !== undefined &&
    populateInitializerArgs(initCallData, contractData);
  constructorArgs !== undefined &&
    populateConstructorArgs(constructorArgs, contractData);

  const saltBytes =
    salt !== undefined && salt.startsWith("0x")
      ? salt
      : calculateSalt(
          salt === undefined ? contractName : salt,
          saltType,
          ethers
        );

  contractData.salt = saltBytes;

  const proxyData = await deployUpgradeableProxyTask(
    contractData,
    hre,
    0,
    factory,
    undefined,
    true
  );

  return await ethers.getContractAt(
    contractName,
    proxyData.proxyAddress as string
  );
};

export const deployFactoryAndBaseTokens =
  async (): Promise<BaseTokensFixture> => {
    // LegacyToken
    const legacyToken = await (
      await ethers.getContractFactory("LegacyToken")
    ).deploy();
    const factory = await deployAliceNetFactory(legacyToken.address);
    // ALCA is deployed on the factory constructor
    const alca = await ethers.getContractAt(
      "ALCA",
      await factory.lookup(ethers.utils.formatBytes32String("ALCA"))
    );

    const centralRouter = await (
      await ethers.getContractFactory("CentralBridgeRouterMock")
    ).deploy(1000);

    const alcbSalt = calculateSalt("ALCB", undefined, ethers);

    await deployCreateAndRegister(
      "ALCB",
      factory,
      ethers,
      [centralRouter.address],
      alcbSalt
    );
    // finally attach ALCB to the address of the deployed contract above
    const alcb = await ethers.getContractAt(
      "ALCB",
      await factory.lookup(alcbSalt)
    );

    // PublicStaking
    const publicStaking = (await deployUpgradeableWithFactory(
      factory,
      "PublicStaking",
      "PublicStaking",
      []
    )) as PublicStaking;

    return {
      factory,
      alca,
      alcb,
      legacyToken,
      publicStaking,
    };
  };

export const deployAliceNetFactory = async (
  legacyTokenAddress_: string
): Promise<AliceNetFactory> => {
  const hre = await require("hardhat");
  const Factory = await deployFactory(legacyTokenAddress_, hre.ethers);
  return await Factory.deployed();
};

export const preFixtureSetup = async () => {
  await network.provider.send("evm_setAutomine", [true]);
  // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
  // being sent as input to the function (the contract bytecode), so we need to increase the block
  // gas limit temporally in order to deploy the template
  const hre = await require("hardhat");
  if (hre.__SOLIDITY_COVERAGE_RUNNING !== true) {
    await network.provider.send("evm_setBlockGasLimit", ["0x3000000000000000"]);
  }
};

export const posFixtureSetup = async (factory: AliceNetFactory, alca: ALCA) => {
  // finish workaround, putting the blockgas limit to the previous value 30_000_000
  const hre = await require("hardhat");
  if (hre.__SOLIDITY_COVERAGE_RUNNING !== true) {
    await network.provider.send("evm_setBlockGasLimit", ["0x1C9C380"]);
  }
  await network.provider.send("hardhat_setNextBlockBaseFeePerGas", ["0x1"]);
  const [admin] = await ethers.getSigners();

  // transferring those ALCAs to the admin
  await factoryCallAny(factory, alca, "transfer", [
    admin.address,
    ethers.utils.parseEther("200000000"),
  ]);
};

export const getBaseTokensFixture = async (): Promise<BaseTokensFixture> => {
  await preFixtureSetup();

  const fixture = await deployFactoryAndBaseTokens();
  await posFixtureSetup(fixture.factory, fixture.alca);
  return fixture;
};

export const getFixture = async (
  mockValidatorPool?: boolean,
  mockSnapshots?: boolean,
  mockETHDKG?: boolean
): Promise<Fixture> => {
  await preFixtureSetup();
  const namedSigners = await ethers.getSigners();

  // Deploy the base tokens
  const { factory, alca, alcb, legacyToken, publicStaking } =
    await deployFactoryAndBaseTokens();
  // ValidatorStaking is not considered a base token since is only used by validators
  const validatorStaking = (await deployUpgradeableWithFactory(
    factory,
    "ValidatorStaking",
    "ValidatorStaking",
    []
  )) as ValidatorStaking;
  // LiquidityProviderStaking
  const liquidityProviderStaking = (await deployUpgradeableWithFactory(
    factory,
    "LiquidityProviderStaking",
    "LiquidityProviderStaking",
    []
  )) as LiquidityProviderStaking;
  // Foundation
  const foundation = (await deployUpgradeableWithFactory(
    factory,
    "Foundation",
    undefined
  )) as Foundation;
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
        8192,
      ]
    )) as ValidatorPool;
  }

  // ETHDKG Accusations
  await deployUpgradeableWithFactory(factory, "ETHDKGAccusations");

  // StakingPositionDescriptor
  const stakingPositionDescriptor = (await deployUpgradeableWithFactory(
    factory,
    "StakingPositionDescriptor"
  )) as StakingPositionDescriptor;

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

  const alcaMinter = (await deployUpgradeableWithFactory(
    factory,
    "ALCAMinter",
    "ALCAMinter"
  )) as ALCAMinter;

  // mint some alcas
  await factoryCallAny(factory, alcaMinter, "mint", [
    factory.address,
    ethers.utils.parseEther("100000000"),
  ]);

  const alcaBurner = (await deployUpgradeableWithFactory(
    factory,
    "ALCABurner",
    "ALCABurner"
  )) as ALCABurner;

  const invalidTxConsumptionAccusation = (await deployUpgradeableWithFactory(
    factory,
    "InvalidTxConsumptionAccusation",
    "InvalidTxConsumptionAccusation",
    undefined,
    undefined,
    "Accusation"
  )) as InvalidTxConsumptionAccusation;

  const multipleProposalAccusation = (await deployUpgradeableWithFactory(
    factory,
    "MultipleProposalAccusation",
    "MultipleProposalAccusation",
    undefined,
    undefined,
    "Accusation"
  )) as MultipleProposalAccusation;

  // distribution contract for distributing ALCBs yields
  const distribution = (await deployUpgradeableWithFactory(
    factory,
    "Distribution",
    undefined,
    undefined,
    [332, 332, 332, 4]
  )) as Distribution;

  const dynamics = (await deployUpgradeableWithFactory(
    factory,
    "Dynamics",
    "Dynamics",
    [4000]
  )) as Dynamics;

  await posFixtureSetup(factory, alca);
  const blockNumber = BigInt(await ethers.provider.getBlockNumber());
  const phaseLength = (await ethdkg.getPhaseLength()).toBigInt();
  if (phaseLength >= blockNumber) {
    await mineBlocks(phaseLength);
  }

  return {
    alca,
    alcb,
    legacyToken,
    publicStaking,
    validatorStaking,
    validatorPool,
    snapshots,
    ethdkg,
    factory,
    namedSigners,
    alcaMinter,
    alcaBurner,
    liquidityProviderStaking,
    foundation,
    stakingPositionDescriptor,
    invalidTxConsumptionAccusation,
    multipleProposalAccusation,
    distribution,
    dynamics,
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

export async function factoryCallAnyFixture(
  fixture: BaseFixture,
  contractName: string,
  functionName: string,
  args?: Array<any>
) {
  const factory = fixture.factory;
  const contract: Contract = fixture[contractName];
  return await factoryCallAny(factory, contract, functionName, args);
}

export async function factoryCallAny(
  factory: AliceNetFactory,
  contract: Contract,
  functionName: string,
  args?: Array<any>
) {
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

export async function callFunctionAndGetReturnValues(
  contract: Contract,
  functionName: any,
  account: SignerWithAddress,
  inputParameters: any[],
  messageValue?: BigNumber
): Promise<[any, ContractTransaction]> {
  try {
    let returnValues;
    let tx;
    if (messageValue !== undefined) {
      returnValues = await contract
        .connect(account)
        .callStatic[functionName](...inputParameters, { value: messageValue });
      tx = await contract
        .connect(account)
        [functionName](...inputParameters, { value: messageValue });
    } else {
      returnValues = await contract
        .connect(account)
        .callStatic[functionName](...inputParameters);
      tx = await contract.connect(account)[functionName](...inputParameters);
    }
    return [returnValues, tx];
  } catch (error) {
    throw new Error(
      `Couldn't call function '${functionName}' with account '${account.address}' and input parameters '${inputParameters}'\n${error}`
    );
  }
}

export const getMetamorphicAddress = (
  factoryAddress: string,
  salt: string
): string => {
  const initCode = "0x6020363636335afa1536363636515af43d36363e3d36f3";
  return ethers.utils.getCreate2Address(
    factoryAddress,
    ethers.utils.formatBytes32String(salt),
    ethers.utils.keccak256(initCode)
  );
};

export const getReceiptForFailedTransaction = async (
  tx: Promise<any>
): Promise<any> => {
  let receipt: any;
  try {
    await tx;
  } catch (error: any) {
    receipt = await ethers.provider.getTransactionReceipt(
      error.transactionHash
    );

    if (receipt === null) {
      throw new Error(`Transaction ${error.transactionHash} failed`);
    }
  }
  return receipt;
};

export const getBridgePoolSalt = (
  tokenContractAddr: string,
  tokenType: number,
  chainID: number,
  version: number
): string => {
  return ethers.utils.keccak256(
    ethers.utils.solidityPack(
      ["bytes32", "bytes32", "bytes32", "bytes32"],
      [
        ethers.utils.solidityKeccak256(["address"], [tokenContractAddr]),
        ethers.utils.solidityKeccak256(["uint8"], [tokenType]),
        ethers.utils.solidityKeccak256(["uint256"], [chainID]),
        ethers.utils.solidityKeccak256(["uint16"], [version]),
      ]
    )
  );
};

export const getStakingSVG = (
  shares: string,
  freeAfter: string,
  withdrawFreeAfter: string
) => {
  return `<svg width="500" height="645" viewBox="0 0 500 645" fill="none" xmlns="http://www.w3.org/2000/svg"> <mask id="mask0_8_62560" style="mask-type:alpha" maskUnits="userSpaceOnUse" x="0" y="0" width="500" height="440"> <rect width="500" height="440" fill="#280029"/> </mask> <g mask="url(#mask0_8_62560)"> <rect width="500" height="440" fill="#F2EAE9"/> <path d="M75.0879 194.011L85.4299 176L95.83 194.011H75.0879Z" fill="#F812C6"/> <path d="M87.4043 219.75L98.9664 199.647L110.587 219.75H87.4043Z" fill="#F812C6"/> <path d="M53.5303 231.487L73.9237 196.103L94.3172 231.487H53.5303Z" fill="#FA1A22"/> <path d="M112.329 272.332L126.68 247.406L141.031 272.332H112.329Z" fill="#F812C6"/> <path d="M98.9082 247.464L112.852 223.294L126.797 247.464H98.9082Z" fill="#F812C6"/> <path d="M70.7881 272.273L94.319 231.487L117.792 272.273H70.7881Z" fill="#FA1A22"/> <path d="M30 272.332L57.3075 224.979L84.615 272.332H30Z" fill="#FF9C00"/> <path d="M450.245 203.947L459.018 197.323V212.255H469.999V219.75H459.018V238.807C459.018 243.746 460.994 244.734 464.654 244.734C466.165 244.734 468.256 244.559 469.999 243.688V251.473C467.908 252.519 465.468 252.693 463.202 252.693C454.661 252.693 450.187 249.091 450.187 240.027V219.75H442.924V212.255H450.187V203.947H450.245ZM409.807 228.349C410.388 222.539 414.455 218.588 420.439 218.588C426.424 218.588 430.491 222.481 431.072 228.349H409.807ZM421.311 252.926C428.748 252.926 435.894 249.963 439.438 244.094L433.163 238.517C431.014 243.572 426.54 245.431 421.543 245.431C416.14 245.431 410.562 242.642 409.749 235.031H439.961C440.019 233.694 440.136 232.416 440.136 231.254C440.136 219.983 432.641 211.093 420.672 211.093C409.11 211.093 400.511 219.924 400.511 232.126C400.394 244.269 408.993 252.926 421.311 252.926ZM347.987 195.929H360.421L383.778 240.492V195.929H392.551V251.764H380.117L356.761 207.2V251.764H347.987V195.929ZM309.931 228.349C310.512 222.539 314.579 218.588 320.564 218.588C326.548 218.588 330.615 222.481 331.196 228.349H309.931ZM321.435 252.926C328.872 252.926 336.018 249.963 339.563 244.094L333.288 238.517C331.138 243.572 326.664 245.431 321.668 245.431C316.264 245.431 310.686 242.642 309.873 235.031H340.086C340.144 233.694 340.26 232.416 340.26 231.254C340.26 219.983 332.765 211.093 320.796 211.093C309.234 211.093 300.635 219.924 300.635 232.126C300.519 244.269 309.176 252.926 321.435 252.926ZM265.658 231.951C265.658 239.911 270.713 244.966 277.917 244.966C282.798 244.966 287.097 242.003 288.318 237.239L295.987 241.887C293.489 248.278 286.226 252.984 277.976 252.984C265.426 252.984 256.652 244.269 256.652 232.068C256.652 219.866 265.426 211.151 277.976 211.151C286.691 211.151 293.547 215.857 295.987 222.248L288.143 226.955C286.923 222.19 282.798 219.111 277.917 219.111C270.713 219.111 265.658 223.933 265.658 231.951ZM240.152 251.764H248.925V212.197H240.152V251.764ZM239.745 205.051H249.216V195.987H239.745V205.051ZM222.431 251.764H231.204V195.929H222.431V251.764ZM180.25 230.092L189.488 206.329L198.668 230.092H180.25ZM162.122 251.764H171.999L177.17 238.052H201.573L206.86 251.764H216.853L194.426 195.929H184.723L162.122 251.764Z" fill="#49004B"/> </g> <rect y="440" width="500" height="205" fill="#280029"/> <path d="M16.0201 619.964C14.2441 619.964 12.7441 620.984 12.7441 622.592C12.7441 626.084 18.0361 624.332 18.0361 626.576C18.0361 627.428 17.2201 627.944 16.1041 627.944C14.7121 627.944 13.8961 627.104 13.8961 625.796L12.5401 626.444C12.7681 628.184 14.2201 629.18 16.1041 629.18C18.0361 629.18 19.5001 628.208 19.5001 626.516C19.5001 622.964 14.1961 624.644 14.1961 622.496C14.1961 621.68 14.9521 621.2 16.0201 621.2C17.2921 621.2 18.0721 621.968 18.0721 623.036L19.3801 622.472C19.2121 621.104 17.9761 619.964 16.0201 619.964Z" fill="#F812C6"/> <path d="M24.2515 622.544C23.3035 622.544 22.6675 622.928 22.2835 623.576V620.18H20.9035V629H22.2835V625.676C22.2835 624.488 22.8835 623.816 23.8675 623.816C24.7915 623.816 25.2115 624.332 25.2115 625.46V629H26.5915V625.268C26.5915 623.516 25.7755 622.544 24.2515 622.544Z" fill="#F812C6"/> <path d="M27.8917 627.368C27.8917 628.484 28.7677 629.168 30.0157 629.168C30.8797 629.168 31.6597 628.856 32.1157 628.232L32.3797 629H33.6037L33.3877 627.764V624.704C33.3877 623.312 32.3677 622.544 30.7477 622.544C29.2957 622.544 28.1917 623.372 28.0117 624.548L29.2357 625.028C29.2357 624.14 29.8597 623.636 30.7477 623.636C31.4797 623.636 32.0437 623.912 32.0437 624.524C32.0437 625.1 31.7677 625.148 29.9677 625.46C28.7077 625.688 27.8917 626.3 27.8917 627.368ZM29.3437 627.296C29.3437 626.744 29.7637 626.504 30.3517 626.384C31.4917 626.168 31.8517 626.12 32.0437 625.952V626.492C32.0437 627.416 31.4557 628.124 30.3277 628.124C29.7037 628.124 29.3437 627.764 29.3437 627.296Z" fill="#F812C6"/> <path d="M35.0349 622.76V629H36.4149V625.988C36.4149 624.608 37.0749 623.96 38.0349 623.96C38.2989 623.96 38.5029 623.996 38.7429 624.068V622.7C38.5509 622.628 38.4069 622.616 38.1909 622.616C37.6149 622.616 36.8709 622.844 36.4149 623.744V622.76H35.0349Z" fill="#F812C6"/> <path d="M42.3886 622.544C40.5766 622.544 39.2566 623.96 39.2566 625.892C39.2566 627.824 40.5046 629.18 42.4726 629.18C43.6246 629.18 44.6926 628.736 45.2806 627.788L44.2846 626.9C43.9486 627.668 43.3846 628.016 42.5086 628.016C41.5606 628.016 40.8526 627.5 40.7086 626.312H45.3406C45.3646 626.096 45.3766 625.94 45.3766 625.76C45.3766 623.852 44.2246 622.544 42.3886 622.544ZM40.7206 625.304C40.8886 624.224 41.5486 623.708 42.3646 623.708C43.3246 623.708 43.8886 624.344 43.8886 625.304H40.7206Z" fill="#F812C6"/> <path d="M49.0582 622.544C47.6662 622.544 46.4902 623.372 46.4902 624.524C46.4902 627.032 50.4022 625.988 50.4022 627.308C50.4022 627.824 49.8622 628.076 49.2022 628.076C48.3382 628.076 47.6782 627.668 47.6302 626.744L46.2982 627.308C46.5142 628.472 47.6062 629.18 49.2022 629.18C50.7262 629.18 51.8782 628.532 51.8782 627.224C51.8782 624.74 47.9062 625.724 47.9062 624.44C47.9062 623.972 48.3262 623.636 49.0582 623.636C49.9222 623.636 50.4502 624.08 50.4982 624.836L51.8062 624.344C51.6022 623.252 50.5582 622.544 49.0582 622.544Z" fill="#F812C6"/> <path d="M12.1201 590.18L14.4841 599H16.1881L18.0241 592.436L19.8361 599H21.5401L23.9041 590.18H22.3801L20.6281 596.996L18.7681 590.3H17.2681L15.4081 596.996L13.6441 590.18H12.1201Z" fill="#F812C6"/> <path d="M24.8396 592.76V599H26.2196V592.76H24.8396ZM24.7796 591.632H26.2676V590.18H24.7796V591.632Z" fill="#F812C6"/> <path d="M28.4393 591.44V592.76H27.3233V593.948H28.4393V597.164C28.4393 598.592 29.1833 599.156 30.5393 599.156C30.7433 599.156 31.2953 599.108 31.4993 599.024V597.812C31.2953 597.86 30.9353 597.908 30.7433 597.908C30.1193 597.908 29.8193 597.752 29.8193 596.984V593.948H31.5953V592.76H29.8193V590.408L28.4393 591.44Z" fill="#F812C6"/> <path d="M36.2983 592.544C35.3503 592.544 34.7143 592.928 34.3303 593.576V590.18H32.9503V599H34.3303V595.676C34.3303 594.488 34.9303 593.816 35.9143 593.816C36.8383 593.816 37.2583 594.332 37.2583 595.46V599H38.6383V595.268C38.6383 593.516 37.8223 592.544 36.2983 592.544Z" fill="#F812C6"/> <path d="M46.5506 590.18H45.1706V593.516C44.6786 592.928 43.9586 592.544 42.9986 592.544C41.1146 592.544 39.8426 593.996 39.8426 595.88C39.8426 597.752 41.1266 599.18 42.9986 599.18C43.9706 599.18 44.6906 598.796 45.1826 598.208V599H46.5506V590.18ZM41.2706 595.88C41.2706 594.668 42.0746 593.792 43.2506 593.792C44.3786 593.792 45.2186 594.656 45.2186 595.88C45.2186 597.092 44.3786 597.92 43.2506 597.92C42.0746 597.92 41.2706 597.068 41.2706 595.88Z" fill="#F812C6"/> <path d="M48.2419 592.76V599H49.6219V595.988C49.6219 594.608 50.2819 593.96 51.2419 593.96C51.5059 593.96 51.7099 593.996 51.9499 594.068V592.7C51.7579 592.628 51.6139 592.616 51.3979 592.616C50.8219 592.616 50.0779 592.844 49.6219 593.744V592.76H48.2419Z" fill="#F812C6"/> <path d="M52.6886 597.368C52.6886 598.484 53.5646 599.168 54.8126 599.168C55.6766 599.168 56.4566 598.856 56.9126 598.232L57.1766 599H58.4006L58.1846 597.764V594.704C58.1846 593.312 57.1646 592.544 55.5446 592.544C54.0926 592.544 52.9886 593.372 52.8086 594.548L54.0326 595.028C54.0326 594.14 54.6566 593.636 55.5446 593.636C56.2766 593.636 56.8406 593.912 56.8406 594.524C56.8406 595.1 56.5646 595.148 54.7646 595.46C53.5046 595.688 52.6886 596.3 52.6886 597.368ZM54.1406 597.296C54.1406 596.744 54.5606 596.504 55.1486 596.384C56.2886 596.168 56.6486 596.12 56.8406 595.952V596.492C56.8406 597.416 56.2526 598.124 55.1246 598.124C54.5006 598.124 54.1406 597.764 54.1406 597.296Z" fill="#F812C6"/> <path d="M59.0666 592.76L61.0706 599H62.6426L63.9506 594.764L65.2946 599H66.8546L68.8706 592.76H67.4306L66.0266 597.368L64.5986 592.808H63.3146L61.8866 597.368L60.5306 592.76H59.0666Z" fill="#F812C6"/> <path d="M71.8497 599H73.3977L74.2377 596.852H78.1017L78.9177 599H80.4897L76.9497 590.18H75.4137L71.8497 599ZM74.7297 595.58L76.1817 591.848L77.6097 595.58H74.7297Z" fill="#F812C6"/> <path d="M83.9985 589.964C82.7745 589.964 81.9825 590.852 81.9825 592.088V592.76H80.8545V593.936H81.9825V599H83.3625V593.936H85.1625V592.76H83.3625V592.124C83.3625 591.464 83.6985 591.176 84.2745 591.176C84.5145 591.176 84.7785 591.2 85.0065 591.248V590.072C84.6945 589.988 84.3345 589.964 83.9985 589.964Z" fill="#F812C6"/> <path d="M86.951 591.44V592.76H85.835V593.948H86.951V597.164C86.951 598.592 87.695 599.156 89.051 599.156C89.255 599.156 89.807 599.108 90.011 599.024V597.812C89.807 597.86 89.447 597.908 89.255 597.908C88.631 597.908 88.331 597.752 88.331 596.984V593.948H90.107V592.76H88.331V590.408L86.951 591.44Z" fill="#F812C6"/> <path d="M94.1269 592.544C92.3149 592.544 90.9949 593.96 90.9949 595.892C90.9949 597.824 92.2429 599.18 94.2109 599.18C95.3629 599.18 96.4309 598.736 97.0189 597.788L96.0229 596.9C95.6869 597.668 95.1229 598.016 94.2469 598.016C93.2989 598.016 92.5909 597.5 92.4469 596.312H97.0789C97.1029 596.096 97.1149 595.94 97.1149 595.76C97.1149 593.852 95.9629 592.544 94.1269 592.544ZM92.4589 595.304C92.6269 594.224 93.2869 593.708 94.1029 593.708C95.0629 593.708 95.6269 594.344 95.6269 595.304H92.4589Z" fill="#F812C6"/> <path d="M98.4685 592.76V599H99.8485V595.988C99.8485 594.608 100.508 593.96 101.468 593.96C101.732 593.96 101.936 593.996 102.176 594.068V592.7C101.984 592.628 101.84 592.616 101.624 592.616C101.048 592.616 100.304 592.844 99.8485 593.744V592.76H98.4685Z" fill="#F812C6"/> <path d="M13.0201 569H14.4481V565.328H18.4561V564.056H14.4481V561.476H18.8881V560.18H13.0201V569Z" fill="#F812C6"/> <path d="M20.2341 562.76V569H21.6141V565.988C21.6141 564.608 22.2741 563.96 23.2341 563.96C23.4981 563.96 23.7021 563.996 23.9421 564.068V562.7C23.7501 562.628 23.6061 562.616 23.3901 562.616C22.8141 562.616 22.0701 562.844 21.6141 563.744V562.76H20.2341Z" fill="#F812C6"/> <path d="M27.5878 562.544C25.7758 562.544 24.4558 563.96 24.4558 565.892C24.4558 567.824 25.7038 569.18 27.6718 569.18C28.8238 569.18 29.8918 568.736 30.4798 567.788L29.4838 566.9C29.1478 567.668 28.5838 568.016 27.7078 568.016C26.7598 568.016 26.0518 567.5 25.9078 566.312H30.5398C30.5638 566.096 30.5758 565.94 30.5758 565.76C30.5758 563.852 29.4238 562.544 27.5878 562.544ZM25.9198 565.304C26.0878 564.224 26.7478 563.708 27.5638 563.708C28.5238 563.708 29.0878 564.344 29.0878 565.304H25.9198Z" fill="#F812C6"/> <path d="M34.6894 562.544C32.8774 562.544 31.5574 563.96 31.5574 565.892C31.5574 567.824 32.8054 569.18 34.7734 569.18C35.9254 569.18 36.9934 568.736 37.5814 567.788L36.5854 566.9C36.2494 567.668 35.6854 568.016 34.8094 568.016C33.8614 568.016 33.1534 567.5 33.0094 566.312H37.6414C37.6654 566.096 37.6774 565.94 37.6774 565.76C37.6774 563.852 36.5254 562.544 34.6894 562.544ZM33.0214 565.304C33.1894 564.224 33.8494 563.708 34.6654 563.708C35.6254 563.708 36.1894 564.344 36.1894 565.304H33.0214Z" fill="#F812C6"/> <path d="M41.0294 569H42.5774L43.4174 566.852H47.2814L48.0974 569H49.6694L46.1294 560.18H44.5934L41.0294 569ZM43.9094 565.58L45.3614 561.848L46.7894 565.58H43.9094Z" fill="#F812C6"/> <path d="M53.1782 559.964C51.9542 559.964 51.1622 560.852 51.1622 562.088V562.76H50.0342V563.936H51.1622V569H52.5422V563.936H54.3422V562.76H52.5422V562.124C52.5422 561.464 52.8782 561.176 53.4542 561.176C53.6942 561.176 53.9582 561.2 54.1862 561.248V560.072C53.8742 559.988 53.5142 559.964 53.1782 559.964Z" fill="#F812C6"/> <path d="M56.1307 561.44V562.76H55.0147V563.948H56.1307V567.164C56.1307 568.592 56.8747 569.156 58.2307 569.156C58.4347 569.156 58.9867 569.108 59.1907 569.024V567.812C58.9867 567.86 58.6267 567.908 58.4347 567.908C57.8107 567.908 57.5107 567.752 57.5107 566.984V563.948H59.2867V562.76H57.5107V560.408L56.1307 561.44Z" fill="#F812C6"/> <path d="M63.3066 562.544C61.4946 562.544 60.1746 563.96 60.1746 565.892C60.1746 567.824 61.4226 569.18 63.3906 569.18C64.5426 569.18 65.6106 568.736 66.1986 567.788L65.2026 566.9C64.8666 567.668 64.3026 568.016 63.4266 568.016C62.4786 568.016 61.7706 567.5 61.6266 566.312H66.2586C66.2826 566.096 66.2946 565.94 66.2946 565.76C66.2946 563.852 65.1426 562.544 63.3066 562.544ZM61.6386 565.304C61.8066 564.224 62.4666 563.708 63.2826 563.708C64.2426 563.708 64.8066 564.344 64.8066 565.304H61.6386Z" fill="#F812C6"/> <path d="M67.6481 562.76V569H69.0281V565.988C69.0281 564.608 69.6881 563.96 70.6481 563.96C70.9121 563.96 71.1161 563.996 71.3561 564.068V562.7C71.1641 562.628 71.0201 562.616 70.8041 562.616C70.2281 562.616 69.4841 562.844 69.0281 563.744V562.76H67.6481Z" fill="#F812C6"/> <text fill="#F812C6" xml:space="preserve" style="white-space: pre" font-family="Arial" font-size="13" font-weight="bold" letter-spacing="0em"><tspan x="294" y="628.007">${shares}</tspan></text> <text fill="#F812C6" xml:space="preserve" style="white-space: pre" font-family="Arial" font-size="13" font-weight="bold" letter-spacing="0em"><tspan x="432.109" y="598.007">${withdrawFreeAfter}</tspan></text> <text fill="#F812C6" xml:space="preserve" style="white-space: pre" font-family="Arial" font-size="13" font-weight="bold" letter-spacing="0em"><tspan x="432.109" y="568.007">${freeAfter}</tspan></text> <line x1="12" y1="608.5" x2="488" y2="608.5" stroke="#F2EAE9" stroke-opacity="0.06"/> <line x1="12" y1="578.5" x2="488" y2="578.5" stroke="#F2EAE9" stroke-opacity="0.06"/> </svg>`;
};

export const getStakingSVGBase64 = (
  shares: string,
  freeAfter: string,
  withdrawFreeAfter: string
) => {
  const svg = getStakingSVG(shares, freeAfter, withdrawFreeAfter);
  return Buffer.from(svg).toString("base64");
};
