# Bounded background retirement load screen

Repeat DERIVED_NOSYNC_PROTOCOL.md with only the research serving wrapper changed
to AsyncPartitionServing. Same8 partitions,2 TOTAL retired handles,8 leases,
64 TOTAL delta entries,32 trigger, single compactor and original arrival/deadlines.
One cleanup worker; queue capacity includes running work. Failed cleanup remains
charged. Release reports scheduling success, not successful reclamation.

Close refuses active users, then stops admission, joins every queued cleanup and
closes current graphs. Record this drain duration separately; do not erase deferred
work. Cleanup errors surface at shutdown and fail the audit gate. Error handling
does not claim an uncertain resource reclaimed. Original synchronous variant stays
unchanged as control. No retries, cap increases or public/production behavior change.

Success requires all six original combined gates, including reopen and shutdown
audits. Moving checkpoint work is not itself evidence of better throughput.
