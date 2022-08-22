package dynamics

import (
	"context"
	"testing"

	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/stretchr/testify/assert"
)

func TestCanonicalVersionCheckTask_ShouldExecute_False(t *testing.T) {
	t.Parallel()
	localVersion := utils.GetLocalVersion()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	task := NewVersionCheckTask(localVersion)
	err := task.Initialize(ctx, nil, nil, logger, nil, nil, "", "", nil)
	assert.Nil(t, err)
	taskErr := task.Prepare(ctx)
	assert.Nil(t, taskErr)
	shouldExecute, taskErr := task.ShouldExecute(ctx)
	assert.Nil(t, taskErr)
	assert.False(t, shouldExecute)
}

func TestCanonicalVersionCheckTask_Execute_PatchOutdated(t *testing.T) {
	t.Parallel()
	localVersion := utils.GetLocalVersion()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	outdatedPatchVersion := localVersion
	outdatedPatchVersion.Patch++
	task := NewVersionCheckTask(outdatedPatchVersion)
	err := task.Initialize(ctx, nil, nil, logger, nil, nil, "", "", nil)
	assert.Nil(t, err)
	shouldExecute, taskErr := task.ShouldExecute(ctx)
	assert.Nil(t, taskErr)
	assert.True(t, shouldExecute)
	txn, taskErr := task.Execute(ctx)
	assert.Nil(t, txn)
	assert.NotNil(t, taskErr)
	assert.Equal(t, taskErr.Error(), "printed update message")
}

func TestCanonicalVersionCheckTask_Execute_MinorOutdated(t *testing.T) {
	t.Parallel()
	localVersion := utils.GetLocalVersion()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	outdatedMinorVersion := localVersion
	outdatedMinorVersion.Minor++
	task := NewVersionCheckTask(outdatedMinorVersion)
	err := task.Initialize(ctx, nil, nil, logger, nil, nil, "", "", nil)
	assert.Nil(t, err)
	shouldExecute, taskErr := task.ShouldExecute(ctx)
	assert.Nil(t, taskErr)
	assert.True(t, shouldExecute)
	txn, taskErr := task.Execute(ctx)
	assert.Nil(t, txn)
	assert.NotNil(t, taskErr)
	assert.Equal(t, taskErr.Error(), "printed update message")
}

func TestCanonicalVersionCheckTask_Execute_MajorOutdated(t *testing.T) {
	t.Parallel()
	localVersion := utils.GetLocalVersion()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	outdatedMajorVersion := localVersion
	outdatedMajorVersion.Major++
	task := NewVersionCheckTask(outdatedMajorVersion)
	err := task.Initialize(ctx, nil, nil, logger, nil, nil, "", "", nil)
	assert.Nil(t, err)
	shouldExecute, taskErr := task.ShouldExecute(ctx)
	assert.Nil(t, taskErr)
	assert.True(t, shouldExecute)
	txn, taskErr := task.Execute(ctx)
	assert.Nil(t, txn)
	assert.NotNil(t, taskErr)
	assert.Equal(t, taskErr.Error(), "printed update message")
}
