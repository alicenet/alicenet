import { ethers } from "hardhat";
import { END_POINT, PROXY } from "../../scripts/lib/constants";
import { expect } from "../chai-setup";
import {
  deployFactory,
  expectTxSuccess,
  getAccounts,
  getSalt,
} from "../factory/Setup";

describe("PROXY", async () => {
  it("deploy proxy through factory", async () => {
    const factory = await deployFactory();
    const salt = getSalt();
    const txResponse = await factory.deployProxy(salt);
    await expectTxSuccess(txResponse);
  });

  it("deploy proxy raw and upgrades to endPointLockable logic", async () => {
    const accounts = await getAccounts();
    const proxyFactory = await ethers.getContractFactory(PROXY);
    const proxy = await proxyFactory.deploy();
    const endPointLockableFactory = await ethers.getContractFactory(
      "MockEndPointLockable"
    );
    const endPointLockable = await endPointLockableFactory.deploy(accounts[0]);
    expect(proxy.deployed());
    const abicoder = new ethers.utils.AbiCoder();
    const encodedAddress = abicoder.encode(
      ["address"],
      [endPointLockable.address]
    );
    const txReq = {
      data: "0xca11c0de11" + encodedAddress.substring(2),
    };
    const txResponse = await proxy.fallback(txReq);
    await txResponse.wait();
    const expectedAddress = await getProxyImplementation(proxy.address);
    expect(expectedAddress).to.equal(endPointLockable.address);
  });

  it("deploys proxy and attempts to upgrade it with non-admin account", async () => {
    const accounts = await ethers.getSigners();
    const proxyFactory = await ethers.getContractFactory(PROXY);
    const proxy = await proxyFactory.deploy();
    const endPointLockableFactory = await ethers.getContractFactory(
      "MockEndPointLockable"
    );
    const abicoder = new ethers.utils.AbiCoder();
    const endPointLockable = await endPointLockableFactory.deploy(
      accounts[0].address
    );
    const encodedAddress = abicoder.encode(
      ["address"],
      [endPointLockable.address]
    );
    const txReq = {
      data: "0xca11c0de11" + encodedAddress.substring(2),
      gasLimit: 1000000,
    };
    const txResponse = proxy.connect(accounts[1]).fallback(txReq);
    await expect(txResponse).to.be.revertedWith("unauthorized");
  });

  it("deploys proxy and attempts to call logic without upgrading", async () => {
    const proxyFactory = await ethers.getContractFactory(PROXY);
    const proxy = await proxyFactory.deploy();
    const mockLockable = await ethers.getContractAt(
      "MockEndPointLockable",
      proxy.address
    );
    const txResponse = mockLockable.owner();
    await expect(txResponse).to.be.revertedWith("logic not set");
  });

  it("locks the proxy upgradeability, prevents the proxy from being updated", async () => {
    const accounts = await getAccounts();
    const proxyFactory = await ethers.getContractFactory(PROXY);
    const proxy = await proxyFactory.deploy();
    const endPointLockableFactory = await ethers.getContractFactory(
      "MockEndPointLockable"
    );
    const endPointLockable = await endPointLockableFactory.deploy(accounts[0]);
    const endPointFactory = await ethers.getContractFactory(END_POINT);
    const endPoint = await endPointFactory.deploy();
    expect(proxy.deployed());
    const abicoder = new ethers.utils.AbiCoder();
    let encodedAddress = abicoder.encode(
      ["address"],
      [endPointLockable.address]
    );
    let txReq = {
      data: "0xca11c0de11" + encodedAddress.substring(2),
      gasLimit: 1000000,
    };
    let txResponse = await proxy.fallback(txReq);
    let receipt = await txResponse.wait();
    expect(receipt.status).to.equal(1);
    const proxyImplAddr = await getProxyImplementation(proxy.address);
    expect(proxyImplAddr).to.equal(endPointLockable.address);
    // interface of logic connected to logic contract
    const proxyContract = endPointLockableFactory.attach(proxy.address);
    // lock the implementation
    txResponse = await proxyContract.upgradeLock();
    receipt = await txResponse.wait();
    expect(receipt.status).to.equal(1);
    let response = proxy.fallback(txReq);
    await expect(response).to.be.revertedWith("update locked");
    txResponse = await proxyContract.upgradeUnlock();
    receipt = await txResponse.wait();
    expect(receipt.status).to.equal(1);
    encodedAddress = abicoder.encode(["address"], [endPoint.address]);
    txReq = {
      data: "0xca11c0de11" + encodedAddress.substring(2),
      gasLimit: 1000000,
    };
    response = proxy.fallback(txReq);
    const expectedAddress = await getProxyImplementation(proxy.address);
    expect(expectedAddress).to.equal(endPoint.address);
  });

  it("locks the proxy upgradeability even with poluted implementation address", async () => {
    const accounts = await getAccounts();
    const factory = await deployFactory();
    const proxyFactory = await ethers.getContractFactory(PROXY);
    const proxy = await proxyFactory.deploy();
    const endPointLockableFactory = await ethers.getContractFactory(
      "MockEndPointLockable"
    );
    const endPointLockable = await endPointLockableFactory.deploy(accounts[0]);
    const endPointFactory = await ethers.getContractFactory(END_POINT);
    const endPoint = await endPointFactory.deploy();
    expect(proxy.deployed());
    const abicoder = new ethers.utils.AbiCoder();
    let encodedAddress = abicoder.encode(
      ["address"],
      [endPointLockable.address]
    );
    let txReq = {
      data: "0xca11c0de11" + encodedAddress.substring(2),
      gasLimit: 1000000,
    };
    const txResponse = await proxy.fallback(txReq);
    const receipt = await txResponse.wait();
    expect(receipt.status).to.equal(1);
    const proxyImplAddr = await getProxyImplementation(proxy.address);
    expect(proxyImplAddr).to.equal(endPointLockable.address);
    expect(await factory.getProxyImplementation(proxy.address)).to.equal(
      endPointLockable.address
    );
    // interface of logic connected to logic contract
    const proxyContract = endPointLockableFactory.attach(proxy.address);
    await proxyContract.poluteImplementationAddress();
    expect(proxyImplAddr).to.equal(endPointLockable.address);
    expect(await factory.getProxyImplementation(proxy.address)).to.equal(
      endPointLockable.address
    );
    // lock the implementation
    expect((await (await proxyContract.upgradeLock()).wait()).status).to.equal(
      1
    );
    expect(proxyImplAddr).to.equal(endPointLockable.address);
    expect(await factory.getProxyImplementation(proxy.address)).to.equal(
      endPointLockable.address
    );
    let response = proxy.fallback(txReq);
    await expect(response).to.be.revertedWith("update locked");
    expect(
      (await (await proxyContract.upgradeUnlock()).wait()).status
    ).to.equal(1);
    encodedAddress = abicoder.encode(["address"], [endPoint.address]);
    txReq = {
      data: "0xca11c0de11" + encodedAddress.substring(2),
      gasLimit: 1000000,
    };
    response = proxy.fallback(txReq);
    const expectedAddress = await getProxyImplementation(proxy.address);
    expect(expectedAddress).to.equal(endPoint.address);
    expect(await factory.getProxyImplementation(proxy.address)).to.equal(
      endPoint.address
    );
  });

  it("should not be able to call proxy function with incorrect signature", async () => {
    const proxyFactory = await ethers.getContractFactory(PROXY);
    const proxy = await proxyFactory.deploy();
    const abicoder = new ethers.utils.AbiCoder();
    const encodedAddress = abicoder.encode(["address"], [proxy.address]);
    let txReq = {
      data: "0xaabbccddee" + encodedAddress.substring(2),
      gasLimit: 1000000,
    };
    await expect(proxy.fallback(txReq)).to.be.rejectedWith(
      "function not found"
    );
    txReq = {
      data: "0xaabbccddee",
      gasLimit: 1000000,
    };
    await expect(proxy.fallback(txReq)).to.be.rejectedWith(
      "function not found"
    );
  });
});

export async function getProxyImplementation(proxyAddress: string) {
  //   const proxy = await ethers.getContractAt("Proxy", proxyAddress);
  const txReq = {
    data: "0x0cbcae703c",
    to: proxyAddress,
  };
  const signers = await ethers.getSigners();
  const implementationAddress = ethers.utils.getAddress(
    await signers[0].call(txReq)
  );
  return implementationAddress;
}
