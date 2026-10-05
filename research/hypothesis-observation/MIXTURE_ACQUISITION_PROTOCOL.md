# Mixture-aware acquisition v7

Frozen before execution. Test the v6 joint reliability model with its own exact
one-step target-Gini selector; no additional evidence or hidden truth access.

For target class c and outcome y at candidate test t, compute joint mass
a[c,y] = sum_R,h:class(h)=c w_R w[h|R] P(y|t,h,R,history).
The expected next posterior squared norm is sum_y,c a[c,y]^2 / sum_c a[c,y].
Subtract the current squared norm to obtain expected target-Gini reduction.
Select its maximum among unused source slots, with v3's rounded14-digit and
smallest-test/smallest-slot tie rule. Zero-probability branches contribute zero.
This preserves shared-R dependence; it is not Gini of independently averaged
per-test mode priors. No source-mode entropy bonus or lookahead is used.

Seven v6 cases; two splits,128 episodes per case, new base seed202609170417.
All arms share environmental tapes and signals,16 queries and eight signal
checks (cost24). Controls: fixed acquisition/fixed scoring, fixed acquisition/
reliable scoring, fixed acquisition/averaged scoring. Treatment: joint acquisition
and averaged scoring. Ground-truth modes remain simulator-only data.

Frozen confirmation screen:

- Mean curve/final harm <=0.01 against fixed in all seven cases.
- Mean curve/final harm <=0.01 against reliable-only in independent20.
- Misleading20 final gain against reliable-only >=0.03 with paired z3.3 lower >0.
- Independent20 curve gain against fixed-acquisition averaging >=0.005 with
  paired z3.3 lower >0, so unchanged output cannot count as acquisition rescue.

These are finite screening gates, not simultaneous uncertainty guarantees.
Report all cases, directions and failures. No runtime adoption or latency claim.
Additional computation is O(T*R*H) per selection under the finite caps, with
T=8,R=3,H=16; measure production costs separately before integration.

Verify analytical scores against independently copied model updates for both
counterfactual outcomes; selection must not mutate state. Test exhausted slots,
normalization, unique acquisition, environmental tape agreement, full replay
and source hashes. Preserve prior artifacts; exclusive-create new outputs.
