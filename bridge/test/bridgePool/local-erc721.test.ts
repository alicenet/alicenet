import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  callFunctionAndGetReturnValues,
  Fixture,
  getContractAddressFromEventLog,
  getFixture,
} from "../setup";
import {
  getMockBlockClaimsForStateRoot,
  getState,
  showState,
  state,
} from "./setup";

let fixture: Fixture;
let expectedState: state;
let firstOwner: SignerWithAddress;
let user: SignerWithAddress;
let user2: SignerWithAddress;
let bridgePool: any;

const bTokenFeeInETH = 10;
const erc20Amount = BigNumber.from(100).toBigInt();
const bTokenAmount = BigNumber.from(100).toBigInt();
const tokenId = 100;

// The following merkle proof and stateRoot values can be obtained from accusation_builder_test.go execution
const merkleProof =
  "0x010005cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000af469f3b9864a5132323df8bdd9cbd59ea728cd7525b65252133a5a02f1566ee00010003a8793650a7050ac58cf53ea792426b97212251673788bf0b4045d0bb5bdc3843aafb9eb5ced6edc2826e734abad6235c8cf638c812247fd38f04e7080d431933b9c6d6f24756341fde3e8055dd3a83743a94dddc122ab3f32a3db0c4749ff57bad"; // capnproto
const stateRoot =
  "0x0d66a8a0babec3d38b67b5239c1683f15a57e087f3825fac3d70fd6a243ed30b"; // stateRoot
