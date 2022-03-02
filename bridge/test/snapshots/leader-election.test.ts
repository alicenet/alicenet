import { getFixture } from "../setup"
import { expect } from "chai"
import { Snapshots } from "../../typechain-types/Snapshots"

const numValidators = 10
const desperationFactor = 40

// this blocksignature happens to coincide with a starting index of 7 in the case of 10 validators
const blockSignature = `0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563`

describe("Snapshots: MayValidatorSnapshot", () => {
  let mayValidatorSnapshot: Snapshots["mayValidatorSnapshot"]

  before(async () => {
    const fixture = await getFixture()
    mayValidatorSnapshot = fixture.snapshots.mayValidatorSnapshot
  })

  describe("When desperation has not been reached", () => {
    const blocksSinceDesperation = 0

    it("Allows one validator to snapshot", async () => {
      expect(await mayValidatorSnapshot(numValidators, 7, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
    })

    it("Does not allow the second validator to snapshot yet", async () => {
      expect(await mayValidatorSnapshot(numValidators, 8, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
    })

  })

  describe("When desperation has just been reached", () => {
    const blocksSinceDesperation = 1

    it("Allows two validators to snapshot", async () => {
      expect(await mayValidatorSnapshot(numValidators, 7, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 8, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
    })

    it("Does not allow the third validator to snapshot yet", async () => {
      expect(await mayValidatorSnapshot(numValidators, 9, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
    })

  })

  describe("When desperationFactor has not been reached yet", () => {
    const blocksSinceDesperation = desperationFactor

    it("Allows two validators to snapshot", async () => {
      expect(await mayValidatorSnapshot(numValidators, 7, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 8, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
    })

    it("Does not allow the third validator to snapshot yet", async () => {
      expect(await mayValidatorSnapshot(numValidators, 9, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
    })

  })

  describe("When desperationFactor has just been reached", () => {
    const blocksSinceDesperation = desperationFactor+1

    it("Allows three validators to snapshot", async () => {
      expect(await mayValidatorSnapshot(numValidators, 7, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 8, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 9, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
    })

    it("Does not allow the fourth validator to snapshot yet", async () => {
      expect(await mayValidatorSnapshot(numValidators, 0, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
    })

  })

  describe("When desperationFactor has not been reached twice yet", () => {
    const blocksSinceDesperation = desperationFactor + Math.floor(desperationFactor/2)

    it("Allows three validators to snapshot", async () => {
      expect(await mayValidatorSnapshot(numValidators, 7, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 8, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 9, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
    })

    it("Does not allow the fourth validator to snapshot yet", async () => {
      expect(await mayValidatorSnapshot(numValidators, 0, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
    })

  })

  describe("When desperationFactor has just been reached twice", () => {
    const blocksSinceDesperation = desperationFactor + Math.floor(desperationFactor/2) + 1 // note: desperationFactor works exponentially

    it("Allows four validators to snapshot", async () => {
      expect(await mayValidatorSnapshot(numValidators, 7, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 8, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 9, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 0, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
    })

    it("Does not allow the fifth validator to snapshot yet", async () => {
      expect(await mayValidatorSnapshot(numValidators, 1, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
    })

  })

  describe("When desperationFactor been reached many times over", () => {
    const blocksSinceDesperation = 1e6

    it("Still only allows at most one third of the validator pool to snapshot", async () => {
      expect(await mayValidatorSnapshot(numValidators, 7, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 8, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 9, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 0, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 1, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
      expect(await mayValidatorSnapshot(numValidators, 2, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
      expect(await mayValidatorSnapshot(numValidators, 3, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
      expect(await mayValidatorSnapshot(numValidators, 4, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
      expect(await mayValidatorSnapshot(numValidators, 5, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
      expect(await mayValidatorSnapshot(numValidators, 6, blocksSinceDesperation, blockSignature, desperationFactor)).to.equal(false)
    })

  })

  describe("When blockSignature changes", () => {
    const blockSignature2 = `0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470`
    const blocksSinceDesperation = 1

    it("Elects a different range of validators", async () => {
      expect(await mayValidatorSnapshot(numValidators, 1, blocksSinceDesperation, blockSignature2, desperationFactor)).to.equal(false)
      expect(await mayValidatorSnapshot(numValidators, 2, blocksSinceDesperation, blockSignature2, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 3, blocksSinceDesperation, blockSignature2, desperationFactor)).to.equal(true)
      expect(await mayValidatorSnapshot(numValidators, 4, blocksSinceDesperation, blockSignature2, desperationFactor)).to.equal(false)
    })

  })

})