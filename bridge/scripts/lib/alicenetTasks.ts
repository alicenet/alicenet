import toml from "@iarna/toml";
import { spawn } from "child_process";
import {
  BigNumber,
  BigNumberish,
  BytesLike,
  ContractTransaction,
} from "ethers";
import fs from "fs";
import { task, types } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
// import { ValidatorPool } from "../../typechain-types";
import axios from "axios";
import { getEventVar } from "./alicenetFactory";
import { getGasPrices } from "./alicenetFactoryTasks";
import { CONTRACT_ADDR, DEFAULT_CONFIG_DIR, DEPLOYED_RAW } from "./constants";
import { readDeploymentArgs } from "./deployment/deploymentConfigUtil";
export type MultiCallArgsStruct = {
  target: string;
  value: BigNumberish;
  data: BytesLike;
};
function delay(milliseconds: number) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

export async function getTokenIdFromTx(ethers: any, tx: ContractTransaction) {
  const abi = [
    "event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)",
  ];
  const iface = new ethers.utils.Interface(abi);
  const receipt = await tx.wait();
  const log = iface.parseLog(receipt.logs[2]);
  return log.args[2];
}

task(
  "deploy-legacy-token-and-update-deployment-args",
  "Computes factory address and to the deploymentArgs file"
)
  .addOptionalParam(
    "deploymentArgsTemplatePath",
    "path of the deploymentArgsTemplate file",
    DEFAULT_CONFIG_DIR + "/deploymentArgsTemplate"
  )
  .addOptionalParam(
    "outputFolder",
    "path of the output folder where new deploymentArgsTemplate file will be saved",
    "../scripts/generated"
  )
  .setAction(async (taskArgs, hre) => {
    if (!fs.existsSync(taskArgs.deploymentArgsTemplatePath)) {
      throw new Error(
        `Error: Could not find deployment Args file expected at ${taskArgs.deploymentArgsTemplatePath}`
      );
    }
    if (!fs.existsSync(taskArgs.outputFolder)) {
      throw new Error(
        `Error: Output folder  ${taskArgs.outputFolder} doesn't exist!`
      );
    }
    console.log(
      `Loading deploymentArgs from: ${taskArgs.deploymentArgsTemplatePath}`
    );

    const deploymentConfig: any = await readDeploymentArgs(
      taskArgs.deploymentArgsTemplatePath
    );

    const expectedContract = "contracts/AliceNetFactory.sol:AliceNetFactory";
    const expectedField = "legacyToken_";
    if (deploymentConfig.constructor[expectedContract] === undefined) {
      throw new Error(
        `Couldn't find ${expectedField} in the constructor area for` +
          ` ${expectedContract} inside the ${taskArgs.deploymentArgsTemplatePath}`
      );
    }

    // Make sure that admin is the named account at position 0
    const [admin] = await hre.ethers.getSigners();
    console.log(`Admin address: ${admin.address}`);

    const legacyToken = await (
      await (await hre.ethers.getContractFactory("LegacyToken"))
        .connect(admin)
        .deploy()
    ).deployed();

    console.log(
      `Minted ${await legacyToken.balanceOf(admin.address)} tokens for user: ${
        admin.address
      }`
    );

    console.log(`Deployed legacy token at: ${legacyToken.address}`);
    deploymentConfig.constructor[expectedContract][0] = {
      legacyToken_: legacyToken.address,
    };

    const data = toml.stringify(deploymentConfig);
    fs.writeFileSync(taskArgs.outputFolder + "/deploymentArgsTemplate", data);
  });

task("create-local-seed-node", "start and syncs a node with mainnet")
  .addOptionalParam(
    "configPath",
    "path to the nodes config file",
    "~/Desktop/seedValidatorConfig.toml"
  )
  .setAction(async (taskArgs) => {
    const valNode = spawn(
      "./alicenet",
      ["--config", taskArgs.configPath, "validator"],
      {
        cwd: "../",
        shell: true,
      }
    );
    valNode.stdout.on("data", (data) => {
      console.log(data.toString());
    });
    valNode.stderr.on("data", (data) => {
      console.log(data.toString());
    });
    valNode.on("close", (code) => {
      console.log(`child process exited with code ${code}`);
    });
    let synced = false;
    let alicenetHeight;
    while (!synced) {
      try {
        const requestConfig = {
          timeout: 2000,
        };
        const response = await axios.post(
          "http://0.0.0.0:8885/v1/" + "get-block-number",
          {},
          requestConfig
        );
        if (response.status === 200) {
          alicenetHeight = response.data;
          synced = true;
          break;
        }
      } catch (err: any) {
        if (err) {
          await new Promise((resolve) => setTimeout(resolve, 5000));
        }
      }
    }
    valNode.kill(1);
    console.log(alicenetHeight.BlockHeight);
    return alicenetHeight.BlockHeight.toString();
  });
// task("fork-external-chain", "")
//   .addOptionalParam("rpcUrl")
//   .setAction(async () => {
//     const hardhatNode = spawn("npm", ["run", "fork-testnet"]);
//     hardhatNode.stdout.on("data", (data) => {
//       console.log(data.toString());
//     });
//     hardhatNode.stderr.on("data", (data) => {
//       console.log(data.toString());
//     });
//     hardhatNode.on("close", (code) => {
//       console.log(`child process exited with code ${code}`);
//     });

//     while (1) {
//       continue;
//     }
//   });

task(
  "start-local-seed-node",
  "starts a node already synce with remote testnet on local testnet"
).setAction(async () => {
  spawn(
    "./alicenet",
    [
      "--config",
      "./scripts/base-files/localTestNetBaseConfig.toml",
      "validator",
    ],
    {
      cwd: "../",
      shell: true,
    }
  );

  // valNode.stdout.on("data", (data) => {
  //   console.log(data.toString());
  // });
  // valNode.stderr.on("data", (data) => {
  //   console.log(data.toString());
  // });
  // valNode.on("close", (code) => {
  //   console.log(`child process exited with code ${code}`);
  // });
});

task("enable-local-environment-impersonate")
  .addParam(
    "account",
    "account to impersonate",
    "0xb9670e38d560c5662f0832cacaac3282ecffddb1"
  )
  .setAction(async (taskArgs, hre) => {
    await hre.network.provider.request({
      method: "hardhat_impersonateAccount",
      params: [taskArgs.account],
    });
  });

task("mine-num-blocks")
  .addParam("numBlocks", "number of blocks to mine")
  .setAction(async (taskArgs, hre) => {
    const numBlocks = parseInt(taskArgs.numBlocks, 10);
    await hre.network.provider.send("hardhat_mine", [
      "0x" + numBlocks.toString(16),
    ]);
  });

