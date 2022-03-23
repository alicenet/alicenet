import {
  Fixture,
  getFixture,
  getTokenIdFromTx,
  getValidatorEthAccount
} from "../setup"
import { ethers } from "hardhat"
import { expect } from "../chai-setup"
import { BigNumber, BigNumberish, Signer } from "ethers"
import { StakeNFT, StakeNFTPositionDescriptor } from "../../typechain-types"

describe("StakeNFTPositionDescriptor: Tests StakeNFTPositionDescriptor methods", async () => {
  let fixture: Fixture
  let adminSigner: Signer
  let stakeNFT: StakeNFT
  let stakeNFTPositionDescriptor: StakeNFTPositionDescriptor
  let stakeAmount = 20000
  let stakeAmountMadWei = ethers.utils.parseUnits(stakeAmount.toString(), 18)
  let lockTime = 1
  let tokenId: BigNumberish

  beforeEach(async function () {
    fixture = await getFixture(true, true)

    const [admin] = fixture.namedSigners
    adminSigner = await getValidatorEthAccount(admin.address)

    stakeNFT = fixture.stakeNFT
    stakeNFTPositionDescriptor = fixture.stakeNFTPositionDescriptor

    await fixture.madToken.approve(
      fixture.stakeNFT.address,
      BigNumber.from(stakeAmountMadWei)
    )
    let tx = await fixture.stakeNFT
      .connect(adminSigner)
      .mintTo(admin.address, stakeAmountMadWei, lockTime)
    tokenId = await getTokenIdFromTx(tx)
  })

  it("Fails if token at id does not exist", async function () {
    const invalidTokenId = 1234

    await expect(stakeNFT.tokenURI(invalidTokenId)).to.be.revertedWith(
      "StakeNFT: Error, NFT token doesn't exist!"
    )
  })

  describe("Given valid token id", async () => {
    let positionData
    let svg: string
    let tokenUriJson: string
    let expectedTokenUriData: string
    beforeEach(async function () {
      positionData = await stakeNFT.getPosition(tokenId)

      svg =
        `<svg width="500" height="500" viewBox="0 0 500 500" xmlns="http://www.w3.org/2000/svg" xmlns:xlink='http://www.w3.org/1999/xlink'>` +
        `<text x='10' y='20'>Shares: ${positionData.shares.toString()}</text>` +
        `<text x='10' y='40'>Free after: ${positionData.freeAfter.toString()}</text>` +
        `<text x='10' y='60'>Withdraw Free After: ${positionData.withdrawFreeAfter.toString()}</text>` +
        `<text x='10' y='80'>Accumulator (ETH): ${positionData.accumulatorEth.toString()}</text>` +
        `<text x='10' y='100'>Accumulator (Token): ${positionData.accumulatorToken.toString()}</text></svg>`

      tokenUriJson =
        `{"name":"MadNET Staked token for position #1", ` +
        `"description":"This NFT represents a staked position on MadNET.` +
        `\\nThe owner of this NFT can modify or redeem the position.` +
        `\\n Shares: ${positionData.shares.toString()}` +
        `\\nFree After: ${positionData.freeAfter.toString()}` +
        `\\nWithdraw Free After: ${positionData.withdrawFreeAfter.toString()}` +
        `\\nAccumulator Eth: ${positionData.accumulatorEth.toString()}` +
        `\\nAccumulator Token: ${positionData.accumulatorToken.toString()}` +
        `\\nToken ID: ${tokenId.toString()}", ` +
        `"image": "data:image/svg+xml;base64,${btoa(svg)}"}`

      expectedTokenUriData = `data:application/json;base64,${btoa(
        tokenUriJson
      )}`
    })

    it("StakeNFTPositionDescriptor should return correct token uri", async function () {
      const tokenUri = await stakeNFTPositionDescriptor.tokenURI(
        stakeNFT.address,
        tokenId
      )

      const parsedJson = JSON.parse(
        atob(tokenUri.replace("data:application/json;base64,", ""))
      )

      await expect(tokenUri).to.be.equal(expectedTokenUriData)
      await expect(
        atob(parsedJson.image.replace("data:image/svg+xml;base64,", ""))
      ).to.be.equal(svg)
    })

    it("StakeNFT contract Should return correct token uri", async function () {
      const tokenUri = await stakeNFT.tokenURI(tokenId)

      const parsedJson = JSON.parse(
        atob(tokenUri.replace("data:application/json;base64,", ""))
      )

      await expect(tokenUri).to.be.equal(expectedTokenUriData)
      await expect(
        atob(parsedJson.image.replace("data:image/svg+xml;base64,", ""))
      ).to.be.equal(svg)
    })
  })
})
