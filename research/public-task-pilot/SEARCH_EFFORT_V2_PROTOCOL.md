# Search effort above the backend floor

The first search-effort diagnostic requested0/100/200/400. Inspection of pinned
libravdb hnsw.go shows ef=max(EfSearch,k,2*EfConstruction,efOverride). Therefore
all four arms used400 with the frozen construction parameter200. That run is
a same-effective-effort control, NOT evidence that larger effort fails.

Retain it unchanged. This follow-up uses0/800/1600/3200, effective400/800/1600/3200,
with all other STATIC_RECALL and SEARCH_EFFORT protocol conditions unchanged.
The direct default adapter remains a per-query control. This is still discovery
on the known failing corpus, not independent confirmation or production tuning.
Compare miss counts, full candidates and unloaded timing. More search is not a
rescue unless it actually recovers the target; record all negative results.
