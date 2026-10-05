import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';

const protocol = 'docs/experiments/mmm-batch-motion-model-v1-protocol.md';

class BatchMachine {
  constructor() {
    this.backend = { runtime: 10, evidence: 10, events: new Map() };
    this.durable = { runtime: 10, evidence: 10, motion: new Map(), touches: new Map() };
    this.published = { runtime: 10, evidence: 10, motion: new Map() };
    this.guard = false;
    this.quarantined = false;
    this.pending = null;
    this.operations = { durableCommits: 0, publications: 0, cleanAborts: 0 };
  }

  capture() {
    return { runtime: this.published.runtime, evidence: this.published.evidence, policy: 1 };
  }

  begin(writes) {
    if (this.guard || this.quarantined || !Array.isArray(writes) || writes.length < 1 || writes.length > 16) {
      throw Error('invalid batch begin');
    }
    const tenant = writes[0].tenant, at = writes[0].at;
    if (!tenant || !Number.isSafeInteger(at) || at <= 0) throw Error('invalid batch identity or time');
    const seen = new Map(), accepted = [], duplicate = [];
    for (const w of writes) {
      if (w.tenant !== tenant || w.at !== at || !w.id || !w.digest) throw Error('mixed or invalid batch');
      const key = `${tenant}\u0000${w.id}`;
      const previous = seen.get(key) ?? this.backend.events.get(key)?.digest;
      if (previous !== undefined && previous !== w.digest) throw Error('changed-digest conflict');
      if (previous === undefined) accepted.push({ key, id: w.id, digest: w.digest, at });
      seen.set(key, w.digest);
      duplicate.push(previous !== undefined);
    }
    if (this.backend.runtime + accepted.length > Number.MAX_SAFE_INTEGER || this.backend.evidence + accepted.length > Number.MAX_SAFE_INTEGER) {
      throw Error('version overflow');
    }
    this.guard = true;
    this.pending = { phase: 'begun', before: this.backend.runtime, at, accepted, duplicate };
  }

  backendCommit() {
    const p = this.pending;
    if (!this.guard || p?.phase !== 'begun') throw Error('backend commit out of order');
    const events = new Map(this.backend.events);
    for (let i = 0; i < p.accepted.length; i++) {
      const event = p.accepted[i];
      events.set(event.key, { digest: event.digest, version: p.before + i + 1, at: event.at });
    }
    this.backend = { runtime: this.backend.runtime + p.accepted.length,
      evidence: this.backend.evidence + p.accepted.length, events };
    p.phase = 'backend';
  }

  durableCommit() {
    const p = this.pending;
    if (!this.guard || p?.phase !== 'backend') throw Error('durable commit out of order');
    const n = p.accepted.length;
    if (this.durable.runtime !== p.before || this.backend.runtime !== p.before + n ||
        this.backend.evidence !== this.durable.evidence + n) throw Error('invalid backend transition');
    const motion = new Map(this.durable.motion), touches = new Map(this.durable.touches);
    for (let i = 0; i < n; i++) {
      const event = p.accepted[i], version = p.before + i + 1;
      if (this.backend.events.get(event.key)?.version !== version || motion.has(version)) {
        throw Error('unbound accepted event');
      }
      motion.set(version, event.at);
      touches.set(event.key, version);
    }
    this.durable = { runtime: this.backend.runtime, evidence: this.backend.evidence, motion, touches };
    p.phase = 'durable';
    this.operations.durableCommits++;
  }

  publish() {
    const p = this.pending;
    if (!this.guard || p?.phase !== 'durable') throw Error('publication out of order');
    if (this.backend.runtime !== this.durable.runtime || this.backend.evidence !== this.durable.evidence ||
        this.published.runtime !== p.before || this.published.evidence + p.accepted.length !== this.backend.evidence) {
      this.quarantined = true;
      throw Error('publication checkpoint mismatch');
    }
    for (let i = 0; i < p.accepted.length; i++) {
      const event = p.accepted[i], version = p.before + i + 1;
      if (this.durable.motion.get(version) !== event.at || this.durable.touches.get(event.key) !== version) {
        this.quarantined = true;
        throw Error('incomplete per-event authority');
      }
    }
    this.published.runtime = this.backend.runtime;
    this.published.evidence = this.backend.evidence;
    this.published.motion = new Map(this.durable.motion);
    p.phase = 'published';
    this.operations.publications++;
  }

  acknowledge() {
    const p = this.pending;
    if (!this.guard || p?.phase !== 'published' || this.quarantined) throw Error('premature acknowledgement');
    const result = { version: this.backend.runtime, duplicate: [...p.duplicate],
      accepted: p.accepted.map(event => ({ id: event.id, version: this.backend.events.get(event.key).version })) };
    this.pending = null;
    this.guard = false;
    return result;
  }