// Mock a merkle proof for a burned UTXO on alicenet
const burnedUTXO = {
  chainId: 0,
  owner: "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
  value: tokenId,
  fee: 1,
  txHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
};
const encodedBurnedUTXO = ethers.utils.defaultAbiCoder.encode(
  [
    "tuple(uint256 chainId, address owner, uint256 value, uint256 fee, bytes32 txHash)",
  ],
  [burnedUTXO]
);
const networkId = 1337;
describe("Testing BridgePool", () => {
  beforeEach(async function () {
    [firstOwner, user, user2] = await ethers.getSigners();
    fixture = await getFixture(true, true, false);
    await fixture.factory.setDelegator(fixture.bridgePoolFactory.address);
    const ethIn = ethers.utils.parseEther(bTokenFeeInETH.toString());
    // Deploy a new pool
    const deployNewPoolTransaction =
      await fixture.bridgePoolFactory.deployNewPool(
        fixture.erc721Mock.address,
        1
      );
    const eventSignature = "event BridgePoolCreated(address contractAddr)";
    const eventName = "BridgePoolCreated";
    const bridgePoolAddress = await getContractAddressFromEventLog(
      deployNewPoolTransaction,
      eventSignature,
      eventName
    );
    bridgePool = (
      await ethers.getContractFactory("LocalERC721BridgePoolV1")
    ).attach(bridgePoolAddress);
    // Mint and approve an ERC721 token to deposit
    fixture.erc721Mock.mint(user.address, tokenId);
    fixture.erc721Mock.connect(user).approve(bridgePool.address, tokenId);
    // Mint and approve some bTokens to deposit as fee
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      firstOwner,
      [user.address, 0],
      ethIn
    );
    await fixture.bToken
      .connect(user)
      .approve(bridgePool.address, BigNumber.from(bTokenAmount));
    const encodedMockBlockClaims = getMockBlockClaimsForStateRoot(stateRoot);
    // Take a mock snapshot
    await fixture.snapshots.snapshot(
      Buffer.from("0x0"),
      encodedMockBlockClaims
    );
    showState("Initial", await getState(fixture, bridgePool));
  });

  describe("Testing BridgePool Router", async () => {
    it("Should make a deposit with parameters and emit correspondent event", async () => {
      expectedState = await getState(fixture, bridgePool);
      expectedState.Balances.nft.user -= BigNumber.from(1).toBigInt();
      expectedState.Balances.nft.bridgePool += BigNumber.from(1).toBigInt();
      expectedState.Balances.bToken.user -= bTokenAmount;
      expectedState.Balances.bToken.totalSupply -= bTokenAmount;
      const nonce = 1;
      await expect(
        bridgePool
          .connect(user)
          .deposit(1, user.address, erc20Amount, bTokenAmount)
      )
        .to.emit(fixture.bridgePoolDepositNotifier, "Deposited")
        .withArgs(
          BigNumber.from(nonce),
          fixture.erc721Mock.address,
          user.address,
          BigNumber.from(tokenId),
          BigNumber.from(networkId)
        );
      showState("After Deposit", await getState(fixture, bridgePool));
      expect(await getState(fixture, bridgePool)).to.be.deep.equal(
        expectedState
      );
    });

    it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
      // Make first a deposit to withdraw afterwards
      await bridgePool
        .connect(user)
        .deposit(1, user.address, erc20Amount, bTokenAmount);
      showState("After Deposit", await getState(fixture, bridgePool));
      expectedState = await getState(fixture, bridgePool);
      expectedState.Balances.nft.user += BigNumber.from(1).toBigInt();
      expectedState.Balances.nft.bridgePool -= BigNumber.from(1).toBigInt();
      await bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXO);
      showState("After withdraw", await getState(fixture, bridgePool));
      expect(await getState(fixture, bridgePool)).to.be.deep.equal(
        expectedState
      );
    });

    it("Should not make a withdraw for amount specified on informed burned UTXO with not verified merkle proof", async () => {
      const wrongMerkleProof =
        "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000";
      expectedState = await getState(fixture, bridgePool);
      const reason = ethers.utils.parseBytes32String(
        await fixture.bridgePoolErrorCodesContract.BRIDGEPOOL_COULD_NOT_VERIFY_PROOF_OF_BURN()
      );
      await expect(
        bridgePool.connect(user).withdraw(wrongMerkleProof, encodedBurnedUTXO)
      ).to.be.revertedWith(reason);
      expect(await getState(fixture, bridgePool)).to.be.deep.equal(
        expectedState
      );
    });

    it("Should not make a withdraw for amount specified on informed burned UTXO with wrong root", async () => {
      const wrongStateRoot =
        "0x0000000000000000000000000000000000000000000000000000000000000000";
      const encodedMockBlockClaims =
        getMockBlockClaimsForStateRoot(wrongStateRoot);
      await fixture.snapshots.snapshot(
        Buffer.from("0x0"),
        encodedMockBlockClaims
      );
      expectedState = await getState(fixture, bridgePool);
      const reason = ethers.utils.parseBytes32String(
        await fixture.bridgePoolErrorCodesContract.BRIDGEPOOL_COULD_NOT_VERIFY_PROOF_OF_BURN()
      );
      await expect(
        bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXO)
      ).to.be.revertedWith(reason);
      expect(await getState(fixture, bridgePool)).to.be.deep.equal(
        expectedState
      );
    });

    it("Should not make a withdraw to an address that is not the owner in informed burned UTXO", async () => {
      const reason = ethers.utils.parseBytes32String(
        await fixture.bridgePoolErrorCodesContract.BRIDGEPOOL_RECEIVER_IS_NOT_OWNER_ON_PROOF_OF_BURN_UTXO()
      );
      expectedState = await getState(fixture, bridgePool);
      await expect(
        bridgePool.connect(user2).withdraw(merkleProof, encodedBurnedUTXO)
      ).to.be.revertedWith(reason);
      expect(await getState(fixture, bridgePool)).to.be.deep.equal(
        expectedState
      );
    });
  });

  describe("Testing BridgePoolDepositNotifier", async () => {
    it("Should emit an event if called from a BridgePool", async () => {
      const nonce = 1;
      await expect(
        bridgePool
          .connect(user)
          .deposit(1, user.address, erc20Amount, bTokenAmount)
      )
        .to.emit(fixture.bridgePoolDepositNotifier, "Deposited")
        .withArgs(
          BigNumber.from(nonce),
          fixture.erc721Mock.address,
          user.address,
          BigNumber.from(tokenId),
          BigNumber.from(networkId)
        );
    });

    it("Should not emit an event if not called from a BridgePool", async () => {
      const salt = ethers.utils.keccak256("0x00");
      const reason = ethers.utils.parseBytes32String(
        await fixture.immutableAuthErrorCodesContract.IMMUTEABLEAUTH_ONLY_BRIDGEPOOL()
      );
      await expect(
        fixture.bridgePoolDepositNotifier.doEmit(
          salt,
          fixture.erc721Mock.address,
          tokenId,
          user.address
        )
      ).to.be.rejectedWith(reason);
    });
  });
});
