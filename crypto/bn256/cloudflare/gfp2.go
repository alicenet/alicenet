package cloudflare

// For details of the algorithms used, see "Multiplication and Squaring on
// Pairing-Friendly Fields, Devegili et al.
// http://eprint.iacr.org/2006/471.pdf.

// gfP2 implements a field of size p² as a quadratic extension of the base field
// where i²=-1.
type gfP2 struct {
	x, y gfP // value is xi+y.
}

func gfP2Decode(in *gfP2) *gfP2 {
	out := &gfP2{}
	montDecode(&out.x, &in.x)
	montDecode(&out.y, &in.y)
	return out
}

func gfP2Encode(in *gfP2) *gfP2 {
	out := &gfP2{}
	montEncode(&out.x, &in.x)
	montEncode(&out.y, &in.y)
	return out
}

func (e *gfP2) String() string {
	return "(" + e.x.String() + ", " + e.y.String() + ")"
}

func (e *gfP2) IsEqual(a *gfP2) bool {
	return e.String() == a.String()
}

func (e *gfP2) Set(a *gfP2) *gfP2 {
	e.x.Set(&a.x)
	e.y.Set(&a.y)
	return e
}

func (e *gfP2) SetZero() *gfP2 {
	e.x = gfP{0}
	e.y = gfP{0}
	return e
}

func (e *gfP2) SetOne() *gfP2 {
	e.x = gfP{0}
	e.y = *newGFp(1)
	return e
}

func (e *gfP2) IsZero() bool {
	zero := gfP{0}
	return e.x == zero && e.y == zero
}

func (e *gfP2) IsOne() bool {
	zero, one := gfP{0}, *newGFp(1)
	return e.x == zero && e.y == one
}

func (e *gfP2) Conjugate(a *gfP2) *gfP2 {
	e.y.Set(&a.y)
	gfpNeg(&e.x, &a.x)
	return e
}

func (e *gfP2) Neg(a *gfP2) *gfP2 {
	gfpNeg(&e.x, &a.x)
	gfpNeg(&e.y, &a.y)
	return e
}

func (e *gfP2) Add(a, b *gfP2) *gfP2 {
	gfpAdd(&e.x, &a.x, &b.x)
	gfpAdd(&e.y, &a.y, &b.y)
	return e
}

func (e *gfP2) Sub(a, b *gfP2) *gfP2 {
	gfpSub(&e.x, &a.x, &b.x)
	gfpSub(&e.y, &a.y, &b.y)
	return e
}

// See "Multiplication and Squaring in Pairing-Friendly Fields",
// http://eprint.iacr.org/2006/471.pdf
func (e *gfP2) Mul(a, b *gfP2) *gfP2 {
	tx, t := &gfP{}, &gfP{}
	gfpMul(tx, &a.x, &b.y)
	gfpMul(t, &b.x, &a.y)
	gfpAdd(tx, tx, t)

	ty := &gfP{}
	gfpMul(ty, &a.y, &b.y)
	gfpMul(t, &a.x, &b.x)
	gfpSub(ty, ty, t)

	e.x.Set(tx)
	e.y.Set(ty)
	return e
}

func (e *gfP2) MulScalar(a *gfP2, b *gfP) *gfP2 {
	gfpMul(&e.x, &a.x, b)
	gfpMul(&e.y, &a.y, b)
	return e
}

// MulXi sets e=ξa where ξ=i+9 and then returns e.
func (e *gfP2) MulXi(a *gfP2) *gfP2 {
	// (xi+y)(i+9) = (9x+y)i+(9y-x)
	tx := &gfP{}
	gfpAdd(tx, &a.x, &a.x)
	gfpAdd(tx, tx, tx)
	gfpAdd(tx, tx, tx)
	gfpAdd(tx, tx, &a.x)

	gfpAdd(tx, tx, &a.y)

	ty := &gfP{}
	gfpAdd(ty, &a.y, &a.y)
	gfpAdd(ty, ty, ty)
	gfpAdd(ty, ty, ty)
	gfpAdd(ty, ty, &a.y)

	gfpSub(ty, ty, &a.x)

	e.x.Set(tx)
	e.y.Set(ty)
	return e
}

