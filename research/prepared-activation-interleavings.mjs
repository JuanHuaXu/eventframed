import assert from 'node:assert/strict';

// Bounded state model. A prepared row is durable but invisible; the accepted
// marker, source uniqueness, and final ID allocation commit atomically.
const protocols = {
  guarded: { guardCommit: true, checkID: true },
  releaseBeforeCommit: { guardCommit: false, checkID: true },
  omitIDCheck: { guardCommit: true, checkID: false },
};

function copy(s) {
  return { ...s, accepted: { ...s.accepted }, trace: [...s.trace] };
}

function explore(protocol, rivalSource) {
  const terminals = [];
  function visit(s) {
    if (s.phase === 'done' && s.mutated && s.rivalDone) {
      terminals.push(s);
      return;
    }
    if (!s.locked && !s.mutated) {
      const n = copy(s);
      n.version++;
      n.mutated = true;
      n.trace.push('publish incompatible source version');
      visit(n);
    }
    if (!s.locked && !s.rivalDone) {
      const n = copy(s);
      n.rivalDone = true;
      if (!n.accepted[rivalSource]) {
        const id = ++n.nextID;
        n.accepted[rivalSource] = { id, bytes: 'rival', version: n.version, marker: true };
      }
      n.trace.push(`rival accepts ${rivalSource}`);
      visit(n);
    }
    if (!s.crashed && s.prepared && s.phase !== 'done') {
      const n = copy(s);
      n.crashed = true;
      n.locked = false;
      n.phase = 'done';
      n.trace.push('crash and reopen');
      visit(n);
    }
    if (s.phase === 'done') return;
    const n = copy(s);
    switch (s.phase) {
      case 'prepare':
        n.prepared = { id: s.nextID + 1, version: s.version, bytes: 'original-A' };
        n.phase = 'acquire';
        n.trace.push('durable prepare, invisible to lookup');
        break;
      case 'acquire':
        n.locked = true;
        n.phase = 'validate';
        n.trace.push('acquire publication guard');
        break;
      case 'validate':
        if (s.version !== s.prepared.version || s.accepted.A ||
            (protocol.checkID && s.nextID + 1 !== s.prepared.id)) {
          n.phase = 'release';
          n.rejected = true;
          n.trace.push('reject stale prepared record');
        } else {
          n.phase = protocol.guardCommit ? 'commit' : 'earlyRelease';
          n.trace.push('validate version, source, and ID');
        }
        break;
      case 'earlyRelease':
        n.locked = false;
        n.phase = 'commit';
        n.trace.push('release guard before commit');
        break;
      case 'commit':
        if (s.accepted.A) {
          n.rejected = true;
        } else {
          n.nextID = Math.max(n.nextID, s.prepared.id);
          n.accepted.A = { id: s.prepared.id, bytes: s.prepared.bytes,
            version: s.prepared.version, committedVersion: s.version, marker: true };
        }
        n.phase = s.locked ? 'release' : 'ack';
        n.trace.push('atomic accepted marker and source index commit');
        break;
      case 'release':
        n.locked = false;
        n.phase = s.rejected ? 'done' : 'ack';
        n.trace.push('release publication guard');
        break;
      case 'ack':
        n.acked = true;
        n.phase = 'done';
        n.trace.push('acknowledge original');
        break;
      default:
        throw new Error(`unknown phase ${s.phase}`);
    }
    visit(n);
  }
  visit({ phase: 'prepare', locked: false, version: 0, nextID: 0,
    prepared: null, accepted: {}, mutated: false, rivalDone: false,
    crashed: false, rejected: false, acked: false, trace: [] });

  function defects(s) {
    const accepted = Object.values(s.accepted);
    const ids = accepted.map(r => r.id);
    const a = s.accepted.A;
    return {
      duplicateID: new Set(ids).size !== ids.length,
      staleVersion: a?.bytes === 'original-A' && a.version !== a.committedVersion,
      falseAck: s.acked && (!a || a.bytes !== 'original-A'),
      missingMarker: !!a && !a.marker,
    };
  }
  const invalid = terminals.filter(s => Object.values(defects(s)).some(Boolean));
  const crashBeforeAcceptance = terminals.filter(s => s.crashed &&
    s.prepared && s.accepted.A?.bytes !== 'original-A');
  for (const s of crashBeforeAcceptance) {
    assert.notEqual(s.accepted.A?.bytes, 'original-A', 'orphaned prepare became visible');
  }
  for (const s of terminals.filter(s => s.crashed && s.accepted.A?.bytes === 'original-A')) {
    assert.equal(s.accepted.A.bytes, s.prepared.bytes, 'uncertain ack lost original bytes');
  }
  return { schedules: terminals.length, invalid: invalid.length,
    defects: Object.fromEntries(['duplicateID', 'staleVersion', 'falseAck', 'missingMarker']
      .map(k => [k, terminals.filter(s => defects(s)[k]).length])),
    acceptedA: terminals.filter(s => s.accepted.A?.bytes === 'original-A').length,
    rejectedA: terminals.filter(s => s.rejected).length,
    orphanedPrepares: crashBeforeAcceptance.length,
    counterexample: invalid[0]?.trace ?? null };
}

const results = Object.fromEntries(Object.entries(protocols).map(([name, protocol]) => [name, {
  differentSource: explore(protocol, 'B'),
  sameSource: explore(protocol, 'A'),
}]));
for (const arm of Object.values(results.guarded)) {
  assert.equal(arm.invalid, 0);
  assert(arm.acceptedA > 0 && arm.rejectedA > 0 && arm.orphanedPrepares > 0);
}
assert(results.releaseBeforeCommit.differentSource.invalid > 0);
assert(results.omitIDCheck.differentSource.invalid > 0);
console.log(JSON.stringify({
  scope: 'Atomic model steps: one prepared A, one incompatible publication, one rival, optional crash. No SQL or service performance proof.',
  invariant: 'Only accepted originals are visible; their source/version/ID are valid at commit, and uncertain ack replays exact bytes.',
  results,
}, null, 2));
