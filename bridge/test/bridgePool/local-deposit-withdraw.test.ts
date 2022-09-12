import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber, Contract } from "ethers";
import { ethers } from "hardhat";
import { IBridgePool } from "../../typechain-types";
import { expect } from "../chai-setup";


import { Fixture, getFixture } from "../setup";
import {
  
  getMockBlockClaimsForStateRoot,
  getSimulatedBridgeRouter,
  getState,
  showState,
  state,
  tokenTypes,
  valueOrId
} from "./setup";
import { assert } from "console";

let fixture: Fixture;
let expectedState: state;
let admin: SignerWithAddress;
let user: SignerWithAddress;
let user2: SignerWithAddress;
let bridgePool: IBridgePool;
let depositCallData: any;



const bTokenFeeInWEI = 1000;

// The following merkle proof and stateRoot values can be obtained from accusation_builder_test.go execution
const merkleProof =
  "0x010005cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000af469f3b9864a5132323df8bdd9cbd59ea728cd7525b65252133a5a02f1566ee00010003a8793650a7050ac58cf53ea792426b97212251673788bf0b4045d0bb5bdc3843aafb9eb5ced6edc2826e734abad6235c8cf638c812247fd38f04e7080d431933b9c6d6f24756341fde3e8055dd3a83743a94dddc122ab3f32a3db0c4749ff57bad"; // capnproto
const wrongMerkleProof =
  "0x010005cda80a6c60e1215c1882b25b4744bd9d95c1218a2fd17827ab809c68196fd9bf0000000000000000000000000000000000000000000000000000000000000000af469f3b9864a5132323df8bdd9cbd59ea728cd7525b65252133a5a02f1566ee00010003a8793650a7050ac58cf53ea792426b97212251673788bf0b4045d0bb5bdc3843aafb9eb5ced6edc2826e734abad6235c8cf638c812247fd38f04e7080d431933b9c6d6f24756341fde3e8055dd3a83743a94dddc122ab3f32a3db0c4749ff57fff";
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
let merkleProofLibraryErrors: Contract;
  let  bridgeRouter:any;

  describe(
    "Testing BridgePool Router Deposit/Withdraw for tokenType ERC20",
    async () => {
      beforeEach(async () => {
        [admin, user, user2] = await ethers.getSigners();
        fixture = await getFixture(true, true, false);
        const encodedMockBlockClaims =
          getMockBlockClaimsForStateRoot(stateRoot);
        // Take a mock snapshot
        await fixture.snapshots.snapshot(
          Buffer.from("0x0"),
          encodedMockBlockClaims
        );
        merkleProofLibraryErrors = await (
          await (
            await ethers.getContractFactory("MerkleProofLibraryErrors")
          ).deploy()
        ).deployed();

        //Simulate a bridge router with some gas for transactions
        const bridgeRouterAddress = getSimulatedBridgeRouter(fixture.factory.address);
        depositCallData = {
          ercContract: fixture[tokenTypes[0].options.ercContractName].address,
          destinationAccountType: 1,
          destinationAccount: ethers.constants.AddressZero,
          tokenType: tokenTypes[0].options.tokenType,
          number: valueOrId,
          chainID: 1337,
          poolVersion: 1,
        };
        bridgePool = fixture[tokenTypes[0].options.bridgeImpl] as IBridgePool;
          fixture[tokenTypes[0].options.ercContractName].mint(user.address, valueOrId);
          fixture[tokenTypes[0].options.ercContractName]
            .connect(user)
            .approve(bridgePool.address, valueOrId);
        showState("Initial", await getState(fixture, bridgePool.address));
         erc = tokenTypes[0].it as keyof typeof expectedState.Balances;
      });

      it("Should make a deposit", async () => {
        expectedState = await getState(fixture, bridgePool.address);
        expectedState.Balances[erc].user -= BigInt(tokenTypes[0].options.quantity);
        expectedState.Balances[erc].bridgePool += BigInt(tokenTypes[0].options.quantity);
        await bridgePool.connect(bridgeRouter).deposit(user.address, valueOrId)
        showState("After Deposit", await getState(fixture, bridgePool.address));
        expect(await getState(fixture, bridgePool.address)).to.be.deep.equal(
          expectedState
        );
      });

/*       it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
        // Make first a deposit to withdraw afterwards
        await bridgePool.connect(bridgeRouter).deposit(user.address, valueOrId)
        showState("After Deposit", await getState(fixture, bridgePool.address));
        expectedState = await getState(fixture, bridgePool.address);
        expectedState.Balances[erc].user += BigInt(run.options.quantity);
        expectedState.Balances[erc].bridgePool -= BigInt(run.options.quantity);
        await bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXO);
        showState(
          "After withdraw",
          await getState(fixture, bridgePool.address)
        );
        expect(await getState(fixture, bridgePool.address)).to.be.deep.equal(
          expectedState
        );
      });

      it("Should not make a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
        await expect(
          bridgePool.connect(user).withdraw(wrongMerkleProof, encodedBurnedUTXO)
        ).to.be.revertedWithCustomError(
          merkleProofLibraryErrors,
          "ProofDoesNotMatchTrieRoot"
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
        await expect(
          bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXO)
        ).to.be.revertedWithCustomError(
          merkleProofLibraryErrors,
          "ProofDoesNotMatchTrieRoot"
        );
      });

      it("Should not make a withdraw to an address that is not the owner in informed burned UTXO", async () => {
        await expect(
          bridgePool.connect(user2).withdraw(merkleProof, encodedBurnedUTXO)
        ).to.be.revertedWithCustomError(
          bridgePool,
          "ReceiverIsNotOwnerOnProofOfBurnUTXO"
        );
      });
 */    
});


