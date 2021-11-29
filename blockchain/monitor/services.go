package monitor

/*
// Services just a bundle of requirements common for monitoring functionality
type Services struct {
	logger            *logrus.Logger
	eth               interfaces.Ethereum
	dph               *deposit.Handler
	ah                interfaces.AdminHandler
	contractAddresses []common.Address
	batchSize         int
	eventMap          *objects.EventMap
}

// NewServices creates a new Services struct
func NewServices(eth interfaces.Ethereum, cdb *db.Database, db *db.Database, dph *deposit.Handler, ah interfaces.AdminHandler, batchSize int) *Services {

	c := eth.Contracts()

	contractAddresses := []common.Address{
		c.DepositAddress(), c.EthdkgAddress(), c.RegistryAddress(),
		c.StakingTokenAddress(), c.UtilityTokenAddress(), c.ValidatorsAddress(),
		c.GovernorAddress()}

	serviceLogger := logging.GetLogger("services")

	svcs := &Services{
		ah:                ah,
		batchSize:         batchSize,
		contractAddresses: contractAddresses,
		dph:               dph,
		eth:               eth,
		eventMap:          objects.NewEventMap(),
		logger:            serviceLogger,
		// taskManager:       tasks.NewManager(),
	}

	// Register handlers for known events, if this failed we really can't continue
	if err := SetupEventMap(svcs.eventMap, cdb, ah, dph); err != nil {
		panic(err)
	}

	ah.RegisterSnapshotCallback(svcs.PersistSnapshot) // HUNTER: moved out of main func and into constructor

	return svcs
}

// UpdateProgress updates what we know of Ethereum chain height
func (svcs *Services) UpdateProgress(ctx context.Context, state *objects.MonitorState) error {
	height, err := svcs.eth.GetFinalizedHeight(ctx)
	if err != nil {
		return err
	}

	// Only updated single attribute so no need to copy -- Not sure if check is required
	state.HighestBlockFinalized = height
	return nil
}

func (svcs *Services) AccuseDoubleProposal() error {

	return nil
}

// PersistSnapshot records the given block header on Ethereum and increments epoch
// TODO Returning an error kills the main loop, retry forever instead
func (svcs *Services) PersistSnapshot(blockHeader *objs.BlockHeader) error {

	eth := svcs.eth
	c := eth.Contracts()
	logger := svcs.logger

	// pull out the block claims
	bclaims := blockHeader.BClaims
	rawBclaims, err := bclaims.MarshalBinary()
	if err != nil {
		logger.Errorf("Could not extract BClaims from BlockHeader: %v", err)
		return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
	}

	// pull out the sig
	rawSigGroup := blockHeader.SigGroup

	// Do the mechanics
	txnOpts, err := svcs.eth.GetTransactionOpts(context.Background(), svcs.eth.GetDefaultAccount())
	if err != nil {
		logger.Errorf("Could not create transaction for snapshot: %v", err)
		return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
	}

	txn, err := c.Validators().Snapshot(txnOpts, rawSigGroup, rawBclaims)
	if err != nil {
		logger.Errorf("Failed to take snapshot: %v", err)
		return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
	}

	toCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rcpt, err := eth.Queue().QueueAndWait(toCtx, txn)
	if err != nil {
		logger.Errorf("Failed to retrieve snapshot receipt: %v", err)
		return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
	}

	if rcpt == nil {
		logger.Warnf("No receipt from snapshot")
	} else {
		if rcpt.Status != uint64(1) {
			logger.Errorf("Snapshot receipt shows failure.")
			return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
		}
	}

	return nil
}
*/
