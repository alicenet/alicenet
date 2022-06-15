package constants

const (
	//Task names
	RegisterTaskName                        = "RegisterTask"
	DisputeMissingRegistrationTaskName      = "DisputeMissingRegistrationTask"
	ShareDistributionTaskName               = "ShareDistributionTask"
	DisputeMissingShareDistributionTaskName = "DisputeMissingShareDistributionTask"
	DisputeShareDistributionTaskName        = "DisputeShareDistributionTask"
	KeyShareSubmissionTaskName              = "KeyShareSubmissionTask"
	DisputeMissingKeySharesTaskName         = "DisputeMissingKeySharesTask"
	MPKSubmissionTaskName                   = "MPKSubmissionTask"
	GPKjSubmissionTaskName                  = "GPKjSubmissionTask"
	DisputeMissingGPKjTaskName              = "DisputeMissingGPKjTask"
	DisputeGPKjTaskName                     = "DisputeGPKjTask"
	CompletionTaskName                      = "CompletionTask"
	SnapshotTaskName                        = "SnapshotTask"

	//Common errors
	ErrorLoadingDkgState  = "error loading dkgState: %v"
	FailedGettingTxnOpts  = "failed getting txn opts: %v"
	FailedGettingCallOpts = "failed getting call opts: %v"

	ETHDKGMaxStaleBlocks uint64 = 6
)
