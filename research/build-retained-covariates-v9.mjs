import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const [mode, outputDir] = process.argv.slice(2);
assert.ok(['biased', 'latent'].includes(mode) && outputDir);
const originalPath = path.resolve('internal/observationlearners/retained.go');
const original = fs.readFileSync(originalPath, 'utf8');
const originalSHA256 = 'ecbc832eb2ea3d71957fe0ecd0da226325c60c5b4e247cc7ab245ddcdda86370';
assert.equal(hash(original), originalSHA256);
const oldDraw = 'uint16(rng.Intn(512))';
const sites = original.split(oldDraw);
assert.equal(sites.length, 3, 'only the online and base-fit draw sites may change');

const helper = `
// This overlay is experiment-only. The v8 implementation is not modified.
const researchCovariateMode = ${JSON.stringify(mode)}

func researchCovariateDraw(rng *rand.Rand) uint16 {
	var bits uint16
	switch researchCovariateMode {
	case "biased":
		for bit := 0; bit < 9; bit++ {
			probability := 0.5
			if bit < 3 {
				probability = 0.75
			} else if bit >= 6 {
				probability = 0.25
			}
			if rng.Float64() < probability {
				bits |= 1 << bit
			}
		}
	case "latent":
		latent := rng.Intn(2) == 1
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
	default:
		panic("undeclared research covariate mode")
	}
	return bits
}
`;
const generated = sites.join('researchCovariateDraw(rng)') + helper;
assert.equal(generated.split('researchCovariateDraw(rng)').length, 3);
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
  protocolSHA256: hash(fs.readFileSync('docs/experiments/mmm-retained-covariates-v9-protocol.md')),
  builderSHA256: hash(fs.readFileSync(new URL(import.meta.url))),
  replacedSites: 2,
};
fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2) + '\n', {flag: 'wx'});
console.log(JSON.stringify(manifest));
