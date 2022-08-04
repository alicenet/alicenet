package cloudflare

import (
	"errors"
)

const (
	g1UncompFlag byte = 0x10
	g1CompFlag   byte = 0x08
	g1YOddFlag   byte = 0x01

	g2UncompFlag byte = 0x40
	g2CompFlag   byte = 0x20
	g2YXOddFlag  byte = 0x01
	g2YYOddFlag  byte = 0x02
)

// SerializeCompressed returns the compressed serialization of the G1 point
func (e *G1) SerializeCompressed() []byte {
	ret := make([]byte, 0, 1+numBytes)
	slice := e.Marshal()
	xBytes := slice[:numBytes]
	yIsOdd := (slice[2*numBytes-1] & 1) == 1
	bitFlag := g1CompFlag
	if yIsOdd {
		bitFlag |= g1YOddFlag
	}
	ret = append(ret, bitFlag)
	ret = append(ret, xBytes...)
	return ret
}

// SerializeUncompressed returns the uncompressed serialization of the G1 point
func (e *G1) SerializeUncompressed() []byte {
	ret := make([]byte, 0, 1+2*numBytes)
	ret = append(ret, g1UncompFlag)
	slice := e.Marshal()
	ret = append(ret, slice...)
	return ret
}

// Deserialize converts a byte slice into the corresponding G1 point
func (e *G1) Deserialize(m []byte) error {
	if e.p == nil {
		e.p = &curvePoint{}
	}
	if len(m) == 0 {
		return errors.New("Invalid byte slice")
	}
	format := m[0]
	yIsOdd := (format & g1YOddFlag) == g1YOddFlag
	flagByte := format & (^g1YOddFlag)
	data := m[1:]

	switch flagByte {
	case g1UncompFlag:
		// Ensure data has proper length; store both x and y
		if len(data) != 2*numBytes {
			return errors.New("Invalid byte slice")
		}
		_, err := e.Unmarshal(data)
		if err != nil {
			return err
		}
		return nil

	case g1CompFlag:
		// Ensure data has proper length; only store x
		if len(data) != numBytes {
			return errors.New("Invalid byte slice")
		}
		// Need to unmarshal bytes then encode
		if err := e.p.x.Unmarshal(data); err != nil {
			return err
		}
		montEncode(&e.p.x, &e.p.x)
		zero := gfP{0}
		if e.p.x == zero {
			if yIsOdd {
				// x bytes are zero but we y is odd; invalid
				return errors.New("Invalid byte slice: improper encoding of identity")
			}
			// We finish encoding the identity element
			e.p.y = *newGFp(1)
			e.p.z = gfP{0}
			e.p.t = gfP{0}
			return nil
		}

		// We now need to compute correct y value
		t := computeG1YValue(&e.p.x, yIsOdd)
		e.p.y.Set(t)
		e.p.z = *newGFp(1)
		e.p.t = *newGFp(1)
		// Confirm that we have a valid curve point
		if !e.p.IsOnCurve() {
			return ErrMalformedPoint
		}
		return nil
	default:
		return errors.New("Invalid byte slice: improper formatting")
	}
}

// computeG1YValue computes the correct y value;
// that is, if x is a valid coordinate, then compute y such that
// (x, y) is on the curve and y is odd or even as desired.
func computeG1YValue(x *gfP, yIsOdd bool) *gfP {
	t := &gfP{}
	// Compute t == x^3 + curveB
	gfpMul(t, x, x)
	gfpMul(t, t, x)
	gfpAdd(t, t, curveB)
	y := &gfP{}
	// Sqrt computation always succeeds even if sqrt does not exist;
	// validity of the (x, y) pair is not checked here
	y.Sqrt(t)
	montDecode(t, y)
	currentlyOdd := (t[0] & 1) == 1
	if currentlyOdd != yIsOdd {
		gfpNeg(y, y)
	}
	return y
}

