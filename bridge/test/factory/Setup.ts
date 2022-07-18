// const { contracts } from"@openzeppelin/cli/lib/prompts/choices");
import { expect } from "chai";
import {
  BytesLike,
  ContractFactory,
  ContractReceipt,
  ContractTransaction,
} from "ethers";
import { ethers } from "hardhat";
import { ALICENET_FACTORY, END_POINT, MOCK } from "../../scripts/lib/constants";
import { AliceNetFactory } from "../../typechain-types";

export async function getAccounts() {
  const signers = await ethers.getSigners();
  const accounts = [];
  for (const signer of signers) {
    accounts.push(signer.address);
  }
  return accounts;
}

export async function predictFactoryAddress(ownerAddress: string) {
  const txCount = await ethers.provider.getTransactionCount(ownerAddress);
  // console.log(txCount)
  const futureFactoryAddress = ethers.utils.getContractAddress({
    from: ownerAddress,
    nonce: txCount,
  });
  return futureFactoryAddress;
}

export async function proxyMockLogicTest(
  contract: ContractFactory,
  salt: string,
  proxyAddress: string,
  mockLogicAddr: string,
  endPointAddr: string,
  factoryAddress: string
) {
  const endPointFactory = await ethers.getContractFactory(END_POINT);
  const factoryBase = await ethers.getContractFactory(ALICENET_FACTORY);
  const factory = factoryBase.attach(factoryAddress);
  const mockProxy = contract.attach(proxyAddress);
  let txResponse = await mockProxy.setFactory(factoryAddress);
  await txResponse.wait();
  const testArg = 4;
  await expectTxSuccess(txResponse);
  const fa = await mockProxy.callStatic.getFactory();
  expect(fa).to.equal(factoryAddress);
  txResponse = await mockProxy.setV(testArg);
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  let v = await mockProxy.callStatic.getVar();
  expect(v.toNumber()).to.equal(testArg);
  let i = await mockProxy.callStatic.getImut();
  expect(i.toNumber()).to.equal(2);
  // upgrade the proxy
  txResponse = await factory.upgradeProxy(salt, endPointAddr, "0x");
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  // endpoint interface connected to proxy address
  const proxyEndpoint = endPointFactory.attach(proxyAddress);
  i = await proxyEndpoint.i();
  let num = i.toNumber() + 2;
  txResponse = await proxyEndpoint.addTwo();
  await txResponse.wait();
  let test = await getEventVar(txResponse, "AddedTwo", "i");
  expect(test.toNumber()).to.equal(num);
  // lock the proxy upgrade
  txResponse = await factory.upgradeProxy(salt, mockLogicAddr, "0x");
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  txResponse = await mockProxy.setV(testArg + 2);
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  v = await mockProxy.getVar();
  await txResponse.wait();
  expect(v.toNumber()).to.equal(testArg + 2);
  // lock the upgrade functionality
  txResponse = await mockProxy.lock();
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  const txRes = factory.upgradeProxy(salt, endPointAddr, "0x");
  await expect(txRes).to.be.revertedWith(
    "reverted with an unrecognized custom error"
  );
  // unlock the proxy
  txResponse = await mockProxy.unlock();
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  txResponse = await factory.upgradeProxy(salt, endPointAddr, "0x");
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  i = await proxyEndpoint.i();
  num = i.toNumber() + 2;
  txResponse = await proxyEndpoint.addTwo();
  await txResponse.wait();
  test = await getEventVar(txResponse, "AddedTwo", "i");
  expect(test.toNumber()).to.equal(num);
}

export async function metaMockLogicTest(
  contract: ContractFactory,
  address: string,
  factoryAddress: string
) {
  const Contract = contract.attach(address);
  let txResponse = await Contract.setFactory(factoryAddress);
  const test = 4;
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  const fa = await Contract.getFactory.call();
  expect(fa).to.equal(factoryAddress);
  txResponse = await Contract.setV(test);
  await txResponse.wait();
  await expectTxSuccess(txResponse);
  const v = await Contract.getVar();
  expect(v.toNumber()).to.equal(test);
  const i = await Contract.getImut();
  expect(i.toNumber()).to.equal(2);
}

