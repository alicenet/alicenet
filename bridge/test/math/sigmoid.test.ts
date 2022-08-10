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
      const x = 3 * 2 ** 16 + 5;
      const trueSqrt = 443;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 8:   5*2**31  - 27", async function () {
      const x = 5 * 2 ** 31 - 27;
      const trueSqrt = 103621;
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 9:   7*2**32  + 9", async function () {
      const x = 7 * 2 ** 32 + 9;
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
    it("Integer Square Root 17:    (2**128 - 1)**2", async function () {
      const x = BigNumber.from(
        "0xfffffffffffffffffffffffffffffffe00000000000000000000000000000001"
      );
      const trueSqrt = BigNumber.from("0xffffffffffffffffffffffffffffffff");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
    it("Integer Square Root 18:    (2**128 - 1)**2 - 1", async function () {
      const x = BigNumber.from(
        "0xfffffffffffffffffffffffffffffffe00000000000000000000000000000000"
      );
      const trueSqrt = BigNumber.from("0xfffffffffffffffffffffffffffffffe");
      const retSqrt = await sigmoid.sqrt(x);
      expect(retSqrt).to.be.equal(trueSqrt);
    });
  });

  describe("Integer Square Root Tests: Gas", async () => {
    it("Integer Square Root 0:  0", async function () {
      const x = 0;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 1:  1", async function () {
      const x = 1;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 2:  4", async function () {
      const x = 4;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 3:  5", async function () {
      const x = 5;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 4:  10", async function () {
      const x = 10;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 5:  257", async function () {
      const x = 257;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 6:     2**15  - 19", async function () {
      const x = 2 ** 15 - 19;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 7:   3*2**16  + 5", async function () {
      const x = 3 * 2 ** 16 + 5;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 8:   5*2**31  - 27", async function () {
      const x = 5 * 2 ** 31 - 27;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 9:   7*2**32  + 9", async function () {
      const x = 7 * 2 ** 32 + 9;
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 10: 11*2**63  - 9", async function () {
      const x = BigNumber.from("0x57ffffffffffffff7");
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 11: 13*2**64  + 43", async function () {
      const x = BigNumber.from("0xd000000000000002b");
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 12: 17*2**127 - 23", async function () {
      const x = BigNumber.from("0x87fffffffffffffffffffffffffffffe9");
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 13: 19*2**128 + 109", async function () {
      const x = BigNumber.from("0x130000000000000000000000000000006d");
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 14:    2**130 - 5", async function () {
      const x = BigNumber.from("0x3fffffffffffffffffffffffffffffffb");
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 15:    2**255 - 19", async function () {
      const x = BigNumber.from(
        "0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed"
      );
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 16:    2**256 - 1", async function () {
      const x = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 17:    (2**128 - 1**2", async function () {
      const x = BigNumber.from(
        "0xfffffffffffffffffffffffffffffffe00000000000000000000000000000001"
      );
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
    });
    it("Integer Square Root 18:    (2**128 - 1**2 - 1", async function () {
      const x = BigNumber.from(
        "0xfffffffffffffffffffffffffffffffe00000000000000000000000000000000"
      );
      const tx = await sigmoid.sqrtPublic(x);
      const receipt = await tx.wait();
      expect(receipt.status).to.eq(1);
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
    it("safeAbsSub Test 5", async function () {
      const a = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const b = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const trueValue = BigNumber.from(0);
      const retValue = await sigmoid.safeAbsSub(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("safeAbsSub Test 6", async function () {
      const a = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const b = BigNumber.from(0);
      const trueValue = a;
      const retValue = await sigmoid.safeAbsSub(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("safeAbsSub Test 7", async function () {
      const a = BigNumber.from(0);
      const b = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const trueValue = b;
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
    it("max Test 3", async function () {
      const a = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const b = 0;
      const trueValue = a;
      const retValue = await sigmoid.max(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("max Test 4", async function () {
      const a = 0;
      const b = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
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
    it("min Test 3", async function () {
      const a = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const b = 0;
      const trueValue = b;
      const retValue = await sigmoid.min(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
    it("min Test 4", async function () {
      const a = 0;
      const b = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const trueValue = a;
      const retValue = await sigmoid.min(a, b);
      expect(retValue).to.be.equal(trueValue);
    });
  });

  describe("P Constant Tests", async () => {
    it("P Constant A", async function () {
      const trueA = BigNumber.from(200);
      const retA = await sigmoid.pConstA();
      expect(retA).to.be.equal(trueA);
    });
    it("P Constant B", async function () {
      const trueB = BigNumber.from(2500).mul(
        BigNumber.from("1000000000000000000")
      ); // 2500 * 10**18
      const retB = await sigmoid.pConstB();
      expect(retB).to.be.equal(trueB);
    });
    it("P Constant C", async function () {
      const trueC = BigNumber.from("5611050234958650739260304").add(
        BigNumber.from(125).mul(
          BigNumber.from("1000000000000000000000000000000000000000")
        ) // 125 * 10**39
      );
      const retC = await sigmoid.pConstC();
      expect(retC).to.be.equal(trueC);
    });
    it("P Constant D", async function () {
      const trueD = 1;
      const retD = await sigmoid.pConstD();
      expect(retD).to.be.equal(trueD);
    });
    it("P Constant S", async function () {
      const b = await sigmoid.pConstB();
      const c = await sigmoid.pConstC();
      const retS = await sigmoid.pConstS();
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
      const trueA = await sigmoid.pConstA();
      const trueB = await sigmoid.pConstB();
      const trueD = await sigmoid.pConstD();
      const trueS = await sigmoid.pConstS();
      const tmp1 = trueA.add(trueD);
      const tmp2 = trueA.mul(trueB);
      const tmp3 = tmp1.mul(trueS);
      const tmp4 = tmp3.add(tmp2);
      const trueC1 = trueA.mul(tmp4);
      const retC1 = await sigmoid.pInverseConstC1();
      expect(retC1).to.be.equal(trueC1);
    });
    it("P Inverse Constant C2", async function () {
      // Ensure
      //    c2 == a + d
      const trueA = await sigmoid.pConstA();
      const trueD = await sigmoid.pConstD();
      const trueC2 = trueA.add(trueD);
      const retC2 = await sigmoid.pInverseConstC2();
      expect(retC2).to.be.equal(trueC2);
    });
    it("P Inverse Constant C3", async function () {
      // Ensure
      //    c3 == d*(2a + d)
      const big2 = BigNumber.from(2);
      const trueA = await sigmoid.pConstA();
      const trueD = await sigmoid.pConstD();
      const trueC3 = trueD.mul(trueD.add(big2.mul(trueA)));
      const retC3 = await sigmoid.pInverseConstC3();
      expect(retC3).to.be.equal(trueC3);
    });
    it("P Inverse Constant D0", async function () {
      // Ensure
      //    d0 == [(a+d)*sqrt(c+b**2) + a*b]**2
      const trueA = await sigmoid.pConstA();
      const trueB = await sigmoid.pConstB();
      const trueD = await sigmoid.pConstD();
      const trueS = await sigmoid.pConstS();
      const tmp1 = trueA.add(trueD);
      const tmp2 = trueA.mul(trueB);
      const tmp3 = tmp1.mul(trueS);
      const tmp4 = tmp3.add(tmp2);
      const trueD0 = tmp4.mul(tmp4);
      const retD0 = await sigmoid.pInverseConstD0();
      expect(retD0).to.be.equal(trueD0);
    });
    it("P Inverse Constant D1", async function () {
      // Ensure
      //    d1 == 2[a*sqrt(c+b**2) + (a+d)*b]
      const big2 = BigNumber.from(2);
      const trueA = await sigmoid.pConstA();
      const trueB = await sigmoid.pConstB();
      const trueD = await sigmoid.pConstD();
      const trueS = await sigmoid.pConstS();
      const tmp1 = trueA.add(trueD);
      const tmp2 = trueA.mul(trueS);
      const tmp3 = tmp1.mul(trueB);
      const tmp4 = tmp2.add(tmp3);
      const trueD1 = big2.mul(tmp4);
      const retD1 = await sigmoid.pInverseConstD1();
      expect(retD1).to.be.equal(trueD1);
    });
  });

  describe("P Function Tests", async () => {
    it("Evaluate P(0)", async function () {
      // Confirm
      //      P(0) == 0
      const value = BigNumber.from(0);
      const trueValue = BigNumber.from(0);
      const retValue = await sigmoid.p(value);
      expect(retValue).to.be.equal(trueValue);
    });
    it("Evaluate P(b)", async function () {
      // Confirm
      //      P(b) == a*sqrt(c+b^2) - sqrt(a**2 * c) + (a+d)*b
      const trueA = await sigmoid.pConstA();
      const trueB = await sigmoid.pConstB();
      const trueC = await sigmoid.pConstC();
      const trueD = await sigmoid.pConstD();
      const trueS = await sigmoid.pConstS();
      const value = trueB;
      const tmp1 = trueA.mul(trueS);
      const tmp2P = trueA.mul(trueA.mul(trueC));
      const tmp2 = await sigmoid.sqrt(tmp2P);
      const tmp3 = trueB.mul(trueA.add(trueD));
      const trueValue = tmp1.add(tmp3.sub(tmp2));
      const retValue = await sigmoid.p(value);
      expect(retValue).to.be.equal(trueValue);
    });
    it("Evaluate P(2b)", async function () {
      // Confirm
      //      P(2b) == 2b*(a+d)
      const big2 = BigNumber.from(2);
      const trueA = await sigmoid.pConstA();
      const trueB = await sigmoid.pConstB();
      const trueD = await sigmoid.pConstD();
      const value = trueB.add(trueB);
      const trueValue = big2.mul(trueB.mul(trueA.add(trueD)));
      const retValue = await sigmoid.p(value);
      expect(retValue).to.be.equal(trueValue);
    });
    it("Evaluate P(b+k) - P(b-k); ", async function () {
      // Confirm
      //      P(b) == a*sqrt(c+b^2) - sqrt(a**2 * c) + (a+d)*b
      const iterations = 10;
      for (let i = 0; i < iterations; i++) {
        const K = BigNumber.from(10 ** (i + 1));
        const big2 = BigNumber.from(2);
        const trueA = await sigmoid.pConstA();
        const trueB = await sigmoid.pConstB();
        const trueD = await sigmoid.pConstD();
        const value1 = trueB.add(K);
        const value2 = trueB.sub(K);
        const trueDiff = big2.mul(K.mul(trueA.add(trueD)));
        const retValue1 = await sigmoid.p(value1);
        const retValue2 = await sigmoid.p(value2);
        const compDiff = retValue1.sub(retValue2);
        expect(compDiff).to.be.equal(trueDiff);
      }
    });
    it("Evaluate P; passes", async function () {
      // Confirm valid input for 2**120 - 1
      const value = BigNumber.from("0xffffffffffffffffffffffffffffff");
      const retValue = await sigmoid.p(value);
      await expect(retValue).to.not.equal(0);
    });
    it("Evaluate P; overflow 1", async function () {
      // Confirm overflow: 2**256 - 1
      const value = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const tx = sigmoid.p(value);
      await expect(tx).to.be.reverted;
    });
    it("Evaluate P; overflow 2", async function () {
      // Confirm overflow: 2**128 - 1
      const value = BigNumber.from("0xffffffffffffffffffffffffffffffff");
      const tx = sigmoid.p(value);
      await expect(tx).to.be.reverted;
    });
  });

  describe("P Inverse Function Tests", async () => {
    it("Evaluate P_inv(0)", async function () {
      // Confirm
      //      P_inv(0) == 0
      const value = BigNumber.from(0);
      const trueValue = BigNumber.from(0);
      const retValue = await sigmoid.pInverse(value);
      expect(retValue).to.be.equal(trueValue);
    });
    it("Evaluate P_inv(P(b))", async function () {
      // Confirm
      //      P_inv(P(b)) == b
      const trueB = await sigmoid.pConstB();
      const initialValue = trueB;
      const value = await sigmoid.p(initialValue);
      const retValue = await sigmoid.pInverse(value);
      expect(retValue).to.be.equal(initialValue);
    });
    it("Evaluate P_inv(P(2b))", async function () {
      // Confirm
      //      P_inv(P(2b)) == 2b
      const big2 = BigNumber.from(2);
      const trueB = await sigmoid.pConstB();
      const initialValue = big2.mul(trueB);
      const value = await sigmoid.p(initialValue);
      const retValue = await sigmoid.pInverse(value);
      expect(retValue).to.be.equal(initialValue);
    });
    it("Evaluate P_inv(P(3b))", async function () {
      // Confirm
      //      P_inv(P(3b)) == 3b
      const big3 = BigNumber.from(3);
      const trueB = await sigmoid.pConstB();
      const initialValue = big3.mul(trueB);
      const value = await sigmoid.p(initialValue);
      const retValue = await sigmoid.pInverse(value);
      expect(retValue).to.be.equal(initialValue);
    });
    it("Evaluate P_inv(P(4b))", async function () {
      // Confirm
      //      P_inv(P(4b)) == 4b
      const big4 = BigNumber.from(4);
      const trueB = await sigmoid.pConstB();
      const initialValue = big4.mul(trueB);
      const value = await sigmoid.p(initialValue);
      const retValue = await sigmoid.pInverse(value);
      expect(retValue).to.be.equal(initialValue);
    });
    it("Evaluate P_inv; passes", async function () {
      // Confirm valid input for 2**120 - 1
      const value = BigNumber.from("0xffffffffffffffffffffffffffffff");
      const retValue = await sigmoid.pInverse(value);
      await expect(retValue).to.not.equal(0);
    });
    it("Evaluate P_inv; overflow 1", async function () {
      // Confirm overflow: 2**256 - 1
      const value = BigNumber.from(
        "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
      );
      const tx = sigmoid.pInverse(value);
      await expect(tx).to.be.reverted;
    });
    it("Evaluate P_inv; overflow 2", async function () {
      // Confirm overflow: 2**128 - 1
      const value = BigNumber.from("0xffffffffffffffffffffffffffffffff");
      const tx = sigmoid.pInverse(value);
      await expect(tx).to.be.reverted;
    });
  });
});
