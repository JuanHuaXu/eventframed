import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync, existsSync } from 'node:fs';

const [input, output] = process.argv.slice(2);
if (!input || !output || existsSync(output)) throw new Error('usage: node research/check-notification-repeat-v9.mjs input.jsonl new-output.json');
const hash = b => createHash('sha256').update(b).digest('hex');
const raw = readFileSync(input);
const lines = raw.toString('utf8').trimEnd().split('\n').map(line => JSON.parse(line));
if (lines.length !== 37 || lines[0].kind !== 'header' || lines[0].expected_arms !== 36) throw new Error('incomplete artifact');
for (const [name, expected] of Object.entries(lines[0].Hashes)) {
  const embedded = lines[0].Sources[name];
  if (typeof embedded !== 'string' || hash(Buffer.from(embedded)) !== expected || hash(readFileSync(name)) !== expected) {
    throw new Error(`source mismatch: ${name}`);
  }
}
const nearest = (samples, fraction) => {
  if (!samples.length) return null;
  const sorted = [...samples].sort((a, b) => a - b);
  return sorted[Math.ceil(fraction * sorted.length) - 1];
};
const arms = [];
const seen = new Set();
for (const arm of lines.slice(1)) {
  const { Mode: mode, Queue: queue, Result: r } = arm;
  if (!['off', 'notified', 'polling'].includes(mode) || ![16, 64].includes(queue) ||
      !Number.isInteger(r.Trial) || r.Trial < 0 || r.Trial >= 6 || r.Requests !== 192 ||
      r.Capacity !== (mode === 'off' ? 0 : queue) || r.Enabled !== (mode !== 'off')) throw new Error('invalid arm identity');
  const id = `${queue}/${r.Trial}/${mode}`;
  if (seen.has(id)) throw new Error(`duplicate arm ${id}`);
  seen.add(id);
  if (r.ReadNS.length !== 192 || r.WriteNS.length !== 96 || r.P99 !== nearest(r.ReadNS, .99) ||
      [...r.ReadNS, ...r.WriteNS].some(ns => !Number.isFinite(ns) || ns < 0) || r.Overlap <= 0) throw new Error(`invalid latency samples ${id}`);
  if (mode === 'off') {
    if (r.Admitted !== 0 || r.Completed !== 0 || r.Dropped !== 0 || r.AgeNS?.length) throw new Error(`invalid off state ${id}`);
  } else if (r.Admitted + r.Dropped !== 192 || r.Completed !== 50 * r.Admitted || r.AgeNS.length !== r.Admitted || r.AgeNS.some(ns => !Number.isFinite(ns) || ns < 0)) {
    throw new Error(`learning accounting ${id}`);
  }
  arms.push({ id, mode, queue, trial: r.Trial, p99NS: r.P99, ageP95NS: mode === 'off' ? null : nearest(r.AgeNS, .95), admitted: r.Admitted, dropped: r.Dropped, completed: r.Completed, failed: r.Failed, overlap: r.Overlap, errors: r.Errors?.length ?? 0 });
}
if (seen.size !== 36) throw new Error('missing arms');
const pairs = [];
for (const queue of [16, 64]) for (let trial = 0; trial < 6; trial++) {
  const off = arms.find(a => a.queue === queue && a.trial === trial && a.mode === 'off');
  for (const mode of ['notified', 'polling']) {
    const a = arms.find(x => x.queue === queue && x.trial === trial && x.mode === mode);
    const ratio = a.p99NS / off.p99NS;
    const pass = off.errors === 0 && a.errors === 0 && a.failed === 0 && a.admitted >= 154 && a.completed === 50 * a.admitted && a.ageP95NS <= 250_000_000 && ratio <= 1.10;
    pairs.push({ queue, trial, mode, admitted: a.admitted, completed: a.completed, ageP95MS: a.ageP95NS / 1e6, offP99MS: off.p99NS / 1e6, onP99MS: a.p99NS / 1e6, p99Ratio: ratio, pass });
  }
}
const result = { rawSHA256: hash(raw), protocolSHA256: hash(readFileSync('docs/experiments/mmm-notification-repeat-v9-protocol.md')), checkerSHA256: hash(readFileSync('research/check-notification-repeat-v9.mjs')), arms, pairs };
writeFileSync(output, JSON.stringify(result, null, 2) + '\n', { flag: 'wx' });
console.log(JSON.stringify({ arms: arms.length, passes: pairs.filter(p => p.pass).length, failures: pairs.filter(p => !p.pass).length, queue64NotifiedPasses: pairs.filter(p => p.queue === 64 && p.mode === 'notified' && p.pass).length }));
