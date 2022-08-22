package dynamics

import (
	"context"
	"fmt"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/core/types"
)

// CanonicalVersionCheckTask contains required state for the task.
type CanonicalVersionCheckTask struct {
	*tasks.BaseTask
	//Version info
	Version bindings.CanonicalVersion
}

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

	logger.Infof(
		"Received a new Canonical Node Version %d.%d.%d to be analyzed",
		t.Version.Major,
		t.Version.Minor,
		t.Version.Patch,
	)

	return nil
}

// Execute executes the task business logic.
func (t *CanonicalVersionCheckTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	newMajorIsGreater, newMinorIsGreater, newPatchIsGreater, localVersion := utils.CompareCanonicalVersion(t.Version)
	logger = logger.WithField("currentVersion", fmt.Sprintf("%d.%d.%d", localVersion.Major, localVersion.Minor, localVersion.Patch))

	text := fmt.Sprintf(
		"There's a new version of the aliceNet node available: %d.%d.%d",
		t.Version.Major,
		t.Version.Minor,
		t.Version.Patch,
	)

	if newMajorIsGreater {
		text = fmt.Sprintf(
			"CRITICAL: %s. Please update your node, otherwise it will be killed after height %d.",
			text,
			t.Version.ExecutionEpoch*constants.EpochLength,
		)
	} else if newMinorIsGreater {
		text = fmt.Sprintf("WARNING: %s. Please update your node.", text)
	} else if newPatchIsGreater {
		text = fmt.Sprintf("WARNING: %s. Please update your node.", text)
	}

	logger.Warn(text)
	return nil, tasks.NewTaskErr("printed update message", true)
}

// ShouldExecute checks if it makes sense to execute the task.
func (t *CanonicalVersionCheckTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")
	newMajorIsGreater, newMinorIsGreater, newPatchIsGreater, _ := utils.CompareCanonicalVersion(t.Version)
	if newMajorIsGreater || newMinorIsGreater || newPatchIsGreater {
		return true, nil
	}
	return false, nil
}