task(
  "deploy-state-migration-contract",
  "Deploy state migration contract and run migrations"
)
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addFlag(
    "skipFirstTransaction",
    "The task executes 2 tx to execute the migrations." +
      " Use this flag if you want to skip the first tx where we mint the NFT."
  )
  .setAction(async (taskArgs, hre) => {
    if (
      taskArgs.factoryAddress === undefined ||
      taskArgs.factoryAddress === ""
    ) {
      throw new Error("Expected a factory address to be passed!");
    }
    // Make sure that admin is the named account at position 0
    const [admin] = await hre.ethers.getSigners();
    console.log(`Admin address: ${admin.address}`);
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const alcaAddress = await factoryLookupAddress(
      factory.address,
      "AToken",
      hre
    );
    const publicStakingAddress = await factoryLookupAddress(
      factory.address,
      "PublicStaking",
      hre
    );
    const validatorPoolAddress = await factoryLookupAddress(
      factory.address,
      "ValidatorPool",
      hre
    );
    const snapshotAddress = await factoryLookupAddress(
      factory.address,
      "Snapshots",
      hre
    );
    const ethDKGAddress = await factoryLookupAddress(
      factory.address,
      "ETHDKG",
      hre
    );
    const tokenIds: Array<BigNumber> = [];
    const epoch = BigNumber.from(101);
    const masterPublicKey = [
      BigNumber.from(
        "19973864405227474494428046886218960395017286398286997273859673757240376592503"
      ),
      BigNumber.from(
        "3666564623138203565215945530726129772754495167220101411282198856069539995091"
      ),
      BigNumber.from(
        "16839568499509318396654065605910525620620734593528167444141729662886794883267"
      ),
      BigNumber.from(
        "4238394038888761339041070176707923282936512397513542246905279490862584218353"
      ),
    ];
    console.log("Master Public Key: " + masterPublicKey);
    const validatorAccounts = [
      "0xb80d6653f7e5b80dbbe8d0aa9f61b5d72e8028ad",
      "0x25489d6a720663f7e5253df68948edb302dfdcb6",
      "0x322e8f463b925da54a778ed597aef41bc4fe4743",
      "0xadf2a338e19c12298a3007cbea1c5276d1f746e0",
    ];

    const blockClaims: Array<string> = [
      "0x000000000100040015000000008001000d00000002010000190000000201000025000000020100003100000002010000417e8cd8049656c9e1b2cbf90ae9dcc7fa1d70f2aaea7df7a2a4cfcd8b4a02a8c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4702a5187985e08d03011645d5072d0aadb9b50eafb0014ad93cbc8a2c869a4139054ef9bd198da2c45c75bbecd8062d39451112d53cf60c02b48f144778f40b1b5",
      "0x000000000100040015000000008401000d0000000201000019000000020100002500000002010000310000000201000034003720c36c4406bd8fd6cb7b6f9ec3877dcdffb08e74058e54f07015e89e99c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4703139e7a7362b5e743a2013b7f33d360530c5195a402693f74c78200e90afcef57f9e10df3d7cf622c1f2497043de80d3abf4450d05355571d732c1ea47987432",
      "0x000000000100040015000000008801000d000000020100001900000002010000250000000201000031000000020100007be9eda53b983d0fa3e17f6f8f051b658e165f39b1eb27da02eafa09467f76f8c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470b28d67ca4ba36e663354920c301f855eb83703445d503c80183ee877041bd05b9936d9df3a955c09140093ea74a33151249848b964d0bb579e24377535b56316",
      "0x000000000100040015000000008c01000d00000002010000190000000201000025000000020100003100000002010000a9272bb0fd1998d2c98d8c449efe48f9dca1fe7d5c59b15e13aa1844e07a0aedc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4704b4c04b0bacc06565ec44607848f4c16f7bc9b5bcf2a0622d0200f7d4e57f663735108c50c69dc004a21f95c1787b9109fba494481e504d8d87a4133b8b5057d",
      "0x000000000100040015000000009001000d00000002010000190000000201000025000000020100003100000002010000b11dfff48b9cd76b85e669f10e4cb8e64ab7d5f6a44d3aab1747c49b16626620c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4708925a50edd15ded4b0076b2d682aba6cabdf9b3682005254082b808b6eeefb4782fd2260282b659817b3f6e725e13f743cfbda9c27b34271e4a003607c743144",
      "0x000000000100040015000000009401000d0000000201000019000000020100002500000002010000310000000201000072e4379e7ab07923d3047755c3eee288116d911f7ff92532e6faaff39f4b4b58c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470c9e957c9edde4dc5e9766a1ec7524d412b456e0fcaba46d1f41d262768806b0827def5d9058edfb39cb9025033c0a3855d5c3de0850b52a446f907a7231ee744",
    ];
    const validatorShares: Array<Array<string>> = [
      [
        "0x109f68dde37442959baa4b16498a6fd19c285f9355c23d8eef900876e8536a12",
        "0x2c11cec2ce4e17afffcc105f9bd0e646f6274f562f1c93f93545fc22c74a2cdc",
        "0x0430024aa1619a117e74425481c44f4628f45af7b389e4d5f84fc41227e1829e",
        "0x163cb9abb41800ba5cc1955fd72c0983edc9869a21925006e691d4947451c9fd",
      ],
      [
        "0x13f8a33ff7ef3cb5536b2223195b7b652a3533d309ad3887bcf570c9b1dbe142",
        "0x170fe500681ff96e84a6dd7d1e4698f0ad6dd3ef17520a7cac29b29b84f86aa7",
        "0x19d29ec38a1d7d8d7284a76b214bf5b818eaccf47cd37ab8bc08c20833e586e9",
        "0x27c0f981d1bbc1667ea520341b7fa65fc79815ab8122a814e3714bfdbacd84db",
      ],
      [
        "0x1064dd800716a7e80ed5d40d5563940cd25be4451976a28c237e1426d40eae5a",
        "0x0e1c4d27ca7672e0662aaecbc5f2d62ec23e58cf63ffa9b6fabd41d3cff7c927",
        "0x106d1f91c4b77d5c9bb485aeea784e9acf0c91702eefb766e94aefc92043a004",
        "0x281971fd391a560142b8d796018afc31131a668b2ca6f62b304564d6422bb03f",
      ],
      [
        "0x2228b7dd85ddae13994fa85f42df1833da3b9468a1e65b987142d62f125a9754",
        "0x0c5682ae7cd22a3c3daff06ce469f318025845e90254d9d05cecaeba45f445a5",
        "0x2cdac99ed82ffc83fc17213e96d56400db23f08d05418936cb352f0e179cf971",
        "0x06371376125bb2b96a5e427fac829f5c3919296aac4c42ddc44eb7c773369c2b",
      ],
    ];
    const groupSignatures: Array<string> = await getGroupSignatures(epoch);
    if (
      taskArgs.skipFirstTransaction === undefined ||
      taskArgs.skipFirstTransaction === false
    ) {
      console.log("Staking Atoken!");
      const contractTx = await stakeValidators(
        4,
        taskArgs.factoryAddress,
        alcaAddress,
        publicStakingAddress,
        hre
      );
      const receipt = await contractTx.wait(8);
      if (receipt.events === undefined) {
        throw new Error("receipt has no events");
      }
      const events = receipt.events;

      const transferEventHash =
        "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef";
      for (const event of events) {
        if (
          event.address === publicStakingAddress &&
          event.topics[0] === transferEventHash
        ) {
          tokenIds.push(BigNumber.from(BigInt(event.topics[3])));
        }
      }
      console.log("minted the following tokens: ", tokenIds);
    }

    console.log("registering and migrating state");
    const ethHeight = await hre.ethers.provider.getBlockNumber();
    const validatorIndexes = [1, 2, 3, 4];
    const contractTx = await migrateSnapshotsAndValidators(
      factory.address,
      publicStakingAddress,
      snapshotAddress,
      ethDKGAddress,
      validatorPoolAddress,
      tokenIds,
      validatorAccounts,
      validatorIndexes,
      validatorShares,
      ethHeight,
      masterPublicKey,
      blockClaims,
      groupSignatures,
      hre
    );
    console.log("finished migration");
    await contractTx.wait(8);
  });
