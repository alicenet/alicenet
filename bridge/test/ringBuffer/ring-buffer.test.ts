import { ethers } from "hardhat"

describe("Ring Buffer Library",async () => {
    let ringBuffer
    beforeEach(async () => {
        const ringBufferBase = await ethers.getContractFactory("SnapshotsRingBufferMock")
        ringBuffer = await ringBufferBase.deploy();
    })
    it("stores a snapshot on the ring buffer",async () => {
        
    })
    it("stores 7 snapshot on the ring buffer",async () => {
        
    })
    it("attempts to get a snapshot that is no longer in the ring buffer",async () => {
        
    })
    it("unsafeSet into an arbitrary location in ring buffer",async () => {
        
    })
    
})