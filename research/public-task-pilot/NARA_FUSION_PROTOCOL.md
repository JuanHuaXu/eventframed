# Untouched National Archives packet-fusion transfer

Frozen on 2026-10-01 before any service run on this set. The proposed rule
comes from a diagnostic on the already-consumed RFC transfer, not from this
set. This set uses 15 National Archives milestone-document facts, each with
literal and paraphrased questions, two absent controls, and the same 13 NASA
distractors used in the RFC screen. The National Archives [milestone list](https://www.archives.gov/milestone-documents/list)
is the primary source for document years. The questions and oracle are frozen
before inspecting model output. Paired wordings are one fact cluster, not two
independent observations.

Use the same Nomic embedding model, opt-in task-plus-lexical overlay, source
order, `CaptureTurn`/`Recall` calls, recall50, pack10, and diversity setting
as RFC transfer. Capture raw focus and priority outputs before reading the
oracle. No learning, feedback, LLM generation, or production OpenClaw.

The third arm is an offline, fixed-budget packet fusion. Let B be the ordered
focus packet of at most ten distinct fixture IDs and P the priority packet.
If P's first ID is already in B, move that ID to the front of B, retaining
the relative order of every other B member. Otherwise return B unchanged.
An empty P also returns B. Never add a candidate absent from B. This
guarantees exact support-set parity with focus; it does not guarantee top1
non-harm or recover a focus packing miss. The third arm is a permutation of
the baseline packet, not a new posterior or law.

Score per wording and fact cluster, separately for focus, priority, fusion:
top1 target ID and target survival anywhere in pack10. A finite transfer pass
for fusion requires (1) zero top1 losses against focus for either wording,
(2) at least one top1 gain across positive questions, (3) exact pack membership
parity with focus and at most ten unique IDs, and (4) identical full-frontier
numeric scores and forecast laws between focus and priority, with journals
matching packet explanations. Report absent controls as unsupported packs,
not successful abstentions. Also report priority independently; no relabeling
of its RFC failure.

Sequential isolated Recall timings do not measure integrated fusion latency,
agent answer quality, concurrent p99, or OpenClaw performance. If the finite
gate fails, preserve that result. Even a pass is a small public retrieval
screen, not completion of research goal 5.
