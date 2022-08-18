package constants

import (
	"time"
)

// Original from constants/consensus.go
const (
	MaxBytes        = 3000000
	MaxProposalSize = MaxBytes // Parameterize: equal to maxBytes
	MsgTimeout      = 4 * time.Second
	SrvrMsgTimeout  = (3 * MsgTimeout) / 4 // Parameterize: 0.75*MsgTimeout
	ProposalStepTO  = 4 * time.Second
	PreVoteStepTO   = 3 * time.Second
	PreCommitStepTO = 3 * time.Second
	DBRNRTO         = (5 * (ProposalStepTO + PreVoteStepTO + PreCommitStepTO)) / 2 // Parameterize: make 2.5 times Prop, PV, PC timeouts
	DownloadTO      = ProposalStepTO + PreVoteStepTO + PreCommitStepTO             // Parameterize: sum of Prop, PV, PC timeouts
)
