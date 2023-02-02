package dynamics

import (
	"context"
	"sync"
	"testing"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var globalVersion sync.Once

func setVersion() {
	config.Configuration.Version = "v0.0.0"
}

func TestCanonicalVersionCheckTask_ShouldExecute_False(t *testing.T) {
	t.Parallel()
	globalVersion.Do(setVersion)
	localVersion, err := utils.GetLocalVersion()
	require.Nil(t, err)

	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	task := NewVersionCheckTask(localVersion)
	err = task.Initialize(nil, logger, nil, nil, "", "", 1, 10, false, nil, nil)
	assert.Nil(t, err)
	taskErr := task.Prepare(ctx)
	assert.Nil(t, taskErr)
	shouldExecute, taskErr := task.ShouldExecute(ctx)
	assert.Nil(t, taskErr)
	assert.False(t, shouldExecute)
}

func TestCanonicalVersionCheckTask_Execute_PatchOutdated(t *testing.T) {
	t.Parallel()
	globalVersion.Do(setVersion)
	localVersion, err := utils.GetLocalVersion()
	require.Nil(t, err)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	outdatedPatchVersion := localVersion
	outdatedPatchVersion.Patch++
	task := NewVersionCheckTask(outdatedPatchVersion)
	err = task.Initialize(nil, logger, nil, nil, "", "", 1, 10, false, nil, nil)
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
	globalVersion.Do(setVersion)
	localVersion, err := utils.GetLocalVersion()
	require.Nil(t, err)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	outdatedMinorVersion := localVersion
	outdatedMinorVersion.Minor++
	task := NewVersionCheckTask(outdatedMinorVersion)
	err = task.Initialize(nil, logger, nil, nil, "", "", 1, 10, false, nil, nil)
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
	globalVersion.Do(setVersion)
	localVersion, err := utils.GetLocalVersion()
	require.Nil(t, err)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	outdatedMajorVersion := localVersion
	outdatedMajorVersion.Major++
	task := NewVersionCheckTask(outdatedMajorVersion)
	err = task.Initialize(nil, logger, nil, nil, "", "", 1, 10, false, nil, nil)
	assert.Nil(t, err)
	shouldExecute, taskErr := task.ShouldExecute(ctx)
	assert.Nil(t, taskErr)
	assert.True(t, shouldExecute)
	txn, taskErr := task.Execute(ctx)
	assert.Nil(t, txn)
	assert.NotNil(t, taskErr)
	assert.Equal(t, taskErr.Error(), "printed update message")
}