async function getGroupSignatures(epoch: BigNumber) {
  let start = BigNumber.from(1);
  const bufferSize = BigNumber.from(5);
  const groupSignatures: Array<string> = [];
  if (epoch.gt(bufferSize)) {
    start = epoch.sub(bufferSize);
  }
  console.log("epoch: ", epoch);
  console.log("epoch start: ", start);
  console.log("is epoch start less than epoch: ", start.lt(epoch));
  for (let i = start; i.lte(epoch); i = i.add(1)) {
    const height = i.mul(1024);
    console.log("retrieved alicenet height from rpc request: ", height);
    const response = await axios.post(
      "https://edge.alice.net/v1/get-block-header",
      { Height: height.toString() }
    );
    groupSignatures.push("0x" + response.data.BlockHeader.SigGroup);
  }
  console.log("snapshot group signatures: ", groupSignatures);
  return groupSignatures;
}

task("register-validators", "registers validators")
  .addFlag("test")
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addVariadicPositionalParam(
    "addresses",
    "validators' addresses",
    undefined,
    types.string,
    false
  )
  .setAction(async (taskArgs, hre) => {
    console.log("\nRegistering Validators\n", taskArgs.addresses);
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const publicStakingBase = await hre.ethers.getContractFactory(
      "PublicStaking"
    );
    const validatorPoolBase = await hre.ethers.getContractFactory(
      "ValidatorPool"
    );
    const alcaAddress = await factoryLookupAddress(
      factory.address,
      "AToken",
      hre
    );
    const publicStakingAddress = await factoryLookupAddress(
      factory.address,
      "PublicStaking",
      hre
    );
    const validatorPoolAddress = await factoryLookupAddress(
      factory.address,
      "ValidatorPool",
      hre
    );
    const validatorAddresses: string[] = taskArgs.addresses;

    if (taskArgs.test) {
      await hre.network.provider.send("hardhat_mine", [
        hre.ethers.utils.hexValue(3),
      ]);
    }
    let tx = await stakeValidators(
      validatorAddresses.length,
      factory.address,
      alcaAddress,
      publicStakingAddress,
      hre
    );
    const receipt = await tx.wait();
    if (receipt.events === undefined) {
      throw new Error("receipt has no events");
    }
    const events = receipt.events;
    const tokenIds: Array<BigNumber> = [];
    const approveTokens: Array<MultiCallArgsStruct> = [];
    const transferEventHash =
      "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef";
    for (const event of events) {
      if (
        event.address === publicStakingAddress &&
        event.topics[0] === transferEventHash
      ) {
        const id = BigNumber.from(BigInt(event.topics[3]));
        tokenIds.push(id);
        const approveCallData = publicStakingBase.interface.encodeFunctionData(
          "approve",
          [validatorPoolAddress, id]
        );
        const approve = encodeMultiCallArgs(
          publicStakingAddress,
          0,
          approveCallData
        );
        approveTokens.push(approve);
      }
    }

    if (taskArgs.test) {
      await hre.network.provider.send("hardhat_mine", [
        hre.ethers.utils.hexValue(3),
      ]);
    } else {
      await tx.wait(3);
    }

    const regValidatorsCallData =
      validatorPoolBase.interface.encodeFunctionData("registerValidators", [
        validatorAddresses,
        tokenIds,
      ]);
    const regValidators = encodeMultiCallArgs(
      validatorPoolAddress,
      0,
      regValidatorsCallData
    );
    tx = await factory.multiCall(
      [...approveTokens, regValidators],
      await getGasPrices(hre)
    );
    if (taskArgs.test) {
      await hre.network.provider.send("hardhat_mine", [
        hre.ethers.utils.hexValue(3),
      ]);
    } else {
      await tx.wait(3);
    }
  });

task("unregister-validators", "unregister validators")
  .addFlag("test")
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addVariadicPositionalParam(
    "addresses",
    "validators' addresses",
    undefined,
    types.string,
    false
  )
  .setAction(async (taskArgs, hre) => {
    console.log("Unregistering Validators\n", taskArgs.addresses);
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );

    // checking factory address
    factory
      .lookup(hre.ethers.utils.formatBytes32String("AToken"))
      .catch((error: any) => {
        throw new Error(
          `Invalid factory-address ${taskArgs.factoryAddress}!\n${error}`
        );
      });
    const validatorAddresses: string[] = taskArgs.addresses;
    console.log(validatorAddresses);
    // Make sure that admin is the named account at position 0
    const [admin] = await hre.ethers.getSigners();
    console.log(`Admin address: ${admin.address}`);

    const validatorPool = await hre.ethers.getContractAt(
      "ValidatorPool",
      await factory.lookup(
        hre.ethers.utils.formatBytes32String("ValidatorPool")
      )
    );
    console.log(`validatorPool Address: ${validatorPool.address}`);

    const input = validatorPool.interface.encodeFunctionData(
      "unregisterValidators",
      [validatorAddresses]
    );
    const recpt = await (
      await factory.connect(admin).callAny(validatorPool.address, 0, input)
    ).wait(3);
    if (recpt.status !== 1) {
      throw new Error(`Receipt indicates failure: ${recpt}`);
    }
  });

