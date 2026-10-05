// New full-cohort screen: all old18 controls plus12 joint-mean policy arms.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),outputs={};
let source='internal/researchdispersion/dynvarcache_v72_test.go',raw=fs.readFileSync(source),s=raw.toString().replaceAll('DynVarCacheV72','MeanJointV74').replaceAll('dynVarCacheV72','meanJointV74').replaceAll('researchdynvarcache"','researchmeanjoint"').replaceAll('researchdynvarianceref"','researchmeanjointref"').replaceAll('EVENTFRAME_DYNVARCACHE_V72','EVENTFRAME_MEANJOINT_V74');
function replace(a,b){assert.equal(s.split(a).length-1,1,a);s=s.replace(a,b)}
const a=s.indexOf('func configMeanJointV74'),b=s.indexOf('func chooseMeanJointV74',a);assert(a>0&&b>a);
s=s.slice(0,a)+`func configMeanJointV74(name string) joint.Config {
 c:=joint.Config{Mode:"noise",Family:"learn",Means:"learn",Hazard:1./16}
 switch name {case "mean":case "meanlocal":c.Mode="local";case "meanindividual":c.Mode="individual";default:c.Mode="invalid"}
 return c
}

`+s.slice(b);
replace('w := p.World\n\tif mode == "full"', 'w := p.World\n\tif !strings.HasPrefix(mode,"mean") {return runDynVarCacheV72(p,mode,schedule)}\n\tif mode == "full"');
replace('func independentMeanJointV74(p populationPairedV60, a armPairedV60) error {','func independentMeanJointV74(p populationPairedV60, a armPairedV60) error {\n if !strings.HasPrefix(a.Mode,"mean") {return independentDynVarCacheV72(p,a)}');
replace('configMeanJointV74(configuration).Family, 1./16','configMeanJointV74(configuration).Family, configMeanJointV74(configuration).Means, 1./16');
const old=JSON.parse(fs.readFileSync('research/dynvariance-v71-generation.json')).modes,newModes=['mean','meanlocal','meanindividual'].flatMap(c=>['no_pair','random','uncertainty','falsification'].map(p=>c+'_'+p)),modes=[...old,...newModes];
const i=s.indexOf('var modesMeanJointV74 ='),j=s.indexOf('\n',i);assert(i>0);s=s.slice(0,i)+'var modesMeanJointV74 = []string{'+modes.map(x=>JSON.stringify(x)).join(', ')+'}'+s.slice(j);
replace('[]string{"learn_random", "learn_model_class", "learn_noise_class"}','[]string{"mean_random", "mean_falsification", "meanlocal_uncertainty"}');
s=s.replaceAll('policy != "learn_random"','policy != "mean_random"');
replace('x.Decisions[0].Values[0] += .1','x.Decisions[0].Options[0].Uncertainty += .1');
replace('t.Fatal("class score corruption")','t.Fatal("option score corruption")');
// Only the three detailed fixture arms are independently replayed here. Full
// cohort scores are separately recomputed; no full new-arm replay is claimed.
const out='internal/researchdispersion/mean_joint_v74_test.go';fs.writeFileSync(out,s,{flag:'wx',mode:0o600});outputs[out]={source,sourceSHA256:hash(raw),initialSHA256:hash(s)};
fs.writeFileSync('research/mean-joint-v74-study-generation.json',JSON.stringify({outputs,oldModes:old,newModes,modes,newArmCount:1440,totalArmCount:3600,changes:'full30-arm consumed-cohort screen; old18 controls delegated unchanged; new27-mean common/independent noise and fully independent member hypotheses with no-pair/random/uncertainty/falsification; allAPI query modes retain unit/fulljournal/bench testing; scientific gates unchanged'},null,2)+'\n',{flag:'wx',mode:0o600});
