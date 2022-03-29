import { BigNumber, ContractTransaction, Signer } from "ethers";
import { ethers } from "hardhat";
import { ValidatorRawData } from "../ethdkg/setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getTokenIdFromTx,
  getValidatorEthAccount,
} from "../setup";

interface Contract {
  PublicStaking: bigint;
  ValNFT: bigint;
  MAD: bigint;
  ETH: bigint;
  Addr: string;
}

interface Admin {
  PublicStaking: bigint;
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
  PublicStaking: Contract;
  ValidatorStaking: Contract;
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
  const state: State = {
    Admin: {
      PublicStaking: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      Addr: "0x0",
    },
    PublicStaking: {
      PublicStaking: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      ETH: BigInt(0),
      Addr: "0x0",
    },
    ValidatorStaking: {
      PublicStaking: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      ETH: BigInt(0),
      Addr: "0x0",
    },
    ValidatorPool: {
      PublicStaking: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      ETH: BigInt(0),
      Addr: "0x0",
    },
    Factory: {
      PublicStaking: BigInt(0),
      ValNFT: BigInt(0),
      MAD: BigInt(0),
      ETH: BigInt(0),
      Addr: "0x0",
    },
    validators: [],
  };
  const [adminSigner] = await ethers.getSigners();
  // Get state for admin
  state.Admin.PublicStaking = (
    await fixture.publicStaking.balanceOf(adminSigner.address)
  ).toBigInt();
  state.Admin.ValNFT = (
    await fixture.validatorStaking.balanceOf(adminSigner.address)
  ).toBigInt();
  state.Admin.MAD = (
    await fixture.madToken.balanceOf(adminSigner.address)
  ).toBigInt();
  state.Admin.Addr = adminSigner.address;

  // Get state for validators
  for (let i = 0; i < _validators.length; i++) {
    const validator: Validator = {
      Idx: i,
      NFT: (await fixture.publicStaking.balanceOf(_validators[i])).toBigInt(),
      MAD: (await fixture.madToken.balanceOf(_validators[i])).toBigInt(),
      Addr: _validators[i],
      Reg: await fixture.validatorPool.isValidator(_validators[i]),
      ExQ: await fixture.validatorPool.isInExitingQueue(_validators[i]),
      Acc: await fixture.validatorPool.isAccusable(_validators[i]),
    };
    state.validators.push(validator);
  }
  // Contract data
  const contractData = [
    {
      contractState: state.PublicStaking,
      contractAddress: fixture.publicStaking.address,
    },
    {
      contractState: state.ValidatorStaking,
      contractAddress: fixture.validatorStaking.address,
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
    contractData[i].contractState.PublicStaking = (
      await fixture.publicStaking.balanceOf(contractData[i].contractAddress)
    ).toBigInt();
    contractData[i].contractState.ValNFT = (
      await fixture.validatorStaking.balanceOf(contractData[i].contractAddress)
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
  if (process.env.npm_config_detailed === "true") {
    // execute "npm --detailed=true  run test" to see this output
    console.log(title);
    console.log(state);
  }
};

export const createValidators = async (
  fixture: Fixture,
  _validatorsSnapshots: ValidatorRawData[]
): Promise<string[]> => {
  const validators: string[] = [];
  const stakeAmountMadWei = await fixture.validatorPool.getStakeAmount();
  const [adminSigner] = await ethers.getSigners();
  // Approve ValidatorPool to withdraw MAD tokens of validators
  await fixture.madToken.approve(
    fixture.validatorPool.address,
    stakeAmountMadWei.mul(_validatorsSnapshots.length)
  );
  for (let i = 0; i < _validatorsSnapshots.length; i++) {
    const validator = _validatorsSnapshots[i];
    await getValidatorEthAccount(validator);
    validators.push(validator.address);
    // Send MAD tokens to each validator
    await fixture.madToken.transfer(validator.address, stakeAmountMadWei);
  }
  await fixture.madToken
    .connect(adminSigner)
    .approve(
      fixture.publicStaking.address,
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
  const stakingTokenIds: BigNumber[] = [];
  const [adminSigner] = await ethers.getSigners();
  const stakeAmountMadWei = await fixture.validatorPool.getStakeAmount();
  const lockTime = 1;
  for (let i = 0; i < validators.length; i++) {
    // Stake all MAD tokens
    const tx = await fixture.publicStaking
      .connect(adminSigner)
      .mintTo(fixture.factory.address, stakeAmountMadWei, lockTime);
    // Get the proof of staking (NFT's tokenID)
    const tokenID = await getTokenIdFromTx(tx);
    stakingTokenIds.push(tokenID);
    await factoryCallAnyFixture(fixture, "publicStaking", "approve", [
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
  const claimTx = (await fixture.validatorPool
    .connect(await getValidatorEthAccount(validator))
    .claimExitingNFTPosition()) as ContractTransaction;
  const receipt = await ethers.provider.getTransactionReceipt(claimTx.hash);
  return BigNumber.from(receipt.logs[0].topics[3]);
};

export const getPublicStakingFromMinorSlashEvent = async (
  tx: ContractTransaction
): Promise<bigint> => {
  const receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  const intrface = new ethers.utils.Interface([
    "event ValidatorMinorSlashed(address indexed account, uint256 publicStaking)",
  ]);
  const data = receipt.logs[receipt.logs.length - 1].data;
  const topics = receipt.logs[receipt.logs.length - 1].topics;
  const event = intrface.decodeEventLog("ValidatorMinorSlashed", data, topics);
  return event.publicStaking;
};

/**
 * Mint a publicStaking and burn it to the ValidatorPool contract. Besides a contract self destructing
 * itself, this is a method to send eth accidentally to the validatorPool contract
 * @param fixture
 * @param etherAmount
 * @param madTokenAmount
 * @param adminSigner
 */
export const burnStakeTo = async (
  fixture: Fixture,
  etherAmount: BigNumber,
  madTokenAmount: BigNumber,
  adminSigner: Signer
) => {
  await fixture.madToken
    .connect(adminSigner)
    .approve(fixture.publicStaking.address, madTokenAmount);
  const tx = await fixture.publicStaking
    .connect(adminSigner)
    .mint(madTokenAmount);
  const tokenID = await getTokenIdFromTx(tx);
  await fixture.publicStaking.depositEth(42, {
    value: etherAmount,
  });
  await fixture.publicStaking
    .connect(adminSigner)
    .burnTo(fixture.validatorPool.address, tokenID);
};

/**
 * Mint a publicStaking
 * @param fixture
 * @param etherAmount
 * @param madTokenAmount
 * @param adminSigner
 */
export const mintPublicStaking = async (
  fixture: Fixture,
  madTokenAmount: BigNumber,
  adminSigner: Signer
) => {
  await fixture.madToken
    .connect(adminSigner)
    .approve(fixture.publicStaking.address, madTokenAmount);
  const tx = await fixture.publicStaking
    .connect(adminSigner)
    .mint(madTokenAmount);
  return await getTokenIdFromTx(tx);
};
