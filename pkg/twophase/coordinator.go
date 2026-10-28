package twophase

import (
	"errors"
	"fmt"
)

type PhaseState string

const (
	StateInitiated PhaseState = "INITIATED"
	StatePrepared  PhaseState = "PREPARED"
	StateCommitted PhaseState = "COMMITTED"
	StateAborted   PhaseState = "ABORTED"
)

type ParticipantVote struct {
	NodeID   string
	Prepared bool
	Reason   string
}

type BatchDisbursementTx struct {
	BatchID      string
	TotalMinor   int64
	Currency     string
	State        PhaseState
	Participants []ParticipantVote
}

func EvaluatePreparePhase(tx *BatchDisbursementTx) error {
	if tx.TotalMinor <= 0 {
		tx.State = StateAborted
		return errors.New("2pc: batch total must be strictly positive minor units")
	}
	if len(tx.Participants) == 0 {
		tx.State = StateAborted
		return errors.New("2pc: at least one participant node is required")
	}

	for _, vote := range tx.Participants {
		if !vote.Prepared {
			tx.State = StateAborted
			return fmt.Errorf("2pc: participant %s voted ABORT (%s)", vote.NodeID, vote.Reason)
		}
	}

	tx.State = StatePrepared
	return nil
}

func CommitBatch(tx *BatchDisbursementTx) error {
	if tx.State != StatePrepared {
		return fmt.Errorf("2pc: cannot commit batch in state %s", tx.State)
	}
	tx.State = StateCommitted
	return nil
}
