import { BigNumber } from "ethers";
import fs from "fs";
import { task, types } from "hardhat/config";
import { env } from "./constants";

export async function getTokenIdFromTx(ethers: any, tx: any) {
  const abi = [
    "event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)",
  ];
  const iface = new ethers.utils.Interface(abi);
  const receipt = await tx.wait();
  const log = iface.parseLog(receipt.logs[2]);
  return log.args[2];
}

task(
  "updateDeploymentArgsWithFactory",
  "Computes factory address and to the deploymentArgs file"
).setAction(async (taskArgs, hre) => {
  const path: string = `./deployments/${env()}/deploymentArgsTemplate.json`;
  if (!fs.existsSync(path)) {
    throw new Error(
      `Error: Could not find deployment Args file expected at ${path}`
    );
  }
  console.log(`Loading deploymentArgs from: ${path}`);
  const rawData = fs.readFileSync(path);
  // Make sure that admin is the named account at position 0
  const [admin] = await hre.ethers.getSigners();
  const txCount = await hre.ethers.provider.getTransactionCount(admin.address);
  // calculate the factory address for the constructor arg
  const factoryAddress = hre.ethers.utils.getContractAddress({
    from: admin.address,
    nonce: txCount,
  });
  console.log(`Future factory Address: ${factoryAddress}`);
  const replaceStringsPair = [
    [`"selfAddr_": "UNDEFINED"`, `"selfAddr_": "${factoryAddress}"`],
    [`"owner_": "UNDEFINED"`, `"owner_": "${admin.address}"`],
  ];
  let outputData = rawData.toString();
  for (const pair of replaceStringsPair) {
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
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const lockTime = 1;
    const validatorAddresses: string[] = taskArgs.addresses;
    const stakingTokenIds: BigNumber[] = [];

    const aToken = await hre.ethers.getContractAt(
      "AToken",
      await factory.lookup(hre.ethers.utils.formatBytes32String("AToken"))
    );
    console.log(`AToken Address: ${aToken.address}`);
    const publicStaking = await hre.ethers.getContractAt(
      "PublicStaking",
      await factory.lookup(
        hre.ethers.utils.formatBytes32String("PublicStaking")
      )
    );
    console.log(`publicStaking Address: ${publicStaking.address}`);
    const validatorPool = await hre.ethers.getContractAt(
      "ValidatorPool",
      await factory.lookup(
        hre.ethers.utils.formatBytes32String("ValidatorPool")
      )
    );
    console.log(`validatorPool Address: ${validatorPool.address}`);
    console.log(await validatorPool.getMaxNumValidators());
    const stakeAmountATokenWei = await validatorPool.getStakeAmount();

    console.log(
      `Minimum amount ATokenWei to stake: ${stakeAmountATokenWei.toNumber()}`
    );

    // Make sure that admin is the named account at position 0
    const [admin] = await hre.ethers.getSigners();
    console.log(`Admin address: ${admin.address}`);
    let iface = new hre.ethers.utils.Interface([
      "function transfer(address,uint256)",
    ]);
    const totalStakeAmt = stakeAmountATokenWei.mul(validatorAddresses.length);
    let input = iface.encodeFunctionData("transfer", [
      admin.address,
      totalStakeAmt,
    ]);
    let tx = await factory.connect(admin).callAny(aToken.address, 0, input);
    await tx.wait();
    // approve tokens
    tx = await aToken
      .connect(admin)
      .approve(
        publicStaking.address,
        stakeAmountATokenWei.mul(validatorAddresses.length)
      );
    await tx.wait();
    console.log(
      `Approved allowance to validatorPool of: ${stakeAmountATokenWei
        .mul(validatorAddresses.length)
        .toNumber()} ATokenWei`
    );

    console.log("Starting the registration process...");
    // mint PublicStaking positions to validators
    for (let i = 0; i < validatorAddresses.length; i++) {
      let tx = await publicStaking
        .connect(admin)
        .mintTo(factory.address, stakeAmountATokenWei, lockTime);
      await tx.wait();
      const tokenId = BigNumber.from(await getTokenIdFromTx(hre.ethers, tx));
      console.log(`Minted PublicStaking.tokenID ${tokenId}`);
      stakingTokenIds.push(tokenId);
      const iface = new hre.ethers.utils.Interface([
        "function approve(address,uint256)",
      ]);
      const input = iface.encodeFunctionData("approve", [
        validatorPool.address,
        tokenId,
      ]);
      tx = await factory
        .connect(admin)
        .callAny(publicStaking.address, 0, input);

      await tx.wait();
      console.log(`Approved tokenID:${tokenId} to ValidatorPool`);
    }

    console.log(
      `registering ${validatorAddresses.length} validators with ValidatorPool...`
    );
    // add validators to the ValidatorPool
    // await validatorPool.registerValidators(validatorAddresses, stakingTokenIds)
    iface = new hre.ethers.utils.Interface([
      "function registerValidators(address[],uint256[])",
    ]);
    input = iface.encodeFunctionData("registerValidators", [
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
    const iface = new ethers.utils.Interface(["function initializeETHDKG()"]);
    const input = iface.encodeFunctionData("initializeETHDKG");
    console.log("input", input);
  }
);

task(
  "virtualMintDeposit",
  "Virtually creates a deposit on the side chain"
).setAction(async (taskArgs, hre) => {
  const { ethers } = hre;
  const iface = new ethers.utils.Interface([
    "function virtualMintDeposit(uint8 accountType_,address to_,uint256 amount_)",
  ]);
  const input = iface.encodeFunctionData("virtualMintDeposit", [
    1,
    "0x546F99F244b7B58B855330AE0E2BC1b30b41302F",
    1001,
  ]);
  console.log("input", input);
  const [admin] = await ethers.getSigners();
  const adminSigner = await ethers.getSigner(admin.address);
  const factory = await ethers.getContractAt(
    "AliceNetFactory",
    "0x0b1f9c2b7bed6db83295c7b5158e3806d67ec5bc"
  );
  const bToken = await ethers.getContractAt(
    "BToken",
    await factory.lookup(hre.ethers.utils.formatBytes32String("BToken"))
  );
  const tx = await factory
    .connect(adminSigner)
    .callAny(bToken.address, 0, input);
  await tx.wait();
  const receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  console.log(receipt);
  const intrface = new ethers.utils.Interface([
    "event DepositReceived(uint256 indexed depositID, uint8 indexed accountType, address indexed depositor, uint256 amount)",
  ]);
  const data = receipt.logs[0].data;
  const topics = receipt.logs[0].topics;
  const event = intrface.decodeEventLog("DepositReceived", data, topics);
  console.log(event);
});

task("scheduleMaintenance", "Calls schedule Maintenance")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;
    const iface = new ethers.utils.Interface([
      "function scheduleMaintenance()",
    ]);
    const input = iface.encodeFunctionData("scheduleMaintenance", []);
    console.log("input", input);
    const [admin] = await ethers.getSigners();
    const adminSigner = await ethers.getSigner(admin.address);
    const factory = await ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const validatorPool = await hre.ethers.getContractAt(
      "ValidatorPool",
      await factory.lookup(
        hre.ethers.utils.formatBytes32String("ValidatorPool")
      )
    );
    await (
      await factory
        .connect(adminSigner)
        .callAny(validatorPool.address, 0, input)
    ).wait();
  });

task(
  "pauseEthdkgArbitraryHeight",
  "Forcing consensus to stop on block number defined by --input"
)
  .addParam("alicenetHeight", "The block number after the latest block mined")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;
    const iface = new ethers.utils.Interface([
      "function pauseConsensusOnArbitraryHeight(uint256)",
    ]);
    const input = iface.encodeFunctionData("pauseConsensusOnArbitraryHeight", [
      taskArgs.alicenetHeight,
    ]);
    const [admin] = await ethers.getSigners();
    const adminSigner = await ethers.getSigner(admin.address);
    const factory = await ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const validatorPool = await hre.ethers.getContractAt(
      "ValidatorPool",
      await factory.lookup(
        hre.ethers.utils.formatBytes32String("ValidatorPool")
      )
    );
    await (
      await factory
        .connect(adminSigner)
        .callAny(validatorPool.address, 0, input)
    ).wait();
  });
