import { MadnetFactory } from '../../typechain-types/MadnetFactory';
import { ContractFactory, ContractTransaction } from 'ethers';
//const { contracts } from"@openzeppelin/cli/lib/prompts/choices");
import { expect } from "chai";
import { BytesLike } from "ethers";
import { ethers } from "hardhat";
import {
  MOCK,
  END_POINT,
  MADNET_FACTORY,
} from '../../scripts/lib/constants';


export async function getAccounts() {
  let signers = await ethers.getSigners();
  let accounts = [];
  for (let signer of signers) {
    accounts.push(signer.address);
  }
  return accounts;
}

export async function predictFactoryAddress(ownerAddress: string) {
  let txCount = await ethers.provider.getTransactionCount(ownerAddress);
  // console.log(txCount)
  let futureFactoryAddress = ethers.utils.getContractAddress({
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
  const factoryBase = await ethers.getContractFactory(MADNET_FACTORY);
  const factory = factoryBase.attach(factoryAddress);
  const mockProxy = contract.attach(proxyAddress);
  let txResponse = await mockProxy.setFactory(factoryAddress);
  let receipt = await txResponse.wait();
  const testArg = 4;
  await expectTxSuccess(txResponse);
  let fa = await mockProxy.callStatic.getFactory();
  expect(fa).to.equal(factoryAddress);
  txResponse = await mockProxy.setv(testArg);
  receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  let v = await mockProxy.callStatic.v();
  expect(v.toNumber()).to.equal(testArg);
  let i = await mockProxy.callStatic.i();
  expect(i.toNumber()).to.equal(2);
  //upgrade the proxy
  txResponse = await factory.upgradeProxy(salt, endPointAddr, "0x");
  receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  //endpoint interface connected to proxy address
  let proxyEndpoint = endPointFactory.attach(proxyAddress);
  i = await proxyEndpoint.callStatic.i();
  let num = i.toNumber() + 2;
  txResponse = await proxyEndpoint.addTwo();
  receipt = await txResponse.wait();
  let test = await getEventVar(txResponse, "addedTwo", "i");
  expect(test.toNumber()).to.equal(num);
  //lock the proxy upgrade
  txResponse = await factory.upgradeProxy(salt, mockLogicAddr, "0x");
  receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  txResponse = await mockProxy.setv(testArg + 2);
  receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  v = await mockProxy.callStatic.v();
  receipt = await txResponse.wait();
  expect(v.toNumber()).to.equal(testArg + 2);
  //lock the upgrade functionality
  txResponse = await mockProxy.lock();
  receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  let txRes = factory.upgradeProxy(salt, endPointAddr, "0x");
  await expect(txRes).to.be.revertedWith("reverted with an unrecognized custom error");
  //unlock the proxy
  txResponse = await mockProxy.unlock();
  receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  txResponse = await factory.upgradeProxy(salt, endPointAddr, "0x");
  receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  i = await proxyEndpoint.callStatic.i();
  num = i.toNumber() + 2;
  txResponse = await proxyEndpoint.addTwo();
  receipt = await txResponse.wait();
  test = await getEventVar(txResponse, "addedTwo", "i");
  expect(test.toNumber()).to.equal(num);
}

export async function metaMockLogicTest(
  contract: ContractFactory,
  address: string,
  factoryAddress: string
) {
  const Contract = contract.attach(address);
  let txResponse = await Contract.setFactory(factoryAddress);
  let test = 4;
  let receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  let fa = await Contract.getFactory.call();
  expect(fa).to.equal(factoryAddress);
  txResponse = await Contract.setv(test);
  receipt = await txResponse.wait();
  await expectTxSuccess(txResponse);
  let v = await Contract.callStatic.v();
  expect(v.toNumber()).to.equal(test);
  let i = await Contract.callStatic.i();
  expect(i.toNumber()).to.equal(2);
}

export async function getEventVar(txResponse: ContractTransaction, eventName: string, varName: string) {
  let result:any;
  let receipt = await txResponse.wait();
  if (receipt.events !== undefined) {
    let events = receipt.events
    for (let i = 0; i < events.length; i++) {
      //look for the event
      if (events[i].event === eventName) {
        if (events[i].args !== undefined) {
          let args = events[i].args
          //extract the deployed mock logic contract address from the event
          result = args !== undefined ? args[varName] : undefined;
          if (result !== undefined) {
            return result;
          }
        } else {
          throw new Error(`failed to extract ${varName} from event: ${eventName}`)
        }
      }
    }
  }
  throw new Error(`failed to find event: ${eventName}`)
}

export async function expectTxSuccess(txResponse: ContractTransaction) {
  let receipt = await txResponse.wait();
  expect(receipt.status).to.equal(1);
}

export function getCreateAddress(Address: string, nonce: number) {
  return ethers.utils.getContractAddress({
    from: Address,
    nonce: nonce,
  });
}
export function bytes32ArrayToStringArray(bytes32Array: Array<any>) {
  let ret = [];
  for (let i = 0; i < bytes32Array.length; i++) {
    ret.push(ethers.utils.parseBytes32String(bytes32Array[i]));
  }
  return ret;
}

export function getSalt() {
  //set a new salt
  let salt = new Date();
  //use the time as the salt
  let Salt = salt.getTime();
  return ethers.utils.formatBytes32String(Salt.toString());
}

export async function getDeployTemplateArgs(contractName: string) {
  let contract = await ethers.getContractFactory(contractName);
  let deployByteCode = contract.getDeployTransaction();
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
  let contract = await ethers.getContractFactory(contractName);
  let ret: DeployStaticArgs = {
    salt: getSalt(),
    initCallData: contract.interface.encodeFunctionData(
      "initialize",
      argsArray
    ),
  };
  return ret;
}

export async function checkMockInit(target: string, initVal: number) {
  let mockFactory = await ethers.getContractFactory(MOCK);
  let mock = await mockFactory.attach(target);
  let i = await mock.callStatic.i();
  expect(i.toNumber()).to.equal(initVal);
}

export async function deployFactory() {
  let factoryBase = await ethers.getContractFactory(MADNET_FACTORY);
  let accounts = await getAccounts();
  let firstOwner = accounts[0];
  //gets the initial transaction count for the address
  let transactionCount = await ethers.provider.getTransactionCount(firstOwner);
  //pre calculate the address of the factory contract
  let futureFactoryAddress = ethers.utils.getContractAddress({
    from: firstOwner,
    nonce: transactionCount,
  });
  //deploy the factory with its address as a constructor input
  let factory = await factoryBase.deploy(futureFactoryAddress);
  expect(factory.address).to.equal(futureFactoryAddress);
  return factory;
}

export async function deployCreate2Initable(factory: MadnetFactory, salt:BytesLike) {
  let mockInitFactory = await ethers.getContractFactory("MockInitializable")
  let txResponse = await factory.deployCreate2(0, salt, mockInitFactory.bytecode)
  return txResponse;
}

export function getMetamorphicAddress(factoryAddress: string, salt: string) {
  let initCode = "0x6020363636335afa1536363636515af43d36363e3d36f3";
  return ethers.utils.getCreate2Address(
    factoryAddress,
    salt,
    ethers.utils.keccak256(initCode)
  );
}
