import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import readline from 'node:readline';
import {execFileSync} from 'node:child_process';

const root = process.cwd(), dir = path.join(root, 'research/durable-witness-v23');
const hash = x => crypto.createHash('sha256').update(x).digest('hex');
const jsonHash = x => hash(JSON.stringify(x));
const equal = (a, b) => JSON.stringify(a) === JSON.stringify(b);
const requireThat = (ok, why) => { if (!ok) throw Error(why); };
const ns = x => {
  requireThat(typeof x === 'string' && /^\d{4}-.*Z$/.test(x), 'missing UTC timestamp');
  const m = x.match(/^(.*?)(?:\.(\d+))?Z$/);
  return BigInt(Date.parse(m[1] + 'Z')) * 1000000n + BigInt((m[2] ?? '').padEnd(9, '0'));
};
const delta = (a, b) => Number(ns(a) - ns(b));
const quantile = (a, p) => [...a].sort((x, y) => x - y)[Math.max(0, Math.ceil(a.length * p) - 1)];
const unique = (rows, n, label) => requireThat(rows.length === n && new Set(rows.map(r => r.Index)).size === n && rows.every(r => r.Index >= 0 && r.Index < n), label + ' counts/IDs');
const snapshotKeys = ['policy_version', 'contract_version', 'graph_version', 'abstraction_version', 'agency_version'];

