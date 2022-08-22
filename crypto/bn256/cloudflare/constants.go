// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloudflare

import (
	"math/big"
)

func bigFromBase10(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 10)
	return n
}

// u is the BN parameter.
var u = bigFromBase10("4965661367192848881")

// Order is the number of elements in both G₁ and G₂: 36u⁴+36u³+18u²+6u+1.
// Needs to be highly 2-adic for efficient SNARK key and proof generation.
// Order - 1 = 2^28 * 3^2 * 13 * 29 * 983 * 11003 * 237073 * 405928799 * 1670836401704629 * 13818364434197438864469338081.
// Refer to https://eprint.iacr.org/2013/879.pdf and https://eprint.iacr.org/2013/507.pdf for more information on these parameters.
var Order = bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495617")

// P is a prime over which we form a basic field: 36u⁴+36u³+24u²+6u+1.
var P = bigFromBase10("21888242871839275222246405745257275088696311157297823662689037894645226208583")

// p2 is p, represented as little-endian 64-bit words.
var p2 = [4]uint64{0x3c208c16d87cfd47, 0x97816a916871ca8d, 0xb85045b68181585d, 0x30644e72e131a029}

// np is the negative inverse of p, mod 2^256.
var np = [4]uint64{0x87d20782e4866389, 0x9ede7d651eca6ac9, 0xd8afcbd01833da80, 0xf57a22b791888c6b}

// rN1 is R^-1 where R = 2^256 mod p.
var rN1 = &gfP{0xed84884a014afa37, 0xeb2022850278edf8, 0xcf63e9cfb74492d9, 0x2e67157159e5c639}

// r2 is R^2 where R = 2^256 mod p.
var r2 = &gfP{0xf32cfc5b538afa89, 0xb5e71911d44501fb, 0x47ab1eff0a417ff6, 0x06d89f71cab8351f}

// r3 is R^3 where R = 2^256 mod p.
var r3 = &gfP{0xb1cd6dafda1530df, 0x62f210e6a7283db6, 0xef7f0b0c0ada0afb, 0x20fd6e902d592544}

// xiToPMinus1Over6 is ξ^((p-1)/6) where ξ = i+9.
var xiToPMinus1Over6 = &gfP2{gfP{0xa222ae234c492d72, 0xd00f02a4565de15b, 0xdc2ff3a253dfc926, 0x10a75716b3899551}, gfP{0xaf9ba69633144907, 0xca6b1d7387afb78a, 0x11bded5ef08a2087, 0x02f34d751a1f3a7c}}

// xiToPMinus1Over3 is ξ^((p-1)/3) where ξ = i+9.
var xiToPMinus1Over3 = &gfP2{gfP{0x6e849f1ea0aa4757, 0xaa1c7b6d89f89141, 0xb6e713cdfae0ca3a, 0x26694fbb4e82ebc3}, gfP{0xb5773b104563ab30, 0x347f91c8a9aa6454, 0x7a007127242e0991, 0x1956bcd8118214ec}}

// xiToPMinus1Over2 is ξ^((p-1)/2) where ξ = i+9.
var xiToPMinus1Over2 = &gfP2{gfP{0xa1d77ce45ffe77c7, 0x07affd117826d1db, 0x6d16bd27bb7edc6b, 0x2c87200285defecc}, gfP{0xe4bbdd0c2936b629, 0xbb30f162e133bacb, 0x31a9d1b6f9645366, 0x253570bea500f8dd}}

// xiToPSquaredMinus1Over3 is ξ^((p²-1)/3) where ξ = i+9.
var xiToPSquaredMinus1Over3 = &gfP{0x3350c88e13e80b9c, 0x7dce557cdb5e56b9, 0x6001b4b8b615564a, 0x2682e617020217e0}

// xiTo2PSquaredMinus2Over3 is ξ^((2p²-2)/3) where ξ = i+9 (a cubic root of unity, mod p).
var xiTo2PSquaredMinus2Over3 = &gfP{0x71930c11d782e155, 0xa6bb947cffbe3323, 0xaa303344d4741444, 0x2c3b3f0d26594943}

// xiToPSquaredMinus1Over6 is ξ^((1p²-1)/6) where ξ = i+9 (a cubic root of -1, mod p).
var xiToPSquaredMinus1Over6 = &gfP{0xca8d800500fa1bf2, 0xf0c5d61468b39769, 0x0e201271ad0d4418, 0x04290f65bad856e6}

// xiTo2PMinus2Over3 is ξ^((2p-2)/3) where ξ = i+9.
var xiTo2PMinus2Over3 = &gfP2{gfP{0x5dddfd154bd8c949, 0x62cb29a5a4445b60, 0x37bc870a0c7dd2b9, 0x24830a9d3171f0fd}, gfP{0x7361d77f843abe92, 0xa5bb2bd3273411fb, 0x9c941f314b3e2399, 0x15df9cddbb9fd3ec}}

