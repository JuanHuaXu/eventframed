"""Independent known-copy planning using original likelihood and loss APIs."""
from functools import cache
import json
import experiment


def evaluate(name):
    table = experiment.likelihood(name)

    @cache
    def posterior(observations):
        w = [1 / 16] * 16
        for t, y in enumerate(observations):
            if y >= 0:
                w = experiment.update(w, table[t], bool(y))
        return tuple(w)

    def risk(observations):
        w = posterior(observations)
        classes = experiment.classes(w, name)
        return sum(p * experiment.brier(classes, h % 4)
                   for h, p in enumerate(w))

    def branches(observations, t):
        p = sum(w * q for w, q in zip(posterior(observations), table[t]))
        for y, weight in ((0, 1-p), (1, p)):
            child = list(observations)
            child[t] = y
            yield tuple(child), weight

    @cache
    def plan(observations, depth):
        if depth == 0 or -1 not in observations:
            return risk(observations), None
        best, chosen = float('inf'), None
        for t, y in enumerate(observations):
            if y < 0:
                value = sum(p * plan(child, depth-1)[0]
                            for child, p in branches(observations, t))
                if value < best - 1e-14:
                    best, chosen = value, t
        return best, chosen

    @cache
    def policy(observations, budget, depth):
        if budget == 0 or -1 not in observations:
            return risk(observations)
        chosen = plan(observations, min(budget, depth))[1]
        return sum(p * policy(child, budget-1, depth)
                   for child, p in branches(observations, chosen))

    return {str(d): [policy((-1,) * 8, b, d) for b in range(9)]
            for d in (1, 2, 3, 8)}


print(json.dumps({name: evaluate(name)
                  for name in ('noise05', 'noise20', 'null')}, indent=2))