function audit(row) {
  unique(row.Reads, 128, 'Recall'); unique(row.Writes, 128, 'write'); unique(row.Outcomes, 16, 'outcome');
  requireThat((row.Errors ?? []).length === 0, 'collector correctness errors');
  const init = row.InitialSnapshot, final = row.FinalSnapshot;
  requireThat(init.evidence_epoch === 217 && final.evidence_epoch === init.evidence_epoch + 128 && final.runtime_version === init.runtime_version + 144 && final.posterior_version === init.posterior_version + 16, 'runtime/epoch conservation');
  requireThat(snapshotKeys.every(k => final[k] === init[k]), 'unexpected semantic version motion');
  requireThat(equal(row.FinalWitness.Head, final), 'witness head');
  let head = row.FinalWitness.Base;
  let chain = jsonHash({Head:head, Hash:'', Log:null, Sources:{}, Journals:{}, Certificate:null, Base:head});
  let stateHash = chain, log = [], mutations = [];
  requireThat(equal(head, init), 'witness genesis');
  for (const [i, c] of row.Commits.entries()) {
    const tr = JSON.parse(c.Payload);
    requireThat(c.Seq === i + 1 && c.Prior === chain && c.Digest === jsonHash([chain, c.Payload]) && equal(tr.Before, head), 'commit chain gap/hash');
    const changes = tr.Mutations ?? [];
    requireThat(tr.After.runtime_version - tr.Before.runtime_version === changes.length, 'commit mutation completeness');
    let epoch = head.evidence_epoch;
    for (const [j, m] of changes.entries()) {
      requireThat(m.RuntimeVersion === head.runtime_version + j + 1, 'runtime witness order');
      epoch += m.Kind === 1 ? 1 : 0;
      requireThat(m.EvidenceEpochAfter === epoch && [1,2].includes(m.Kind), 'epoch witness order/kind');
      if (m.Kind === 2) requireThat(m.AffectedKnown && m.AffectedKeys.length === 1, 'affected-key closure');
      mutations.push(m);
    }
    requireThat(tr.After.evidence_epoch === epoch, 'transition evidence motion');
    head = tr.After; chain = c.Digest; stateHash = tr.StateHash;
    log.push(...changes);
  }
  const copiedState = {...row.FinalWitness, Hash:''};
  requireThat(equal(head, final) && chain === row.WitnessHash && chain === row.FinalWitness.Hash && stateHash === jsonHash(copiedState) && equal(log, row.FinalWitness.Log), 'durable provenance checksum');
  requireThat(mutations.filter(m => m.Kind === 1).length === 128 && mutations.filter(m => m.Kind === 2).length === 16, 'mutation counts');
  const sourceByKey = new Map(row.DurableSources.map(p => [p.posterior_key, p]));
  requireThat(sourceByKey.size === 16 && Object.keys(row.FinalWitness.Sources).length === 16, 'source cardinality');
  const byEvent = new Map();
  for (const o of row.Outcomes) {
    const p = o.Response.posterior, source = row.FinalWitness.Sources[p.posterior_key];
    requireThat(o.Error === '' && !o.Response.duplicate && o.Useful === (o.Index % 2 === 0) && o.Request.useful === o.Useful && o.Request.event_id === o.EventID && o.Request.source === 'full_stream' && o.Request.inclusion_probability === 1, 'mixed outcome provenance');
    requireThat(o.Request.available_at === o.Offer && o.Request.observed_at === o.Offer && p.updated_at === o.Offer && delta(o.Published, o.Offer) >= 0, 'outcome availability');
    requireThat(p.alpha === 1 + Number(o.Useful) && p.beta === 2 - Number(o.Useful) && p.effective_support === 1 && p.certified, 'ordinary single-trial Beta arithmetic');
    requireThat(source && equal(source.Posterior, p) && equal(sourceByKey.get(p.posterior_key), p) && source.Record.SourceID === o.Request.idempotency_key && source.Record.SourceAt === o.Offer && equal(source.Record.Base, o.Response.snapshot), 'durable source differs from commit');
    requireThat(source.Binding.Snapshot.evidence_epoch === p.evidence_epoch, 'source epoch retagged');
    byEvent.set(o.EventID, o);
  }
  const sortedWrites = [...row.Writes].sort((a,b) => a.Index-b.Index);
  for (const w of sortedWrites) {
    const epoch = w.Snapshot.evidence_epoch - init.evidence_epoch;
    requireThat(w.Error === '' && epoch >= w.Index + 1 && epoch <= w.Index + 16 && delta(w.Ack,w.Offer) >= 0 && delta(w.Available,row.Origin) === (row.Visible ? 0 : 3600e9), 'write availability/ack/batch');
  }
  const metricValues = {call:[], offer:[], view:[], write:[], outcome:[]};
  let learned = 0, crossEpoch = 0, maxDrift = 0, firstUse = new Map();
  for (const r of row.Reads) {
    requireThat(r.Error === '' && r.Journal !== '' && r.Request.as_of === r.Offer && r.Request.session_id === 'witness-session' && equal(r.Snapshot, r.PinSnapshot), 'served pin/request/journal');
    requireThat(r.Request.recall_k === 50 && r.Request.pack_k === 10 && r.Request.token_budget === 10000, 'selection settings');
    const ins = r.Snapshot.evidence_epoch - init.evidence_epoch;
    const outcomes = r.Snapshot.posterior_version - init.posterior_version;
    requireThat(ins >= 0 && ins <= 128 && outcomes >= 0 && outcomes <= 16 && r.Snapshot.runtime_version === init.runtime_version + ins + outcomes, 'pinned state conservation');
    requireThat(ins === 0 || sortedWrites.some(w => w.Snapshot.evidence_epoch === r.Snapshot.evidence_epoch), 'not a completed batch snapshot');
    const expected = [{id:'past100',angle:0}, {id:'at120',angle:0}, ...Array.from({length:198},(_,i) => ({id:`eligible-${String(i).padStart(3,'0')}`,angle:.005*(i+1)}))];
    if (row.Visible) for (let i=0; i<ins; i++) expected.push({id:`write-${String(i).padStart(3,'0')}`,angle:.00213*(i+1)});
    expected.sort((a,b) => a.angle-b.angle || a.id.localeCompare(b.id));
    const ids = r.Decisions.map(d=>d.event_id).sort();
    requireThat(ids.length === 150 && equal(ids,expected.slice(0,150).map(d=>d.id).sort()), 'exact top150 oracle/as-of');
    requireThat(new Set(r.Packed.map(p=>p.event.id)).size === r.Packed.length && r.Packed.length <= 10 && r.Packed.every(p=>ids.includes(p.event.id) && ns(p.event.available_at) <= ns(r.Offer)), 'packed nomination/as-of');
    for (const d of r.Decisions) {
      const f = d.forecast, o = byEvent.get(d.event_id);
      for (const law of [f.base_law, f.pre_residual_law, f.corrected_law, f.belief_law].filter(Boolean)) requireThat(law.useful >= 0 && law.useful <= 1 && Math.abs(law.useful+law.not_useful-1) < 1e-14, 'invalid forecast law');
      const available = o && o.Response.snapshot.runtime_version <= r.Snapshot.runtime_version && ns(o.Offer) <= ns(r.Offer);
      const sameEpoch = available && o.Response.posterior.evidence_epoch === r.Snapshot.evidence_epoch;
      const hasBelief = !!f.belief_law;
      if (hasBelief) {
        requireThat(available && r.Selection && r.Omitted && d.activated, 'future/unadmitted label served');
        const p = o.Response.posterior, target = p.alpha/(p.alpha+p.beta);
        maxDrift = Math.max(maxDrift, Math.abs(f.belief_law.useful-target));
        requireThat(Math.abs(f.belief_law.useful-target) < 1e-14, 'served posterior predictive arithmetic');
        requireThat(sameEpoch || (row.Enabled && !row.Visible), 'unsafe visible/disabled transport');
        learned++; if (!sameEpoch) crossEpoch++;
        if (!firstUse.has(o.EventID) || ns(r.End) < firstUse.get(o.EventID)) firstUse.set(o.EventID,ns(r.End));
      }
      if (row.Enabled && !row.Visible && available) requireThat(hasBelief, 'available certified future-mode source not served');
      if (row.Visible && ins > 0) requireThat(!r.Selection && !hasBelief, 'visible motion accepted');
    }
    for (const p of r.Packed) requireThat(equal(p.forecast,r.Decisions.find(d=>d.event_id===p.event.id).forecast), 'packed law differs from journal decision');
    metricValues.call.push(delta(r.End,r.Start)); metricValues.offer.push(delta(r.End,r.Offer)); metricValues.view.push(delta(r.End,r.ViewAt));
  }
  metricValues.write = row.Writes.map(w=>delta(w.Ack,w.Offer));
  metricValues.outcome = row.Outcomes.map(o=>delta(o.Published,o.Offer));
  for (const [name, vals] of Object.entries(metricValues)) {
    requireThat(vals.every(v=>v>=0), 'negative timing');
    requireThat(row.Metrics[name+'_p99_ns'] === quantile(vals,.99) && row.Metrics[name+'_max_ns'] === Math.max(...vals), 'timing recomputation');
  }
  requireThat(learned === row.Metrics.learned_decisions && crossEpoch === row.Transported, 'served-use/transport counts');
  const m = row.Metrics;
  const timing = m.call_p99_ns < 100e6 && m.offer_p99_ns < 100e6 && m.write_p99_ns < 250e6 && m.outcome_p99_ns < 100e6 && m.outcome_max_ns < 250e6 && m.view_max_ns < 250e6;
  const pass = timing && (!row.Enabled || row.Visible || learned>0 && crossEpoch>0) && (!row.Visible || crossEpoch===0);
  requireThat(row.Pass === pass, 'gate bookkeeping');
  const freshness = row.Outcomes.map(o=>({event:o.EventID,used:firstUse.has(o.EventID),first_use_ns:firstUse.has(o.EventID)?Number(firstUse.get(o.EventID)-ns(o.Offer)):null}));
  const offerGaps = rows => { const a=[...rows].sort((a,b)=>a.Index-b.Index); return a.slice(1).map((r,i)=>delta(r.Offer,a[i].Offer)); };
  return {trial:row.Trial,enabled:row.Enabled,visible:row.Visible,pass,timing,learned,cross_epoch:crossEpoch,max_predictive_error:maxDrift,metrics:m,freshness,
    offers:{read_max_gap_ns:Math.max(...offerGaps(row.Reads)),write_max_gap_ns:Math.max(...offerGaps(row.Writes)),outcome_max_gap_ns:Math.max(...offerGaps(row.Outcomes))},
    mutations:mutations.length,commits:row.Commits.length};
}

