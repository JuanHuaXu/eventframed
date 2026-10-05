# Concurrent pre-search snapshot v1: frozen full-Recall screen

Use two fresh rotated pairs of private LibraVDB stores. Each arm has 64
synthetic tenants, with one as-of-visible seed and one future-available
sentinel preloaded per tenant. Offer 64 `Service.Recall` requests at 4 ms
intervals to four workers. For each request, a context-scoped test wrapper
inserts one additional as-of-visible event immediately after the first
underlying Search. Compare the current Search-then-Snapshot order with a
pre-search pin that presents the captured pre-search snapshot to that
attempt's subsequent snapshot read. Production code is unchanged.

Record offer-to-response and service-call p50/p95/p99, search attempts,
successful responses, retry-exhausted stale errors, other errors, silent
omissions, journal/packet mismatches and future-event leaks. A silent
omission means the packet claims a runtime version at least as new as its
own injected write but does not include that event. Every offered request
must actually commit its injected write exactly once. The pin arm passes
this finite screen only with zero omissions and future/journal violations,
at least 99% successful responses, and offer-to-response p99 below 100 ms.

The current-order arm is a negative control, not a safe latency baseline.
The two pairs use the same synthetic event design and are timing repeats,
not independent outcome populations. Even a pass here would not establish
the published-LSN backend, journal-through-gate integration, durable learner
freshness, external retrieval, large-corpus scale, or production readiness.
