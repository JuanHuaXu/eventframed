# Guide authority and component quality

Post-hoc analysis of the frozen v11 artifact, not new confirmation or a changed
success criterion. Python reconstruction reproduces all1,179,648 journaled
mixture forecasts within1e-12 and the adaptive inner mixture at each frame.
Weights advance only after the forecast, using recorded delivery order and
recorded pre-outcome expert probabilities. Guide selection uses raw stored
weights, matching the runner, rather than the fixed-share predictive weights.

For clustered shift128, adaptive-retained empirical mode:

| Split | All frames | Supported-fit frames | Tree-guided frames | Average tree forecast mass |
| --- | --- | --- | --- | --- |
| Design | 4096 | 3109 | 96 | 2.99% |
| Confirmation | 4096 | 3021 | 177 | 3.78% |

Confirmation guides: incumbent1249, short969, long1701, tree177. Tree mass is
outer inner-slot mass times the inner tree slots' combined predictive mass,
averaged across all frames. It measures forecast influence, not factual
confidence or proof of optimal observation behavior.

Post-change, supported-fit-only component Brier on the SAME adaptive arm masks:

| Empirical mode | Short count | Long count | Tree | Mixture |
| --- | --- | --- | --- | --- |
| Design | 0.132406 | 0.136679 | 0.191283 | 0.136417 |
| Confirmation | 0.125932 | 0.127965 | 0.179900 | 0.127553 |

These component scores exclude pre-fit frames, so their mixture values differ
from v11's complete post-window headline. Each stream contributes equally.
The corrected tree is substantially worse than short/long counts on the observed
masks, even though it improves over its uniform-integration counterpart.

## Consequence

Limited guide authority is real, but simply granting more authority to this tree
is not supported. Low authority coincides with poorer observed predictions. This
does not prove the tree would be worse on its own unchosen observation paths;
selection and model quality remain coupled. An alternate-path diagnostic or a
different sample-efficient challenger is a better next discriminator than
lowering the existing guide threshold and calling that a rescue.

All seven roadmap directions remain in scope. This narrows one lead in1/4 and
does not establish success or exhaustion for2/3/5/6/7. No implementation behavior,
frozen experiment, paper claim or production configuration changed.

Artifacts: [guide reconstruction](mmm-acquisition-v11-guides.json),
[component losses](mmm-acquisition-v11-experts.json). Both bind the raw v11 hash.

```sh
python3 research/acquisition_guides.py docs/experiments/mmm-acquisition-v11.json.gz NEW-guides.json
python3 research/acquisition_experts.py docs/experiments/mmm-acquisition-v11.json.gz NEW-experts.json
```
