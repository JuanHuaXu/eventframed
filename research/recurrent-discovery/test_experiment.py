import unittest
import json
from pathlib import Path
import tempfile

import torch

from experiment import Learner, make_data, parameter_hash, signature, update
from verify_evidence import verify


class ExperimentTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        torch.set_num_threads(1)

    def test_partition_and_reverse_pairs(self):
        train, test, meta = make_data(123, "addition")
        keys = lambda data: {tuple(sorted(row[:2])) for row in data.tolist()}
        self.assertFalse(keys(train) & keys(test))
        self.assertEqual(len(train) + len(test), 31 ** 2)
        self.assertEqual(len(keys(train)), meta["train_groups"])
        self.assertEqual(len(keys(test)), meta["test_groups"])
        random_train, random_test, _ = make_data(123, "random_labels")
        self.assertTrue(torch.equal(train[:, :2], random_train[:, :2]))
        self.assertTrue(torch.equal(test[:, :2], random_test[:, :2]))

    def test_no_cross_example_state_or_parameter_mutation(self):
        model = Learner()
        pairs = torch.tensor([[1, 2], [3, 4]])
        before = parameter_hash(model)
        a = model(pairs, "LLHLLH")
        model(torch.tensor([[8, 9]]), "LLHLLH")
        self.assertTrue(torch.equal(a, model(pairs, "LLHLLH")))
        self.assertTrue(torch.allclose(a.flip(0), model(pairs.flip(0), "LLHLLH")))
        self.assertEqual(before, parameter_hash(model))

    def test_training_reaches_both_modules(self):
        train, _, _ = make_data(123, "addition")
        for schedule in ["LH", "LHLHLH", "LLHLLH"]:
            model = Learner()
            old = {k: v.clone() for k, v in model.state_dict().items()}
            update(model, torch.optim.AdamW(model.parameters()), train, schedule)
            for key in ["embedding.weight", "low.weight", "high.weight", "output.weight"]:
                self.assertFalse(torch.equal(old[key], model.state_dict()[key]), key)
                self.assertTrue(torch.isfinite(model.state_dict()[key]).all())

    def test_signature_does_not_relabel_early_success_as_delayed(self):
        curve = [{"step": n * 100, "train": {"accuracy": 1.0},
                  "test": {"accuracy": 1.0}} for n in range(12)]
        self.assertFalse(signature(curve)["delayed_generalization_signature"])
        for row in curve[:6]:
            row["test"]["accuracy"] = 0.1
        self.assertTrue(signature(curve)["delayed_generalization_signature"])
        curve[-2]["test"]["accuracy"] = 0.1
        curve[-3]["test"]["accuracy"] = 0.1
        curve[-4]["test"]["accuracy"] = 0.1
        self.assertFalse(signature(curve)["delayed_generalization_signature"])

    def test_evidence_rejects_incomplete_duplicate_and_changed_results(self):
        artifact = Path(__file__).resolve().parents[2] / "docs/experiments/recurrent-discovery-v1.json"
        if not artifact.exists():
            self.skipTest("full pilot evidence not generated yet")
        self.assertEqual(verify(artifact)["verified_runs"], 16)
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "tampered.json"
            for mutation in ["incomplete", "duplicate", "changed_final", "changed_source"]:
                data = json.loads(artifact.read_text())
                if mutation == "incomplete":
                    data["complete"] = False
                elif mutation == "duplicate":
                    data["runs"].append(data["runs"][0])
                elif mutation == "changed_final":
                    data["runs"][0]["final"]["test"]["accuracy"] = 1.0
                else:
                    data["source_sha256"] = "not the source"
                path.write_text(json.dumps(data))
                with self.assertRaises(AssertionError, msg=mutation):
                    verify(path)
if __name__ == "__main__":
    torch.set_num_threads(1)
    unittest.main()
