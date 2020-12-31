package gossip

import (
	"context"

	"github.com/MadBase/MadNet/consensus/objs"
)

type transactionMsg struct {
	ctx    context.Context
	cancel func()
	msg    []byte
	errC   chan error
}

type proposalMsg struct {
	ctx    context.Context
	cancel func()
	msg    *objs.Proposal
	errC   chan error
}

type preVoteMsg struct {
	ctx    context.Context
	cancel func()
	msg    *objs.PreVote
	errC   chan error
}

type preVoteNilMsg struct {
	ctx    context.Context
	cancel func()
	msg    *objs.PreVoteNil
	errC   chan error
}

type preCommitMsg struct {
	ctx    context.Context
	cancel func()
	msg    *objs.PreCommit
	errC   chan error
}

type preCommitNilMsg struct {
	ctx    context.Context
	cancel func()
	msg    *objs.PreCommitNil
	errC   chan error
}

type nextHeightMsg struct {
	ctx    context.Context
	cancel func()
	msg    *objs.NextHeight
	errC   chan error
}

type nextRoundMsg struct {
	ctx    context.Context
	cancel func()
	msg    *objs.NextRound
	errC   chan error
}

type blockHeaderMsg struct {
	ctx    context.Context
	cancel func()
	msg    *objs.BlockHeader
	errC   chan error
}
