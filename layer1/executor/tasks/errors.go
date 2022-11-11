package tasks

import "errors"

type TaskErr struct {
	message       string
	isRecoverable bool
}

func (e *TaskErr) Error() string {
	return e.message
}

func (e *TaskErr) IsRecoverable() bool {
	return e.isRecoverable
}

func NewTaskErr(message string, isRecoverable bool) *TaskErr {
	return &TaskErr{message: message, isRecoverable: isRecoverable}
}

const (
	// Common errors.
	ErrorLoadingDkgState              = "error loading dkgState: %v"
	ErrorDuringPreparation            = "error during the preparation: %v"
	ErrorGettingAccusableParticipants = "error getting accusableParticipants: %v"
	ErrorGettingValidators            = "error getting validators: %v"
	FailedGettingTxnOpts              = "failed getting txn opts: %v"
	FailedGettingCallOpts             = "failed getting call opts: %v"
	FailedGettingIsValidator          = "failed getting isValidator: %v"
	NobodyToAccuse                    = "nobody to accuse"
)

var (
	ErrTaskKilled                   = errors.New("the task was killed, aborting execution")
	ErrTaskExecutionMechanismClosed = errors.New("tasks execution mechanism is closed, aborting execution")
)
