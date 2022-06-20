package tasks

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