func (e *gfP2) Square(a *gfP2) *gfP2 {
	// Complex squaring algorithm:
	// (xi+y)² = (x+y)(y-x) + 2*i*x*y
	tx, ty := &gfP{}, &gfP{}
	gfpSub(tx, &a.y, &a.x)
	gfpAdd(ty, &a.x, &a.y)
	gfpMul(ty, tx, ty)

	gfpMul(tx, &a.x, &a.y)
	gfpAdd(tx, tx, tx)

	e.x.Set(tx)
	e.y.Set(ty)
	return e
}

// Invert computes the multiplicative inverse in gfP2
func (e *gfP2) Invert(a *gfP2) *gfP2 {
	// See "Implementing cryptographic pairings", M. Scott, section 3.2.
	// ftp://136.206.11.249/pub/crypto/pairings.pdf
	t1, t2 := &gfP{}, &gfP{}
	gfpMul(t1, &a.x, &a.x)
	gfpMul(t2, &a.y, &a.y)
	gfpAdd(t1, t1, t2)

	inv := &gfP{}
	inv.Invert(t1)

	gfpNeg(t1, &a.x)

	gfpMul(&e.x, t1, inv)
	gfpMul(&e.y, &a.y, inv)
	return e
}

// Legendre determines if a square exists in gfP2
func (e *gfP2) Legendre() int {
	ap1 := &gfP{}
	gfpMul(ap1, &e.x, &e.x)
	ap2 := &gfP{}
	gfpMul(ap2, &e.y, &e.y)
	alpha := &gfP{}
	gfpAdd(alpha, ap1, ap2)
	return alpha.Legendre()
}

// exp performs modular exponentiation using square and multiply method.
func (e *gfP2) exp(f *gfP2, bits [4]uint64) {
	sum, power := &gfP2{}, &gfP2{}
	sum.SetOne()
	power.Set(f)

	for word := 0; word < 4; word++ {
		for bit := uint(0); bit < 64; bit++ {
			if (bits[word]>>bit)&1 == 1 {
				sum.Mul(sum, power)
			}
			power.Square(power)
		}
	}

	e.Set(sum)
}

// Sqrt computes the square roots in gfP2.
// This assumes a square root of e exists; use Legendre to determine existence.
// This algorithm is taken from Algorithm 9 in the 2012 paper
//
//	Square root computation over even extension fields
//	by Gora Adj and Francisco Rodrı́guez-Henrı́quez
//
// Nonexistence of a square root will lead to incorrect results.
func (e *gfP2) Sqrt(f *gfP2) {
	t := &gfP2{}
	t.exp(f, pMinus3Over4)
	y := &gfP2{}
	y.Mul(t, f)
	alpha := &gfP2{}
	alpha.Mul(t, y)

	gfp2One := &gfP2{}
	gfp2One.SetOne()
	gfp2NegOne := &gfP2{}
	gfp2NegOne.Neg(gfp2One)
	gfp2Zero := &gfP2{}
	gfp2Zero.SetZero()

	gfp2I := &gfP2{}
	gfp2I.x.Set(newGFp(1))
	gfp2I.y.Set(newGFp(0))

	x := &gfP2{}
	b := &gfP2{}

	if *alpha == *gfp2NegOne {
		gfp2One.Add(gfp2One, gfp2Zero)
		gfp2Zero.exp(gfp2Zero, pMinus1Over2)
		x.Mul(gfp2I, y)
	} else {
		b.Add(gfp2One, alpha)
		b.exp(b, pMinus1Over2)
		x.Mul(b, y)
	}

	e.Set(x)
}

// We use the sign convention from Wahby and Boneh 2019 paper,
// which takes the sign of the imaginary component;
// if zero, use the sign of the real component.
func sign0GFp2(e *gfP2) int {
	gfpZero := &gfP{}
	if !e.x.IsEqual(gfpZero) {
		return sign0(&e.x)
	}
	return sign0(&e.y)
}
