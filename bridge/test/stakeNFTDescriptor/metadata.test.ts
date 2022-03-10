import { ethers } from "hardhat"
import { expect } from "../chai-setup"
import { StakeNFTDescriptorMock } from "../../typechain-types/StakeNFTDescriptorMock"

describe("StakeNFTDescriptor: Tests StakeNFTDescriptor methods", async () => {
  let stakeNFTDescriptor: StakeNFTDescriptorMock

  beforeEach(async function () {
    const StakeNFTDescriptorFactory = await ethers.getContractFactory(
      "StakeNFTDescriptorMock"
    )
    stakeNFTDescriptor = await StakeNFTDescriptorFactory.deploy()
    await stakeNFTDescriptor.deployed()
  })

  it("Should return correct token uri", async function () {
    const inputData = {
      tokenId: 123,
      shares: 456,
      freeAfter: 789,
      withdrawFreeAfter: 987,
      accumulatorEth: 123456789,
      accumulatorToken: 123456789
    }
    const expectedTokenUriData =
      "data:application/json;base64,eyJuYW1lIjoiTWFkTkVUIFN0YWtlZCB0b2tlbiBmb3IgcG9zaXRpb24gIzEyMyIsICJkZXNjcmlwdGlvbiI6IlRoaXMgTkZUIHJlcHJlc2VudHMgYSBzdGFrZWQgcG9zaXRpb24gb24gTWFkTkVULlxuVGhlIG93bmVyIG9mIHRoaXMgTkZUIGNhbiBtb2RpZnkgb3IgcmVkZWVtIHRoZSBwb3NpdGlvbi5cbiBTaGFyZXM6IDQ1NlxuRnJlZSBBZnRlcjogNzg5XG5XaXRoZHJhdyBGcmVlIEFmdGVyOiA5ODdcbkFjY3VtdWxhdG9yIEV0aDogMTIzNDU2Nzg5XG5BY2N1bXVsYXRvciBUb2tlbjogMTIzNDU2Nzg5XG5Ub2tlbiBJRDogMTIzIiwgImltYWdlIjogImRhdGE6aW1hZ2Uvc3ZnK3htbDtiYXNlNjQsUEhOMlp5QjNhV1IwYUQwaU5UQXdJaUJvWldsbmFIUTlJalV3TUNJZ2RtbGxkMEp2ZUQwaU1DQXdJRFV3TUNBMU1EQWlJSGh0Ykc1elBTSm9kSFJ3T2k4dmQzZDNMbmN6TG05eVp5OHlNREF3TDNOMlp5SWdlRzFzYm5NNmVHeHBibXM5SjJoMGRIQTZMeTkzZDNjdWR6TXViM0puTHpFNU9Ua3ZlR3hwYm1zblBqeDBaWGgwSUhnOUp6RXdKeUI1UFNjeU1DYytVMmhoY21Wek9pQTBOVFk4TDNSbGVIUStQSFJsZUhRZ2VEMG5NVEFuSUhrOUp6UXdKejVHY21WbElHRm1kR1Z5T2lBM09EazhMM1JsZUhRK1BIUmxlSFFnZUQwbk1UQW5JSGs5SnpZd0p6NVhhWFJvWkhKaGR5QkdjbVZsSUVGbWRHVnlPaUE1T0RjOEwzUmxlSFErUEhSbGVIUWdlRDBuTVRBbklIazlKemd3Sno1QlkyTjFiWFZzWVhSdmNpQW9SVlJJS1RvZ01USXpORFUyTnpnNVBDOTBaWGgwUGp4MFpYaDBJSGc5SnpFd0p5QjVQU2N4TURBblBrRmpZM1Z0ZFd4aGRHOXlJQ2hVYjJ0bGJpazZJREV5TXpRMU5qYzRPVHd2ZEdWNGRENDhMM04yWno0PSJ9"
    const tokenUri = await stakeNFTDescriptor.constructTokenURI(inputData)

    await expect(tokenUri).to.be.equal(expectedTokenUriData)
  })
})
