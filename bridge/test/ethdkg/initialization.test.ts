import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BigNumber, BytesLike } from "ethers";
import { ethers, network } from "hardhat";
import {
  Fixture,
  getContractAddressFromDeployedRawEvent,
  getFixture,
} from "../setup";
import { expect } from "./setup";

describe("Contract Initialization", () => {
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await loadFixture(getFixture);
  });

  it("Should not allow initialize more than once", async () => {
    await expect(
      fixture.factory.callAny(
        fixture.ethdkg.address,
        0,
        fixture.ethdkg.interface.encodeFunctionData("initialize", [1, 2])
      )
    ).to.revertedWith("Initializable: contract is already initialized");
  });

  it("Should be able to reinitialize more than once", async () => {
    await network.provider.send("evm_setBlockGasLimit", ["0x3000000000000000"]);
    const deployCode = (
      await ethers.getContractFactory("ETHDKGMock")
    ).getDeployTransaction().data as BytesLike;

    const transaction = await fixture.factory.deployCreate(deployCode);
    const newETHDKGAddress = await getContractAddressFromDeployedRawEvent(
      transaction
    );
    expect(await fixture.ethdkg.getPhaseLength()).to.equal(BigNumber.from(40));
    expect(await fixture.ethdkg.getConfirmationLength()).to.equal(
      BigNumber.from(6)
    );
    await fixture.factory.upgradeProxy(
      ethers.utils.formatBytes32String("ETHDKG"),
      newETHDKGAddress,
      "0x"
    );
    await expect(fixture.ethdkg.initialize(1, 2)).to.revertedWith(
      "Initializable: contract is already initialized"
    );
    const ethdkgMockBase = await ethers.getContractFactory("ETHDKGMock");
    const ethdkgMock = await ethers.getContractAt(
      "ETHDKGMock",
      fixture.ethdkg.address
    );
    await fixture.factory.callAny(
      fixture.ethdkg.address,
      0,
      ethdkgMockBase.interface.encodeFunctionData("reinitialize", [
        BigNumber.from(1),
        BigNumber.from(2),
      ])
    );
    await expect(ethdkgMock.reinitialize(1, 2)).to.revertedWith(
      "Initializable: contract is already initialized"
    );
    expect(await fixture.ethdkg.getPhaseLength()).to.equal(BigNumber.from(1));
    expect(await fixture.ethdkg.getConfirmationLength()).to.equal(
      BigNumber.from(2)
    );
  });
});
