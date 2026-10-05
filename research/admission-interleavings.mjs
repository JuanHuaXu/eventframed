import assert from 'node:assert/strict';
import fs from 'node:fs';

// Bounded specification model, not execution of the Go implementation.
// One incompatible publication races one admission. Commit is the durable
// visibility boundary. A recheck is not atomic with a later ledger commit.
const protocols = {
  guarded: ['acquire', 'capture', 'commit', 'release'],
  earlyRelease: ['acquire', 'capture', 'release', 'commit'],
  recheckOnly: ['acquire', 'capture', 'release', 'check', 'commit'],
  reacquireThroughCommit: ['acquire', 'capture', 'release', 'acquire', 'check', 'commit', 'release'],
};

function explore(steps) {
  const terminals = [];
  function visit(s, trace) {
    if (s.pc === steps.length && s.mutated) {
      terminals.push({ ...s, trace });
      return;
    }
    if (!s.locked && !s.mutated) {
      visit({ ...s, version: s.version + 1, mutated: true }, [...trace, 'publish incompatible version']);
    }
    if (s.pc === steps.length) return;
    const op = steps[s.pc];
    if (op === 'acquire' && s.locked) return;
    const n = { ...s, pc: s.pc + 1 };
    if (op === 'acquire') n.locked = true;
    if (op === 'release') { assert(s.locked); n.locked = false; }
    if (op === 'capture') { assert(s.locked); n.captured = s.version; }
    if (op === 'check' && s.version !== s.captured) n.rejected = true;
    if (op === 'commit' && !s.rejected) {
      n.committed = true;
      n.invalid = s.version !== s.captured;
    }
    visit(n, [...trace, op]);
  }
  visit({ pc: 0, locked: false, version: 0, captured: null, mutated: false,
    rejected: false, committed: false, invalid: false }, []);
  return {
    terminalSchedules: terminals.length,
    invalidCommits: terminals.filter(s => s.invalid).length,
    rejectedSchedules: terminals.filter(s => s.rejected).length,
    validCommits: terminals.filter(s => s.committed && !s.invalid).length,
    counterexample: terminals.find(s => s.invalid)?.trace ?? null,
  };
}

const results = Object.fromEntries(Object.entries(protocols).map(([name, steps]) => [name, explore(steps)]));
assert.equal(results.guarded.invalidCommits, 0);
assert(results.guarded.validCommits > 0);
assert(results.earlyRelease.invalidCommits > 0);
assert(results.recheckOnly.invalidCommits > 0);
assert.equal(results.reacquireThroughCommit.invalidCommits, 0);
assert(results.reacquireThroughCommit.validCommits > 0);
assert(results.reacquireThroughCommit.rejectedSchedules > 0);
const output = {
  scope: 'One incompatible mutation, one admission, atomic model steps. No crash, retry, storage or performance proof.',
  invariant: 'At durable admission visibility, captured service version remains compatible.',
  results,
};
if (process.argv[2]) fs.writeFileSync(process.argv[2], JSON.stringify(output, null, 2) + '\n', { flag: 'wx' });
console.log(JSON.stringify(output, null, 2));
