# Whole-learner feedback pairing

Consumed v120 diagnostic, not a new learner or oracle ceiling. Pair immediate
and delayed/missing schedules by phase/case/index. Require identical Initial,
and all256 X/Y/Q values. Validate immediate Delay=0 and Missing=false. Retain
all1344 pairs and all15 original experts/mixers; recompute expected Brier directly
from issued probabilities, check against original Metrics. Hash raw data/script.

For each phase/case/arm and all256/terminal64 window, report immediate, delayed,
paired gain delayed-minus-immediate, and mean +/-3.5SE over32 pairs. Count positive
gain with lower>0 and negative gain with upper<0 descriptively, not as familywise
or anytime guarantees. Do not pool two schedules as independent observations.

Compare original four-expert Markov (arm12) to itself across schedules, not to
later switch mixtures. If full feedback fails to rescue cases, acquisition timing
alone is not a sufficient explanation; this does not prove information is useless.
More frequent labels change sample ages under count caps and may change learning
adversely. Preserve all cases rather than selecting favorable schedules.
