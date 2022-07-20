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
    it("Integer Square Root 6:     2**15  - 19", async function () {
      const x = 2 ** 15 - 19;
      const trueSqrt = 180;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 7:   3*2**16  + 5", async function () {
      const x = 3*(2 ** 16) + 5;
      const trueSqrt = 443;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 8:   5*2**31  - 27", async function () {
      const x = 5*(2 ** 31) - 27;
      const trueSqrt = 103621;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 9:   7*2**32  + 9", async function () {
      const x = 7*(2 ** 32) + 9;
      const trueSqrt = 173391;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 10: 11*2**63  - 9", async function () {
      const x = BigNumber.from("0x57ffffffffffffff7");
      const trueSqrt = BigNumber.from("0x2585f8b2a");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 11: 13*2**64  + 43", async function () {
      const x = BigNumber.from("0xd000000000000002b");
      const trueSqrt = BigNumber.from("0x39b05688c");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 12: 17*2**127 - 23", async function () {
      const x = BigNumber.from("0x87fffffffffffffffffffffffffffffe9");
      const trueSqrt = BigNumber.from("0x2ea5ca1b674743636");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 13: 19*2**128 + 109", async function () {
      const x = BigNumber.from("0x130000000000000000000000000000006d");
      const trueSqrt = BigNumber.from("0x45be0cd19137e2179");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 14:    2**130 - 5", async function () {
      const x = BigNumber.from("0x3fffffffffffffffffffffffffffffffb");
      const trueSqrt = BigNumber.from("0x1ffffffffffffffff");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 15:    2**255 - 19", async function () {
      const x = BigNumber.from(
        "0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed"
      );
      const trueSqrt = BigNumber.from("0xb504f333f9de6484597d89b3754abe9f");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 16:    2**256 - 1", async function () {
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
      const trueB = BigNumber.from(2500).mul(BigNumber.from("1000000000000000000")); // 2500 * 10**18
      const retB = await sigmoid.p_b();
      expect(retB).to.be.equal(trueB);
    });
    it("P Constant C", async function () {
      const trueC = BigNumber.from("5611050234958650739260304").add(
        BigNumber.from(125).mul(BigNumber.from("1000000000000000000000000000000000000000")) // 125 * 10**39
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
      const retS = await sigmoid.p_inv_s();
      const trueSSquared = c.add(b.mul(b));
      const trueS = await sigmoid.sqrt(trueSSquared);
      expect(retS).to.be.equal(trueS);
      expect(retS.mul(retS)).to.be.equal(trueSSquared);
    });
  });

  describe("P Inverse Constant Tests", async () => {
    it("P Inverse Constant C1", async function () {
      // Ensure
      //    c1 == a[(a+d)*sqrt(c+b**2) + a*b]
      const trueA = await sigmoid.p_a();
      const trueB = await sigmoid.p_b();
      const trueD = await sigmoid.p_d();
      const trueS = await sigmoid.p_inv_s();
      const tmp1 = trueA.add(trueD);
      const tmp2 = trueA.mul(trueB);
      const tmp3 = tmp1.mul(trueS);
      const tmp4 = tmp3.add(tmp2);
      const trueC1 = trueA.mul(tmp4);
      const retC1 = await sigmoid.p_inv_c1();
      expect(retC1).to.be.equal(trueC1);
    });
    it("P Inverse Constant C2", async function () {
      // Ensure
      //    c2 == a + d
      const trueA = await sigmoid.p_a();
      const trueD = await sigmoid.p_d();
      const trueC2 = trueA.add(trueD);
      const retC2 = await sigmoid.p_inv_c2();
      expect(retC2).to.be.equal(trueC2);
    });
    it("P Inverse Constant C3", async function () {
      // Ensure
      //    c3 == d*(2a + d)
      const big2 = BigNumber.from(2);
      const trueA = await sigmoid.p_a();
      const trueD = await sigmoid.p_d();
      const trueC3 = trueD.mul(trueD.add(big2.mul(trueA)));
      const retC3 = await sigmoid.p_inv_c3();
      expect(retC3).to.be.equal(trueC3);
    });
    it("P Inverse Constant D0", async function () {
      // Ensure
      //    d0 == [(a+d)*sqrt(c+b**2) + a*b]**2
      const trueA = await sigmoid.p_a();
      const trueB = await sigmoid.p_b();
      const trueD = await sigmoid.p_d();
      const trueS = await sigmoid.p_inv_s();
      const tmp1 = trueA.add(trueD);
      const tmp2 = trueA.mul(trueB);
      const tmp3 = tmp1.mul(trueS);
      const tmp4 = tmp3.add(tmp2);
      const trueD0 = tmp4.mul(tmp4);
      const retD0 = await sigmoid.p_inv_d0();
      expect(retD0).to.be.equal(trueD0);
    });
    it("P Inverse Constant D1", async function () {
      // Ensure
      //    d1 == 2[a*sqrt(c+b**2) + (a+d)*b]
      const big2 = BigNumber.from(2);
      const trueA = await sigmoid.p_a();
      const trueB = await sigmoid.p_b();
      const trueD = await sigmoid.p_d();
      const trueS = await sigmoid.p_inv_s();
      const tmp1 = trueA.add(trueD);
      const tmp2 = trueA.mul(trueS);
      const tmp3 = tmp1.mul(trueB);
      const tmp4 = tmp2.add(tmp3);
      const trueD1 = big2.mul(tmp4);
      const retD1 = await sigmoid.p_inv_d1();
      expect(retD1).to.be.equal(trueD1);
    });
  });
});
