import { ethers } from "hardhat";

async function main() {
  const contractFactory = await ethers.getContractFactory("ALCB");
  const initCallData =
    contractFactory.interface.encodeFunctionData("initialize");
  console.log(initCallData);
}

main()
  .then(() => {
    return 0;
  })
  .catch((error) => {
    console.error(error);
    throw new Error("unexpected error");
  });
