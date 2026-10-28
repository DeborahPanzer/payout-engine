----------------------- MODULE RollingReserveInvariant -----------------------
EXTENDS Naturals

VARIABLES
  grossCaptured,
  heldReserve,
  disbursedNet

vars == <<grossCaptured, heldReserve, disbursedNet>>

Init ==
  /\ grossCaptured = 10000
  /\ heldReserve = 1000
  /\ disbursedNet = 9000

ReleaseMaturedReserve ==
  /\ heldReserve > 0
  /\ disbursedNet' = disbursedNet + heldReserve
  /\ heldReserve' = 0
  /\ UNCHANGED grossCaptured

Next == ReleaseMaturedReserve \/ UNCHANGED vars

Spec == Init /\ [][Next]_vars

\* CONSERVATION INVARIANT: Held reserve plus disbursed net always equals gross captured.
ReserveConservationInvariant == heldReserve + disbursedNet = grossCaptured
=============================================================================
