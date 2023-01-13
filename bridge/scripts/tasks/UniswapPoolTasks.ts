import { abi as UNIV3_FACTORY_ABI } from "@uniswap/v3-core/artifacts/contracts/UniswapV3Factory.sol/UniswapV3Factory.json";
import { abi as POOL_ABI } from "@uniswap/v3-core/artifacts/contracts/UniswapV3Pool.sol/UniswapV3Pool.json";
import { abi as nonFungiblePositionManager_ABI } from "@uniswap/v3-periphery/artifacts/contracts/NonfungiblePositionManager.sol/NonfungiblePositionManager.json";
import { task } from "hardhat/config";
import { BigintIsh, sqrt } from '@uniswap/sdk-core'
import JSBI from 'jsbi'

const UNISWAPV3FACTORY = "0x1F98431c8aD98523631AE4a59f267346ea31F984";
const UNISWAPV3NFTMANAGER = "0xC36442b4a4522E871399CD717aBDD847Ab11FE88";

export function encodeSqrtRatioX96(amount1: BigintIsh, amount0: BigintIsh): JSBI {
  const numerator = JSBI.leftShift(JSBI.BigInt(amount1), JSBI.BigInt(192))
  const denominator = JSBI.BigInt(amount0)
  const ratioX192 = JSBI.divide(numerator, denominator)
  return sqrt(ratioX192)
}
task("deployUniswapV3Pool", "deploys a uniswap v3 pool")
  .addParam("token0", "address of token0")
  .addParam("token1", "address of token1")
  .addParam("amount0", "amount of token0 not in wei")
  .addParam("amount1", "amount of token1")
  .addParam("fee", "fee of the pool")
  .setAction(async (taskArgs, hre) => {
    //get instance of multi call contract
    const amount0 = hre.ethers.utils.parseEther(taskArgs.amount0);
    const amount1 = hre.ethers.utils.parseEther(taskArgs.amount1);
    const sqrtRatioX96 = encodeSqrtRatioX96(amount1, amount0)
    //approve NonFungiblePositionManager contract to spend erc token
    // on the pool, call the mint function with either eth amout of token amout
    //createAndInitializePoolIFNecessary
    //mint
    const token0 = taskArgs.token0;
    const token1 = taskArgs.token1;
    let tokenA
    let tokenB
    if (taskArgs.token0 < taskArgs.token1) {
      tokenA = token0
      tokenB = token1
    }else {
      tokenB = token0
      tokenA = token1
    }

    const UNIV3factory = await hre.ethers.getContractAt(
      UNIV3_FACTORY_ABI,
      UNISWAPV3FACTORY
    );
    const NFTPositionManager = await hre.ethers.getContractAt(
      nonFungiblePositionManager_ABI,
      UNISWAPV3NFTMANAGER
    );
    const multiCallData = [];
    const sqrtPriceX96 = 0;
    NFTPositionManager.createAndInitializePoolIfNecessary();
    const createAndInitializePool = NFTPositionManager.interface.encodeFunctionData(
      "createAndInitializePoolIfNecessary",
      [token0, token1, fee, sqrtPriceX96]
    );
    multiCallData.push(createAndInitializePool);
    NFTPositionManager.mint();
    const mintPosition = NFTPositionManager.interface.encodeFunctionData("mint", []);
    const txResponse = await UNIV3factory.createPool(token0, token1, fee);
    const receipt = await txResponse.wait();
  });

task("add_Liquity", "add liquidity to a uniswap v3 pool")
  .addParam("poolAddress", "address of the pool")
  .addParam("token0", "address of token0")
  .addParam("token1", "address of token1")
  .addParam("fee", "fee of the pool")
  .addParam("amount0", "amount of token0 not in wei")
  .addParam("amount1", "amount of token1")
  .setAction(async (taskArgs, hre) => {
    const amount0 = hre.ethers.utils.parseEther(taskArgs.amount0);
    const amount1 = hre.ethers.utils.parseEther(taskArgs.amount1);

    const NFTPositionManager = await hre.ethers.getContractAt(
      nonFungiblePositionManager_ABI,
      UNISWAPV3NFTMANAGER
    );
    const pool = await hre.ethers.getContractAt(POOL_ABI, taskArgs.poolAddress);

    const txResponse = await pool.mint(
      amount0, ,
      hre.ethers.utils.parseUnits(amount1, 18),
      0,
      0,
      hre.ethers.constants.AddressZero
    );
    const receipt = await txResponse.wait();
  });
