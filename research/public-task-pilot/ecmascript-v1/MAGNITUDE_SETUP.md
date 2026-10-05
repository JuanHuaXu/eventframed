# Preserved Pre-Collection Build Repair

The first `go test -race ./internal/researchmagnitude
./cmd/public-ecma-magnitude` passed the fusion package(1.247s) but failed the
collector build: ResearchFrontierCandidate has no Score or Forecast field.
This was a collector authoring error, not a runtime defect or data failure.
No fit/design/confirmation retrieval existed at this point.

Repair binds tapped EventID to the packet's recorded ForecastBundle, exactly
as the earlier public Git collector does. Corrected laws/rank scores come
from that journal, never fabricated fields or a changed service contract.
Full journal/report and tapped snapshot identities remain required. The
extraneous accepted CLI argument count was also removed before freezing.
Subsequent tests and source freeze must cover the repaired collector.
