import unittest
import prepare


class Tests(unittest.TestCase):
    def test_separation(self):
        d = prepare.prepare()
        self.assertEqual(len(d["queries"]), 16)
        for q in d["queries"]:
            self.assertNotIn("support", q)
            self.assertNotIn("answer", q)
            self.assertNotIn(d["oracle"][q["case_id"]]["answer"], q["question"])
        for row in d["corpus"]:
            self.assertEqual(set(row), {"fixture_id", "text", "source"})

    def test_relation_and_clusters(self):
        d = prepare.prepare()
        texts = {v["fixture_id"]: v["text"] for v in d["corpus"]}
        self.assertIn("contact with Earth", texts["m10-contact"])
        self.assertIn("launch from Earth", texts["v2-launch"])
        clusters = [{d["oracle"][q["case_id"]]["cluster"] for q in d["queries"] if q["split"] == split} for split in ["design", "confirmation"]]
        self.assertFalse(clusters[0] & clusters[1])


if __name__ == "__main__":
    unittest.main()