task("ethdkg-input", "calculate the initializeETHDKG selector").setAction(
  async (taskArgs, hre) => {
    const { ethers } = hre;
    const iface = new ethers.utils.Interface(["function initializeETHDKG()"]);
    const input = iface.encodeFunctionData("initializeETHDKG");
    console.log("input", input);
  }
);

task("virtual-mint-deposit", "Virtually creates a deposit on the side chain")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set",
    undefined,
    types.string
  )
  .addParam(
    "depositOwnerAddress",
    "the address of the account that will have ownership over the newly created deposit",
    undefined,
    types.string
  )
  .addParam(
    "depositAmount",
    "Amount of ALCB to be deposited",
    undefined,
    types.string
  )
  .addParam(
    "accountType",
    "For ethereum based address use number: 1  For BN curve addresses user number: 2",
    1,
    types.int
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;
    const iface = new ethers.utils.Interface([
      "function virtualMintDeposit(uint8 accountType_,address to_,uint256 amount_)",
    ]);
    const input = iface.encodeFunctionData("virtualMintDeposit", [
      taskArgs.accountType,
      taskArgs.depositOwnerAddress,
      BigNumber.from(taskArgs.depositAmount),
    ]);
    const [admin] = await ethers.getSigners();
    const adminSigner = await ethers.getSigner(admin.address);
    const factory = await ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const alcb = await ethers.getContractAt(
      "BToken",
      await factory.lookup(hre.ethers.utils.formatBytes32String("BToken"))
    );
    const tx = await factory
      .connect(adminSigner)
      .callAny(alcb.address, 0, input);
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

task("schedule-maintenance", "Calls schedule Maintenance")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .setAction(async (taskArgs, hre) => {
    console.log(`scheduling maintenance after the next snapshot`);
    const { ethers } = hre;
    const iface = new ethers.utils.Interface([
      "function scheduleMaintenance()",
    ]);
    const input = iface.encodeFunctionData("scheduleMaintenance", []);

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
    console.log(`calling the contract and waiting receipt`);
    await (
      await factory
        .connect(adminSigner)
        .callAny(validatorPool.address, 0, input)
    ).wait(3);
  });

task(
  "aggregate-lockup-profits",
  "Aggregate the profits of the locked positions in the lockup contract"
)
  .addParam("factoryAddress", "the AliceNet factory address")
  .addFlag(
    "onlyOnce",
    "only execute aggregateProfits once instead of executing" +
      " it until is safe to unlock (very gas consuming)"
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;
    const factory = await ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const lockup = await ethers.getContractAt(
      "Lockup",
      await factory.lookup(ethers.utils.formatBytes32String("Lockup"))
    );
    let safeToUnlock = await lockup.payoutSafe();
    while (!safeToUnlock) {
      await (await lockup.aggregateProfits()).wait(8);
      safeToUnlock = await lockup.payoutSafe();
      console.log("Is safe to unlock: " + safeToUnlock);
      if (taskArgs.onlyOnce) break;
    }
    console.log("Done!");
  });

task(
  "create-bonus-pool-position",
  "Transfer and stake the ALCA that will be used to pay the bonus shares to the users that lock a position"
)
  .addParam("factoryAddress", "the AliceNet factory address")
  .addParam(
    "bonusAmount",
    "the amount of ALCA that will transferred and staked as bonus"
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;
    const factory = await ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const lockup = await ethers.getContractAt(
      "Lockup",
      await factory.lookup(ethers.utils.formatBytes32String("Lockup"))
    );
    const bonusPool = await ethers.getContractAt(
      "BonusPool",
      await lockup.getBonusPoolAddress()
    );
    const alca = await ethers.getContractAt(
      "AToken",
      await factory.lookup(ethers.utils.formatBytes32String("AToken"))
    );
    const transferCall = encodeMultiCallArgs(
      alca.address,
      0,
      alca.interface.encodeFunctionData("transfer", [
        bonusPool.address,
        ethers.utils.parseEther(taskArgs.bonusAmount),
      ])
    );
    const createBonusStakeCall = encodeMultiCallArgs(
      bonusPool.address,
      0,
      bonusPool.interface.encodeFunctionData("createBonusStakedPosition")
    );
    await (
      await factory.multiCall([transferCall, createBonusStakeCall])
    ).wait(8);
  });

task(
  "pause-ethdkg-arbitrary-height",
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

task(
  "change-interval-to-evict-validators",
  "Task to change the interval to evict validators from the validator pool in case of no snapshots"
)
  .addParam(
    "interval",
    "The block internal without snapshots to evict validators from the validator pool"
  )
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;
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
    const input = validatorPool.interface.encodeFunctionData(
      "setMaxIntervalWithoutSnapshots",
      [taskArgs.interval]
    );
    await (
      await factory
        .connect(adminSigner)
        .callAny(validatorPool.address, 0, input)
    ).wait();
  });

task(
  "unregister-all-validators",
  "Task to unregister all validators (in case of pause-consensus-on-arbitrary-height)"
)
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;
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
    const input = validatorPool.interface.encodeFunctionData(
      "unregisterAllValidators"
    );
    await (
      await factory
        .connect(adminSigner)
        .callAny(validatorPool.address, 0, input)
    ).wait();
  });

