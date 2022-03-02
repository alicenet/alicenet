import { ethers } from "hardhat";

async function main() {
  let contractFactory = await ethers.getContractFactory("MadByte");
  let initCallData = contractFactory.interface.encodeFunctionData("initialize");
  console.log(initCallData);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