  cleanAbort() {
    const p = this.pending;
    if (!this.guard || !p || p.phase !== 'begun' && !(p.phase === 'backend' && p.accepted.length === 0)) {
      throw Error('cannot clean-abort possibly committed mutation');
    }
    this.pending = null;
    this.guard = false;
    this.operations.cleanAborts++;
  }

  uncertainError() {
    if (!this.guard || this.pending?.phase !== 'backend') throw Error('uncertain error without backend attempt');
    this.quarantined = true;
    this.guard = false;
    this.pending = null;
  }

  protectedAsOf(captured, asOf) {
    if (this.guard || this.quarantined || !Number.isSafeInteger(asOf) || asOf <= 0 || captured.policy !== 1 ||
        this.backend.runtime !== this.durable.runtime || this.published.runtime !== this.backend.runtime ||
        this.backend.evidence !== this.durable.evidence || this.published.evidence !== this.backend.evidence ||
        captured.runtime > this.backend.runtime || captured.evidence > this.backend.evidence ||
        this.backend.runtime - captured.runtime !== this.backend.evidence - captured.evidence) return false;
    for (let v = captured.runtime + 1; v <= this.backend.runtime; v++) {
      const at = this.published.motion.get(v);
      if (at === undefined || at <= asOf) return false;
    }
    return true;
  }

  crashReopen() {
    const reopened = new BatchMachine();
    reopened.backend = { runtime: this.backend.runtime, evidence: this.backend.evidence,
      events: new Map(this.backend.events) };
    reopened.durable = { runtime: this.durable.runtime, evidence: this.durable.evidence,
      motion: new Map(this.durable.motion), touches: new Map(this.durable.touches) };
    reopened.quarantined = reopened.backend.runtime !== reopened.durable.runtime ||
      reopened.backend.evidence !== reopened.durable.evidence;
    if (!reopened.quarantined) {
      reopened.published = { runtime: reopened.durable.runtime, evidence: reopened.durable.evidence,
        motion: new Map(reopened.durable.motion) };
    }
    return reopened;
  }
}

const event = (id, digest = id, at = 20) => ({ tenant: 'tenant-a', id, digest, at });
const cases = [];
function check(name, run) {
  run();
  cases.push(name);
}
function complete(machine, writes) {
  machine.begin(writes);
  machine.backendCommit();
  if (machine.pending.accepted.length === 0) {
    const result = { version: machine.backend.runtime, duplicate: [...machine.pending.duplicate], accepted: [] };
    machine.cleanAbort();
    return result;
  }
  machine.durableCommit();
  machine.publish();
  return machine.acknowledge();
}

check('ordered versions, future-only as-of and visible backfill', () => {
  const m = new BatchMachine(), captured = m.capture();
  const result = complete(m, [event('a'), event('b'), event('c')]);
  assert.deepEqual(result.accepted.map(x => x.version), [11, 12, 13]);
  assert(m.protectedAsOf(captured, 19));
  assert(!m.protectedAsOf(captured, 20));
  assert(!m.protectedAsOf({ ...captured, policy: 2 }, 19));
  const newer = m.capture();
  complete(m, [event('backfill', 'backfill', 5)]);
  assert(!m.protectedAsOf(newer, 19));
});

check('mixed duplicate and new input preserves accepted order', () => {
  const m = new BatchMachine();
  complete(m, [event('a')]);
  const result = complete(m, [event('a'), event('b'), event('b'), event('c')]);
  assert.deepEqual(result.duplicate, [true, false, true, false]);
  assert.deepEqual(result.accepted, [{ id: 'b', version: 12 }, { id: 'c', version: 13 }]);
  assert.equal(m.durable.touches.get('tenant-a\u0000a'), 11);
  assert.equal(m.durable.touches.get('tenant-a\u0000b'), 12);
  assert.equal(m.durable.touches.get('tenant-a\u0000c'), 13);
});

check('exact duplicate-only retry advances no version', () => {
  const m = new BatchMachine();
  complete(m, [event('a'), event('b')]);
  const captured = m.capture();
  const before = { ...m.operations };
  const retry = complete(m, [event('a'), event('b')]);
  assert.equal(retry.version, captured.runtime);
  assert.deepEqual(retry.duplicate, [true, true]);
  assert.deepEqual(retry.accepted, []);
  assert.equal(m.operations.cleanAborts, before.cleanAborts + 1);
  assert.equal(m.operations.durableCommits, before.durableCommits);
  assert.equal(m.operations.publications, before.publications);
  assert(m.protectedAsOf(captured, 20));
});

