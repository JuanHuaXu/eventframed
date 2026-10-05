"""Full copied-report information ceiling using the original likelihood API."""
import itertools
import json
import experiment

results = {}
for name in ('noise05', 'noise20', 'null'):
    table = experiment.likelihood(name)
    mass = risk = error = 0.0
    for reports in itertools.product((False, True), repeat=8):
        weights = [1 / 16] * 16
        for t, y in enumerate(reports):
            weights = [w * (table[t][h] if y else 1 - table[t][h])
                       for h, w in enumerate(weights)]
        p = sum(weights)
        posterior = [w / p for w in weights]
        classes = experiment.classes(posterior, name)
        # Direct conditional expected multiclass Brier, not the Gini shortcut.
        risk += sum(w * experiment.brier(classes, h % 4)
                    for h, w in enumerate(weights))
        error += p * (1 - max(classes))
        mass += p
    results[name] = dict(mass=mass, brier=risk, error=error)
print(json.dumps(results, indent=2))
