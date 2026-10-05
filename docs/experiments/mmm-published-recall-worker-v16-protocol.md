# Native batch journal worker-capacity v16: frozen screen

Date: 2026-10-02. V15's native batch journal reduced full-Recall
offer-to-done p99 to about 155 ms but failed the unchanged <100 ms
gate with four Recall workers. Its read-admission wait p50 was about
10 ms and occupied-call p50 about 21 ms. This is a separate
capacity hypothesis, not a retroactive v15 threshold change.

Copy v15's test-only native batch architecture and fixture exactly:
256D corpus, 128 visible writes and 128 full Recalls at nominal
4 ms offers, cap-16/16-ms event writer, published-LSN Search,
read-to-journal admission, full residual-enabled service, native
journal batch cap four, 1 ms dwell, 500 ms internal transaction
bound, exact top-150 oracle and durable-after-readback marker
publication. Change only Recall workers from four to eight. Do
not change the journal worker, batch cap, dwell, offered load,
limits or verification. Run two normal fresh trials; rerun the
unchanged v15 control nearby for environmental context. Record
batch-size distribution, queue/admission/call/offer tails, writer
age, published-view age, stale retries, and resource-relevant worker
count. The nearby control is descriptive, not a new fit set.

The frozen gates remain 128/128 durable writes and successful
Recalls; zero stale, oracle, packet/journal, future or
acknowledged-before-offer violations; final READY; write age p99
<250 ms, Recall call and offer-to-done p99 <100 ms, published-view
max age <250 ms. Run `-race` correctness-only mode plus ordinary
package tests and vet. A finite pass would not complete Goal 6:
loaded durable learning, crash recovery, larger corpora and
real-agent outcomes remain separate requirements. Production stays
untouched.
