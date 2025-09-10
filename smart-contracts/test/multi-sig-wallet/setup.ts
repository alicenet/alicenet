import { ethers } from "hardhat";
import { getFixture } from "../legacy/setup";

export async function getMultiSigFixture() {
  const owner = await ethers.provider.getSigner(0);
  const fixture = await getFixture();
  const multiSigWallet = await (
    await ethers.getContractFactory("AlicenetMultiSigWallet")
  ).deploy();
  const accounts = [
    await (await ethers.provider.getSigner(1)).getAddress(),
    await (await ethers.provider.getSigner(2)).getAddress(),
    await (await ethers.provider.getSigner(3)).getAddress(),
  ];
  const sortedAccount = accounts.sort((a, b) => {
    return a.localeCompare(b);
  });

  const user1 = await ethers.provider.getSigner(sortedAccount[0]);
  const user2 = await ethers.provider.getSigner(sortedAccount[1]);
  const user3 = await ethers.provider.getSigner(sortedAccount[2]);
  await multiSigWallet.waitForDeployment();
  await multiSigWallet.setup(
    [
      await user1.getAddress(),
      await user2.getAddress(),
      await user3.getAddress(),
    ],
    2,
    ethers.ZeroAddress,
    "0x",
    ethers.ZeroAddress,
    ethers.ZeroAddress,
    0,
    ethers.ZeroAddress
  );
  await fixture.alcb.setShareHolders([
    {
      account: await fixture.publicStaking.getAddress(),
      percentage: 1000n,
      isMagicTransfer: true,
    },
  ]);
  await fixture.alcb.setOperationalWallet(await multiSigWallet.getAddress());
  // changing owner to multiSigWallet
  await fixture.factory.setOwner(await multiSigWallet.getAddress());
  await fixture.alcb.transferOwnership(await multiSigWallet.getAddress());
  return { fixture, multiSigWallet, owner, user1, user2, user3 };
}
