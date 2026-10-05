import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const [mode, outputDir] = process.argv.slice(2);
assert.ok(['uniform', 'latent'].includes(mode) && outputDir);
const originalPath = path.resolve('internal/observationlearners/retained.go');
const original = fs.readFileSync(originalPath, 'utf8');
const originalSHA256 = 'ecbc832eb2ea3d71957fe0ecd0da226325c60c5b4e247cc7ab245ddcdda86370';
assert.equal(hash(original), originalSHA256);

const changes = [
  ['uint16(rng.Intn(512))', 'researchTruthInput(rng)', 2],
  ['truth(x, t, s, rng)', 'researchTruthLabel(x, t, s, rng)', 1],
  ['truth(x, -1, s, rng)', 'researchTruthLabel(x, -1, s, rng)', 1],
  ['2026093001*1000000 + int64(j*1000+fit)',
    '(int64(2026100122+k)*1000000 + int64(j*1000+fit))', 1],
  ['int64(2026093002+k)', 'int64(2026100124+k)', 1],
];
let generated = original;
for (const [oldText, newText, expected] of changes) {
  const parts = generated.split(oldText);
  assert.equal(parts.length, expected + 1, `unexpected source sites: ${oldText}`);
  generated = parts.join(newText);
}

const helper = `
// Experiment-only outcome family; the archived v8 source is unchanged.
const researchTruthMode = ${JSON.stringify(mode)}

func researchTruthInput(rng *rand.Rand) uint16 {
	if researchTruthMode == "uniform" {
		return uint16(rng.Intn(512))
	}
	if researchTruthMode != "latent" {
		panic("undeclared research input mode")
	}
	latent := rng.Intn(2) == 1
	var bits uint16
	for bit := 0; bit < 9; bit++ {
		value := false
		switch bit {
		case 0, 1, 6, 7:
			value = latent
			if rng.Float64() < 0.15 {
				value = !value
			}
		default:
			value = rng.Intn(2) == 1
		}
		if value {
			bits |= 1 << bit
		}
	}
	return bits
}

func researchTruthBit(bits uint16, bit uint) bool {
	return bits&(1<<bit) != 0
}

func researchTruthMajority(a, b, c bool) bool {
	return (a && b) || (a && c) || (b && c)
}

func researchTruthLabel(bits uint16, t int, scenario Scenario, rng *rand.Rand) bool {
	if scenario.Name == "null" {
		return rng.Intn(2) == 1
	}
	local := t >= scenario.Change
	if scenario.Name == "gradual" {
		probability := float64(t-128) / 256
		if probability < 0 {
			probability = 0
		}
		if probability > 1 {
			probability = 1
		}
		local = rng.Float64() < probability
	}
	if scenario.Name == "recurring" {
		local = t >= 0 && (t/128)%2 == 1
	}
	y := researchTruthMajority(
		researchTruthBit(bits, 0), researchTruthBit(bits, 3), researchTruthBit(bits, 6))
	if local {
		y = researchTruthMajority(
			researchTruthBit(bits, 2), researchTruthBit(bits, 5), researchTruthBit(bits, 8))
		if scenario.Name == "interaction" {
			y = (researchTruthBit(bits, 0) && researchTruthBit(bits, 1)) ||
				(researchTruthBit(bits, 2) && researchTruthBit(bits, 3))
		}
	}
	if rng.Float64() < scenario.Noise {
		y = !y
	}
	return y
}
`;
generated += helper;
const dir = path.resolve(outputDir);
fs.mkdirSync(dir, {recursive: true});
const generatedPath = path.join(dir, 'retained.go');
const overlayPath = path.join(dir, 'overlay.json');
const manifestPath = path.join(dir, 'manifest.json');
for (const p of [generatedPath, overlayPath, manifestPath]) {
  assert.ok(!fs.existsSync(p), `output exists: ${p}`);
}
fs.writeFileSync(generatedPath, generated, {flag: 'wx'});
fs.writeFileSync(overlayPath, JSON.stringify({Replace: {[originalPath]: generatedPath}}, null, 2) + '\n', {flag: 'wx'});
const manifest = {
  mode,
  originalPath,
  originalSHA256,
  generatedPath,
  generatedSHA256: hash(generated),
  overlayPath,
  protocolSHA256: hash(fs.readFileSync('docs/experiments/mmm-retained-truth-v10-protocol.md')),
  builderSHA256: hash(fs.readFileSync(new URL(import.meta.url))),
  replacementCounts: changes.map(([oldText, , count]) => ({oldText, count})),
};
fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2) + '\n', {flag: 'wx'});
console.log(JSON.stringify(manifest));
