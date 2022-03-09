import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BaseContract, BigNumber } from "ethers/lib/ethers";
import { ethers, network } from "hardhat";
import { staticDeployment } from "../../scripts/lib/constants";
import { MadToken, MadToken__factory } from "../../typechain-types";
import { AToken } from "../../typechain-types/AToken";
import { deployStaticWithFactory, expect } from "../setup";

describe("aToken: Testing Migration", async () => {
  let aToken: AToken;
  let madToken: MadToken;
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let minter: SignerWithAddress;
  let burner: SignerWithAddress;
  let expectedState : state
  let currentState : state
  // interface StateContract  {
  //   contract: BaseContract
  //       balance: number
  // }
  // interface StateActor {
  //   actor: SignerWithAddress,
  //       balance: number
  // }
  // interface state  {
  //   contracts: StateContract[]
  //   actors: StateActor[]
  // }
  interface state {
    Balances: {
      madToken: {
        address: string;
        admin: number;
        user: number;
        aToken: number
      };
      aToken: {
        address: string;
        admin: number;
        user: number;
        madToken: number;
      };
    };
  }


  beforeEach(async function () {
    await network.provider.send("evm_setAutomine", [true]);
    // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
    // being sent as input to the function (the contract bytecode), so we need to increase the block
    // gas limit temporally in order to deploy the template
    await network.provider.send("evm_setBlockGasLimit", ["0x3000000000000000"]);

    const namedSigners = await ethers.getSigners();
    [admin,user,minter,burner] = namedSigners;

    let txCount = await ethers.provider.getTransactionCount(admin.address);
    //calculate the factory address for the constructor arg
    let futureFactoryAddress = ethers.utils.getContractAddress({
      from: admin.address,
      nonce: txCount,
    });

    const Factory = await ethers.getContractFactory("MadnetFactory");
    const factory = await Factory.deploy(futureFactoryAddress);
    await factory.deployed();

    // MadToken
    madToken = (await deployStaticWithFactory(factory, "MadToken", [
      admin.address,
    ])) as MadToken;

    aToken = (await deployStaticWithFactory(
      factory,
      "AToken",
      [madToken.address],
    )) as AToken;

    // finish workaround, putting the blockgas limit to the previous value 30_000_000
    await network.provider.send("evm_setBlockGasLimit", ["0x1C9C380"]);
    await madToken.connect(admin).approve(admin.address, 1000);
    await madToken.connect(admin).transferFrom(admin.address, user.address, 1000);

    showState("Initial", await getState());
    expectedState = await getState();

  });

  it("Should migrate user tokens", async function () {
    let amount = 1000
    await madToken.connect(user).approve(aToken.address, amount);
    await aToken.connect(user).migrate(amount);
    expectedState.Balances.madToken.user -= amount
    expectedState.Balances.aToken.user += amount
    expectedState.Balances.madToken.aToken += amount
    currentState = await getState()
    showState("Expected", expectedState)
    showState("Current", currentState)
    expect(currentState).to.be.deep.eq(expectedState);
  });

  async function getState() {
    let state: state = {
      Balances: {
        madToken: {
          address: madToken.address.slice(-4),
          admin: format(await madToken.balanceOf(admin.address)),
          user: format(await madToken.balanceOf(user.address)),
          aToken: format(await madToken.balanceOf(aToken.address)),
        },
        aToken: {
          address: aToken.address.slice(-4),
          admin: format(await aToken.balanceOf(admin.address)),
          user: format(await aToken.balanceOf(user.address)),
          madToken: format(await aToken.balanceOf(madToken.address)),
        },
      },
    };
    return state;
  }

  function showState(title: string, state: state) {
    if (process.env.npm_config_detailed == "true") {
      // execute "npm --detailed=true test" to see this output
      console.log(title, state);
    }
  }

  function format(number: BigNumber) {
    return parseInt(ethers.utils.formatUnits(number, 0));
  }
});