task("change-dynamic-value", "Change a certain dynamic value")
  .addParam("factoryAddress", "the alicenet factory address")
  .addOptionalParam(
    "relativeEpoch",
    "How many epochs from the value will be updated on the side chain"
  )
  .addOptionalParam("maxBlockSize", "new max block size value")
  .addOptionalParam("proposalTimeout", "new proposal Timeout value")
  .addOptionalParam("preVoteTimeout", "new preVote Timeout value")
  .addOptionalParam("preCommitTimeout", "new preCommit Timeout value")
  .addOptionalParam("dataStoreFee", "new preVote Timeout value")
  .addOptionalParam("valueStoreFee", "new preVote Timeout value")
  .addOptionalParam(
    "minScaledTransactionFee",
    "new minScaledTransaction fee value"
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;
    const [admin] = await ethers.getSigners();
    const adminSigner = await ethers.getSigner(admin.address);
    const factory = await ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const dynamics = await hre.ethers.getContractAt(
      "Dynamics",
      await factory.lookup(hre.ethers.utils.formatBytes32String("Dynamics"))
    );
    const currentValue = await dynamics.getLatestDynamicValues();
    const newValue = { ...currentValue };

    newValue.maxBlockSize =
      taskArgs.maxBlockSize !== undefined
        ? taskArgs.maxBlockSize
        : currentValue.maxBlockSize;

    newValue.proposalTimeout =
      taskArgs.proposalTimeout !== undefined
        ? taskArgs.proposalTimeout
        : currentValue.proposalTimeout;

    newValue.preVoteTimeout =
      taskArgs.preVoteTimeout !== undefined
        ? taskArgs.preVoteTimeout
        : currentValue.preVoteTimeout;

    newValue.preCommitTimeout =
      taskArgs.preCommitTimeout !== undefined
        ? taskArgs.preCommitTimeout
        : currentValue.preCommitTimeout;

    newValue.dataStoreFee =
      taskArgs.dataStoreFee !== undefined
        ? taskArgs.dataStoreFee
        : currentValue.dataStoreFee;

    newValue.valueStoreFee =
      taskArgs.valueStoreFee !== undefined
        ? taskArgs.valueStoreFee
        : currentValue.valueStoreFee;

    newValue.minScaledTransactionFee =
      taskArgs.minScaledTransactionFee !== undefined
        ? taskArgs.minScaledTransactionFee
        : currentValue.minScaledTransactionFee;

    let epoch;
    if (taskArgs.relativeEpoch !== undefined && taskArgs.relativeEpoch >= 2) {
      epoch = taskArgs.relativeEpoch;
    } else {
      epoch = 2;
      console.log(
        `Epoch not sent or it's less than minimum epoch allowed, therefore scheduling changes in 2 epochs from now.`
      );
    }

    const input = dynamics.interface.encodeFunctionData("changeDynamicValues", [
      epoch,
      newValue,
    ]);
    await (
      await factory
        .connect(adminSigner)
        .callAny(dynamics.address, 0, input, await getGasPrices(hre))
    ).wait(8);

    const allKeys = Object.keys(currentValue);
    const allValues = Object.values(newValue);
    const keys: string[] = [];
    const newValuesArray = [];
    for (let i = 0; i < allKeys.length; i++) {
      if (isNaN(parseFloat(allKeys[i]))) {
        keys.push(allKeys[i]);
        newValuesArray.push(allValues[i]);
      }
    }

    for (let i = 0; i < currentValue.length; i++) {
      console.log(
        `Changed dynamics value ${keys[i]} from ${currentValue[i]} to ${newValuesArray[i]}`
      );
    }
  });

task(
  "lookup-contract-address",
  "Task to get address of contract deployed by AliceNet factory"
)
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addParam("salt", "contract salt, usually the contract name")
  .setAction(async (taskArgs, hre) => {
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    console.log(
      await factory.lookup(hre.ethers.utils.formatBytes32String(taskArgs.salt))
    );
  });

task("initialize-ethdkg", "Start the ethdkg process")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;

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

    console.log("Initializing ETHDKG");
    const recpt = await (
      await factory
        .connect(adminSigner)
        .callAny(
          validatorPool.address,
          0,
          validatorPool.interface.encodeFunctionData("initializeETHDKG")
        )
    ).wait(3);
    console.log("ETHDKG trigger at block number:" + recpt.blockNumber);
  });

task("transfer-eth", "transfers eth from default account to receiver")
  .addParam("receiver", "address of the account to fund")
  .addParam("amount", "amount of eth to transfer")
  .setAction(async (taskArgs, hre) => {
    const accounts = await hre.ethers.getSigners();
    const ownerBal = await hre.ethers.provider.getBalance(accounts[0].address);
    const wei = BigNumber.from(parseInt(taskArgs.amount, 16)).mul(
      BigNumber.from("10").pow(BigInt(18))
    );
    const amount = wei;
    const target = taskArgs.receiver;
    console.log(`previous owner balance: ${ownerBal.toString()}`);
    let receiverBal = await hre.ethers.provider.getBalance(target);
    console.log(`previous receiver balance: ${receiverBal.toString()}`);
    const txRequest = await accounts[0].populateTransaction({
      from: accounts[0].address,
      value: amount,
      to: target,
    });
    const txResponse = await accounts[0].sendTransaction(txRequest);
    await txResponse.wait();
    receiverBal = await hre.ethers.provider.getBalance(target);
    console.log(`new receiver balance: ${receiverBal}`);
    const ownerBal2 = await hre.ethers.provider.getBalance(accounts[0].address);
    console.log(`new owner balance: ${ownerBal.sub(ownerBal2).toString()}`);
  });

task("mint-alca-To", "mints ALCA to an address")
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addParam("amount", "amount to mint")
  .addParam("to", "address of the recipient")
  .addOptionalParam("nonce", "nonce to send tx with")
  .setAction(async (taskArgs, hre) => {
    const signers = await hre.ethers.getSigners();
    const nonce =
      taskArgs.nonce === undefined
        ? hre.ethers.provider.getTransactionCount(signers[0].address)
        : taskArgs.nonce;
    const aTokenMinterBase = await hre.ethers.getContractFactory(
      "ATokenMinter"
    );
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const aTokenMinterAddr = await factory.callStatic.lookup(
      hre.ethers.utils.formatBytes32String("ATokenMinter")
    );
    const aToken = await hre.ethers.getContractAt(
      "AToken",
      await factory.callStatic.lookup(
        hre.ethers.utils.formatBytes32String("AToken")
      )
    );
    const bal1 = await aToken.callStatic.balanceOf(taskArgs.to);
    const calldata = aTokenMinterBase.interface.encodeFunctionData("mint", [
      taskArgs.to,
      taskArgs.amount,
    ]);
    // use the factory to call the A token minter
    const txResponse = await factory.callAny(aTokenMinterAddr, 0, calldata, {
      nonce,
    });
    await txResponse.wait();
    const bal2 = await aToken.callStatic.balanceOf(taskArgs.to);
    console.log(
      `Minted ${bal2.sub(bal1).toString()} to account ${taskArgs.to}`
    );
  });

task("get-alca-balance", "gets ALCA balance of account")
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addParam("account", "address of account to get balance of")
  .setAction(async (taskArgs, hre) => {
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const alca = await hre.ethers.getContractAt(
      "AToken",
      await factory.callStatic.lookup(
        hre.ethers.utils.formatBytes32String("AToken")
      )
    );
    const bal = await alca.callStatic.balanceOf(taskArgs.account);
    console.log(bal);
    return bal;
  });

