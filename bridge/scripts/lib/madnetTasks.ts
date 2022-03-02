import { env } from "./constants";
import { task, types } from "hardhat/config";
import fs from "fs";

import { BigNumber } from "ethers";

export async function getTokenIdFromTx(ethers: any, tx: any) {
  let abi = [
    "event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)",
  ];
  let iface = new ethers.utils.Interface(abi);
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  //console.log("receipt", receipt, tx)
  let log = iface.parseLog(receipt.logs[2]);
  const { from, to, tokenId } = log.args;
  return tokenId;
}

task(
  "updateDeploymentArgsWithFactory",
  "Computes factory address and to the deploymentArgs file"
).setAction(async (taskArgs, hre) => {
  const path: string = `./deployments/${env()}/deploymentArgsTemplate.json`;
  if (!fs.existsSync(path)) {
    throw "Error: Could not find deployment Args file expected at " + path;
  }
  console.log(`Loading deploymentArgs from: ${path}`);
  let rawData = fs.readFileSync(path);
  // Make sure that admin is the named account at position 0
  const [admin] = await hre.ethers.getSigners();
  let txCount = await hre.ethers.provider.getTransactionCount(admin.address);
  //calculate the factory address for the constructor arg
  const factoryAddress = hre.ethers.utils.getContractAddress({
    from: admin.address,
    nonce: txCount,
  });
  console.log(`Future factory Address: ${factoryAddress}`);
  let replaceStringsPair = [
    [`"selfAddr_": "UNDEFINED"`, `"selfAddr_": "${factoryAddress}"`],
    [`"owner_": "UNDEFINED"`, `"owner_": "${admin.address}"`],
  ];
  let outputData = rawData.toString();
  for (let pair of replaceStringsPair) {
    outputData = outputData.replace(pair[0], pair[1]);
  }
  console.log(`Saving file at: ${path}`);
  fs.writeFileSync(`./deployments/${env()}/deploymentArgs.json`, outputData);
});

task("registerValidators", "registers validators")
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addVariadicPositionalParam(
    "addresses",
    "validators' addresses",
    undefined,
    types.string,
    false
  )
  .setAction(async (taskArgs, hre) => {
    console.log("registerValidators", taskArgs.addresses);

    const factory = await hre.ethers.getContractAt(
      "MadnetFactory",
      taskArgs.factoryAddress
    );

    const lockTime = 1;
    const validatorAddresses: string[] = taskArgs.addresses;
    const stakingTokenIds: BigNumber[] = [];

    const madToken = await hre.ethers.getContractAt(
      "MadToken",
      await factory.lookup("MadToken")
    );
    console.log(`MadToken Address: ${madToken.address}`);
    const stakeNFT = await hre.ethers.getContractAt(
      "StakeNFT",
      await factory.lookup("StakeNFT")
    );
    console.log(`stakeNFT Address: ${stakeNFT.address}`);
    const validatorPool = await hre.ethers.getContractAt(
      "ValidatorPool",
      await factory.lookup("ValidatorPool")
    );
    console.log(`validatorPool Address: ${validatorPool.address}`);

    const stakeAmountMadWei = await validatorPool.getStakeAmount();

    console.log(
      `Minimum amount MadWei to stake: ${stakeAmountMadWei.toNumber()}`
    );

    // Make sure that admin is the named account at position 0
    const [admin] = await hre.ethers.getSigners();

    console.log(`Admin address: ${admin.address}`);

    // approve tokens
    let tx = await madToken
      .connect(admin)
      .approve(
        stakeNFT.address,
        stakeAmountMadWei.mul(validatorAddresses.length)
      );
    await tx.wait();

    console.log(
      `Approved allowance to validatorPool of: ${stakeAmountMadWei
        .mul(validatorAddresses.length)
        .toNumber()} MadWei`
    );

    console.log("Starting the registration process...");

    // mint StakeNFT positions to validators
    for (let i=0; i<validatorAddresses.length; i++) {
      let tx = await stakeNFT
        .connect(admin)
        .mintTo(factory.address, stakeAmountMadWei, lockTime);
      await tx.wait();

      let tokenId = BigNumber.from(await getTokenIdFromTx(hre.ethers, tx));
      console.log(`Minted StakeNFT.tokenID ${tokenId}`);
      stakingTokenIds.push(tokenId);

      let iface = new hre.ethers.utils.Interface([
        "function approve(address,uint256)",
      ]);
      let input = iface.encodeFunctionData("approve", [
        validatorPool.address,
        tokenId,
      ]);
      tx = await factory
        .connect(admin)
        .callAny(stakeNFT.address, 0, input);

      await tx.wait();


      console.log(`Approved tokenID:${tokenId} to ValidatorPool`);
    }

    console.log(
      `registering ${validatorAddresses.length} validators with ValidatorPool...`
    );

    // add validators to the ValidatorPool
    // await validatorPool.registerValidators(validatorAddresses, stakingTokenIds)
    let iface = new hre.ethers.utils.Interface([
      "function registerValidators(address[],uint256[])",
    ]);
    let input = iface.encodeFunctionData("registerValidators", [
      validatorAddresses,
      stakingTokenIds,
    ]);
    tx = await factory.connect(admin).callAny(validatorPool.address, 0, input);
    await tx.wait();
    console.log("done");
  });

task("ethdkgInput", "calculate the initializeETHDKG selector").setAction(
  async (taskArgs, hre) => {
    const { ethers } = hre;
    let iface = new ethers.utils.Interface(["function initializeETHDKG()"]);
    let input = iface.encodeFunctionData("initializeETHDKG");
    console.log("input", input);
  }
);

task(
  "virtualMintDeposit",
  "Virtually creates a deposit on the side chain"
).setAction(async (taskArgs, hre) => {
  const { ethers } = hre;
  let iface = new ethers.utils.Interface([
    "function virtualMintDeposit(uint8 accountType_,address to_,uint256 amount_)",
  ]);
  let input = iface.encodeFunctionData("virtualMintDeposit", [
    1,
    "0x546F99F244b7B58B855330AE0E2BC1b30b41302F",
    1001,
  ]);
  console.log("input", input);
  const [admin] = await ethers.getSigners();
  const adminSigner = await ethers.getSigner(admin.address);
  const factory = await ethers.getContractAt(
    "MadnetFactory",
    "0x0b1f9c2b7bed6db83295c7b5158e3806d67ec5bc"
  );
  const madByte = await ethers.getContractAt(
    "MadByte",
    await factory.lookup("MadByte")
  );
  let tx = await factory
    .connect(adminSigner)
    .callAny(madByte.address, 0, input);
  await tx.wait();
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  console.log(receipt);
  let intrface = new ethers.utils.Interface([
    "event DepositReceived(uint256 indexed depositID, uint8 indexed accountType, address indexed depositor, uint256 amount)",
  ]);
  let data = receipt.logs[0].data;
  let topics = receipt.logs[0].topics;
  let event = intrface.decodeEventLog("DepositReceived", data, topics);
  console.log(event);
});
