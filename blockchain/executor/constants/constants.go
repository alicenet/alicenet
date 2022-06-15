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
	ErrorLoadingDkgState              = "error loading dkgState: %v"
	ErrorDuringPreparation            = "error during the preparation: %v"
	ErrorGettingAccusableParticipants = "error getting accusableParticipants: %v"
	ErrorGettingValidators            = "error getting validators: %v"
	FailedGettingTxnOpts              = "failed getting txn opts: %v"
	FailedGettingCallOpts             = "failed getting call opts: %v"
	FailedGettingIsValidator          = "failed getting isValidator: %v"
	NobodyToAccuse                    = "nobody to accuse"

	ETHDKGMaxStaleBlocks uint64 = 6
)
