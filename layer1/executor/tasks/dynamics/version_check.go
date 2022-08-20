package dynamics

import (
	"context"
	"fmt"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"time"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
)

// CanonicalVersionCheckTask contains required state for the task.
type CanonicalVersionCheckTask struct {
	*tasks.BaseTask
	//Version info
	Version bindings.CanonicalVersion
}

const messageFrequency = 1 * time.Second

// asserting that CanonicalVersionCheckTask struct implements interface tasks.Task.
var _ tasks.Task = &CanonicalVersionCheckTask{}

// NewVersionCheckTask creates a background task that attempts to verify the version check.
func NewVersionCheckTask(version bindings.CanonicalVersion) *CanonicalVersionCheckTask {
	return &CanonicalVersionCheckTask{
		BaseTask: tasks.NewBaseTask(0, 0, false, nil),
		Version:  version,
	}
}

// Prepare prepares for work to be done in the CanonicalVersionCheckTask.
func (t *CanonicalVersionCheckTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	logger.Infof("Received a new %d.%d.%d Canonical Node Version to be analized", t.Version.Major, t.Version.Minor, t.Version.Patch)

	return nil
}

// Execute executes the task business logic.
func (t *CanonicalVersionCheckTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	newMajorIsGreater, newMinorIsGreater, newPatchIsGreater, localVersion := utils.CompareCanonicalVersion(t.Version)
	text := ""

	if newMajorIsGreater {
		text = fmt.Sprintf("CRITICAL: your Major Canonical Node Version %d.%d.%d is lower than the latest %d.%d.%d. Please update your node, otherwise it will be killed after epoch %d.",
			localVersion.Major, localVersion.Minor, localVersion.Patch, t.Version.Major, t.Version.Minor, t.Version.Patch, t.Version.ExecutionEpoch)
	} else if newMinorIsGreater {
		text = fmt.Sprintf("WARNING: your Minor Canonical Node Version %d.%d.%d is lower than the latest %d.%d.%d. Please update your node.",
			localVersion.Major, localVersion.Minor, localVersion.Patch, t.Version.Major, t.Version.Minor, t.Version.Patch)
	} else if newPatchIsGreater {
		text = fmt.Sprintf("WARNING: your Patch Canonical Node Version %d.%d.%d is lower than the latest %d.%d.%d. Please update your node.",
			localVersion.Major, localVersion.Minor, localVersion.Patch, t.Version.Major, t.Version.Minor, t.Version.Patch)
	}

	for {
		printingTime := time.After(messageFrequency)
		select {
		case <-ctx.Done():
			return nil, tasks.NewTaskErr(ctx.Err().Error(), false)
		case <-printingTime:
			if shouldPrint, _ := t.ShouldExecute(ctx); shouldPrint {
				logger.Info(text)
				printingTime = time.After(messageFrequency)
				continue
			}
			return nil, nil
		}
	}
}

// ShouldExecute checks if it makes sense to execute the task.
func (t *CanonicalVersionCheckTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	if newMajorIsGreater, newMinorIsGreater, newPatchIsGreater, _ := utils.CompareCanonicalVersion(t.Version); newMajorIsGreater || newMinorIsGreater || newPatchIsGreater {
		return true, nil
	}

	return false, nil
}