task("mint-alcb-to", "mints ALCB to an address")
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addParam("amount", "amount to mint")
  .addParam("numWei", "amount of eth to use")
  .addParam("to", "address of the recipient")
  .setAction(async (taskArgs, hre) => {
    if (
      taskArgs.factoryAddress === undefined ||
      taskArgs.factoryAddress === ""
    ) {
      throw new Error("Expected a factory address to be passed!");
    }
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const alcb = await hre.ethers.getContractAt(
      "BToken",
      await factory.callStatic.lookup(
        hre.ethers.utils.formatBytes32String("BToken")
      )
    );
    const bal1 = await alcb.callStatic.balanceOf(taskArgs.to);
    const txResponse = await alcb.mintTo(taskArgs.to, taskArgs.amount, {
      value: taskArgs.numWei,
    });
    await txResponse.wait();
    const bal2 = await alcb.callStatic.balanceOf(taskArgs.to);
    console.log(
      `Minted ${bal2.sub(bal1).toString()} ALCB to account ${taskArgs.to}`
    );
  });

task("get-alcb-balance", "gets ALCB balance of account")
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addParam("account", "address of account to get balance of")
  .setAction(async (taskArgs, hre) => {
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const alcb = await hre.ethers.getContractAt(
      "BToken",
      await factory.callStatic.lookup(
        hre.ethers.utils.formatBytes32String("BToken")
      )
    );
    const bal = await alcb.callStatic.balanceOf(taskArgs.account);
    console.log(bal);
    return bal;
  });

task(
  "set-min-ethereum-blocks-per-snapshot",
  "Set the minimum number of ethereum blocks that we should wait between snapshots"
)
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addParam(
    "blockNum",
    "Minimum block of ethereum to wait between snapshots",
    -1,
    types.int
  )
  .setAction(async (taskArgs, hre) => {
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const snapshots = await hre.ethers.getContractAt(
      "Snapshots",
      await factory.callStatic.lookup(
        hre.ethers.utils.formatBytes32String("Snapshots")
      )
    );

    if (taskArgs.blockNum < 0) {
      throw new Error("block-num not passed or the value was smaller than 0!");
    }

    const [admin] = await hre.ethers.getSigners();
    const adminSigner = await hre.ethers.getSigner(admin.address);
    const input = snapshots.interface.encodeFunctionData(
      "setMinimumIntervalBetweenSnapshots",
      [taskArgs.blockNum]
    );
    console.log(
      `\nSetting the setMinimumIntervalBetweenSnapshots to ${taskArgs.blockNum}`
    );
    const rept = await (
      await factory.connect(adminSigner).callAny(snapshots.address, 0, input)
    ).wait(3);
    if (rept.status !== 1) {
      throw new Error(`Receipt indicates failure: ${rept}`);
    }
  });

task("get-eth-balance", "gets Ethereum balance of account")
  .addParam("account", "address of account to get balance of")
  .setAction(async (taskArgs, hre) => {
    const bal = await hre.ethers.provider.getBalance(taskArgs.account);
    console.log(bal);
    return bal;
  });

function notSoRandomNumBetweenRange(max: number, min: number): number {
  return Math.floor(Math.random() * (max - min + 1) + min);
}
// WARNING ONLY RUN THIS ON TESTNET TO TESTLOAD
// RUNNING THIS ON MAINNET WILL WASTE ALL YOUR ETH
task(
  "spam-ethereum",
  "inject a bunch of random transactions to simulate regular block usage"
)
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .setAction(async (taskArgs, hre) => {
    // this function deploys the snapshot contract
    const validatorPoolFactory = await hre.ethers.getContractFactory(
      "ValidatorPool"
    );
    const accounts = await hre.ethers.getSigners();
    const defaultVal = BigNumber.from("100000000000000000000");
    // fund accounts
    if ((await accounts[2].getBalance()).lt(defaultVal)) {
      const txRequest = await accounts[0].populateTransaction({
        from: accounts[0].address,
        // nonce: nonce0,
        value: BigNumber.from("100000000000000000000"),
        to: accounts[2].address,
        // gasPrice: gp,
      });
      const txResponse = await accounts[0].sendTransaction(txRequest);
      await txResponse.wait();
    }
    if ((await accounts[1].getBalance()).lt(defaultVal)) {
      const txRequest = await accounts[0].populateTransaction({
        from: accounts[0].address,
        // nonce: nonce0,
        value: BigNumber.from("100000000000000000000"),
        to: accounts[1].address,
        // gasPrice: gp,
      });
      const txResponse = await accounts[0].sendTransaction(txRequest);
      await txResponse.wait();
    }
    const fooTokenBase = await hre.ethers.getContractFactory("FooToken");
    const fooToken = await fooTokenBase.deploy();
    const minter = await hre.ethers.getContractFactory("Minter");
    const gasBomb = async () => {
      return await minter.deploy(fooToken.address, {
        gasLimit: 30000000n,
        gasPrice: 10000n * 10n ** 9n,
      });
    };
    // function to deploy a contract
    const deployContract = async () => {
      const gp = await hre.ethers.provider.getGasPrice();
      await validatorPoolFactory.deploy({
        nonce: await accounts[0].getTransactionCount("pending"),
        gasPrice: gp,
      });
    };
    // function to send eth back and forth
    const sendEth = async () => {
      const transactions: Array<ContractTransaction> = [];
      const wei = 1;
      const gp = await hre.ethers.provider.getGasPrice();
      let txRequest = await accounts[2].populateTransaction({
        from: accounts[2].address,
        nonce: await accounts[2].getTransactionCount("pending"),
        value: wei,
        to: accounts[1].address,
        gasPrice: gp,
      });
      try {
        transactions.push(await accounts[2].sendTransaction(txRequest));
      } catch {}
      txRequest = await accounts[1].populateTransaction({
        from: accounts[1].address,
        nonce: await accounts[1].getTransactionCount("pending"),
        value: wei,
        to: accounts[2].address,
        gasPrice: gp,
      });
      try {
        transactions.push(await accounts[1].sendTransaction(txRequest));
      } catch {}
      return transactions;
    };
    const mintAToken = async () => {
      return mintATokenTo(
        hre,
        taskArgs.factoryAddress,
        accounts[1].address,
        await accounts[0].getTransactionCount("pending")
      );
    };
    // const setBaseFee = async () => {
    //   const increase = notSoRandomNumBetweenRange(6000, 1);
    //   const baseFee = increase * 1000000000;
    //   await hre.network.provider.send("hardhat_setNextBlockBaseFeePerGas", [
    //     "0x" + baseFee.toString(16),
    //   ]);
    // };
    const txSet: Array<ContractTransaction> = [];
    let txSent = 0;

    let previousBlockNum = 0;
    let blocknum = await hre.ethers.provider.getBlockNumber();
    while (1) {
      blocknum = await hre.ethers.provider.getBlockNumber();
      if (blocknum > previousBlockNum) {
        await gasBomb();
        previousBlockNum = blocknum;
      }
      if (txSent > 15) {
        await Promise.all(txSet);
        txSent = 0;
      } else {
        txSent++;
        const bal1 = await hre.ethers.provider.getBalance(accounts[1].address);
        const bal2 = await hre.ethers.provider.getBalance(accounts[2].address);
        const type = notSoRandomNumBetweenRange(3, 0);
        console.log(
          `tx type: ${type}, account1: ${bal1.toString()}, account2: ${bal2.toString()}`
        );
        switch (type) {
          case 0:
            try {
              await deployContract();
            } catch (error) {}
            break;
          case 1:
            try {
              const tx = await sendEth();
              txSet.push(...tx);
              // nonce0++;
              // nonce1++;
            } catch (error) {}
            break;
          case 2:
            try {
              const tx = await mintAToken();
              txSet.push(tx);
            } catch (error) {}
            break;
          case 3:
            try {
              await gasBomb();
            } catch (error) {}
            break;
          default:
            break;
        }
      }
    }
  });

