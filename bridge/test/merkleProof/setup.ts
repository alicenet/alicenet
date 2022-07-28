import { ethers } from "hardhat";
import { MockMerkleProofLibrary } from "../../typechain-types";

export const deployLibrary = async (): Promise<MockMerkleProofLibrary> => {
  const _Contract = await ethers.getContractFactory("MockMerkleProofLibrary");
  const instance = await _Contract.deploy();
  await instance.deployed();
  return instance as MockMerkleProofLibrary;
};
