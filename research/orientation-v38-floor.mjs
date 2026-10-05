// Post-hoc representational floor, never a forecast/selection/training input.
import fs from "node:fs";
import crypto from "node:crypto";
import readline from "node:readline";
import assert from "node:assert/strict";

function floor(template, rates, priority = false) {
  assert.equal(template.length, rates.length);
  let numerator = 0, denominator = 0, total = 0, noise = 0;
  for (let i = 0; i < template.length; i++) {
    const a = template[i], p = rates[i], w = priority && i < 10 ? 3 : 1;
    assert.ok(Number.isFinite(a) && a > 0 && a < 1 && Number.isFinite(p) && p >= 0 && p <= 1);
    const d = 1 - 2 * a;
    numerator += w * d * (p - a);
    denominator += w * d * d;
    noise += w * p * (1 - p);
    total += w;
  }
  const f = denominator === 0 ? 0 : Math.max(0, Math.min(1, numerator / denominator));
  let risk = 0;
  for (let i = 0; i < template.length; i++) {
    const q = template[i] + (1 - 2 * template[i]) * f, p = rates[i], w = priority && i < 10 ? 3 : 1;
    risk += w * (q * q - 2 * q * p + p);
  }
  return { reversal: f, brier: risk / total, trueRateFloor: noise / total };
}

function interval(x) {
  const mean = x.reduce((a, b) => a + b, 0) / x.length;
  const se = Math.sqrt(x.reduce((a, b) => a + (b - mean) ** 2, 0) / ((x.length - 1) * x.length));
  return { mean, se, lower: mean - 3.5 * se, upper: mean + 3.5 * se };
}

const near = (a, b) => assert.ok(Math.abs(a - b) < 1e-12, `${a} != ${b}`);
near(floor([.2, .8], [.2, .8]).brier, .16);
near(floor([.2, .8], [.8, .2]).brier, .16);
near(floor([.2, .8, .2, .8], [.8, .2, .2, .8]).brier, .25);
near(floor([.5, .5], [.2, .8]).brier, .25);
near(floor([.2, .2], [.01, .01]).reversal, 0);
near(floor([.2, .2], [.99, .99]).reversal, 1);
assert.throws(() => floor([NaN], [.5]));

if (process.argv.includes("--self-test")) {
  console.log("Seven independent floor/domain checks passed; post-hoc tool only.");
} else {
  const summary = { study: "orientation-v38-representational-floor", postHoc: true, learnerUsesOracle: false, cohorts: {} };
  for (const split of ["design", "confirmation"]) {
    const path = `docs/experiments/mmm-orientation-v38-${split}.jsonl`;
    const audit = JSON.parse(fs.readFileSync(path.replace(".jsonl", "-audit.json"), "utf8"));
    const hash = crypto.createHash("sha256");
    const stream = fs.createReadStream(path);
    stream.on("data", b => hash.update(b));
    const groups = new Map();
    let worlds = 0;
    for await (const line of readline.createInterface({ input: stream })) {
      const world = JSON.parse(line);
      if (!world.Population) continue;
      const p = world.Population;
      worlds++;
      const key = `${p.Geometry}/${p.Regime}`;
      const rows = groups.get(key) ?? [];
      const anchor = world.Arms[3].Template;
      assert.deepEqual(anchor, world.Arms[7].Template);
      const oracle = floor(anchor, p.Rates.at(-1));
      const prioritized = floor(anchor, p.Rates.at(-1), true);
      rows.push({ ...oracle, priorityBrier: prioritized.brier,
        fullBrier: world.Arms[0].Snapshots.at(-1).Brier,
        adaptiveBrier: world.Arms[1].Snapshots.at(-1).Brier,
        orientationBrier: world.Arms[3].Snapshots.at(-1).Brier });
      groups.set(key, rows);
    }
    assert.equal(worlds, 384);
    assert.equal(groups.size, 24);
    assert.equal(hash.digest("hex"), audit.SHA256);
    const cells = {};
    for (const [key, rows] of groups) {
      assert.equal(rows.length, 16);
      cells[key] = {
        oracleBrier: interval(rows.map(x => x.brier)),
        oraclePriorityBrier: interval(rows.map(x => x.priorityBrier)),
        oracleReversal: interval(rows.map(x => x.reversal)),
        trueRateFloor: interval(rows.map(x => x.trueRateFloor)),
        oracleMinusFull: interval(rows.map(x => x.brier - x.fullBrier)),
        oracleMinusAdaptive: interval(rows.map(x => x.brier - x.adaptiveBrier)),
        orientationMinusOracle: interval(rows.map(x => x.orientationBrier - x.brier)),
        worldsOracleExceedsPoint20: rows.filter(x => x.brier > .20).length,
      };
    }
    summary.cohorts[split] = { rawSHA256: audit.SHA256, worlds, groups: cells };
  }
  summary.scriptSHA256 = crypto.createHash("sha256").update(fs.readFileSync("research/orientation-v38-floor.mjs")).digest("hex");
  fs.writeFileSync("docs/experiments/mmm-orientation-v38-floor.json", JSON.stringify(summary, null, 2) + "\n", { flag: "wx", mode: 0o600 });
  console.log("Post-hoc floor complete; no gate or learner changed.");
}
