import { ethers } from "hardhat";
import { StakingDescriptorMock } from "../../typechain-types";
import { expect } from "../chai-setup";

describe("stakingDescriptor: Tests stakingDescriptor methods", async () => {
  let stakingDescriptor: StakingDescriptorMock;

  beforeEach(async function () {
    const stakingDescriptorFactory = await ethers.getContractFactory(
      "StakingDescriptorMock"
    );
    stakingDescriptor = await stakingDescriptorFactory.deploy();
    await stakingDescriptor.deployed();
  });

  it("Should return correct token uri", async function () {
    const inputData = {
      tokenId: 123,
      shares: 456,
      freeAfter: 789,
      withdrawFreeAfter: 987,
      accumulatorEth: 123456789,
      accumulatorToken: 123456789,
    };
    const expectedTokenUriData =
      "data:application/json;base64,eyJuYW1lIjoiQWxpY2VOZXQgU3Rha2VkIHRva2VuIGZvciBwb3NpdGlvbiAjMTIzIiwgImRlc2NyaXB0aW9uIjoiVGhpcyBORlQgcmVwcmVzZW50cyBhIHN0YWtlZCBwb3NpdGlvbiBvbiBBbGljZU5ldC5cblRoZSBvd25lciBvZiB0aGlzIE5GVCBjYW4gbW9kaWZ5IG9yIHJlZGVlbSB0aGUgcG9zaXRpb24uXG4gU2hhcmVzOiA0NTZcbkZyZWUgQWZ0ZXI6IDc4OVxuV2l0aGRyYXcgRnJlZSBBZnRlcjogOTg3XG5BY2N1bXVsYXRvciBFdGg6IDEyMzQ1Njc4OVxuQWNjdW11bGF0b3IgVG9rZW46IDEyMzQ1Njc4OVxuVG9rZW4gSUQ6IDEyMyIsICJpbWFnZSI6ICJkYXRhOmltYWdlL3N2Zyt4bWw7YmFzZTY0LFBITjJaeUIzYVdSMGFEMGlOVEF3SWlCb1pXbG5hSFE5SWpVd01DSWdkbWxsZDBKdmVEMGlNQ0F3SURVd01DQTFNREFpSUhodGJHNXpQU0pvZEhSd09pOHZkM2QzTG5jekxtOXlaeTh5TURBd0wzTjJaeUlnZUcxc2JuTTZlR3hwYm1zOUoyaDBkSEE2THk5M2QzY3Vkek11YjNKbkx6RTVPVGt2ZUd4cGJtc25QangwWlhoMElIZzlKekV3SnlCNVBTY3lNQ2MrVTJoaGNtVnpPaUEwTlRZOEwzUmxlSFErUEhSbGVIUWdlRDBuTVRBbklIazlKelF3Sno1R2NtVmxJR0ZtZEdWeU9pQTNPRGs4TDNSbGVIUStQSFJsZUhRZ2VEMG5NVEFuSUhrOUp6WXdKejVYYVhSb1pISmhkeUJHY21WbElFRm1kR1Z5T2lBNU9EYzhMM1JsZUhRK1BIUmxlSFFnZUQwbk1UQW5JSGs5Snpnd0p6NUJZMk4xYlhWc1lYUnZjaUFvUlZSSUtUb2dNVEl6TkRVMk56ZzVQQzkwWlhoMFBqeDBaWGgwSUhnOUp6RXdKeUI1UFNjeE1EQW5Qa0ZqWTNWdGRXeGhkRzl5SUNoVWIydGxiaWs2SURFeU16UTFOamM0T1R3dmRHVjRkRDQ4TDNOMlp6ND0ifQ==";
    const tokenUri = await stakingDescriptor.constructTokenURI(inputData);

    await expect(tokenUri).to.be.equal(expectedTokenUriData);
  });
});
