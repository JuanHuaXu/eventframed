# Direct mixture loss: frozen consumed-data screen

Vovk's [Competitive On-Line Linear Regression, Section2](https://papers.neurips.cc/paper/1419-competitive-on-line-linear-regression.pdf)
adds current input covariance before predicting, then adds its target cross
product after the outcome arrives. This differs from ordinary online ridge.
The original protocol supplies each outcome after its prediction. We do not
inherit its guarantees under missing/delayed feedback, bounded weights,
rolling retention, or safety projection.

Our scalar adaptation centers at m_j=(b_j+c_j)/2, d_j=c_j-b_j. For the as-of
arrived set A_t, define G=sum d_j*(y_j-m_j) and H=1+sum d_j^2. Ridge proposes
w=clip_[0,1](.5+G/H). The current-covariance variant uses H+d_t^2 instead.
No current or pending label is included. Both statistics use the same admitted
set; missing samples are not assigned pseudo-outcomes. Empty evidence yields
.5. This minimizes the declared centered ridge objective, not the original
paper's full protocol and not the loss after guard clipping.

Freeze four proposals: all-history ridge, last64 arrived origins ridge,
all-history current-covariance, and last64 current-covariance. The64 retention
matches an existing evidence cap; regularization1 and prior.5 are declared
without a parameter sweep. Score each before and after the unchanged .01
pointwise Brier guard. Retain Fixed Share/pointwise, Fixed Share/local ledger,
and Markov on identical stored fits. Report all variants, not just a winner.

Use consumed independent-v1 data. No fresh quality claim. Test unavailable
outcome invariance and closed-form solutions against direct objective checks.
Charge original fitted-model costs; measure combiner overhead separately.
Reference replay scans historical origins, so it is O(T^2), not an O(1)
serving implementation. Better adaptive recovery remains an open requirement.
