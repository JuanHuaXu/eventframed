import fs from 'node:fs';
import {spawnSync} from 'node:child_process';
const cases=[['duplicate','distinct available stratum evidence'],['probability','nomination probability'],['issuedLaw','pre-outcome law'],['stackRow','original stack row'],['loo','integrated omitted-member row'],['emptyUnobserved','empty unobserved population'],['cost','phase work sum'],['source','source changed']];
const controls=[];
for(const[mode,expected]of cases){
  const script=`
    import fs from 'node:fs';const original=fs.readFileSync;
    fs.readFileSync=function(path,...args){const raw=original.call(this,path,...args);if(String(path)!=='docs/experiments/mmm-curve-v33-design.jsonl')return raw;
      const lines=raw.toString().trim().split('\\n'),manifest=JSON.parse(lines[0]),world=JSON.parse(lines[1]);
      switch(${JSON.stringify(mode)}){
        case'duplicate':world.Trace[1].Index=world.Trace[0].Index;break;
        case'probability':world.Trace[0].Probability=.123;break;
        case'issuedLaw':world.Trace[0].Q[5]-=.01;break;
        case'stackRow':world.Trace[0].Issued[0]-=.01;break;
        case'loo':world.Snapshots.find(s=>s.Model==='crossfit'&&s.Budget===16).LOO[0][0]-=.01;break;
        case'emptyUnobserved':world.Snapshots.find(s=>s.Model==='crossfit'&&s.Budget===150).UnobservedBrier=0;break;
        case'cost':world.Snapshots.find(s=>s.Model==='crossfit'&&s.Budget===16).Costs.AccountedNS+=1;break;
        case'source':manifest.Sources['internal/researchblend/curve_test.go']='bad';break;
      }
      lines[0]=JSON.stringify(manifest);lines[1]=JSON.stringify(world);return Buffer.from(lines.join('\\n')+'\\n');
    };
    await import('./research/curve-v33-verify.mjs');
  `;
  const run=spawnSync(process.execPath,['--input-type=module','-e',script],{encoding:'utf8'});
  if(run.error||run.status===0||!run.stderr.includes(expected))throw Error(`negative ${mode} failed: ${run.error??run.stderr}`);
  controls.push({mode,expected,rejected:true});
}
const result={study:'curve-v33',controls,immutableTapeMutation:false};fs.writeFileSync('docs/experiments/mmm-curve-v33-negative-controls.json',JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
