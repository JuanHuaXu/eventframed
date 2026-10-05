import {readFileSync, openSync, writeFileSync, closeSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {spawnSync} from 'node:child_process';

const output = 'research/public-task-pilot/event-batch-screen.json';
const fd = openSync(output, 'wx', 0o600);
const files = [
  'internal/store/libravdbstore/research_event_batch.go',
  'internal/store/libravdbstore/research_event_batch_test.go',
  'internal/store/libravdbstore/store.go', 'go.mod',
  'research/public-task-pilot/EVENT_BATCH_PROTOCOL.md',
];
const hashes = Object.fromEntries(files.map(f => [f, createHash('sha256').update(readFileSync(f)).digest('hex')]));
const args = ['test', './internal/store/libravdbstore', '-run', '^$', '-bench', '^BenchmarkResearchEventBatch$', '-benchmem', '-benchtime=1x', '-count=3', '-cpu=4', '-timeout=120s'];
const started = new Date().toISOString();
const result = spawnSync('go', args, {encoding: 'utf8', timeout: 150000});
writeFileSync(fd, JSON.stringify({started, ended: new Date().toISOString(), hashes, command: ['go', ...args], status: result.status, signal: result.signal, error: result.error?.message, stdout: result.stdout, stderr: result.stderr}, null, 2)+'\n');
closeSync(fd);
process.stdout.write(result.stdout ?? '');
process.stderr.write(result.stderr ?? '');
process.exitCode = result.status ?? 1;
