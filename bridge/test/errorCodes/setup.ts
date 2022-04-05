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
  expect(parseInt(code)).to.be.equal(expectedNumber);
};