// pMinus2 is the representation of P-2 in [4]uint64 array; this is used in gfP Invert
// pMinus2 == 21888242871839275222246405745257275088696311157297823662689037894645226208581.
var pMinus2 = [4]uint64{0x3c208c16d87cfd45, 0x97816a916871ca8d, 0xb85045b68181585d, 0x30644e72e131a029}

// pPlus1 is the representation of P+1 in [4]uint64 array.
var pPlus1 = [4]uint64{0x3c208c16d87cfd48, 0x97816a916871ca8d, 0xb85045b68181585d, 0x30644e72e131a029}

// pPlus1Over4 is the representation of (P+1)/4 in [4]uint64 array; this is used in gfP Sqrt
// pPlus1Over4 == 5472060717959818805561601436314318772174077789324455915672259473661306552146.
var pPlus1Over4 = [4]uint64{0x4f082305b61f3f52, 0x65e05aa45a1c72a3, 0x6e14116da0605617, 0x0c19139cb84c680a}

// pMinus3Over4 is the representation of (P-3)/4 in [4]uint64 array; this is used in gfP2 Sqrt.
var pMinus3Over4 = [4]uint64{0x4f082305b61f3f51, 0x65e05aa45a1c72a3, 0x6e14116da0605617, 0x0c19139cb84c680a}

// For Legendre symbol; that is, for determining if square roots exist.
var pMinus1Over2Big = bigFromBase10("10944121435919637611123202872628637544348155578648911831344518947322613104291")

// pMinus1Over2 is the representation of (P-1)/2 in [4]uint64 array; this is used in Legendre.
var pMinus1Over2 = [4]uint64{0x9e10460b6c3e7ea3, 0xcbc0b548b438e546, 0xdc2822db40c0ac2e, 0x183227397098d014}

// two256ModP is 2^256 mod P; this is used in hashToBase.
var two256ModP = bigFromBase10("6350874878119819312338956282401532409788428879151445726012394534686998597021")

// gfP constant for use in HashToG1 function; (-1 + sqrt(-3))/2.
var g1HashConst1 = &gfP{0x71930c11d782e155, 0xa6bb947cffbe3323, 0xaa303344d4741444, 0x2c3b3f0d26594943}

// gfP constant for use in HashToG1 function; sqrt(-3).
var g1HashConst2 = &gfP{0x3e424383c39ad5b9, 0x28ed3f00245fdc6a, 0x4a2e7e8c1e5ebdfa, 0x05b858f624573163}

// gfP constant for use in HashToG1 function; 1/3.
var g1HashConst3 = &gfP{0xafd49a8c34aeae4c, 0xe0a8c73e1f684743, 0xb4ea4db753538a2d, 0x14cf9766d3bdd51d}

// gfP constant for use in HashToG1 function; g(1) == 1 + curveB (== 4).
var g1HashConst4 = &gfP{0x115482203dbf392d, 0x926242126eaa626a, 0xe16a48076063c052, 0x07c5909386eddc93}

// gfP2 constant for use in HashToG2 function; (-1 + sqrt(-3))/2.
var g2HashConst1 = &gfP2{gfP{0x0000000000000000, 0x0000000000000000, 0x0000000000000000, 0x0000000000000000}, gfP{0x71930c11d782e155, 0xa6bb947cffbe3323, 0xaa303344d4741444, 0x2c3b3f0d26594943}}

// gfP2 constant for use in HashToG2 function; sqrt(-3).
var g2HashConst2 = &gfP2{gfP{0x0000000000000000, 0x0000000000000000, 0x0000000000000000, 0x0000000000000000}, gfP{0x3e424383c39ad5b9, 0x28ed3f00245fdc6a, 0x4a2e7e8c1e5ebdfa, 0x05b858f624573163}}

// gfP2 constant for use in HashToG2 function; 1/3.
var g2HashConst3 = &gfP2{gfP{0x0000000000000000, 0x0000000000000000, 0x0000000000000000, 0x0000000000000000}, gfP{0xafd49a8c34aeae4c, 0xe0a8c73e1f684743, 0xb4ea4db753538a2d, 0x14cf9766d3bdd51d}}

// gfP2 constant for use in HashToG2 function; g'(1) == 1 + twistB.
var g2HashConst4 = &gfP2{gfP{0x38e7ecccd1dcff67, 0x65f0b37d93ce0d3e, 0xd749d0dd22ac00aa, 0x0141b9ce4a688d4d}, gfP{0xd335f05a64ca12fe, 0x75029bbec388940d, 0xd4d64ba9406d402e, 0x02baef80fc5ae772}}

// Cofactor in the twist curve; twistCofactor == 2P - Order.
var twistCofactor = bigFromBase10("21888242871839275222246405745257275088844257914179612981679871602714643921549")
