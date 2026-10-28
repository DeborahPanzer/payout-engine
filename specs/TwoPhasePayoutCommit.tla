------------------------ MODULE TwoPhasePayoutCommit ------------------------
EXTENDS Naturals

VARIABLES
  coordinatorState,  \* "INIT", "PREPARED", "COMMITTED", "ABORTED"
  wireDispatched,    \* Number of bank wires dispatched for the batch
  escrowDebited      \* Boolean indicating whether internal escrow was debited

vars == <<coordinatorState, wireDispatched, escrowDebited>>

Init ==
  /\ coordinatorState = "INIT"
  /\ wireDispatched = 0
  /\ escrowDebited = FALSE

Prepare ==
  /\ coordinatorState = "INIT"
  /\ coordinatorState' = "PREPARED"
  /\ UNCHANGED <<wireDispatched, escrowDebited>>

Commit ==
  /\ coordinatorState = "PREPARED"
  /\ wireDispatched = 0
  /\ coordinatorState' = "COMMITTED"
  /\ wireDispatched' = 1
  /\ escrowDebited' = TRUE

Abort ==
  /\ coordinatorState \in {"INIT", "PREPARED"}
  /\ coordinatorState' = "ABORTED"
  /\ UNCHANGED <<wireDispatched, escrowDebited>>

Next == Prepare \/ Commit \/ Abort \/ UNCHANGED vars

Spec == Init /\ [][Next]_vars

\* SAFETY INVARIANT: A payout batch is never wired more than once.
SingleDisbursementInvariant == wireDispatched <= 1

\* ATOMICITY INVARIANT: Wire dispatch occurs if and only if escrow is debited.
AtomicSettlementInvariant == (wireDispatched = 1) <=> (escrowDebited = TRUE)
=============================================================================
