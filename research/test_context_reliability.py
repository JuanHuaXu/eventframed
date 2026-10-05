import copy
import unittest

from context_reliability import ContextReliability, replay_record


class ContextTests(unittest.TestCase):
    def test_support_separation_and_bounds(self):
        state = ContextReliability()
        key = (1, 1, 1)
        for _ in range(7):
            state.observe(key, False, .95)
        self.assertEqual(state.status(key), "unknown")
        state.observe(key, False, .95)
        self.assertEqual(state.status(key), "warn")
        self.assertEqual(state.status((3, 1, 1)), "unknown")
        self.assertEqual(state.status((1, 0, 1)), "unknown")
        self.assertEqual(state.status((1, 1, 0)), "unknown")
        for _ in range(64):
            state.observe(key, True, .95)
        self.assertEqual(state.status(key), "clear")
        self.assertEqual(len(state.history[key]), 64)
        for i in range(511):
            state.observe((i, 0, 0), True, .95)
        state.status(key)
        state.observe((511, 0, 0), True, .95)
        self.assertEqual(len(state.history), 512)
        self.assertNotIn(key, state.history)

    def test_no_current_future_or_hidden_input_use(self):
        record = dict(Ticks=[dict(Predictions=[{}, {}, {"P": .95}],
                                 Outcome=False, Missing=False,
                                 Delivered=[i-2] if i >= 2 else []) for i in range(40)],
                      Views=[[{}, {}, {"stop": "confidence", "trace":
                                      [{"observed": 1, "values": 1}]}] for _ in range(40)])
        a = replay_record(record)
        self.assertEqual(a[9]["status"], "unknown")
        self.assertEqual(a[10]["status"], "warn")
        changed = copy.deepcopy(record)
        changed["Inputs"] = [511] * 40
        for i in range(20, 40):
            changed["Ticks"][i]["Outcome"] = True
        b = replay_record(changed)
        self.assertEqual([v["status"] for v in a[:23]], [v["status"] for v in b[:23]])


if __name__ == "__main__":
    unittest.main()
