import { HardhatRuntimeEnvironment } from "hardhat/types";
import {
    abi as FACTORY_ABI,
    bytecode as FACTORY_BYTECODE,
  } from '@uniswap/v3-core/artifacts/contracts/UniswapV3Factory.sol/UniswapV3Factory.json'
import { task, types } from "hardhat/config";

const UNISWAPV3FACTORY = "0x1F98431c8aD98523631AE4a59f267346ea31F984";
task("deployUniswapV3Pool", "deploys a uniswap v3 pool")
    .addParam("token0", "address of token0")
    .addParam("token1", "address of token1")
    .addParam("fee", "fee of the pool")
    .setAction(async (taskArgs, hre) => {
        const { token0, token1, fee } = taskArgs;
        console.log("Deploying pool for", token0, token1, fee)

        const factory = await hre.ethers.getContractAt(
            FACTORY_ABI,
            UNISWAPV3FACTORY
        );

        const txResponse = await factory.createPool(token0, token1, fee);
        const receipt = await txResponse.wait();
        
    }