check('invalid inputs reject before backend motion', () => {
  const invalid = [
    [], Array.from({ length: 17 }, (_, i) => event(`e${i}`)),
    [event('a'), event('a', 'different')],
    [event('a'), { ...event('b'), tenant: 'tenant-b' }],
    [event('a'), event('b', 'b', 21)],
    [event('a', 'a', 0)], [event('a', '')],
  ];
  for (const writes of invalid) {
    const m = new BatchMachine();
    assert.throws(() => m.begin(writes));
    assert.equal(m.backend.runtime, 10);
    assert(!m.guard);
  }
  const m = new BatchMachine();
  complete(m, [event('a')]);
  assert.throws(() => m.begin([event('a', 'changed')]));
  assert.equal(m.backend.runtime, 11);
});

for (const phase of ['begun', 'backend', 'durable', 'published', 'acknowledged']) {
  check(`crash after ${phase} fails closed or reconstructs exact motion`, () => {
    const m = new BatchMachine(), captured = m.capture();
    m.begin([event('a'), event('b')]);
    assert(!m.protectedAsOf(captured, 19));
    if (phase !== 'begun') m.backendCommit();
    assert(!m.protectedAsOf(captured, 19));
    if (['durable', 'published', 'acknowledged'].includes(phase)) m.durableCommit();
    assert(!m.protectedAsOf(captured, 19));
    if (['published', 'acknowledged'].includes(phase)) m.publish();
    assert(!m.protectedAsOf(captured, 19));
    if (phase === 'acknowledged') m.acknowledge();
    const reopened = m.crashReopen();
    const safe = phase !== 'backend';
    assert.equal(reopened.protectedAsOf(captured, 19), safe);
    assert.equal(reopened.protectedAsOf(captured, 20), phase === 'begun');
    if (safe && phase !== 'begun') {
      const retry = complete(reopened, [event('a'), event('b')]);
      assert.deepEqual(retry.duplicate, [true, true]);
      assert.equal(retry.version, 12);
    }
  });
}

check('unknown backend commit cannot clean-abort or acknowledge', () => {
  const m = new BatchMachine(), captured = m.capture();
  m.begin([event('a'), event('b')]);
  m.backendCommit();
  assert.throws(() => m.cleanAbort());
  assert.throws(() => m.acknowledge());
  m.uncertainError();
  assert(!m.protectedAsOf(captured, 19));
  assert(!m.crashReopen().protectedAsOf(captured, 19));
});

check('late durable validation failure leaves no partial motion', () => {
  const m = new BatchMachine(), captured = m.capture();
  m.begin([event('a'), event('b')]);
  m.backendCommit();
  m.backend.events.set('tenant-a\u0000b', { digest: 'b', version: 99, at: 20 });
  assert.throws(() => m.durableCommit());
  assert.equal(m.durable.runtime, 10);
  assert.equal(m.durable.motion.size, 0);
  assert.equal(m.durable.touches.size, 0);
  m.uncertainError();
  assert(!m.protectedAsOf(captured, 19));
});

check('missing intermediate motion quarantines publication', () => {
  const m = new BatchMachine(), captured = m.capture();
  m.begin([event('a'), event('b')]);
  m.backendCommit();
  m.durableCommit();
  m.durable.motion.delete(11);
  assert.throws(() => m.publish());
  assert(!m.protectedAsOf(captured, 19));
  assert(!m.crashReopen().protectedAsOf(captured, 19));
});

check('final-version-only touch mapping quarantines publication', () => {
  const m = new BatchMachine(), captured = m.capture();
  m.begin([event('a'), event('b')]);
  m.backendCommit();
  m.durableCommit();
  m.durable.touches.set('tenant-a\u0000a', 12);
  assert.throws(() => m.publish());
  assert(!m.protectedAsOf(captured, 19));
});

const report = { version: 'batch-motion-model-v1', cases,
  protocolSHA256: crypto.createHash('sha256').update(fs.readFileSync(protocol)).digest('hex'),
  sourceSHA256: crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex') };
const bytes = JSON.stringify(report, null, 2) + '\n';
if (process.argv[2] === '--replay') {
  assert.equal(bytes, fs.readFileSync(process.argv[3], 'utf8'), 'model replay differs');
} else if (process.argv[2]) {
  fs.writeFileSync(process.argv[2], bytes, { flag: 'wx', mode: 0o600 });
}
console.log(`PASS: ${cases.length} finite batch-authority controls`);