export async function getEventVar(
  txResponse: ContractTransaction,
  eventName: string,
  varName: string
) {
  let result: any;
  const receipt = await txResponse.wait();
  if (receipt.events !== undefined) {
    const events = receipt.events;
    for (let i = 0; i < events.length; i++) {
      // look for the event
      if (events[i].event === eventName) {
        if (events[i].args !== undefined) {
          const args = events[i].args;
          // extract the deployed mock logic contract address from the event
          result = args !== undefined ? args[varName] : undefined;
          if (result !== undefined) {
            return result;
          }
        } else {
          throw new Error(
            `failed to extract ${varName} from event: ${eventName}`
          );
        }
      }
    }
  }
  throw new Error(`failed to find event: ${eventName}`);
}

export async function getReceiptEventVar(
  receipt: ContractReceipt,
  eventName: string,
  varName: string
) {
  let result: any;
  if (receipt.events !== undefined) {
    const events = receipt.events;
    for (let i = 0; i < events.length; i++) {
      // look for the event
      if (events[i].event === eventName) {
        if (events[i].args !== undefined) {
          const args = events[i].args;
          // extract the deployed mock logic contract address from the event
          result = args !== undefined ? args[varName] : undefined;
          if (result !== undefined) {
            return result;
          }
        } else {
          throw new Error(
            `failed to extract ${varName} from event: ${eventName}`
          );
        }
      }
    }
  }
  throw new Error(`failed to find event: ${eventName}`);
}

export async function expectTxSuccess(txResponse: ContractTransaction) {
  const receipt = await txResponse.wait();
  expect(receipt.status).to.equal(1);
}

export function getCreateAddress(Address: string, nonce: number) {
  return ethers.utils.getContractAddress({
    from: Address,
    nonce,
  });
}
export function bytes32ArrayToStringArray(bytes32Array: Array<any>) {
  const ret = [];
  for (let i = 0; i < bytes32Array.length; i++) {
    ret.push(ethers.utils.parseBytes32String(bytes32Array[i]));
  }
  return ret;
}

export function getSalt() {
  // set a new salt
  const salt = new Date();
  // use the time as the salt
  const Salt = salt.getTime();
  return ethers.utils.formatBytes32String(Salt.toString());
}

export async function getDeployTemplateArgs(contractName: string) {
  const contract = await ethers.getContractFactory(contractName);
  const deployByteCode = contract.getDeployTransaction();
  return deployByteCode.data as BytesLike;
}

export type DeployStaticArgs = {
  salt: string;
  initCallData: string;
};

export async function getDeployStaticArgs(
  contractName: string,
  argsArray: Array<any>
) {
  const contract = await ethers.getContractFactory(contractName);
  const ret: DeployStaticArgs = {
    salt: getSalt(),
    initCallData: contract.interface.encodeFunctionData(
      "initialize",
      argsArray
    ),
  };
  return ret;
}

export async function checkMockInit(target: string, initVal: number) {
  const mockFactory = await ethers.getContractFactory(MOCK);
  const mock = await mockFactory.attach(target);
  const i = await mock.getImut();
  expect(i.toNumber()).to.equal(initVal);
}

export async function deployFactory() {
  const factoryBase = await ethers.getContractFactory(ALICENET_FACTORY);
  const accounts = await getAccounts();
  const firstOwner = accounts[0];
  // gets the initial transaction count for the address
  const transactionCount = await ethers.provider.getTransactionCount(
    firstOwner
  );
  // pre calculate the address of the factory contract
  const futureFactoryAddress = ethers.utils.getContractAddress({
    from: firstOwner,
    nonce: transactionCount,
  });
  // deploy the factory with its address as a constructor input
  const factory = await factoryBase.deploy(futureFactoryAddress);
  expect(factory.address).to.equal(futureFactoryAddress);
  return factory;
}

export async function deployCreate2Initable(
  factory: AliceNetFactory,
  salt: BytesLike
) {
  const mockInitFactory = await ethers.getContractFactory("MockInitializable");
  const txResponse = await factory.deployCreate2(
    0,
    salt,
    mockInitFactory.bytecode
  );
  return txResponse;
}

export function getMetamorphicAddress(factoryAddress: string, salt: string) {
  const initCode = "0x6020363636335afa1536363636515af43d36363e3d36f3";
  return ethers.utils.getCreate2Address(
    factoryAddress,
    salt,
    ethers.utils.keccak256(initCode)
  );
}