function filesBelow(dir) {
  return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?filesBelow(path.join(dir,e.name)):e.name.endsWith('.go')?[path.join(dir,e.name)]:[]);
}
function verifyFreeze() {
  const frozen=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));
  for (const [file,digest] of Object.entries(frozen.files)) requireThat(hash(fs.readFileSync(path.join(root,file)))===digest,'source changed: '+file);
  return frozen;
}
if (process.argv.includes('--freeze')) {
  fs.mkdirSync(dir,{recursive:true});
  const files=[...filesBelow(path.join(root,'internal')),path.join(root,'go.mod'),path.join(root,'go.sum'),path.join(root,'research/durable-witness-v23-audit.mjs'),path.join(root,'research/durable-witness-v23-run.mjs'),path.join(root,'docs/experiments/mmm-durable-witness-v23-protocol.md')].sort();
  const freeze={time:new Date().toISOString(),runtime:execFileSync('go',['env','GOVERSION','GOOS','GOARCH','CGO_ENABLED'],{encoding:'utf8'}).trim(),files:Object.fromEntries(files.map(file=>[path.relative(root,file),hash(fs.readFileSync(file))]))};
  fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify(freeze,null,2)+'\n',{flag:'wx'});
  console.log('frozen',files.length,'files');
} else {
  const frozen=verifyFreeze(), raw=path.join(dir,'raw.ndjson');
  const stream=fs.createReadStream(raw), lines=readline.createInterface({input:stream,crlfDelay:Infinity});
  const summaries=[]; let header=false,footer=false,first;
  for await (const line of lines) {
    if (!line) continue;
    const row=JSON.parse(line);
    if (row.Type==='header') { requireThat(!header && !summaries.length && row.Trials===8 && row.FreezeSHA256 === hash(fs.readFileSync(path.join(dir,'freeze.json'))),'header/freeze binding');header=true;continue; }
    if (row.Type==='footer') { requireThat(!footer && summaries.length===8 && row.Trials===8,'footer');footer=true;continue; }
    requireThat(header && !footer,'trial outside header/footer');
    summaries.push(audit(row)); if (!first) first=row;
  }
  requireThat(header && footer && summaries.length===8 && new Set(summaries.map(r=>[r.trial,r.enabled,r.visible].join('/'))).size===8,'complete factorial');
  const controls=[];
  for (const [name,mutate] of [
    ['missing-write',r=>r.Writes.pop()],
    ['false-nominee',r=>r.Reads[0].Decisions[0].event_id='future125'],
    ['pin-mismatch',r=>r.Reads[0].PinSnapshot.evidence_epoch++],
    ['future-availability',r=>r.Outcomes[0].Request.available_at='2099-01-01T00:00:00Z'],
    ['chain-prior',r=>r.Commits[0].Prior='corrupt'],
    ['retagged-source',r=>r.DurableSources[0].evidence_epoch++],
    ['timing-fabrication',r=>r.Metrics.offer_p99_ns=0],
    ['gate-fabrication',r=>r.Pass=!r.Pass],
  ]) {
    const copy=structuredClone(first); mutate(copy);
    let failed=false;try {audit(copy);}catch {failed=true;}
    requireThat(failed,'negative control accepted: '+name);controls.push(name);
  }
  verifyFreeze();
  const out={type:'audited-summary',time:new Date().toISOString(),source_files:Object.keys(frozen.files).length,raw_sha256:hash(fs.readFileSync(raw)),controls,trials:summaries,all_seven_goals:'OPEN'};
  const output=process.argv.find(a=>a.startsWith('--output='))?.slice(9);
  if (output) fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
  console.log(JSON.stringify(out,null,2));
}
