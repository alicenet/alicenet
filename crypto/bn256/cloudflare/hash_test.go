package cloudflare

import (
	"fmt"
	"math/big"
	"testing"
)

func TestHashToBase(t *testing.T) {
	s := "MadHive Rocks!"
	bigMH0, _ := new(big.Int).SetString("2127590730004973161070119349205886624782173228737350700310995052292290242301", 10)
	gfpMH0 := bigToGFp(bigMH0)

	bigMH1, _ := new(big.Int).SetString("10639008181382806573651867997001852879340224748829551406507512470328095559889", 10)
	gfpMH1 := bigToGFp(bigMH1)

	bigMH2, _ := new(big.Int).SetString("12992459742347536215318183667697742426663369714341132686474845995619917232662", 10)
	gfpMH2 := bigToGFp(bigMH2)

	bigMH3, _ := new(big.Int).SetString("7495307691886468505150229418639005229672623242804090137413750370085164536860", 10)
	gfpMH3 := bigToGFp(bigMH3)

	byteI := byte(0x00)
	byteJ := byte(0x01)
	resMH0 := hashTB(s, byteI, byteJ)
	if !resMH0.IsEqual(gfpMH0) {
		t.Fatal("Failed to properly hash MH0")
	}

	byteI = byte(0x02)
	byteJ = byte(0x03)
	resMH1 := hashTB(s, byteI, byteJ)
	if !resMH1.IsEqual(gfpMH1) {
		t.Fatal("Failed to properly hash MH1")
	}

	byteI = byte(0x04)
	byteJ = byte(0x05)
	resMH2 := hashTB(s, byteI, byteJ)
	if !resMH2.IsEqual(gfpMH2) {
		t.Fatal("Failed to properly hash MH2")
	}

	byteI = byte(0x06)
	byteJ = byte(0x07)
	resMH3 := hashTB(s, byteI, byteJ)
	if !resMH3.IsEqual(gfpMH3) {
		t.Fatal("Failed to properly hash MH3")
	}

	s = "Cryptography is great"
	bigCIG0, _ := new(big.Int).SetString("9534123261356789366382704928175673799395615303165154346204974948104375713661", 10)
	gfpCIG0 := bigToGFp(bigCIG0)

	bigCIG1, _ := new(big.Int).SetString("10733399798341036526816149945929244685213243613006172177440646485476438343808", 10)
	gfpCIG1 := bigToGFp(bigCIG1)

	bigCIG2, _ := new(big.Int).SetString("870093808875654621755088553154867106018879613397690960535618857056282394476", 10)
	gfpCIG2 := bigToGFp(bigCIG2)

	bigCIG3, _ := new(big.Int).SetString("1699099704154693850647627702645604987586205758891249985618654349926637959111", 10)
	gfpCIG3 := bigToGFp(bigCIG3)

	byteI = byte(0x00)
	byteJ = byte(0x01)
	resCIG0 := hashTB(s, byteI, byteJ)
	if !resCIG0.IsEqual(gfpCIG0) {
		t.Fatal("Failed to properly hash CIG0")
	}

	byteI = byte(0x02)
	byteJ = byte(0x03)
	resCIG1 := hashTB(s, byteI, byteJ)
	if !resCIG1.IsEqual(gfpCIG1) {
		t.Fatal("Failed to properly hash CIG1")
	}

	byteI = byte(0x04)
	byteJ = byte(0x05)
	resCIG2 := hashTB(s, byteI, byteJ)
	if !resCIG2.IsEqual(gfpCIG2) {
		t.Fatal("Failed to properly hash CIG2")
	}

	byteI = byte(0x06)
	byteJ = byte(0x07)
	resCIG3 := hashTB(s, byteI, byteJ)
	if !resCIG3.IsEqual(gfpCIG3) {
		t.Fatal("Failed to properly hash CIG3")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("HashToBase changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("HashToBase changed twistGen")
	}
}

func hashTB(s string, byteI byte, byteJ byte) *gfP {
	msg := []byte(s)
	gfpIJ := hashToBase(msg, byteI, byteJ)
	return gfpIJ
}

func TestBaseToG1(t *testing.T) {
	// Checks that we have a valid curve point for t==0 and others
	var k, kStart, kStop int64
	kStart = 0
	kStop = 10
	for k = kStart; k < kStop; k++ {
		gfpK := newGFp(k)
		g, err := baseToG1(gfpK)
		if err != nil {
			t.Fatal(err)
		}
		if !g.p.IsOnCurve() {
			t.Fatal("baseToG1 fails for t ==", k)
		}
	}

	gfpOne := newGFp(1)

	bigMH0, _ := new(big.Int).SetString("2127590730004973161070119349205886624782173228737350700310995052292290242301", 10)
	gfpMH0 := bigToGFp(bigMH0)
	bigXMH0, _ := new(big.Int).SetString("1167674509398230065903161433158119414119359249984755520813941226833228172754", 10)
	gfpXMH0 := bigToGFp(bigXMH0)
	bigYMH0, _ := new(big.Int).SetString("14279810437359713948805394484560530836827185547964297688885307207257460366524", 10)
	gfpYMH0 := bigToGFp(bigYMH0)
	curveB2CMH0 := &curvePoint{
		x: *gfpXMH0,
		y: *gfpYMH0,
		z: *gfpOne,
		t: *gfpOne,
	}
	g1B2CMH0 := &G1{}
	g1B2CMH0.p = &curvePoint{}
	g1B2CMH0.p.Set(curveB2CMH0)

	gPMH0, err := baseToG1(gfpMH0)
	if err != nil {
		t.Fatal(err)
	}
	if !g1B2CMH0.IsEqual(gPMH0) {
		t.Fatal("Error in baseToG1 for t = MH0")
	}

	bigMH1, _ := new(big.Int).SetString("10639008181382806573651867997001852879340224748829551406507512470328095559889", 10) // New
	gfpMH1 := bigToGFp(bigMH1)
	bigXMH1, _ := new(big.Int).SetString("1878346344276375820470323620157262927690114369836505673414113378980512268736", 10)
	gfpXMH1 := bigToGFp(bigXMH1)
	bigYMH1, _ := new(big.Int).SetString("5719584678241420409715569454643490978138012211976012834589451671952502577653", 10)
	gfpYMH1 := bigToGFp(bigYMH1)
	curveB2CMH1 := &curvePoint{
		x: *gfpXMH1,
		y: *gfpYMH1,
		z: *gfpOne,
		t: *gfpOne,
	}
	g1B2CMH1 := &G1{}
	g1B2CMH1.p = &curvePoint{}
	g1B2CMH1.p.Set(curveB2CMH1)

	gPMH1, err := baseToG1(gfpMH1)
	if err != nil {
		t.Fatal(err)
	}
	if !g1B2CMH1.IsEqual(gPMH1) {
		t.Fatal("Error in baseToG1 for t = MH1")
	}

	bigCIG0, _ := new(big.Int).SetString("9534123261356789366382704928175673799395615303165154346204974948104375713661", 10)
	gfpCIG0 := bigToGFp(bigCIG0)
	bigXCIG0, _ := new(big.Int).SetString("1925002128151166821532306527125799367465190948487548146582606547078486737495", 10)
	gfpXCIG0 := bigToGFp(bigXCIG0)
	bigYCIG0, _ := new(big.Int).SetString("4162649232132553873360343493589255511248712262515293172723466175905841701718", 10)
	gfpYCIG0 := bigToGFp(bigYCIG0)
	curveB2CCIG0 := &curvePoint{
		x: *gfpXCIG0,
		y: *gfpYCIG0,
		z: *gfpOne,
		t: *gfpOne,
	}
	g1B2CCIG0 := &G1{}
	g1B2CCIG0.p = &curvePoint{}
	g1B2CCIG0.p.Set(curveB2CCIG0)

	gPCIG0, err := baseToG1(gfpCIG0)
	if err != nil {
		t.Fatal(err)
	}
	if !g1B2CCIG0.IsEqual(gPCIG0) {
		t.Fatal("Error in baseToG1 for t = CIG0")
	}

	bigCIG1, _ := new(big.Int).SetString("10733399798341036526816149945929244685213243613006172177440646485476438343808", 10) // New
	gfpCIG1 := bigToGFp(bigCIG1)
	bigXCIG1, _ := new(big.Int).SetString("14945898838857089202610950606197299340042887644017349412467872616710671354169", 10)
	gfpXCIG1 := bigToGFp(bigXCIG1)
	bigYCIG1, _ := new(big.Int).SetString("15495526251907797603899699541291248642340770654202157240162347215649003669974", 10)
	gfpYCIG1 := bigToGFp(bigYCIG1)
	curveB2CCIG1 := &curvePoint{
		x: *gfpXCIG1,
		y: *gfpYCIG1,
		z: *gfpOne,
		t: *gfpOne,
	}
	g1B2CCIG1 := &G1{}
	g1B2CCIG1.p = &curvePoint{}
	g1B2CCIG1.p.Set(curveB2CCIG1)

	gPCIG1, err := baseToG1(gfpCIG1)
	if err != nil {
		t.Fatal(err)
	}
	if !g1B2CCIG1.IsEqual(gPCIG1) {
		t.Fatal("Error in baseToG1 for t = CIG1")
	}

	// chi(x1,x2,x3) = (1, -1, 1); choose x1
	gfp1 := newGFp(1)
	bigX1, _ := new(big.Int).SetString("4377648574367855045771657440140328170590424477155021945122375134257168632896", 10)
	gfpX1 := bigToGFp(bigX1)
	bigY1, _ := new(big.Int).SetString("5462694088646531740394539448374228879154207235258544883567858895414770370417", 10)
	gfpY1 := bigToGFp(bigY1)
	curveB2C1 := &curvePoint{
		x: *gfpX1,
		y: *gfpY1,
		z: *gfpOne,
		t: *gfpOne,
	}
	g1B2C1 := &G1{}
	g1B2C1.p = &curvePoint{}
	g1B2C1.p.Set(curveB2C1)

	gP1, err := baseToG1(gfp1)
	if err != nil {
		t.Fatal(err)
	}
	if !g1B2C1.IsEqual(gP1) {
		t.Fatal("Error in baseToG1 for t = 1")
	}

	// chi(x1,x2,x3) = (1, 1, 1); choose x1
	gfp2 := newGFp(2)
	bigX2, _ := new(big.Int).SetString("10944121435919637611123202872628637544348155578648911831344518947322613104291", 10)
	gfpX2 := bigToGFp(bigX2)
	bigY2, _ := new(big.Int).SetString("4718603453640367770405249522358112449463417117041194427604452040985121683380", 10)
	gfpY2 := bigToGFp(bigY2)
	curveB2C2 := &curvePoint{
		x: *gfpX2,
		y: *gfpY2,
		z: *gfpOne,
		t: *gfpOne,
	}
	g1B2C2 := &G1{}
	g1B2C2.p = &curvePoint{}
	g1B2C2.p.Set(curveB2C2)

	gP2, err := baseToG1(gfp2)
	if err != nil {
		t.Fatal(err)
	}
	if !g1B2C2.IsEqual(gP2) {
		t.Fatal("Error in baseToG1 for t = 2")
	}

	// chi(x1,x2,x3) = (-1, 1, -1); choose x2
	gfp5 := newGFp(5)
	bigX5, _ := new(big.Int).SetString("9057203946967975955628966866593029703936083189203961599749252385249208041182", 10)
	gfpX5 := bigToGFp(bigX5)
	bigY5, _ := new(big.Int).SetString("20402679892741878328321739372072872655278178337258501639575399541518545633396", 10)
	gfpY5 := bigToGFp(bigY5)
	curveB2C5 := &curvePoint{
		x: *gfpX5,
		y: *gfpY5,
		z: *gfpOne,
		t: *gfpOne,
	}
	g1B2C5 := &G1{}
	g1B2C5.p = &curvePoint{}
	g1B2C5.p.Set(curveB2C5)

	gP5, err := baseToG1(gfp5)
	if err != nil {
		t.Fatal(err)
	}
	if !g1B2C5.IsEqual(gP5) {
		t.Fatal("Error in baseToG1 for t = 5")
	}

	// chi(x1,x2,x3) = (-1, -1, 1); choose x3
	gfp13 := newGFp(13)
	bigX13, _ := new(big.Int).SetString("12476730157715089820964913728558880671860421941733867926069293790044320264795", 10)
	gfpX13 := bigToGFp(bigX13)
	bigY13, _ := new(big.Int).SetString("11440613308225403684438094471295479578982765485128228175924347487470986116356", 10)
	gfpY13 := bigToGFp(bigY13)
	curveB2C13 := &curvePoint{
		x: *gfpX13,
		y: *gfpY13,
		z: *gfpOne,
		t: *gfpOne,
	}
	g1B2C13 := &G1{}
	g1B2C13.p = &curvePoint{}
	g1B2C13.p.Set(curveB2C13)

	gP13, err := baseToG1(gfp13)
	if err != nil {
		t.Fatal(err)
	}
	if !g1B2C13.IsEqual(gP13) {
		t.Fatal("Error in baseToG1 for t = 13")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("BaseToG1 changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("BaseToG1 changed twistGen")
	}
}

func BenchmarkHashToG1Bytes32(b *testing.B) {
	msg := make([]byte, 32)
	msg[0] = 0x01
	msg[31] = 0x01
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err := HashToG1(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashToG1Bytes1024(b *testing.B) {
	msg := make([]byte, 1024)
	msg[0] = 0x01
	msg[1023] = 0x01
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err := HashToG1(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestHashToG1(t *testing.T) {
	gfpOne := newGFp(1)

	// Test to ensure curvePoint is correct for string "MadHive Rocks!"
	s := "MadHive Rocks!"
	msg := []byte(s)
	h2cMH, err := HashToG1(msg)
	if err != nil {
		t.Fatal(err)
	}

	bigXMH0, _ := new(big.Int).SetString("1167674509398230065903161433158119414119359249984755520813941226833228172754", 10)
	gfpXMH0 := bigToGFp(bigXMH0)
	bigYMH0, _ := new(big.Int).SetString("14279810437359713948805394484560530836827185547964297688885307207257460366524", 10)
	gfpYMH0 := bigToGFp(bigYMH0)
	curveB2CMH0 := &curvePoint{
		x: *gfpXMH0,
		y: *gfpYMH0,
		z: *gfpOne,
		t: *gfpOne,
	}

	bigXMH1, _ := new(big.Int).SetString("1878346344276375820470323620157262927690114369836505673414113378980512268736", 10)
	gfpXMH1 := bigToGFp(bigXMH1)
	bigYMH1, _ := new(big.Int).SetString("5719584678241420409715569454643490978138012211976012834589451671952502577653", 10)
	gfpYMH1 := bigToGFp(bigYMH1)
	curveB2CMH1 := &curvePoint{
		x: *gfpXMH1,
		y: *gfpYMH1,
		z: *gfpOne,
		t: *gfpOne,
	}
	curveH2CMH := &curvePoint{}
	curveH2CMH.Add(curveB2CMH0, curveB2CMH1)

	if !curveH2CMH.IsEqual(h2cMH.p) {
		t.Fatal("Error in HashToG1 for MH")
	}

	// Test to ensure curvePoint is correct for string "Cryptography is great"
	s = "Cryptography is great"
	msg = []byte(s)
	h2cCIG, err := HashToG1(msg)
	if err != nil {
		t.Fatal(err)
	}

	bigXCIG0, _ := new(big.Int).SetString("1925002128151166821532306527125799367465190948487548146582606547078486737495", 10)
	gfpXCIG0 := bigToGFp(bigXCIG0)
	bigYCIG0, _ := new(big.Int).SetString("4162649232132553873360343493589255511248712262515293172723466175905841701718", 10)
	gfpYCIG0 := bigToGFp(bigYCIG0)
	curveB2CCIG0 := &curvePoint{
		x: *gfpXCIG0,
		y: *gfpYCIG0,
		z: *gfpOne,
		t: *gfpOne,
	}

	bigXCIG1, _ := new(big.Int).SetString("14945898838857089202610950606197299340042887644017349412467872616710671354169", 10)
	gfpXCIG1 := bigToGFp(bigXCIG1)
	bigYCIG1, _ := new(big.Int).SetString("15495526251907797603899699541291248642340770654202157240162347215649003669974", 10)
	gfpYCIG1 := bigToGFp(bigYCIG1)
	curveB2CCIG1 := &curvePoint{
		x: *gfpXCIG1,
		y: *gfpYCIG1,
		z: *gfpOne,
		t: *gfpOne,
	}
	curveH2CCIG := &curvePoint{}
	curveH2CCIG.Add(curveB2CCIG0, curveB2CCIG1)

	if !curveH2CCIG.IsEqual(h2cCIG.p) {
		t.Fatal("Error in HashToG1 for CIG")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("HashToG1 changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("HashToG1 changed twistGen")
	}
}

func TestBaseToG1Constants(t *testing.T) {
	one := newGFp(1)
	negOne := newGFp(-1)
	negThree := newGFp(-3)
	sqrtNegThree := newGFp(0)
	sqrtNegThree.Sqrt(negThree)
	twoInv := newGFp(2)
	twoInv.Invert(twoInv)
	threeInv := newGFp(3)
	threeInv.Invert(threeInv)

	p1 := &gfP{}
	gfpAdd(p1, negOne, sqrtNegThree)
	gfpMul(p1, p1, twoInv)
	if *p1 != *g1HashConst1 {
		t.Fatal("g1HashConst1 is incorrect")
	}

	p2 := &gfP{}
	p2.Set(sqrtNegThree)
	if !p2.IsEqual(g1HashConst2) {
		t.Fatal("g1HashConst2 is incorrect")
	}

	p3 := &gfP{}
	p3.Set(threeInv)
	if !p3.IsEqual(g1HashConst3) {
		t.Fatal("g1HashConst3 is incorrect")
	}

	p4 := &gfP{}
	gfpAdd(p4, one, curveB)
	if !p4.IsEqual(g1HashConst4) {
		t.Fatal("g1HashConst4 is incorrect")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("G1 constants changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("G1 constants changed twistGen")
	}
}

func TestBaseToG2Constants(t *testing.T) {
	gfpOne := newGFp(1)
	gfpTwoInv := newGFp(2)
	gfpTwoInv.Invert(gfpTwoInv)
	gfpThreeInv := newGFp(3)
	gfpThreeInv.Invert(gfpThreeInv)
	gfpSqrtNeg3 := newGFp(-3)
	gfpSqrtNeg3.Sqrt(gfpSqrtNeg3)

	h1 := &gfP{}
	gfpSub(h1, gfpSqrtNeg3, gfpOne)
	gfpMul(h1, h1, gfpTwoInv)
	p1 := &gfP2{}
	p1.x.Set(newGFp(0))
	p1.y.Set(h1)
	if !p1.IsEqual(g2HashConst1) {
		t.Fatal("g2HashConst1 is incorrect")
	}

	h2 := &gfP{}
	h2.Set(gfpSqrtNeg3)
	p2 := &gfP2{}
	p2.x.Set(newGFp(0))
	p2.y.Set(h2)
	if !p2.IsEqual(g2HashConst2) {
		t.Fatal("g2HashConst2 is incorrect")
	}

	h3 := &gfP{}
	h3.Set(gfpThreeInv)
	p3 := &gfP2{}
	p3.x.Set(newGFp(0))
	p3.y.Set(h3)
	if !p3.IsEqual(g2HashConst3) {
		t.Fatal("g2HashConst3 is incorrect")
	}

	p4 := &gfP2{}
	p4.SetOne()
	p4.Add(p4, twistB)
	if !p4.IsEqual(g2HashConst4) {
		t.Fatal("g2HashConst4 is incorrect")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("G2 constants changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("G2 constants changed twistGen")
	}
}

func TestBaseToTwist(t *testing.T) {
	var k int64
	for k = 0; k < 10; k++ {
		gfp2K := &gfP2{}
		gfp2K.x.Set(newGFp(k))
		g, err := baseToTwist(gfp2K)
		if err != nil {
			t.Fatal(err)
		}
		if !g.IsOnTwist() {
			fmt.Println(err)
			t.Fatal("Error occurred in baseToTwist for t = i *", k)
		}
	}

	for k = 0; k < 10; k++ {
		gfp2K := &gfP2{}
		gfp2K.y.Set(newGFp(k))
		g, err := baseToTwist(gfp2K)
		if err != nil {
			t.Fatal(err)
		}
		if !g.IsOnTwist() {
			fmt.Println(err)
			t.Fatal("Error occurred in baseToTwist for t =", k)
		}
	}

	gfp2One := &gfP2{}
	gfp2One.SetOne()

	bigMH2tI, _ := new(big.Int).SetString("12992459742347536215318183667697742426663369714341132686474845995619917232662", 10)
	gfpMH2tI := bigToGFp(bigMH2tI)
	bigMH2t, _ := new(big.Int).SetString("7495307691886468505150229418639005229672623242804090137413750370085164536860", 10)
	gfpMH2t := bigToGFp(bigMH2t)
	gfpMH2 := &gfP2{}
	gfpMH2.x.Set(gfpMH2tI)
	gfpMH2.y.Set(gfpMH2t)

	bigXIMH2, _ := new(big.Int).SetString("21145401971650475132083803431827682122311003338382127200622140855587533426772", 10)
	gfpXIMH2 := bigToGFp(bigXIMH2)
	bigXMH2, _ := new(big.Int).SetString("11135012987201212255812125407313592818805556398652009252710243684056434394879", 10)
	gfpXMH2 := bigToGFp(bigXMH2)
	bigYIMH2, _ := new(big.Int).SetString("9664785971369663690507442992871023551819040655308005206361757427945564081916", 10)
	gfpYIMH2 := bigToGFp(bigYIMH2)
	bigYMH2, _ := new(big.Int).SetString("20428037599366003986825092750315908120496904544443323723938069076600248316906", 10)
	gfpYMH2 := bigToGFp(bigYMH2)
	gfp2XMH2 := &gfP2{}
	gfp2XMH2.x.Set(gfpXIMH2)
	gfp2XMH2.y.Set(gfpXMH2)
	gfp2YMH2 := &gfP2{}
	gfp2YMH2.x.Set(gfpYIMH2)
	gfp2YMH2.y.Set(gfpYMH2)
	twistB2TMH2 := &twistPoint{
		x: *gfp2XMH2,
		y: *gfp2YMH2,
		z: *gfp2One,
		t: *gfp2One,
	}
	gTMH2, err := baseToTwist(gfpMH2)
	if err != nil {
		t.Fatal(err)
	}
	if !twistB2TMH2.IsEqual(gTMH2) {
		t.Fatal("Error in baseToTwist for t = MH2")
	}

	bigMH3tI, _ := new(big.Int).SetString("4855990189322309715028857930699553576595425766283098683930673270898226616477", 10)
	gfpMH3tI := bigToGFp(bigMH3tI)
	bigMH3t, _ := new(big.Int).SetString("5860202962247230303714256300661018643199957095585130563790611644630251126971", 10)
	gfpMH3t := bigToGFp(bigMH3t)
	gfpMH3 := &gfP2{}
	gfpMH3.x.Set(gfpMH3tI)
	gfpMH3.y.Set(gfpMH3t)

	bigXIMH3, _ := new(big.Int).SetString("15073965372935732833548295181126150547167758293531121735095871889509837890754", 10)
	gfpXIMH3 := bigToGFp(bigXIMH3)
	bigXMH3, _ := new(big.Int).SetString("3384978724250253552514293579948138020803671250818521470916826884217893756913", 10)
	gfpXMH3 := bigToGFp(bigXMH3)
	bigYIMH3, _ := new(big.Int).SetString("18421003069321019582712569168793383638438817863219955223643570740010021465384", 10)
	gfpYIMH3 := bigToGFp(bigYIMH3)
	bigYMH3, _ := new(big.Int).SetString("18850425580715494940317398040083556001560765787179425831312772322931042070317", 10)
	gfpYMH3 := bigToGFp(bigYMH3)
	gfp2XMH3 := &gfP2{}
	gfp2XMH3.x.Set(gfpXIMH3)
	gfp2XMH3.y.Set(gfpXMH3)
	gfp2YMH3 := &gfP2{}
	gfp2YMH3.x.Set(gfpYIMH3)
	gfp2YMH3.y.Set(gfpYMH3)
	twistB2TMH3 := &twistPoint{
		x: *gfp2XMH3,
		y: *gfp2YMH3,
		z: *gfp2One,
		t: *gfp2One,
	}
	gTMH3, err := baseToTwist(gfpMH3)
	if err != nil {
		t.Fatal(err)
	}
	if !twistB2TMH3.IsEqual(gTMH3) {
		t.Fatal("Error in baseToTwist for t = MH3")
	}

	bigCIG2tI, _ := new(big.Int).SetString("870093808875654621755088553154867106018879613397690960535618857056282394476", 10)
	gfpCIG2tI := bigToGFp(bigCIG2tI)
	bigCIG2t, _ := new(big.Int).SetString("1699099704154693850647627702645604987586205758891249985618654349926637959111", 10)
	gfpCIG2t := bigToGFp(bigCIG2t)
	gfpCIG2 := &gfP2{}
	gfpCIG2.x.Set(gfpCIG2tI)
	gfpCIG2.y.Set(gfpCIG2t)

	bigXICIG2, _ := new(big.Int).SetString("2615537802000999729721468167898984107939918603158650035781837447257277204181", 10)
	gfpXICIG2 := bigToGFp(bigXICIG2)
	bigXCIG2, _ := new(big.Int).SetString("10459191249732775298381122855876406847603255431402613421416979112458292826106", 10)
	gfpXCIG2 := bigToGFp(bigXCIG2)
	bigYICIG2, _ := new(big.Int).SetString("20778664226857027060761629198085944494639571073867619083966273910657897032404", 10)
	gfpYICIG2 := bigToGFp(bigYICIG2)
	bigYCIG2, _ := new(big.Int).SetString("10207356919238691379594517063819431403173916213473794821600441295688995635735", 10)
	gfpYCIG2 := bigToGFp(bigYCIG2)
	gfp2XCIG2 := &gfP2{}
	gfp2XCIG2.x.Set(gfpXICIG2)
	gfp2XCIG2.y.Set(gfpXCIG2)
	gfp2YCIG2 := &gfP2{}
	gfp2YCIG2.x.Set(gfpYICIG2)
	gfp2YCIG2.y.Set(gfpYCIG2)
	twistB2TCIG2 := &twistPoint{
		x: *gfp2XCIG2,
		y: *gfp2YCIG2,
		z: *gfp2One,
		t: *gfp2One,
	}
	gTCIG2, err := baseToTwist(gfpCIG2)
	if err != nil {
		t.Fatal(err)
	}
	if !twistB2TCIG2.IsEqual(gTCIG2) {
		t.Fatal("Error in baseToTwist for t = CIG2")
	}

	bigCIG3tI, _ := new(big.Int).SetString("9689411673800783835554945506105351324164114714351204485227628806605127271360", 10)
	gfpCIG3tI := bigToGFp(bigCIG3tI)
	bigCIG3t, _ := new(big.Int).SetString("16037481847349735032259994495916192138350361553795986688459504025251782547235", 10)
	gfpCIG3t := bigToGFp(bigCIG3t)
	gfpCIG3 := &gfP2{}
	gfpCIG3.x.Set(gfpCIG3tI)
	gfpCIG3.y.Set(gfpCIG3t)

	bigXICIG3, _ := new(big.Int).SetString("8312853685969527526682211213746548140780920042737393617613910743683877197380", 10)
	gfpXICIG3 := bigToGFp(bigXICIG3)
	bigXCIG3, _ := new(big.Int).SetString("15820970700891088914411651841491074518877183547245822644523607265869969035269", 10)
	gfpXCIG3 := bigToGFp(bigXCIG3)
	bigYICIG3, _ := new(big.Int).SetString("20474213000763683791542108751521435431467069635629846295992198826513389829232", 10)
	gfpYICIG3 := bigToGFp(bigYICIG3)
	bigYCIG3, _ := new(big.Int).SetString("3790623179278728816446686220906254968318420086858566054466068962711600739936", 10)
	gfpYCIG3 := bigToGFp(bigYCIG3)
	gfp2XCIG3 := &gfP2{}
	gfp2XCIG3.x.Set(gfpXICIG3)
	gfp2XCIG3.y.Set(gfpXCIG3)
	gfp2YCIG3 := &gfP2{}
	gfp2YCIG3.x.Set(gfpYICIG3)
	gfp2YCIG3.y.Set(gfpYCIG3)
	twistB2TCIG3 := &twistPoint{
		x: *gfp2XCIG3,
		y: *gfp2YCIG3,
		z: *gfp2One,
		t: *gfp2One,
	}
	gTCIG3, err := baseToTwist(gfpCIG3)
	if err != nil {
		t.Fatal(err)
	}
	if !twistB2TCIG3.IsEqual(gTCIG3) {
		t.Fatal("Error in baseToTwist for t = CIG3")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("BaseToTwist changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("BaseToTwist changed twistGen")
	}
}

func TestHashToG2(t *testing.T) {
	gfp2One := &gfP2{}
	gfp2One.SetOne()

	s := "MadHive Rocks!"
	msg := []byte(s)
	h2cMH, err := HashToG2(msg)
	if err != nil {
		t.Fatal(err)
	}

	bigXIMH2, _ := new(big.Int).SetString("21145401971650475132083803431827682122311003338382127200622140855587533426772", 10)
	gfpXIMH2 := bigToGFp(bigXIMH2)
	bigXMH2, _ := new(big.Int).SetString("11135012987201212255812125407313592818805556398652009252710243684056434394879", 10)
	gfpXMH2 := bigToGFp(bigXMH2)
	bigYIMH2, _ := new(big.Int).SetString("9664785971369663690507442992871023551819040655308005206361757427945564081916", 10)
	gfpYIMH2 := bigToGFp(bigYIMH2)
	bigYMH2, _ := new(big.Int).SetString("20428037599366003986825092750315908120496904544443323723938069076600248316906", 10)
	gfpYMH2 := bigToGFp(bigYMH2)
	gfp2XMH2 := &gfP2{}
	gfp2XMH2.x.Set(gfpXIMH2)
	gfp2XMH2.y.Set(gfpXMH2)
	gfp2YMH2 := &gfP2{}
	gfp2YMH2.x.Set(gfpYIMH2)
	gfp2YMH2.y.Set(gfpYMH2)
	twistB2TMH2 := &twistPoint{
		x: *gfp2XMH2,
		y: *gfp2YMH2,
		z: *gfp2One,
		t: *gfp2One,
	}

	bigXIMH3, _ := new(big.Int).SetString("15073965372935732833548295181126150547167758293531121735095871889509837890754", 10)
	gfpXIMH3 := bigToGFp(bigXIMH3)
	bigXMH3, _ := new(big.Int).SetString("3384978724250253552514293579948138020803671250818521470916826884217893756913", 10)
	gfpXMH3 := bigToGFp(bigXMH3)
	bigYIMH3, _ := new(big.Int).SetString("18421003069321019582712569168793383638438817863219955223643570740010021465384", 10)
	gfpYIMH3 := bigToGFp(bigYIMH3)
	bigYMH3, _ := new(big.Int).SetString("18850425580715494940317398040083556001560765787179425831312772322931042070317", 10)
	gfpYMH3 := bigToGFp(bigYMH3)
	gfp2XMH3 := &gfP2{}
	gfp2XMH3.x.Set(gfpXIMH3)
	gfp2XMH3.y.Set(gfpXMH3)
	gfp2YMH3 := &gfP2{}
	gfp2YMH3.x.Set(gfpYIMH3)
	gfp2YMH3.y.Set(gfpYMH3)
	twistB2TMH3 := &twistPoint{
		x: *gfp2XMH3,
		y: *gfp2YMH3,
		z: *gfp2One,
		t: *gfp2One,
	}

	twistH2CMH := &twistPoint{}
	twistH2CMH.Add(twistB2TMH2, twistB2TMH3)
	twistH2CMH.ClearCofactor(twistH2CMH)
	if !twistH2CMH.IsEqual(h2cMH.p) {
		t.Fatal("Error in HashToG2 for MH")
	}

	s = "Cryptography is great"
	msg = []byte(s)
	h2cCIG, err := HashToG2(msg)
	if err != nil {
		t.Fatal(err)
	}

	bigXICIG2, _ := new(big.Int).SetString("2615537802000999729721468167898984107939918603158650035781837447257277204181", 10)
	gfpXICIG2 := bigToGFp(bigXICIG2)
	bigXCIG2, _ := new(big.Int).SetString("10459191249732775298381122855876406847603255431402613421416979112458292826106", 10)
	gfpXCIG2 := bigToGFp(bigXCIG2)
	bigYICIG2, _ := new(big.Int).SetString("20778664226857027060761629198085944494639571073867619083966273910657897032404", 10)
	gfpYICIG2 := bigToGFp(bigYICIG2)
	bigYCIG2, _ := new(big.Int).SetString("10207356919238691379594517063819431403173916213473794821600441295688995635735", 10)
	gfpYCIG2 := bigToGFp(bigYCIG2)
	gfp2XCIG2 := &gfP2{}
	gfp2XCIG2.x.Set(gfpXICIG2)
	gfp2XCIG2.y.Set(gfpXCIG2)
	gfp2YCIG2 := &gfP2{}
	gfp2YCIG2.x.Set(gfpYICIG2)
	gfp2YCIG2.y.Set(gfpYCIG2)
	twistB2TCIG2 := &twistPoint{
		x: *gfp2XCIG2,
		y: *gfp2YCIG2,
		z: *gfp2One,
		t: *gfp2One,
	}

	bigXICIG3, _ := new(big.Int).SetString("8312853685969527526682211213746548140780920042737393617613910743683877197380", 10)
	gfpXICIG3 := bigToGFp(bigXICIG3)
	bigXCIG3, _ := new(big.Int).SetString("15820970700891088914411651841491074518877183547245822644523607265869969035269", 10)
	gfpXCIG3 := bigToGFp(bigXCIG3)
	bigYICIG3, _ := new(big.Int).SetString("20474213000763683791542108751521435431467069635629846295992198826513389829232", 10)
	gfpYICIG3 := bigToGFp(bigYICIG3)
	bigYCIG3, _ := new(big.Int).SetString("3790623179278728816446686220906254968318420086858566054466068962711600739936", 10)
	gfpYCIG3 := bigToGFp(bigYCIG3)
	gfp2XCIG3 := &gfP2{}
	gfp2XCIG3.x.Set(gfpXICIG3)
	gfp2XCIG3.y.Set(gfpXCIG3)
	gfp2YCIG3 := &gfP2{}
	gfp2YCIG3.x.Set(gfpYICIG3)
	gfp2YCIG3.y.Set(gfpYCIG3)
	twistB2TCIG3 := &twistPoint{
		x: *gfp2XCIG3,
		y: *gfp2YCIG3,
		z: *gfp2One,
		t: *gfp2One,
	}

	twistH2CCIG := &twistPoint{}
	twistH2CCIG.Add(twistB2TCIG2, twistB2TCIG3)
	twistH2CCIG.ClearCofactor(twistH2CCIG)
	if !twistH2CCIG.IsEqual(h2cCIG.p) {
		t.Fatal("Error in HashToG2 for CIG")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("HashToG2 changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("HashToG2 changed twistGen")
	}
}

func TestSafeSigningPoint(t *testing.T) {
	g0 := new(G1).ScalarBaseMult(Order)
	if safeSigningPoint(g0) {
		t.Fatal("safeSigningPoint failed to label Infinity unsafe")
	}

	g1 := new(G1).ScalarBaseMult(big.NewInt(1))
	if safeSigningPoint(g1) {
		t.Fatal("safeSigningPoint failed to label curveGen unsafe")
	}

	g2 := new(G1).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495616"))
	if safeSigningPoint(g2) {
		t.Fatal("safeSigningPoint failed to label -curveGen unsafe")
	}

	g3 := new(G1).ScalarBaseMult(big.NewInt(1234567890))
	if !safeSigningPoint(g3) {
		t.Fatal("safeSigningPoint failed to label 1234567890*curveGen safe")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("safeSigningPoint changed curveGen")
	}
}
