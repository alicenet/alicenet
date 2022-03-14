import { BigNumber, ContractTransaction } from "ethers";
import { ethers } from "hardhat";
import { ValidatorRawData } from "../ethdkg/setup";
import {
  factoryCallAny,
  Fixture,
  getTokenIdFromTx,
  getValidatorEthAccount,
} from "../setup";

interface Contract {
  StakeNFT: bigint;
  ValNFT: bigint;
  MAD: bigint;
  ETH: bigint;
  Addr: string;
}

interface Admin {
  StakeNFT: bigint;
  ValNFT: bigint;
  MAD: bigint;
  Addr: string;
}

interface Validator {
  NFT: bigint;
  MAD: bigint;
  Addr: string;
  Reg: boolean;
  ExQ: boolean;
  Acc: boolean;
  Idx: number;
}
interface State {
  Admin: Admin;
  StakeNFT: Contract;
  ValidatorNFT: Contract;
  ValidatorPool: Contract;
  Factory: Contract;
  validators: Array<Validator>;
}

export const commitSnapshots = async (
  fixture: Fixture,
  numSnapshots: number
) => {
  for (let i = 0; i < numSnapshots; i++) {
    await fixture.snapshots.snapshot("0x00", "0x00");
  }
};

export const getCurrentState = async (
  fixture: Fixture,
  _validators: string[]
): Promise<State> => {
  // System state
  let state: State = {
    Admin: {
      StakeNFT: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      Addr: "0x0",
    },
    StakeNFT: {
      StakeNFT: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      ETH: BigInt(0),
      Addr: "0x0",
    },
    ValidatorNFT: {
      StakeNFT: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      ETH: BigInt(0),
      Addr: "0x0",
    },
    ValidatorPool: {
      StakeNFT: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      ETH: BigInt(0),
      Addr: "0x0",
    },
    Factory: {
      StakeNFT: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      ETH: BigInt(0),
      Addr: "0x0",
    },
    validators: [],
  };
  let [adminSigner] = await ethers.getSigners();
  // Get state for admin
  state.Admin.StakeNFT = (
    await fixture.stakeNFT.balanceOf(adminSigner.address)
  ).toBigInt();
  state.Admin.ValNFT = (
    await fixture.validatorNFT.balanceOf(adminSigner.address)
  ).toBigInt();
  state.Admin.MAD = (
    await fixture.madToken.balanceOf(adminSigner.address)
  ).toBigInt();
  state.Admin.Addr = adminSigner.address;

  // Get state for validators
  for (let i = 0; i < _validators.length; i++) {
    let validator: Validator = {
      Idx: i,
      NFT: (await fixture.stakeNFT.balanceOf(_validators[i])).toBigInt(),
      MAD: (await fixture.madToken.balanceOf(_validators[i])).toBigInt(),
      Addr: _validators[i],
      Reg: await fixture.validatorPool.isValidator(_validators[i]),
      ExQ: await fixture.validatorPool.isInExitingQueue(_validators[i]),
      Acc: await fixture.validatorPool.isAccusable(_validators[i]),
    };
    state.validators.push(validator);
  }
  // Contract data
  let contractData = [
    {
      contractState: state.StakeNFT,
      contractAddress: fixture.stakeNFT.address,
    },
    {
      contractState: state.ValidatorNFT,
      contractAddress: fixture.validatorNFT.address,
    },
    {
      contractState: state.ValidatorPool,
      contractAddress: fixture.validatorPool.address,
    },
    {
      contractState: state.Factory,
      contractAddress: fixture.factory.address,
    },
  ];
  // Get state for contracts
  for (let i = 0; i < contractData.length; i++) {
    contractData[i].contractState.StakeNFT = (
      await fixture.stakeNFT.balanceOf(contractData[i].contractAddress)
    ).toBigInt();
    contractData[i].contractState.ValNFT = (
      await fixture.validatorNFT.balanceOf(contractData[i].contractAddress)
    ).toBigInt();
    contractData[i].contractState.MAD = (
      await fixture.madToken.balanceOf(contractData[i].contractAddress)
    ).toBigInt();
    contractData[i].contractState.ETH = (
      await ethers.provider.getBalance(contractData[i].contractAddress)
    ).toBigInt();
    contractData[i].contractState.Addr = contractData[i].contractAddress;
  }
  return state;
};

export const showState = async (title: string, state: State): Promise<void> => {
  if (process.env.npm_config_detailed == "true") {
    // execute "npm --detailed=true  run test" to see this output
    console.log(title);
    console.log(state);
  }
};

export const createValidators = async (
  fixture: Fixture,
  _validatorsSnapshots: ValidatorRawData[]
): Promise<string[]> => {
  let validators: string[] = [];
  let stakeAmountMadWei = await fixture.validatorPool.getStakeAmount();
  let [adminSigner] = await ethers.getSigners();
  // Approve ValidatorPool to withdraw MAD tokens of validators
  await fixture.madToken.approve(
    fixture.validatorPool.address,
    stakeAmountMadWei.mul(_validatorsSnapshots.length)
  );
  for (let i = 0; i < _validatorsSnapshots.length; i++) {
    let validator = _validatorsSnapshots[i];
    await getValidatorEthAccount(validator);
    validators.push(validator.address);
    // Send MAD tokens to each validator
    await fixture.madToken.transfer(validator.address, stakeAmountMadWei);
  }
  await fixture.madToken
    .connect(adminSigner)
    .approve(
      fixture.stakeNFT.address,
      stakeAmountMadWei.mul(_validatorsSnapshots.length)
    );
  await showState(
    "After creating:",
    await getCurrentState(fixture, validators)
  );
  return validators;
};

export const stakeValidators = async (
  fixture: Fixture,
  validators: string[]
): Promise<BigNumber[]> => {
  let stakingTokenIds: BigNumber[] = [];
  let [adminSigner] = await ethers.getSigners();
  let stakeAmountMadWei = await fixture.validatorPool.getStakeAmount();
  let lockTime = 1;
  for (let i = 0; i < validators.length; i++) {
    // Stake all MAD tokens
    let tx = await fixture.stakeNFT
      .connect(adminSigner)
      .mintTo(fixture.factory.address, stakeAmountMadWei, lockTime);
    // Get the proof of staking (NFT's tokenID)
    let tokenID = await getTokenIdFromTx(tx);
    stakingTokenIds.push(tokenID);
    await factoryCallAny(fixture, "stakeNFT", "approve", [
      fixture.validatorPool.address,
      tokenID,
    ]);
  }
  await showState("After staking:", await getCurrentState(fixture, validators));
  return stakingTokenIds;
};

export const claimPosition = async (
  fixture: Fixture,
  validator: ValidatorRawData
): Promise<BigNumber> => {
  let claimTx = (await fixture.validatorPool
    .connect(await getValidatorEthAccount(validator))
    .claimExitingNFTPosition()) as ContractTransaction;
  let receipt = await ethers.provider.getTransactionReceipt(claimTx.hash);
  return BigNumber.from(receipt.logs[0].topics[3]);
};

export const getStakeNFTFromMinorSlashEvent = async (
  tx: ContractTransaction
): Promise<bigint> => {
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  let intrface = new ethers.utils.Interface([
    "event ValidatorMinorSlashed(address indexed account, uint256 stakeNFT)",
  ]);
  let data = receipt.logs[receipt.logs.length - 1].data;
  let topics = receipt.logs[receipt.logs.length - 1].topics;
  let event = intrface.decodeEventLog("ValidatorMinorSlashed", data, topics);
  return event.stakeNFT;
};
