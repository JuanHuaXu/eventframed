import unittest
from error_budget import population,conditional,components

class ErrorBudgetTests(unittest.TestCase):
    def test_complete_information(self):
        for family in ("majority","multiplexer"):
            for generator in ("fair","biased","clustered"):
                for local in (False,True):
                    for x,_,p in population(family,generator,local):
                        q=conditional(family,generator,local,511,x)
                        self.assertAlmostEqual(q,p,places=12)
                        self.assertAlmostEqual(components(q,q)["expected_brier"],.0475,places=12)

    def test_correlated_conditioning(self):
        self.assertAlmostEqual(conditional("majority","clustered",True,1,1),.869,places=12)
        self.assertAlmostEqual(conditional("majority","clustered",False,1,1),.83984,places=12)
        self.assertAlmostEqual(conditional("multiplexer","clustered",True,4,4),.788,places=12)

    def test_loss_identity(self):
        for p in (0.,.1,.5,.95,1.):
            for q in (.05,.25,.5,.75,.95):
                out=components(p,q)
                self.assertAlmostEqual(out["expected_brier"],out["noise"]+out["observation_deficit"]+out["forecast_deficit"],places=12)

if __name__=="__main__":unittest.main()
