# Available-evidence forecast diagnostic

Replay the256 consumed closed-loop credit schedules without changing a single
issued forecast or state transition. For arm2, capture the monitoring cache BEFORE
any learner arm inspects the frame. Union ONLY this mask with arm2's requested
mask, not with coordinates acquired by other experimental arms. Both masks are
known before the current outcome. No new Reader calls are allowed in the diagnostic.

Evaluate existing incumbent/short/local-or-pooled/subset models with this union,
using exactly the issued outer and inner mixture weights and the same model
version. First reconstruct the original requested-view forecast exactly. Models
and weights are never selected using the revealing outcome or diagnostic score.
The available-view forecast is shadow only: do not train on it, change acquisition,
or write it into the issued-feedback journal.

This changes the forecast's consumed information but not acquisition. Requested,
monitor-available and consumed masks must remain separate. The six-coordinate
foreground acquisition cap remains unchanged; the diagnostic may CONSUME up to
nine already-acquired coordinates. It is not permission to misreport a nine-bit
read as a six-bit read or claim a new closed-loop policy has the old contract.

Retain every scenario/cohort/schedule. Exploratory diagnostic screens: reverse
post-Brier gain mean>=.005 with mean-3.5SE>0 on four cells; full/post upper harm<=.01
on all16 cells. Independent frames are not assumed. This is consumed fixed-state
potential, not an operational improvement, calibration guarantee or adoption test.
