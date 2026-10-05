// Envelope parsing is independent of the scientific checker. Retain every row;
// only the reader changes from one oversized string to incremental JSON lines.
import fs from 'node:fs';
import crypto from 'node:crypto';
import readline from 'node:readline';
import assert from 'node:assert/strict';

export async function readEnvelope(input, visit, trials = 16) {
  const stream = fs.createReadStream(input), digest = crypto.createHash('sha256');
  stream.on('data', bytes => digest.update(bytes));
  const lines = readline.createInterface({input: stream, crlfDelay: Infinity});
  let header, footer, count = 0;
  try {
    for await (const line of lines) {
      assert(line.trim(), 'blank record inside envelope');
      const row = JSON.parse(line);
      if (!header) {
        assert.equal(row.Type, 'header'); assert.equal(row.Trials, trials); header = row;
      } else if (row.Type === 'footer') {
        assert(!footer, 'duplicate footer'); assert.equal(row.Trials, trials);
        assert.equal(count, trials, 'footer before complete trials'); footer = row;
      } else {
        assert(!footer, 'record after footer'); assert(count < trials, 'extra trial');
        await visit(row, count++);
      }
    }
    assert(header && footer && count === trials, 'complete envelope');
    return {header, footer, count, rawSHA256: digest.digest('hex')};
  } finally {
    lines.close(); stream.destroy();
  }
}
