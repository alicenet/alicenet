import { task, types } from "hardhat/config";
import { deployUpgradeableGasSafe } from "../lib/alicenetFactory";
import {
  ALICENET_FACTORY,
  ALICENET_FACTORY_OWNER_ADDRESS,
  ALICE_NET_FACTORY_ADDRESS,
  BRIDGE_POOL_FACTORY,
} from "../lib/constants";
import { getGasPrices } from "../lib/deployment/utils";
import { dryRunTask, processSingleContractEvent } from "./taskUtils";

task(
  "deploy-bridge-factory",
  "deploys the bridge pool factory as a upgrade proxy from alicenet factory"
)
  .addParam(
    "factory-address",
    "address of alicenet smart contract factory, defaults to mainnet factory address",
    ALICE_NET_FACTORY_ADDRESS,
    types.string
  )
  .addFlag(
    "test",
    "this flag must be used with hardhat fork configured in hardhat config, runs the task, impersonating owner of alicenet factory on mainnet"
  )
  .setAction(async (taskArgs, hre) => {
    //get an instance of the factory
    let factory = await hre.ethers.getContractAt(
      ALICENET_FACTORY,
      taskArgs.factoryAddress
    );
    //if test flag is set, impersonate the owner of the factory
    if (taskArgs.test) {
      factory = await dryRunTask(ALICENET_FACTORY_OWNER_ADDRESS, factory, hre);
    }
    //encode initialize function call
    const bridgeFactory = await hre.ethers.getContractFactory(
      BRIDGE_POOL_FACTORY
    );
    const initCallData =
      bridgeFactory.interface.encodeFunctionData("initialize");
    //deploy the bridge pool factory
    const salt = hre.ethers.utils.formatBytes32String("BridgePoolFactory");
    const txResponse = await deployUpgradeableGasSafe(
      BRIDGE_POOL_FACTORY,
      factory,
      hre.ethers,
      initCallData,
      [],
      salt,
      undefined,
      await getGasPrices(hre.ethers)
    );
    const receipt = await txResponse.wait();
    const events = factory.interface.events;
    for (let event in events) {
      event;
    }
    console.log("receipt: ", receipt);
    //parse the events from the receipt to get human readable data
    await processSingleContractEvent(factory, receipt);
  });