// SerializeCompressed returns the compressed serialization of the G2 point
func (e *G2) SerializeCompressed() []byte {
	ret := make([]byte, 0, 1+2*numBytes)
	slice := e.Marshal()
	xBytes := slice[:2*numBytes]
	yXIsOdd := (slice[3*numBytes-1] & 1) == 1
	yYIsOdd := (slice[4*numBytes-1] & 1) == 1
	bitFlag := g2CompFlag
	if yXIsOdd {
		bitFlag |= g2YXOddFlag
	}
	if yYIsOdd {
		bitFlag |= g2YYOddFlag
	}
	ret = append(ret, bitFlag)
	ret = append(ret, xBytes...)
	return ret
}

// SerializeUncompressed returns the uncompressed serialization of the G2 point
func (e *G2) SerializeUncompressed() []byte {
	ret := make([]byte, 0, 1+4*numBytes)
	ret = append(ret, g2UncompFlag)
	slice := e.Marshal()
	ret = append(ret, slice...)
	return ret
}

// Deserialize converts a byte slice into the corresponding G2 point
func (e *G2) Deserialize(m []byte) error {
	if e.p == nil {
		e.p = &twistPoint{}
	}
	if len(m) == 0 {
		return errors.New("Invalid byte slice")
	}
	format := m[0]
	yXIsOdd := (format & g2YXOddFlag) == g2YXOddFlag
	yYIsOdd := (format & g2YYOddFlag) == g2YYOddFlag
	flagByte := format & (^(g2YXOddFlag | g2YYOddFlag))
	data := m[1:]

	switch flagByte {
	case g2UncompFlag:
		// Ensure data has proper length; store both x and y
		if len(data) != 4*numBytes {
			return errors.New("Invalid byte slice")
		}
		_, err := e.Unmarshal(data)
		if err != nil {
			return err
		}
		return nil

	case g2CompFlag:
		// Ensure data has proper length; only store x
		if len(data) != 2*numBytes {
			return errors.New("Invalid byte slice")
		}
		// Need to unmarshal bytes then encode
		if err := e.p.x.x.Unmarshal(data[:numBytes]); err != nil {
			return err
		}
		montEncode(&e.p.x.x, &e.p.x.x)
		if err := e.p.x.y.Unmarshal(data[numBytes:]); err != nil {
			return err
		}
		montEncode(&e.p.x.y, &e.p.x.y)
		if e.p.x.IsZero() {
			if yXIsOdd || yYIsOdd {
				// x bytes are zero but y has an odd component; invalid
				return errors.New("Invalid byte slice: improper encoding of identity")
			}
			// We finish encoding the identity element
			e.p.y.SetOne()
			e.p.z.SetZero()
			e.p.t.SetZero()
			return nil
		}

		// We now need to compute correct y value
		t := computeG2YValue(&e.p.x, yXIsOdd, yYIsOdd)
		e.p.y.Set(t)
		e.p.z.SetOne()
		e.p.t.SetOne()
		// Confirm that we have a valid curve point
		if !e.p.IsOnCurve() {
			return ErrMalformedPoint
		}
		return nil
	default:
		return errors.New("Invalid byte slice: improper formatting")
	}
}

// computeG2YValue computes the correct y value;
// that is, if x is a valid coordinate, then compute y such that
// (x, y) is on the curve and y is odd or even as desired.
func computeG2YValue(x *gfP2, yXIsOdd, yYIsOdd bool) *gfP2 {
	t := &gfP2{}
	// Compute t == x^3 + twistB
	t.Square(x)
	t.Mul(t, x)
	t.Add(t, twistB)
	y := &gfP2{}
	// Sqrt computation always succeeds even if sqrt does not exist;
	// validity of the (x, y) pair is not checked here
	y.Sqrt(t)
	tx := &gfP{}
	ty := &gfP{}
	montDecode(tx, &y.x)
	montDecode(ty, &y.y)
	yXCurrentlyOdd := (tx[0] & 1) == 1
	yYCurrentlyOdd := (ty[0] & 1) == 1
	if (yXCurrentlyOdd != yXIsOdd) || (yYCurrentlyOdd != yYIsOdd) {
		y.Neg(y)
	}
	return y
}
