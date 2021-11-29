package monitor_test

/*
type ServicesSuite struct {
	suite.Suite
	eth interfaces.Ethereum
}

func (s *ServicesSuite) SetupTest() {
	t := s.T()

	eth, err := blockchain.NewEthereumSimulator(
		"../../assets/test/keys",
		"../../assets/test/passcodes.txt",
		3,
		2*time.Second,
		5*time.Second,
		0,
		big.NewInt(9223372036854775807),
		"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac",
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F")

	assert.Nil(t, err, "Error creating Ethereum simulator")

	s.eth = eth
}

func (s *ServicesSuite) TestRegistrationOpenEvent() {
	t := s.T()
	eth := s.eth
	c := eth.Contracts()
	assert.NotNil(t, c, "Need a *Contracts")

	height, err := s.eth.GetCurrentHeight(context.TODO())
	assert.Nil(t, err, "could not get height")
	assert.Equal(t, uint64(0), height, "Height should be 0")

	s.eth.Commit()

	height, err = s.eth.GetCurrentHeight(context.TODO())
	assert.Nil(t, err, "could not get height")
	assert.Equal(t, uint64(1), height, "Height should be 1")
}

func TestServicesSuite(t *testing.T) {
	suite.Run(t, new(ServicesSuite))
}

*/
