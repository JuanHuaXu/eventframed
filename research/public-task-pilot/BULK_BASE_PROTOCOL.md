# Bulk derived-base construction rescue

Change only BuildHNSWBase in a research build overlay: insert the captured base
in one synchronous transaction instead of one durable Insert per record. The
authoritative record store, delta cap64, trigger32, leases8, retirement2, arrival
rates,100ms budget, eight arms and final worker drain remain unchanged from v2.
No module edit or asynchronous durability. Empty bases require no transaction.

The new derived graph remains unpublished until the build succeeds and the
serving pair is atomically replaced. A build failure leaves the old serving pair
available; it must not be treated as a committed authoritative event. Bulk staging
has corpus-sized memory/work and may hit backend transaction limits. This is a
finite200/800-record experiment, not a guarantee for million-record graphs.

Before timing, run all researchindex tests under the overlay with the race
detector. Then run the exact compaction-inclusive fixed-arrival screen, without
concurrent agent-started tests/builds, into generation-bulk-results.json with
sidecars and retained stores. Reuse the existing verifier and full pass gates.
Compare descriptively with generation-load-v2-results.json. Do not relax a gate
or claim EventFrame service validation from this component workload.
