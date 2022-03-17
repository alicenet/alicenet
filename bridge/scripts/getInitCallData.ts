import { ethers } from "hardhat";

async function main() {
  const contractFactory = await ethers.getContractFactory("MadByte");
  const initCallData =
    contractFactory.interface.encodeFunctionData("initialize");
  console.log(initCallData);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
