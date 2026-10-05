# Independent response generators: contract and verification

Status: generator QA passed; learner transfer quality NOT YET TESTED. These are
synthetic constructions, not empirical agent tasks or reproductions of a paper.
The implementation imports no forecasting model or archived Boolean truth helper.

## Frozen construction

Inputs are uniform on the 512 nine-bit vectors. Relevant coordinates are drawn
from a random permutation of bits 0 through 5; bits 6 through 8 are irrelevant.
This preserves the six-coordinate acquisition budget without revealing relevance
to the learner. It does not test dependent input distributions.

- Additive: q = 0.1 + 0.8 sigmoid(z), with z = b + sum_j a_j(2x_j-1).
  The five coefficient magnitudes are 0.4, 0.7, 1.0, 1.3, 1.6, independently
  signed; b is equally likely -0.4 or +0.4.
- Hierarchical: a complete depth-three binary tree with branch coordinates
  [p0,p1,p2,p3,p4,p4,p5] in breadth-first order. Each of its eight leaves
  independently receives Uniform[0.1,0.9) probability. The repeated p4 lies
  on different paths, not twice on one path.
- Local interaction: the first four permuted coordinates address sixteen
  independent Uniform[0.1,0.9) probability cells.

Each trajectory draws two endpoint response models. Stationary cases use only
the first. Abrupt cases switch at an odd tick uniformly selected from 97 to 159.
Gradual cases mix endpoints with weight (step-change+1)/32, capped at one,
starting at the change tick. No change begins on a 32-tick publication boundary.

There are 16 initial observations at ticks -16 through -1 and 256 scored ticks.
Labels are Bernoulli(q). Scored packets carry independently sampled delays
uniform on integers 0 through 31 and missing indicators with probability 0.2.
Initial packets have neither delay nor missingness. An immediate-feedback runner
may override the schedule, but must retain the same latent inputs and labels.

The generation seed is base + phase*1000000 + family*100000 + mode*10000
+ index*10. Offsets 0 through 4 respectively select parameter, input, label,
delay and missingness RNGs. Phases are 0/1, families and modes 0/1/2, indices
0 through 31. QA base 2176111600 is consumed and must not become untouched
confirmation data. A future quality protocol needs a fresh disjoint base.

## Observation boundary

The simulator packet contains X, Y, Delay and Missing only. It excludes teacher
identity, relevant coordinates, change time and target probability. This is a
data-shape check, not a complete non-leakage proof: the future runner must reveal
X only through charged observation reads, deliver Y only when feedback arrives,
and keep future delay/missingness information outside policy decisions.

The Teacher object and its detached Description are evaluator-side only.
Conditional probabilities average over all uniform input completions. No true
conditional probability may enter the forecast policy.

## Completed checks

`go test -race ./internal/transfergenerator -count=1 -v` passed in 1.741s.
`go vet ./internal/transfergenerator` passed. These are scoped component checks,
not a whole-repository test, performance benchmark or model-quality experiment.

- All 576 configurations reproduce exactly and satisfy packet/time bounds.
- At five boundary-sensitive times per configuration, all 512 input probabilities
  agree with independently written formulas, including tanh for the logistic
  family and direct leaf/table indexing. Irrelevant-bit flips preserve q.
- All 19,683 partial input states are checked against independent completion
  enumeration for one gradual-mixture teacher per family: 59,049 comparisons.
- Invalid queries/specifications are rejected and returned descriptions do not
  alias the teacher. The learner packet's field boundary is checked explicitly.
- All 2,880 effective QA seeds are distinct modulo 2147483647 and disjoint from
  the archived learner/null seed ranges explicitly enumerated in the test.
  This is not a claim about unenumerated historical or future seed ranges.

SHA-256 at verification:

```text
5adcfbdc2834f17123f741a2cf63091b944a7e96f3fb51a94e945a6d1499f015  generator.go
b661b11de051365dac2e5aee45a5f01ba017b0dcf783717139258da2ce84037e  generator_test.go
```

Next freeze the learner adapter, acquisition accounting, fresh quality seeds,
comparison gates and score windows before inspecting outputs. Retain unchanged
generic, arrival-log and fixed-Markov controls. Generator correctness does not
rescue v115, establish forecast improvement, or close any research direction.
