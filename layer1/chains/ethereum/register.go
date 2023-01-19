package ethereum

import (
	"github.com/alicenet/alicenet/layer1/chains/ethereum/tasks/dkg"
	esnapshots "github.com/alicenet/alicenet/layer1/chains/ethereum/tasks/snapshots"
	psnapshots "github.com/alicenet/alicenet/layer1/chains/polygon/tasks/snapshots"
	"github.com/alicenet/alicenet/layer1/executor/marshaller"
)

// getTaskRegistry all the Tasks we can handle in the request.
// If you want to create a new task register its instance type here.
func GetTaskRegistry(tr *marshaller.TypeRegistry) *marshaller.TypeRegistry {
	tr.RegisterInstanceType(&dkg.CompletionTask{})
	tr.RegisterInstanceType(&dkg.DisputeShareDistributionTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingShareDistributionTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingKeySharesTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingGPKjTask{})
	tr.RegisterInstanceType(&dkg.DisputeGPKjTask{})
	tr.RegisterInstanceType(&dkg.GPKjSubmissionTask{})
	tr.RegisterInstanceType(&dkg.KeyShareSubmissionTask{})
	tr.RegisterInstanceType(&dkg.MPKSubmissionTask{})
	tr.RegisterInstanceType(&dkg.RegisterTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingRegistrationTask{})
	tr.RegisterInstanceType(&dkg.ShareDistributionTask{})
	tr.RegisterInstanceType(&esnapshots.SnapshotTask{})
	tr.RegisterInstanceType(&psnapshots.SnapshotTask{})
	return tr
}