task("fund-validators", "manually put 100 eth in each validator account")
  .addOptionalParam(
    "configPath",
    "path to validator configs dir",
    "./../scripts/generated/config"
  )
  .setAction(async (taskArgs, hre) => {
    console.log("\nFunding validators");
    const signers = await hre.ethers.getSigners();
    const configPath = taskArgs.configPath;
    let validatorConfigs: Array<string> = [];
    // get all the validator address from their toml config file, possibly check if generated is there
    validatorConfigs = fs.readdirSync(configPath);
    // extract the address out of each validator config file
    const accounts: Array<string> = [];
    validatorConfigs.forEach((val) => {
      accounts.push(getValidatorAccount(`${configPath}/${val}`));
    });

    const minAmount = 90n;
    const maxAmount = 100n;
    for (const account of accounts) {
      const bal = await hre.ethers.provider.getBalance(account);
      if (bal.lt(hre.ethers.utils.parseEther(minAmount.toString()))) {
        const txResponse = await signers[0].sendTransaction({
          to: account,
          value: hre.ethers.utils.parseEther(
            (maxAmount - bal.toBigInt()).toString()
          ),
        });
        await txResponse.wait();
        console.log(
          `account ${account} now has ${await hre.ethers.provider.getBalance(
            account
          )} ether`
        );
      }
    }
  });

function getValidatorAccount(path: string): string {
  const data = fs.readFileSync(path);
  const config: any = toml.parse(data.toString());
  return config.ethereum.defaultAccount;
}

task("get-gas-cost", "gets the current gas cost")
  .addFlag("ludicrous", "over inflate certain blocks")
  .setAction(async (taskArgs, hre) => {
    let lastBlock = 0;
    while (1) {
      const gasPrice = await hre.ethers.provider.getGasPrice();
      await delay(7000);
      const blocknum = await hre.ethers.provider.blockNumber;
      // console.log(`gas price @ blocknum ${blocknum.toString()}: ${gasPrice.toString()}`);
      if (blocknum > lastBlock) {
        console.log(
          `gas price @ blocknum ${blocknum.toString()}: ${gasPrice.toString()}`
        );
      }
      lastBlock = blocknum;
    }
  });

task(
  "set-local-environment-interval-mining",
  "sets the local environment node to mine on a interval and automine off"
)
  .addFlag("enableAutoMine")
  .addOptionalParam("interval", "time between blocks", "15000")
  .setAction(async (taskArgs, hre) => {
    const network = await hre.ethers.provider.getNetwork();
    const interval = parseInt(taskArgs.interval, 10);
    if (network.chainId === 1337) {
      if (taskArgs.enableAutoMine) {
        try {
          await hre.network.provider.send("evm_setAutomine", [true]);
        } catch (error) {}
      } else {
        try {
          await hre.network.provider.send("evm_setIntervalMining", [interval]);
          await hre.network.provider.send("evm_setAutomine", [false]);
        } catch (error) {}
      }
    }
  });

task(
  "set-local-environment-base-fee",
  "sets the local environment node base fee for the next block"
)
  .addParam("baseFee", "base fee value in GWEIs", "500", types.int)
  .setAction(async (taskArgs, hre) => {
    const network = await hre.ethers.provider.getNetwork();
    const baseFee = BigInt(taskArgs.baseFee) * 10n ** 9n;
    if (network.chainId === 1337) {
      try {
        await hre.network.provider.send("hardhat_setNextBlockBaseFeePerGas", [
          "0x" + baseFee.toString(16),
        ]);
      } catch (error) {}
    }
  });

task("update-alicenet-node-version", "Set the Canonical AliceNet Node Version")
  .addParam("factoryAddress", "address of the factory deploying the contract")
  .addParam(
    "relativeEpoch",
    "relativeEpoch Canonical AliceNet version",
    -1,
    types.int
  )
  .addParam("major", "Major Canonical AliceNet version", -1, types.int)
  .addParam("minor", "Minor Canonical AliceNet version", -1, types.int)
  .addParam("patch", "Patch Canonical AliceNet version", -1, types.int)
  .addParam(
    "binaryHash",
    "BinaryHash Canonical AliceNet version",
    "",
    types.string
  )
  .setAction(async (taskArgs, hre) => {
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const dynamics = await hre.ethers.getContractAt(
      "Dynamics",
      await factory.callStatic.lookup(
        hre.ethers.utils.formatBytes32String("Dynamics")
      )
    );

    if (taskArgs.relativeEpoch < 2) {
      throw new Error(
        "relativeEpoch parameter not sent or the value was smaller than 2!"
      );
    }

    if (taskArgs.major < 0) {
      throw new Error(
        "major version parameter not sent or the value was smaller than 0!"
      );
    }

    if (taskArgs.minor < 0) {
      throw new Error(
        "minor parameter not sent or the value was smaller than 0!"
      );
    }

    if (taskArgs.patch < 0) {
      throw new Error(
        "patch parameter not sent or the value was smaller than 0!"
      );
    }

    if (!taskArgs.binaryHash) {
      throw new Error("binaryHash parameter not sent!");
    }

    const [admin] = await hre.ethers.getSigners();
    const adminSigner = await hre.ethers.getSigner(admin.address);
    const input = dynamics.interface.encodeFunctionData(
      "updateAliceNetNodeVersion",
      [
        taskArgs.relativeEpoch,
        taskArgs.major,
        taskArgs.minor,
        taskArgs.patch,
        hre.ethers.utils.formatBytes32String(taskArgs.binaryHash),
      ]
    );
    console.log(
      `Updating the updateAliceNetNodeVersion to ${taskArgs.major}.${taskArgs.minor}.${taskArgs.patch}`
    );
    const rept = await (
      await factory.connect(adminSigner).callAny(dynamics.address, 0, input)
    ).wait(3);
    if (rept.status !== 1) {
      throw new Error(`Receipt indicates failure: ${rept}`);
    }
    console.log("Done");
  });

