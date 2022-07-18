import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { MockSigmoid } from "../../typechain-types";
import { expect } from "../chai-setup";

describe("Sigmoid unit tests", async () => {
  let sigmoid: MockSigmoid;
  beforeEach(async () => {
    const sigmoidFactory = await ethers.getContractFactory("MockSigmoid");
    sigmoid = await sigmoidFactory.deploy();
  });

  describe("Integer Square Root Tests", async () => {
    it("Integer Square Root 0:  0", async function () {
      const x = 0;
      const trueSqrt = 0;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 1:  1", async function () {
      const x = 1;
      const trueSqrt = 1;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 2:  4", async function () {
      const x = 4;
      const trueSqrt = 2;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 3:  5", async function () {
      const x = 5;
      const trueSqrt = 2;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 4:  10", async function () {
      const x = 10;
      const trueSqrt = 3;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 5:  257", async function () {
      const x = 257;
      const trueSqrt = 16;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 6:  2**15  - 19", async function () {
      const x = 2 ** 15 - 19;
      const trueSqrt = 180;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 7:  2**16  + 1", async function () {
      const x = 2 ** 16 + 1;
      const trueSqrt = 2 ** 8;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 8:  2**31  - 1", async function () {
      const x = 2 ** 31 - 1;
      const trueSqrt = 46340;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 9:  2**32  + 15", async function () {
      const x = 2 ** 32 + 15;
      const trueSqrt = 2 ** 16;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 10: 2**63  - 25", async function () {
      const x = BigNumber.from("0x7fffffffffffffe7");
      const trueSqrt = BigNumber.from("0xb504f333");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 11: 2**64  + 13", async function () {
      const x = BigNumber.from("0x1000000000000000d");
      const trueSqrt = BigNumber.from("0x100000000");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 12: 2**127 - 1", async function () {
      const x = BigNumber.from("0x7fffffffffffffffffffffffffffffff");
      const trueSqrt = BigNumber.from("0xb504f333f9de6484");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 13: 2**128 + 51", async function () {
      const x = BigNumber.from("0x100000000000000000000000000000033");
      const trueSqrt = BigNumber.from("0x10000000000000000");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 14: 2**130 - 5", async function () {
      const x = BigNumber.from("0x3fffffffffffffffffffffffffffffffb");
      const trueSqrt = BigNumber.from("0x1ffffffffffffffff");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 15: 2**255 - 19", async function () {
      const x = BigNumber.from(
        "0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed"
      );
      const trueSqrt = BigNumber.from("0xb504f333f9de6484597d89b3754abe9f");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 16: 2**256 - 1", async function () {
      const x = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const trueSqrt = BigNumber.from("0xffffffffffffffffffffffffffffffff");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
  });

  describe("safeAbsSub Tests", async () => {
    it("safeAbsSub Test 0", async function () {
      const a = 0;
      const b = 0;
      const trueValue = 0;
      const retValue = await sigmoid.safeAbsSub(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("safeAbsSub Test 1", async function () {
      const a = 255;
      const b = 0;
      const trueValue = a;
      const retValue = await sigmoid.safeAbsSub(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("safeAbsSub Test 2", async function () {
      const a = 0;
      const b = 255;
      const trueValue = b;
      const retValue = await sigmoid.safeAbsSub(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("safeAbsSub Test 3", async function () {
      const a = 2 ** 31 - 1;
      const b = 2 ** 16 + 1;
      const trueValue = a - b;
      const retValue = await sigmoid.safeAbsSub(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("safeAbsSub Test 4", async function () {
      const a = 2 ** 16 + 1;
      const b = 2 ** 31 - 1;
      const trueValue = b - a;
      const retValue = await sigmoid.safeAbsSub(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
  });

  describe("max Tests", async () => {
    it("max Test 0", async function () {
      const a = 0;
      const b = 0;
      const trueValue = 0;
      const retValue = await sigmoid.max(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("max Test 1", async function () {
      const a = 255;
      const b = 0;
      const trueValue = a;
      const retValue = await sigmoid.max(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("max Test 2", async function () {
      const a = 0;
      const b = 255;
      const trueValue = b;
      const retValue = await sigmoid.max(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
  });

  describe("min Tests", async () => {
    it("min Test 0", async function () {
      const a = 0;
      const b = 0;
      const trueValue = 0;
      const retValue = await sigmoid.min(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("min Test 1", async function () {
      const a = 255;
      const b = 0;
      const trueValue = b;
      const retValue = await sigmoid.min(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("min Test 2", async function () {
      const a = 0;
      const b = 255;
      const trueValue = a;
      const retValue = await sigmoid.min(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
  });

  describe("P Constant Tests", async () => {
    it("P Constant A", async function () {
      const trueA = BigNumber.from(200);
      const retA = await sigmoid.p_a();
      expect(retA).to.be.equal(trueA);
    });
    it("P Constant B", async function () {
      const trueB = BigNumber.from(2500).mul(BigNumber.from(10 ** 18));
      const retB = await sigmoid.p_b();
      expect(retB).to.be.equal(trueB);
    });
    it("P Constant C", async function () {
      const trueC = BigNumber.from("5611050234958650739260304").add(
        BigNumber.from(125).mul(BigNumber.from(10 ** 39))
      );
      const retC = await sigmoid.p_c();
      expect(retC).to.be.equal(trueC);
    });
    it("P Constant D", async function () {
      const trueD = 1;
      const retD = await sigmoid.p_d();
      expect(retD).to.be.equal(trueD);
    });
    it("P Constant B and C constraints", async function () {
      const b = await sigmoid.p_b();
      const c = await sigmoid.p_c();
      const retSqrt = await sigmoid.p_inv_sqrt();
      const trueSqrtSquared = c.add(b.mul(b));
      const trueSqrt = await sigmoid.sqrt(trueSqrtSquared);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
  });
});
