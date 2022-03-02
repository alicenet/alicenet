import { TransactionRequest } from '@ethersproject/abstract-provider';
import { ethers } from 'hardhat';
import { END_POINT, PROXY } from '../../scripts/lib/constants';
import { expect } from '../chai-setup';
import { deployFactory, expectTxSuccess, getAccounts, getSalt} from '../factory/Setup';


describe("PROXY", async () => {
    it("deploy proxy through factory", async () => {
        let factory = await deployFactory();
        let salt = getSalt();
        let txResponse = await factory.deployProxy(salt);
        expectTxSuccess(txResponse);
    });

    it("deploy proxy raw and upgrades to endPointLockable logic", async () => {
        let accounts = await getAccounts();
        let proxyFactory = await ethers.getContractFactory(PROXY);
        let proxy = await proxyFactory.deploy();
        let endPointLockableFactory = await ethers.getContractFactory("MockEndPointLockable");
        let endPointLockable = await endPointLockableFactory.deploy(accounts[0]);
        expect(proxy.deployed());
        let abicoder = new ethers.utils.AbiCoder()
        let encodedAddress = abicoder.encode(["address"], [endPointLockable.address])
        let txReq:TransactionRequest = {
            data: "0xca11c0de" + encodedAddress.substring(2)
        }
        let txResponse = await proxy.fallback(txReq)
        let receipt = await txResponse.wait();
        expect(receipt.status).to.equal(1);
        let proxyImplAddr = await proxy.callStatic.getImplementationAddress();
        expect(proxyImplAddr).to.equal(endPointLockable.address);
    });

    it("locks the proxy upgradeability, prevents the proxy from being updated", async () => {
        let accounts = await getAccounts();
        let proxyFactory = await ethers.getContractFactory(PROXY);
        let proxy = await proxyFactory.deploy();
        let endPointLockableFactory = await ethers.getContractFactory("MockEndPointLockable");
        let endPointLockable = await endPointLockableFactory.deploy(accounts[0]);
        let endPointFactory = await ethers.getContractFactory(END_POINT);
        let endPoint = await endPointFactory.deploy(accounts[0]);
        expect(proxy.deployed());
        let abicoder = new ethers.utils.AbiCoder()
        let encodedAddress = abicoder.encode(["address"], [endPointLockable.address]);
        let txReq:TransactionRequest = {
            data: "0xca11c0de" + encodedAddress.substring(2)
        }
        let txResponse = await proxy.fallback(txReq);
        let receipt = await txResponse.wait();
        expect(receipt.status).to.equal(1);
        let proxyImplAddr = await proxy.callStatic.getImplementationAddress();
        expect(proxyImplAddr).to.equal(endPointLockable.address);
        //interface of logic connected to logic contract
        let proxyContract = endPointLockableFactory.attach(proxy.address);
        //lock the implementation
        txResponse = await proxyContract.upgradeLock();
        receipt = await txResponse.wait();
        expect(receipt.status).to.equal(1);
        encodedAddress = abicoder.encode(["address"], [endPoint.address]);
        txReq = {
            data: "0xca11c0de" + encodedAddress.substring(2)
        }
        let response = proxy.fallback(txReq);
        await expect(response).to.be.revertedWith("reverted with an unrecognized custom error");
        txResponse = await proxyContract.upgradeUnlock();
        receipt = await txResponse.wait();
        expect(receipt.status).to.equal(1);
    });
});