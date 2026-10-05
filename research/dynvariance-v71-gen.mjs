// Mechanical reuse of the consumed-world harness, not its outcomes or verdicts.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source='internal/researchdispersion/class_v69_test.go';
const raw=fs.readFileSync(source);
let s=raw.toString().replaceAll('ClassV69','DynVarianceV71').replaceAll('classV69','dynVarianceV71').replaceAll('researchclasssequenceref','researchdynvarianceref').replaceAll('researchclasssequence','researchdynvariance').replaceAll('EVENTFRAME_CLASS_V69','EVENTFRAME_DYNVARIANCE_V71');
function replace(from,to){assert.equal(s.split(from).length-1,1,from);s=s.replace(from,to)}
replace('joint.Config{Mode: configuration, Hazard: 1. / 16}', 'configDynVarianceV71(configuration)');
replace('ref.New(p.World.Base, configuration, 1./16)', 'ref.New(p.World.Base, configDynVarianceV71(configuration).Mode, configDynVarianceV71(configuration).Family, 1./16)');
const start=s.indexOf('var modesDynVarianceV71 = '),end=s.indexOf('\n',start);assert(start>0);
const modes=['full','adaptive','baseline_no_pair','current_no_pair','current_uncertainty','free_no_pair','individual_no_pair','individual_uncertainty','local_no_pair','local_uncertainty','learn_no_pair','learn_random','learn_uncertainty','learn_information','learn_falsification','learn_predictive','learn_model_class','learn_noise_class'];
s=s.slice(0,start)+'var modesDynVarianceV71 = []string{'+modes.map(x=>JSON.stringify(x)).join(', ')+'}'+s.slice(end);
replace('[]string{"hybrid_random", "hybrid_model_class", "hybrid_noise_class"}', '[]string{"learn_random", "learn_model_class", "learn_noise_class"}');
s=s.replaceAll('"hybrid_random"','"learn_random"');
const insert=s.indexOf('func chooseDynVarianceV71');assert(insert>0);
s=s.slice(0,insert)+`// All model alternatives use the same frozen issue clock and reset hazard.
func configDynVarianceV71(name string) joint.Config {
 c:=joint.Config{Mode:"noise",Family:name,Hazard:1./16}
 if name=="learn" {c.Family="learn"}
 if name=="local"||name=="individual" {c.Mode=name;c.Family="learn"}
 return c
}

`+s.slice(insert);
const output='internal/researchdispersion/dynvariance_v71_test.go';
fs.writeFileSync(output,s,{flag:'wx',mode:0o600});
fs.writeFileSync('research/dynvariance-v71-generation.json',JSON.stringify({source,sourceSHA256:hash(raw),output,initialOutputSHA256:hash(s),modes,changes:'new model/ref/config API, eighteen complete controls and policies; unchanged generators/seeds/schedules/scorer/gates'},null,2)+'\n',{flag:'wx',mode:0o600});
