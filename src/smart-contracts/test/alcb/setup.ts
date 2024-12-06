import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { ALCB } from "../../typechain-types";
import { ethers } from "hardhat";
import { BaseTokensFixture } from "../legacy/setup";

export type Fixture = {
  alcb: ALCB;
  operationalWallet: SignerWithAddress;
  owner: SignerWithAddress;
};

export async function getFixture(
  fixture: undefined | BaseTokensFixture = undefined
): Promise<Fixture> {
  const owner = await ethers.provider.getSigner(0);
  const operationalWallet = await ethers.provider.getSigner(1);
  const ALCB = await ethers.getContractFactory("ALCB");
  const shareHolderAccount = fixture
    ? await fixture.publicStaking.getAddress()
    : await operationalWallet.getAddress();
  const alcb = await ALCB.deploy(await operationalWallet.getAddress(), [
    {
      account: shareHolderAccount,
      percentage: 1000n,
      isMagicTransfer: true,
    },
  ]);
  await alcb.waitForDeployment();
  return { alcb, operationalWallet, owner };
}
