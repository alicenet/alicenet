import {
  getTokenIdFromTx,
  getValidatorEthAccount,
  factoryCallAny,
  Fixture,
} from "../setup";
import { ValidatorRawData } from "../ethdkg/setup";
import { ethers } from "hardhat";
import { BigNumber, ContractTransaction } from "ethers";

interface Contract {
  StakeNFT: number;
  ValNFT: number;
  MAD: number;
  ETH: number;
  Addr: string;
}

interface Validator {
  NFT: number;
  MAD: number;
  ETH: number;
  Addr: string;
  Reg: boolean;
  ExQ: boolean;
  Acc: boolean;
}
interface State {
  Admin: Contract;
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
  validators: string[]
): Promise<State> => {
  // System state
  let state: State = {
    Admin: {
      StakeNFT: 0,
      ValNFT: 0,
      MAD: 0,
      ETH: 0,
      Addr: "0x0",
    },
    StakeNFT: {
      StakeNFT: 0,
      ValNFT: 0,
      MAD: 0,
      ETH: 0,
      Addr: "0x0",
    },
    ValidatorNFT: {
      StakeNFT: 0,
      ValNFT: 0,
      MAD: 0,
      ETH: 0,
      Addr: "0x0",
    },
    ValidatorPool: {
      StakeNFT: 0,
      ValNFT: 0,
      MAD: 0,
      ETH: 0,
      Addr: "0x0",
    },
    Factory: {
      StakeNFT: 0,
      ValNFT: 0,
      MAD: 0,
      ETH: 0,
      Addr: "0x0",
    },
    validators: [],
  };
  let [adminSigner] = await ethers.getSigners();
  // Get state for admin
  state.Admin.StakeNFT = parseFloat(
    (
      await fixture.stakeNFT.balanceOf(await adminSigner.getAddress())
    ).toString()
  );
  state.Admin.ValNFT = parseFloat(
    (
      await fixture.validatorNFT.balanceOf(await adminSigner.getAddress())
    ).toString()
  );
  state.Admin.MAD = parseFloat(
    ethers.utils.formatEther(
      await fixture.madToken.balanceOf(await adminSigner.getAddress())
    )
  );
  state.Admin.ETH = parseFloat(
    (+ethers.utils.formatEther(
      await ethers.provider.getBalance(await adminSigner.getAddress())
    )).toFixed(0)
  );
  state.Admin.Addr = (await adminSigner.getAddress()).slice(-4);
  let maxNumValidators = await fixture.validatorPool.getMaxNumValidators();
  // Get state for validators
  for (let i = 0; i < maxNumValidators.toNumber(); i++) {
    if (typeof validators[i] !== undefined && validators[i]) {
      let validator: Validator = {
        NFT: parseFloat(
          (await fixture.stakeNFT.balanceOf(validators[i])).toString()
        ),
        MAD: parseFloat(
          ethers.utils.formatEther(
            await fixture.madToken.balanceOf(validators[i])
          )
        ),
        ETH: parseFloat(
          (+ethers.utils.formatEther(
            await ethers.provider.getBalance(validators[i])
          )).toFixed(0)
        ),
        Addr: validators[i].slice(-4),
        Reg: await fixture.validatorPool.isValidator(validators[i]),
        ExQ: await fixture.validatorPool.isInExitingQueue(validators[i]),
        Acc: await fixture.validatorPool.isAccusable(validators[i]),
      };
      state.validators.push(validator);
    }
    // state.validators[i].NFT = parseFloat((await fixture.stakeNFT.balanceOf(
    //   validators[i])).toString())
    // state.validators[i].MAD = parseFloat(ethers.utils.formatEther(await fixture.madToken.balanceOf(
    //   validators[i])))
    // state.validators[i].ETH = parseFloat((+ethers.utils.formatEther(await ethers.provider.getBalance(validators[i]))).toFixed(0))
    // state.validators[i].Addr = validators[i].slice(-4)
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
    contractData[i].contractState.StakeNFT = parseInt(
      (
        await fixture.stakeNFT.balanceOf(contractData[i].contractAddress)
      ).toString()
    );
    contractData[i].contractState.ValNFT = parseInt(
      (
        await fixture.validatorNFT.balanceOf(contractData[i].contractAddress)
      ).toString()
    );
    contractData[i].contractState.MAD = parseFloat(
      ethers.utils.formatEther(
        await fixture.madToken.balanceOf(contractData[i].contractAddress)
      )
    );
    contractData[i].contractState.ETH = parseFloat(
      (+ethers.utils.formatEther(
        await ethers.provider.getBalance(contractData[i].contractAddress)
      )).toFixed(0)
    );
    contractData[i].contractState.Addr =
      contractData[i].contractAddress.slice(-4);
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
  for (const validator of _validatorsSnapshots) {
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
  for (let i=0; i< validators.length ; i++) {
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
  return stakingTokenIds
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
