# Bounded Recall worker scaling v25: frozen protocol

Date: 2026-10-01. This is a research-only Goal 6 test, not a production
configuration change. The v24 full Recall diagnostic found mean occupied
call time of about 33.4 ms with four workers under 8 ms nominal offers,
close to or above the nominal four-worker capacity. That observation does
not establish that adding workers helps: the publication guard, SQLite
journal, event writer, or CPU may instead become more contended.

## Matched workload and decision

Use one guarded SQLite WAL/FULL journal per isolated trial, the 200-live-event
fixture, exact `RecallK=200`, `PackK=10`, 192 offers, a monotonic 8 ms ticker,
256 concurrent future-only writes in loaded trials, and 4, 8, or 16 Recall
workers. Run quiet and loaded cases for each count in three rotated trials;
reverse quiet/loaded order in the middle trial. Every variant retains the
same event writer, embedder, service, as-of guard, durable acknowledgement,
and independent SQLite reopen check. Capture the actual 191 inter-offer gaps
per trial, rather than inferring rate from the ticker setting alone.

Require per trial: 192 completed offers, exactly 200 distinct live nominations
per offer, no future packed event, 256 completed future-only writes and
nonzero overlap when loaded, no service error, and 192 acknowledged durable
journals after close/reopen. Report actual offer-gap median/p99, observed
offer rate, mean call, call p99, queue p99, offer p99, journal p99, and
guard-wait p99 for every trial and pooled arm/case. A valid **Goal 6 latency
component pass** requires the 8- or 16-worker loaded arm to have offer
p99 <100 ms both pooled and in all three trials, with quiet p99 no worse
than the four-worker control by more than 10 ms. Passing does not prove
freshness, power-loss recovery, production latency, or real-agent benefit.

If more workers raise guard/journal time or leave loaded offer p99 above
100 ms, worker count alone is not a rescue and the next design must alter
publication/admission ordering. Do not choose a worker count after looking
at the trials and call it preregistered success. Preserve negative results.
