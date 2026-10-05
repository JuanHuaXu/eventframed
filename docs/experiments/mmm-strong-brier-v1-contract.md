# Strong Brier substitution screen

Consumed independent-v1 cohort. No fresh confirmation claim. Research only.
Read Vovk and Zhdanov (2009), Algorithm 1 and Theorem 1:
https://www.jmlr.org/papers/volume10/vovk09a/vovk09a.pdf

Their two-coordinate Brier equals twice our scalar loss. With scalar eta=2,
define z_y=sum_k w_k exp(-2*(p_k-y)^2). The binary substitution is
p=.5+(log(z_1)-log(z_0))/4. It obeys 2*(p-y)^2 <= -log(z_y).
Static immediate-feedback cumulative scalar regret with equal two-expert prior
is <= log(2)/2. Do NOT claim this unmodified bound for delayed/missing outcomes
or for changing expert weights due to Fixed Share.

Freeze two candidates: strong substitution for segment64/Markov and
static64/Markov. Matched controls: linear averages using the SAME eta=2 weights.
Existing eta=1 linear screens remain the reference. Keep prior1/2, existing
share schedule alpha_j=1/(j+1) for j>0, alpha_0=0, and original origin-order
delayed factor refiltering. No rate sweep, fit change, guard, or scenario route.

Report all four arms with the original 672 non-harm and 96 recovery-gain
comparisons each, using the eight consumed indices transparently. Keep .01
upper non-harm and .005 mean/positive-lower gain criteria; approximate +/-3.5SE
is not original32-index confirmation. Include per-window harms and strata.
Verify paper substitution independently, negative control for linear eta2,
normalization, static immediate bound, as-of tests and eta1 reference identity.
No serving claim; measure reference combination overhead separately, retaining
the original 545.32s source-model collection cost.
