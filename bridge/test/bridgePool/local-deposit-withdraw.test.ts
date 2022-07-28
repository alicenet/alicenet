import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { BigNumber } from "ethers";

import {
  callFunctionAndGetReturnValues,
  factoryCallAnyFixture,
  Fixture,
  getContractAddressFromEventLog,
  getFixture,
} from "../setup";
import {
  formatBigInt,
  getMockBlockClaimsForStateRoot,
  getState,
  maxEth,
  maxTokens,
  showState,
  state,
  tokenTypes,
  valueOrId,
  valueSent,
} from "./setup";

let fixture: Fixture;
let expectedState: state;
let firstOwner: SignerWithAddress;
let user: SignerWithAddress;
let user2: SignerWithAddress;
let bridgePool: any;
let depositCallData: any;
let encodedDepositCallData: string;

const bTokenFeeInWEI = 1000;
const ethIn = BigNumber.from(10000);

// The following merkle proof and stateRoot values can be obtained from accusation_builder_test.go execution
const merkleProof =
  "0x010005cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000af469f3b9864a5132323df8bdd9cbd59ea728cd7525b65252133a5a02f1566ee00010003a8793650a7050ac58cf53ea792426b97212251673788bf0b4045d0bb5bdc3843aafb9eb5ced6edc2826e734abad6235c8cf638c812247fd38f04e7080d431933b9c6d6f24756341fde3e8055dd3a83743a94dddc122ab3f32a3db0c4749ff57bad"; // capnproto
const stateRoot =
  "0x0d66a8a0babec3d38b67b5239c1683f15a57e087f3825fac3d70fd6a243ed30b"; // stateRoot
// Mock a merkle proof for a burned UTXO on alicenet
const burnedUTXO = {
  chainId: 0,
  owner: "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
  value: valueOrId,
  fee: 1,
  txHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
};
const encodedBurnedUTXO = ethers.utils.defaultAbiCoder.encode(
  [
    "tuple(uint256 chainId, address owner, uint256 value, uint256 fee, bytes32 txHash)",
  ],
  [burnedUTXO]
);
const ethFee = 2;
const refund = valueSent.sub(ethFee);
const chainId = 1337;

tokenTypes.forEach(function (run) {
  describe(
    "Testing BridgePool Router Deposit/Withdraw for tokenType " + run.it,
    async () => {
      beforeEach(async () => {
        [firstOwner, user, user2] = await ethers.getSigners();
        fixture = await getFixture(true, true, false);
        console.log(await ethers.provider.getBalance(fixture.bToken.address));
        // const ethIn = ethers.utils.parseEther(bTokenFeeInWEI.toString());
        // Deploy a new pool
        await factoryCallAnyFixture(
          fixture,
          "bridgeRouter",
          "togglePublicPoolDeployment",
          []
        );
        const deployNewPoolTransaction =
          await fixture.bridgeRouter.deployNewLocalPool(
            run.options.poolType,
            fixture[run.options.ercContractName].address,
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
          await ethers.getContractFactory(run.options.bridgeImpl)
        ).attach(bridgePoolAddress);
        // Mint and approve an ERC token to deposit
        fixture[run.options.ercContractName].mint(user.address, valueOrId);
        fixture[run.options.ercContractName]
          .connect(user)
          .approve(bridgePool.address, valueOrId);
          console.log(await ethers.provider.getBalance(fixture.bToken.address));

        // // Mint and approve some bTokens to deposit as fee
        // await callFunctionAndGetReturnValues(
        //   fixture.bToken,
        //   "mintTo",
        //   firstOwner,
        //   [user.address, 0],
        //   ethIn
        // );
        console.log(await ethers.provider.getBalance(fixture.bToken.address));

        const encodedMockBlockClaims =
          getMockBlockClaimsForStateRoot(stateRoot);
        // Take a mock snapshot
        await fixture.snapshots.snapshot(
          Buffer.from("0x0"),
          encodedMockBlockClaims
        );
        depositCallData = {
          ERCContract: fixture[run.options.ercContractName].address,
          tokenType: run.options.poolType,
          number: valueOrId,
          chainID: 1337,
          poolVersion: 1,
        };
        encodedDepositCallData = ethers.utils.defaultAbiCoder.encode(
          [
            "tuple(address ERCContract, uint8 tokenType, uint256 number, uint256 chainID, uint16 poolVersion)",
          ],
          [depositCallData]
        );
        showState("Initial", await getState(fixture, bridgePool));
      });

      it.only("Should make a deposit", async () => {
        expectedState = await getState(fixture, bridgePool);
        const erc = run.it as keyof typeof expectedState.Balances;
        expectedState.Balances[erc].user -= BigInt(run.options.quantity);
        expectedState.Balances[erc].bridgePool += BigInt(run.options.quantity);
        console.log(ethFee, bTokenFeeInWEI, valueSent, refund);
        expectedState.Balances.eth.bToken += BigInt(bTokenFeeInWEI);
        expectedState.Balances.eth.user -= formatBigInt(valueSent);
        expectedState.Balances.eth.user += formatBigInt(refund);
        await fixture.bToken
          .connect(user)
          .payAndDeposit(maxEth, maxTokens, encodedDepositCallData, {
            value: valueSent,
          });
        showState("After Deposit", await getState(fixture, bridgePool));
        expect(await getState(fixture, bridgePool)).to.be.deep.equal(
          expectedState
        );
      });

      it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
        // Make first a deposit to withdraw afterwards
        await fixture.bToken
          .connect(user)
          .payAndDeposit(maxEth, maxTokens, encodedDepositCallData, {
            value: valueSent,
          });
        showState("After Deposit", await getState(fixture, bridgePool));
        expectedState = await getState(fixture, bridgePool);
        const erc = run.it as keyof typeof expectedState.Balances;
        expectedState.Balances[erc].user += BigInt(run.options.quantity);
        expectedState.Balances[erc].bridgePool -= BigInt(run.options.quantity);
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
        await expect(
          bridgePool.connect(user).withdraw(wrongMerkleProof, encodedBurnedUTXO)
        ).to.be.revertedWith(
          "MerkleProofLibrary: Invalid Inclusion Merkle proof!"
        );
        expect(await getState(fixture, bridgePool)).to.be.deep.equal(
          expectedState
        );
      });

      it("Should not make a deposit without enough funds for fees", async () => {
        const reason = ethers.utils.parseBytes32String(
          await fixture.bridgeRouterErrorCodesContract.BRIDGEROUTER_INSUFFICIENT_FUNDS()
        );
        await expect(
          fixture.bToken
            .connect(user)
            .payAndDeposit(maxEth, maxTokens - 100, encodedDepositCallData, {
              value: valueSent,
            })
        ).to.be.revertedWith(reason);
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
        await expect(
          bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXO)
        ).to.be.revertedWith(
          "MerkleProofLibrary: The proof doesn't match the root of the trie!"
        );
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

      it("Should emit an event if called from a BridgePool", async () => {
        const nonce = 1;
        await expect(
          fixture.bToken
            .connect(user)
            .payAndDeposit(maxEth, maxTokens, encodedDepositCallData, {
              value: valueSent,
            })
        )
          .to.emit(fixture.bridgePoolDepositNotifier, "Deposited")
          .withArgs(
            BigInt(nonce),
            fixture[run.options.ercContractName].address,
            user.address,
            BigInt(run.options.poolType),
            BigInt(valueOrId),
            BigInt(chainId)
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
            fixture[run.options.ercContractName].address,
            user.address,
            run.options.poolType,
            valueOrId
          )
        ).to.be.rejectedWith(reason);
      });
    }
  );
});
