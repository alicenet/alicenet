import { Contract } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";

export const deployLibrary = async (libName: string): Promise<Contract> => {
  const _Contract = await ethers.getContractFactory(libName);
  const instance = await _Contract.deploy();
  await instance.deployed();
  return instance;
};

export const assertConstantReturnsCorrectErrorCode = async (
  constant: any,
  expectedNumber: number
) => {
  const code = await constant();
  const codeString = ethers.utils.parseBytes32String(code);

  expect(parseInt(codeString)).to.be.equal(expectedNumber);
};