task("deploy-alcb", "Task to deploy ALCB")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .setAction(async (taskArgs, hre) => {
    const factory = await hre.ethers.getContractAt(
      "AliceNetFactory",
      taskArgs.factoryAddress
    );
    const ALCB_BASE = await hre.ethers.getContractFactory("BToken");
    const deploymentCode = ALCB_BASE.getDeployTransaction(factory.address)
      .data as BytesLike;
    const tx = await factory.deployCreate(deploymentCode);
    const receipt = await tx.wait();
    const alcbAddress = getEventVar(receipt, DEPLOYED_RAW, CONTRACT_ADDR);
    console.log("ALCB/BToken address: ", alcbAddress);
    await factory.addNewExternalContract(
      hre.ethers.utils.formatBytes32String("BToken"),
      alcbAddress
    );
  });

async function mintATokenTo(
  hre: HardhatRuntimeEnvironment,
  factoryAddress: string,
  to: string,
  nonce: number
): Promise<ContractTransaction> {
  const aTokenMinterBase = await hre.ethers.getContractFactory("ATokenMinter");
  const factory = await hre.ethers.getContractAt(
    "AliceNetFactory",
    factoryAddress
  );
  const aTokenMinterAddr = await factory.callStatic.lookup(
    hre.ethers.utils.formatBytes32String("ATokenMinter")
  );

  const calldata = aTokenMinterBase.interface.encodeFunctionData("mint", [
    to,
    1,
  ]);
  // use the factory to call the A token minter
  return factory.callAny(aTokenMinterAddr, 0, calldata, { nonce });
}

export function encodeMultiCallArgs(
  targetAddress: string,
  value: BigNumberish,
  callData: BytesLike
): MultiCallArgsStruct {
  const output: MultiCallArgsStruct = {
    target: targetAddress,
    value,
    data: callData,
  };
  return output;
}
export async function stakeValidators(
  numValidators: number,
  factoryAddress: string,
  aTokenAddress: string,
  publicStakingAddress: string,
  hre: HardhatRuntimeEnvironment
): Promise<ContractTransaction> {
  const factory = await hre.ethers.getContractAt(
    "AliceNetFactory",
    factoryAddress
  );
  const atokenBase = await hre.ethers.getContractFactory("AToken");
  const publicStakingBase = await hre.ethers.getContractFactory(
    "PublicStaking"
  );
  const stakeAmt = hre.ethers.utils.parseEther("5000000");
  const approveATokenCallData = atokenBase.interface.encodeFunctionData(
    "approve",
    [publicStakingAddress, BigNumber.from(numValidators).mul(stakeAmt)]
  );
  const approveAToken = encodeMultiCallArgs(
    aTokenAddress,
    0,
    approveATokenCallData
  );
  const stakeNFT: Array<MultiCallArgsStruct> = [];
  for (let i = 0; i < numValidators; i++) {
    const stakeToken = publicStakingBase.interface.encodeFunctionData("mint", [
      stakeAmt,
    ]);
    stakeNFT.push(encodeMultiCallArgs(publicStakingAddress, 0, stakeToken));
  }
  return factory.multiCall(
    [approveAToken, ...stakeNFT],
    await getGasPrices(hre)
  );
}

export async function migrateSnapshotsAndValidators(
  factoryAddress: string,
  publicStakingAddress: string,
  snapshotAddress: string,
  ethDKGAddress: string,
  validatorPoolAddress: string,
  tokenIDs: Array<BigNumber>,
  validatorAccounts: Array<string>,
  validatorIndexes: Array<number> | Array<string>,
  validatorShares: Array<Array<string>>,
  ethHeight: number,
  masterPublicKey: Array<BigNumber> | Array<string>,
  bClaims: Array<string>,
  groupSignatures: Array<string>,
  hre: HardhatRuntimeEnvironment
): Promise<ContractTransaction> {
  const factory = await hre.ethers.getContractAt(
    "AliceNetFactory",
    factoryAddress
  );
  const validatorPoolBase = await hre.ethers.getContractFactory(
    "ValidatorPool"
  );
  const ethdkgBase = await hre.ethers.getContractFactory("ETHDKG");
  const snapshotBase = await hre.ethers.getContractFactory("Snapshots");
  const publicStakingBase = await hre.ethers.getContractFactory(
    "PublicStaking"
  );
  // approve token transfer
  const approveTokens: Array<MultiCallArgsStruct> = [];
  for (const tokenID of tokenIDs) {
    const approveCallData = publicStakingBase.interface.encodeFunctionData(
      "approve",
      [validatorPoolAddress, tokenID]
    );
    const approve = encodeMultiCallArgs(
      publicStakingAddress,
      0,
      approveCallData
    );
    approveTokens.push(approve);
  }
  // register validators
  const validatorCount = 4;
  const epoch = 0;
  const sideChainHeight = 0;
  const registerValidatorsCallData =
    validatorPoolBase.interface.encodeFunctionData("registerValidators", [
      validatorAccounts,
      tokenIDs,
    ]);
  const registerValidators = encodeMultiCallArgs(
    validatorPoolAddress,
    0,
    registerValidatorsCallData
  );
  const migrateValidatorsCallData = ethdkgBase.interface.encodeFunctionData(
    "migrateValidators",
    [
      validatorAccounts,
      validatorIndexes,
      validatorShares,
      validatorCount,
      epoch,
      sideChainHeight,
      ethHeight,
      masterPublicKey,
    ]
  );
  const migrateValidators = encodeMultiCallArgs(
    ethDKGAddress,
    0,
    migrateValidatorsCallData
  );
  const migrateSnapshotsCallData = snapshotBase.interface.encodeFunctionData(
    "migrateSnapshots",
    [groupSignatures, bClaims]
  );
  const migrateSnapshots = encodeMultiCallArgs(
    snapshotAddress,
    0,
    migrateSnapshotsCallData
  );
  return factory.multiCall(
    [...approveTokens, registerValidators, migrateValidators, migrateSnapshots],
    await getGasPrices(hre)
  );
}

async function factoryLookupAddress(
  factoryAdress: string,
  salt: string,
  hre: HardhatRuntimeEnvironment
): Promise<string> {
  const factory = await hre.ethers.getContractAt(
    "AliceNetFactory",
    factoryAdress
  );
  return factory
    .lookup(hre.ethers.utils.formatBytes32String(salt))
    .catch((error: any) => {
      throw new Error(`Invalid factory-address ${factory.address}!\n${error}`);
    });
}
