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
      data: "0xca11c0de" + encodedAddress.substring(2),
    };
    const txResponse = await proxy.fallback(txReq);
    const receipt = await txResponse.wait();
    expect(receipt.status).to.equal(1);
    const proxyImplAddr = await proxy.callStatic.getImplementationAddress();
    expect(proxyImplAddr).to.equal(endPointLockable.address);
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
      data: "0xca11c0de" + encodedAddress.substring(2),
    };
    let txResponse = await proxy.fallback(txReq);
    let receipt = await txResponse.wait();
    expect(receipt.status).to.equal(1);
    const proxyImplAddr = await proxy.callStatic.getImplementationAddress();
    expect(proxyImplAddr).to.equal(endPointLockable.address);
    // interface of logic connected to logic contract
    const proxyContract = endPointLockableFactory.attach(proxy.address);
    // lock the implementation
    txResponse = await proxyContract.upgradeLock();
    receipt = await txResponse.wait();
    expect(receipt.status).to.equal(1);
    encodedAddress = abicoder.encode(["address"], [endPoint.address]);
    txReq = {
      data: "0xca11c0de" + encodedAddress.substring(2),
    };
    const response = proxy.fallback(txReq);
    await expect(response).to.be.reverted;
    txResponse = await proxyContract.upgradeUnlock();
    receipt = await txResponse.wait();
    expect(receipt.status).to.equal(1);
  });
});
