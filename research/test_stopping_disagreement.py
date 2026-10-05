import unittest
from stopping_disagreement import category

class StoppingTests(unittest.TestCase):
    def test_regions(self):
        for p in (.1,.9,0.,1.):self.assertEqual(category("confidence",p),"both_confident")
        for p in (.100001,.5,.899999):self.assertEqual(category("confidence",p),"guide_confident_mixture_uncertain")
        self.assertEqual(category("budget",.5),"budget_or_other")

if __name__=="__main__":unittest.main()
