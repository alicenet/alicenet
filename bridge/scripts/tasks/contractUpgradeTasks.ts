import { task, types } from "hardhat/config";
import { encodeMultiCallArgs } from "../lib/alicenetFactory";
import { DeploymentConfig } from "../lib/deployment/interfaces";
import {
  impersonateFactoryOwner,
  upgradeProxyTask,
} from "../lib/deployment/tasks";
import {
  extractFullContractInfoByContractName,
  promptCheckDeploymentArgs,
} from "../lib/deployment/utils";

task(
  "upgrade-alca-minter",
  "deploy new alcaminter logic contract and upgrade proxy"
)
  .addOptionalParam(
    "alcaMinter",
    "Address of ALCA Minter proxy address",
    "0xc77fED1A4EF911b8649f395B1F458A1888EcF196",
    types.string
  )
  .addOptionalParam(
    "factoryAddress",
    "address of alicenet factory",
    "0x4b6dF6B299fB6414f45719E0d9e1889269a7843E",
    types.string
  )
  .addOptionalParam("waitConfirmation", "Wait for confirmation of transaction")
  .addFlag("skipInitializer", "Skip initializer")
  .addFlag("test", "Test mode")
  .setAction(async (taskArgs, hre) => {
    const deploymentConfigForContract: DeploymentConfig =
      await extractFullContractInfoByContractName(
        "ALCAMinter",
        hre.artifacts,
        hre.ethers
      );
    await promptCheckDeploymentArgs(
      `deploy ALCA minter logic from factory address: ${taskArgs.factoryAddress}, and upgrade proxy: ${taskArgs.alcaMinter}? y/n\n`
    );
    console.log("upgrading proxy....");
    const proxyData = await upgradeProxyTask(
      deploymentConfigForContract,
      hre,
      taskArgs.waitConfirmation,
      undefined,
      taskArgs.factoryAddress,
      undefined,
      taskArgs.skipInitializer,
      taskArgs.test
    );
    console.log(
      "ALCA minter proxy upgraded to new logic contract: ",
      proxyData.logicAddress
    );
    if (taskArgs.test === true) {
      //run post deployment test
      const factory = await impersonateFactoryOwner(
        hre,
        undefined,
        taskArgs.factoryAddress
      );
      const alcaMinter = await hre.ethers.getContractAt(
        "ALCAMinter",
        proxyData.proxyAddress
      );
      // get the current supply
      let currentSupply = await alcaMinter.totalSupply();
      console.log("current supply:", currentSupply.toString());
      //check if able to mint alca up to 1 billion max supply
      const mintAmount = hre.ethers.utils
        .parseEther("1000000000")
        .sub(currentSupply);
      console.log("encoding mint function data");
      let callData = alcaMinter.interface.encodeFunctionData("mint", [
        await factory.signer.getAddress(),
        mintAmount,
      ]);
      console.log("encoded mint function data:", callData);
      console.log("encoding multicall args");
      let multicallArgs = encodeMultiCallArgs(alcaMinter.address, 0, callData);
      //send the mutlticall to the factory
      console.log("sending multicall to factory");
      let tx = factory.multiCall([multicallArgs]);
      await (await tx).wait();
      currentSupply = await alcaMinter.totalSupply();
      hre
        .expect(currentSupply)
        .to.equal(hre.ethers.utils.parseEther("1000000000"));
      callData = alcaMinter.interface.encodeFunctionData("mint", [
        await factory.signer.getAddress(),
        1,
      ]);
      multicallArgs = encodeMultiCallArgs(alcaMinter.address, 0, callData);
      //send the mutlticall to the factory
      tx = factory.multiCall([multicallArgs]);
      //check if revert when minting more than 1 billion
      await hre
        .expect(tx)
        .to.be.revertedWithCustomError(alcaMinter, "MintingExceeds1Billion")
        .withArgs(hre.ethers.utils.parseEther("1000000000"));
    }
  });

task(
  "alca-minter-post-mainnet-deployment-test",
  "test alca minter post deployment on mainnet, sa"
)
  .addParam(
    "alcaMinter",
    "Address of ALCA Minter proxy address",
    undefined,
    types.string
  )
  .setAction(async (taskArgs, hre) => {
    const alcaMinter = await hre.ethers.getContractAt(
      "ALCAMinter",
      taskArgs.alcaMinter
    );
    // get the current supply
    let currentSupply = await alcaMinter.callStatic.totalSupply();
    console.log("current supply:", currentSupply.toString());
    //get the max supply
    let maxSupply = await alcaMinter.callStatic.maxSupply();
    console.log("max supply:", maxSupply.toString());
  });
