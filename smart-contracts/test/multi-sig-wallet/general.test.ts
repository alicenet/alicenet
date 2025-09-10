import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { getMultiSigFixture } from "./setup";
import { Fixture } from "../legacy/setup";
import { AlicenetMultiSigWallet } from "../../typechain-types";
import { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";
import { expect } from "../chai-setup";
import { ethers } from "hardhat";
import { BytesLike } from "ethers";
import { encodeMultiSigCall, signDataSafe } from "../legacy/utils/utils";

describe("Testing Alicenet Multisig wallet", async () => {
  let fixture: Fixture;
  let multiSigWallet: AlicenetMultiSigWallet;
  let owner: HardhatEthersSigner;
  let user1: HardhatEthersSigner;
  let user2: HardhatEthersSigner;
  let user3: HardhatEthersSigner;
  beforeEach(async function () {
    const superFixture = await loadFixture(getMultiSigFixture);
    fixture = superFixture.fixture;
    multiSigWallet = superFixture.multiSigWallet;
    owner = superFixture.owner;
    user1 = superFixture.user1;
    user2 = superFixture.user2;
    user3 = superFixture.user3;
  });

  it("It shouldn't be possible to call setup again", async () => {
    await expect(
      multiSigWallet.setup(
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
      )
    ).to.be.revertedWith("GS200");
  });

  it("Only Owners of contracts must be able to execute a transaction", async () => {
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await fixture.alcb.getAddress(),
      0,
      fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
        await user1.getAddress(),
      ]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    // owner is not one of the allowed accounts to sign
    const signature = await signDataSafe(owner, dataToSign);
    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await fixture.alcb.getAddress(),
          0,
          fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
            await user1.getAddress(),
          ]),
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature + signature.slice(2)
        )
    ).to.be.revertedWith("GS026");

    // user1 is one of the allowed accounts to sign, but owner is not
    const signature2 = await signDataSafe(user1, dataToSign);

    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await fixture.alcb.getAddress(),
          0,
          fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
            await user1.getAddress(),
          ]),
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature + signature2.slice(2)
        )
    ).to.be.revertedWith("GS026");
  });

  it("Shouldn't allow txs that don't meet the threshold requirement", async () => {
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await fixture.alcb.getAddress(),
      0,
      fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
        await user1.getAddress(),
      ]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    // user1 is one of the allowed accounts to sign
    const signature2 = await signDataSafe(user1, dataToSign);

    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await fixture.alcb.getAddress(),
          0,
          fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
            await user1.getAddress(),
          ]),
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature2
        )
    ).to.be.revertedWith("GS020");

    // try to sign twice
    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await fixture.alcb.getAddress(),
          0,
          fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
            await user1.getAddress(),
          ]),
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature2 + signature2.slice(2)
        )
    ).to.be.revertedWith("GS026");
  });

  it("Owners of the contract must be able to add a new owner", async () => {
    const owner4 = await ethers.provider.getSigner(4);
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await multiSigWallet.getAddress(),
      0,
      multiSigWallet.interface.encodeFunctionData("addOwnerWithThreshold", [
        await owner4.getAddress(),
        3,
      ]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    // user1 is one of the allowed accounts to sign
    let signature1 = await signDataSafe(user1, dataToSign);
    let signature2 = await signDataSafe(user2, dataToSign);
    let signature3 = await signDataSafe(user3, dataToSign);

    expect(await multiSigWallet.getThreshold()).to.be.equal(2);
    expect(await multiSigWallet.getOwners()).to.not.include(
      await owner4.getAddress()
    );

    await multiSigWallet
      .connect(owner)
      .execTransaction(
        await multiSigWallet.getAddress(),
        0,
        multiSigWallet.interface.encodeFunctionData("addOwnerWithThreshold", [
          await owner4.getAddress(),
          3,
        ]),
        0,
        0,
        0,
        0,
        ethers.ZeroAddress,
        ethers.ZeroAddress,
        signature1 + signature2.slice(2) + signature3.slice(2)
      );

    expect(await multiSigWallet.getOwners()).to.include(
      await owner4.getAddress()
    );
    expect(await multiSigWallet.getThreshold()).to.be.equal(3);

    // new owner should be able to participate in the multisig
    const dataToSign2 = await multiSigWallet.encodeTransactionData(
      await fixture.alcb.getAddress(),
      0,
      fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
        await user1.getAddress(),
      ]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    signature1 = await signDataSafe(user1, dataToSign2);
    signature2 = await signDataSafe(user2, dataToSign2);
    signature3 = await signDataSafe(user3, dataToSign2);
    const signature4 = await signDataSafe(owner4, dataToSign2);
    await multiSigWallet
      .connect(owner4)
      .execTransaction(
        await fixture.alcb.getAddress(),
        0,
        fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
          await user1.getAddress(),
        ]),
        0,
        0,
        0,
        0,
        ethers.ZeroAddress,
        ethers.ZeroAddress,
        signature4 +
          signature1.slice(2) +
          signature2.slice(2) +
          signature3.slice(2)
      );
  });

  it("Owners of contracts must be able remove an owner", async () => {
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await multiSigWallet.getAddress(),
      0,
      multiSigWallet.interface.encodeFunctionData("removeOwner", [
        await user2.getAddress(),
        await user3.getAddress(),
        1,
      ]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);

    expect(await multiSigWallet.getOwners()).to.include(
      await user3.getAddress()
    );
    expect(await multiSigWallet.getThreshold()).to.be.equal(2);

    await multiSigWallet
      .connect(owner)
      .execTransaction(
        await multiSigWallet.getAddress(),
        0,
        multiSigWallet.interface.encodeFunctionData("removeOwner", [
          await user2.getAddress(),
          await user3.getAddress(),
          1,
        ]),
        0,
        0,
        0,
        0,
        ethers.ZeroAddress,
        ethers.ZeroAddress,
        signature1 + signature2.slice(2)
      );

    expect(await multiSigWallet.getOwners()).to.not.include(
      await user3.getAddress()
    );
    expect(await multiSigWallet.getThreshold()).to.be.equal(1);
  });

  it("Should not allow to set threshold below 1", async () => {
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await multiSigWallet.getAddress(),
      0,
      multiSigWallet.interface.encodeFunctionData("changeThreshold", [0]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);

    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await multiSigWallet.getAddress(),
          0,
          multiSigWallet.interface.encodeFunctionData("changeThreshold", [0]),
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature1 + signature2.slice(2)
        )
    ).to.be.revertedWith("GS013");
  });

  it("Should be able to deploy a contract and update a proxy via Alicenet Factory", async () => {
    const deploymentData = (
      await (
        await ethers.getContractFactory("PublicStaking")
      ).getDeployTransaction()
    ).data as BytesLike;
    const encodeData = fixture.factory.interface.encodeFunctionData(
      "deployCreate",
      [deploymentData]
    );
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await fixture.factory.getAddress(),
      0,
      encodeData,
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);
    const nonce = await ethers.provider.getTransactionCount(
      await fixture.factory.getAddress()
    );
    const expectedAddress = ethers.getCreateAddress({
      from: await fixture.factory.getAddress(),
      nonce,
    });

    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await fixture.factory.getAddress(),
          0,
          encodeData,
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature1 + signature2.slice(2)
        )
    )
      .to.emit(fixture.factory, "DeployedRaw")
      .withArgs(expectedAddress);

    const encodeData2 = fixture.factory.interface.encodeFunctionData(
      "upgradeProxy",
      [ethers.encodeBytes32String("PublicStaking"), expectedAddress, "0x"]
    );
    const dataToSign2 = await encodeMultiSigCall(
      multiSigWallet,
      await fixture.factory.getAddress(),
      encodeData2
    );
    const signature3 = await signDataSafe(user1, dataToSign2);
    const signature4 = await signDataSafe(user2, dataToSign2);

    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await fixture.factory.getAddress(),
          0,
          encodeData2,
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature3 + signature4.slice(2)
        )
    ).to.emit(fixture.factory, "UpgradedProxy");
  });

  it("Should be able to mint ALCA via factory with multisig wallet", async () => {
    const encodeData = fixture.alcaMinter.interface.encodeFunctionData("mint", [
      await user1.getAddress(),
      ethers.parseEther("100"),
    ]);
    const factoryEncodedData = fixture.factory.interface.encodeFunctionData(
      "callAny",
      [await fixture.alcaMinter.getAddress(), 0, encodeData]
    );
    const dataToSign = await encodeMultiSigCall(
      multiSigWallet,
      await fixture.factory.getAddress(),
      factoryEncodedData
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);

    expect(await fixture.alca.balanceOf(await user1.getAddress())).to.be.equal(
      0
    );

    await multiSigWallet
      .connect(owner)
      .execTransaction(
        await fixture.factory.getAddress(),
        0,
        factoryEncodedData,
        0,
        0,
        0,
        0,
        ethers.ZeroAddress,
        ethers.ZeroAddress,
        signature1 + signature2.slice(2)
      );

    expect(await fixture.alca.balanceOf(await user1.getAddress())).to.be.equal(
      ethers.parseEther("100")
    );
  });

  it("Should be able to change ALCB parameters with multisig wallet", async () => {
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await fixture.alcb.getAddress(),
      0,
      fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
        await user1.getAddress(),
      ]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);

    expect(await fixture.alcb.operationalWallet()).to.be.equal(
      await multiSigWallet.getAddress()
    );

    await multiSigWallet
      .connect(owner)
      .execTransaction(
        await fixture.alcb.getAddress(),
        0,
        fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
          await user1.getAddress(),
        ]),
        0,
        0,
        0,
        0,
        ethers.ZeroAddress,
        ethers.ZeroAddress,
        signature1 + signature2.slice(2)
      );
    expect(await fixture.alcb.operationalWallet()).to.be.equal(
      await user1.getAddress()
    );
  });

  it("Should not be able to do a transaction if the signatures are not sorted by address", async () => {
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await fixture.alcb.getAddress(),
      0,
      fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
        await user1.getAddress(),
      ]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);

    expect(await fixture.alcb.operationalWallet()).to.be.equal(
      await multiSigWallet.getAddress()
    );

    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await fixture.alcb.getAddress(),
          0,
          fixture.alcb.interface.encodeFunctionData("setOperationalWallet", [
            await user1.getAddress(),
          ]),
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature2 + signature1.slice(2)
        )
    ).to.be.revertedWith("GS026");
  });

  it("Should be able to collect eth from ALCB burning with multisig wallet", async () => {
    await fixture.alcb.mint({ value: ethers.parseEther("100") });
    expect(
      await ethers.provider.getBalance(await multiSigWallet.getAddress())
    ).to.be.equal(0);
    await fixture.alcb.destroyTokens(ethers.parseEther("100"));
    expect(await fixture.alcb.operationalWallet()).to.be.equal(
      await multiSigWallet.getAddress()
    );
    expect(
      await ethers.provider.getBalance(await multiSigWallet.getAddress())
    ).to.be.equal(ethers.parseEther("100"));
    const randomUser = ethers.Wallet.createRandom();
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await randomUser.getAddress(),
      ethers.parseEther("100"),
      "0x",
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);

    expect(
      await ethers.provider.getBalance(await randomUser.getAddress())
    ).to.be.equal(ethers.parseEther("0"));

    await multiSigWallet
      .connect(owner)
      .execTransaction(
        await randomUser.getAddress(),
        ethers.parseEther("100"),
        "0x",
        0,
        0,
        0,
        0,
        ethers.ZeroAddress,
        ethers.ZeroAddress,
        signature1 + signature2.slice(2)
      );

    expect(
      await ethers.provider.getBalance(await multiSigWallet.getAddress())
    ).to.be.equal(0);
    expect(
      await ethers.provider.getBalance(await randomUser.getAddress())
    ).to.be.equal(ethers.parseEther("100"));
  });

  it("Should be able to call alicenet factory multi call with multisig wallet", async () => {
    const deploymentData = (
      await (
        await ethers.getContractFactory("PublicStaking")
      ).getDeployTransaction()
    ).data as BytesLike;
    const encodeData = fixture.factory.interface.encodeFunctionData(
      "deployCreate",
      [deploymentData]
    );
    const nonce = await ethers.provider.getTransactionCount(
      await fixture.factory.getAddress()
    );
    const expectedAddress = ethers.getCreateAddress({
      from: await fixture.factory.getAddress(),
      nonce,
    });
    const encodeData2 = fixture.factory.interface.encodeFunctionData(
      "upgradeProxy",
      [ethers.encodeBytes32String("PublicStaking"), expectedAddress, "0x"]
    );
    const factoryAddress = await fixture.factory.getAddress();
    const multiCallData = [
      [factoryAddress, 0, encodeData],
      [factoryAddress, 0, encodeData2],
    ];
    const finalEncodedData = fixture.factory.interface.encodeFunctionData(
      "multiCall",
      [multiCallData]
    );
    const dataToSign = await encodeMultiSigCall(
      multiSigWallet,
      await fixture.factory.getAddress(),
      finalEncodedData
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);

    await expect(
      multiSigWallet
        .connect(owner)
        .execTransaction(
          await fixture.factory.getAddress(),
          0,
          finalEncodedData,
          0,
          0,
          0,
          0,
          ethers.ZeroAddress,
          ethers.ZeroAddress,
          signature1 + signature2.slice(2)
        )
    ).to.emit(fixture.factory, "DeployedRaw");
    expect(await ethers.provider.getCode(expectedAddress)).to.not.be.equal(
      "0x"
    );
  });

  it("Should be able to change alicenet factory owner with multisig wallet", async () => {
    const dataToSign = await multiSigWallet.encodeTransactionData(
      await fixture.factory.getAddress(),
      0,
      fixture.factory.interface.encodeFunctionData("setOwner", [
        await user1.getAddress(),
      ]),
      0,
      0,
      0,
      0,
      ethers.ZeroAddress,
      ethers.ZeroAddress,
      await multiSigWallet.nonce()
    );
    const signature1 = await signDataSafe(user1, dataToSign);
    const signature2 = await signDataSafe(user2, dataToSign);

    expect(await fixture.factory.owner()).to.be.equal(
      await multiSigWallet.getAddress()
    );

    await multiSigWallet
      .connect(owner)
      .execTransaction(
        await fixture.factory.getAddress(),
        0,
        fixture.factory.interface.encodeFunctionData("setOwner", [
          await user1.getAddress(),
        ]),
        0,
        0,
        0,
        0,
        ethers.ZeroAddress,
        ethers.ZeroAddress,
        signature1 + signature2.slice(2)
      );

    expect(await fixture.factory.owner()).to.be.equal(await user1.getAddress());
  });
});
