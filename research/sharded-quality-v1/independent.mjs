import fs from "node:fs";
import path from "node:path";
import assert from "node:assert/strict";
import readline from "node:readline";
import {createGunzip} from "node:zlib";
import {fileURLToPath} from "node:url";

// Deliberately does not import the primary evaluator. It checks its essential
// arithmetic and phase verdicts by a second one-pass implementation.
export async function independent(root){
  const run=JSON.parse(fs.readFileSync(path.join(root,"run-results.json"))),primary=JSON.parse(fs.readFileSync(path.join(root,"evaluation.json")));
  const near=(a,b)=>assert.ok(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=1e-12,`${a} != ${b}`);
  const same=(a,b)=>{if(typeof a==="number"&&typeof b==="number")return Math.abs(a-b)<=1e-9;if(a===b)return true;if(!a||!b||typeof a!=="object"||typeof b!=="object")return false;const aa=Object.keys(a),bb=Object.keys(b);return aa.length===bb.length&&aa.every(k=>Object.hasOwn(b,k)&&same(a[k],b[k]));};
  let checkedRows=0,checkedCells=0,checkedPhasePairs=0;
  for(const command of run.commands.filter(c=>c.full)){
    const file=path.join(root,command.transcript),input=fs.existsSync(file)?fs.createReadStream(file):fs.createReadStream(file+".gz").pipe(createGunzip());
    const accum=new Map(),rows=new Map();
    for await(const line of readline.createInterface({input,crlfDelay:Infinity})){
      const index=line.indexOf("SHARD_QUALITY_ROW=");if(index<0)continue;const r=JSON.parse(line.slice(index+"SHARD_QUALITY_ROW=".length));
      const key=r.k+"/"+r.phase,stat=accum.get(key)??{n:0,recall:0,strict:0,regret:0,deficit:0,overlap:0,counts:0,common:0,delta:0};
      const threshold=r.exact_top[r.exact_top.length-1].exact,have=new Set(r.nomination.map(x=>x.id));
      let required=0,found=0,strict=0,ties=0,sumBest=0,sumActual=0;
      for(const x of r.exact_top){sumBest+=x.exact;if(have.has(x.id))strict++;if(x.exact-threshold>1e-6){required++;if(have.has(x.id))found++;}}
      for(const x of r.nomination){sumActual+=x.exact;if(Math.abs(x.exact-threshold)<=1e-6)ties++;}
      const n=r.nomination.length;stat.recall+=(found+Math.min(n-required,ties))/n;stat.strict+=strict/n;stat.regret+=Math.max(0,(sumBest-sumActual)/n);
      const a=r.actual,b=r.reference;
      stat.deficit+=Math.max(0,b.exact_baseline.reduce((s,v)=>s+v,0)/b.ids.length-a.exact_baseline.reduce((s,v)=>s+v,0)/a.ids.length);
      stat.overlap+=a.ids.filter(x=>b.ids.includes(x)).length/a.ids.length;stat.counts+=Number(a.ids.length!==b.ids.length);
      const byID=new Map(b.frontier.map(x=>[x.id,x.forecast]));for(const x of a.frontier){if(!byID.has(x.id))continue;stat.common++;for(const field of["base_law","pre_residual_law","corrected_law"])stat.delta=Math.max(stat.delta,Math.abs(x.forecast[field].useful-byID.get(x.id)[field].useful));}
      stat.n++;accum.set(key,stat);rows.set(`${r.phase}/${r.k}/${r.query}`,{actual:a,reference:b});checkedRows++;
    }
    for(const[key,s]of accum){const[k,phase]=key.split("/"),p=primary.cells.find(c=>c.arm===command.arm&&c.pair===command.pair&&c.k===Number(k)&&c.phase===phase);assert.ok(p);assert.equal(s.n,96);
      for(const[a,b]of[[s.recall/s.n,p.meanTieRecall],[s.strict/s.n,p.meanStrictIDRecall],[s.regret/s.n,p.meanExactScoreRegret],[s.deficit/s.n,p.meanPacketDeficit],[s.overlap/s.n,p.meanPacketIDOverlap],[s.delta,p.maxReferenceLawDelta]])near(a,b);
      assert.equal(s.counts,p.packetCountMismatch);assert.equal(s.common,p.commonReferenceLaws);checkedCells++;
    }
    for(const k of[50,200])for(const[from,to,label]of[["past","unchanged_repeat","past_repeat"],["future","future_repeat","future_repeat"],["unchanged_repeat","future","future_counterfactual"]])for(const field of["actual","reference"]){
      let packets=0,frontiers=0,common=0,full=0,maxLaw=0;
      for(let i=0;i<96;i++){
        const x=rows.get(`${from}/${k}/${i}`)[field],y=rows.get(`${to}/${k}/${i}`)[field];
        const pChanged=!same(x.ids,y.ids),fChanged=!same(x.frontier.map(x=>x.id),y.frontier.map(x=>x.id));packets+=pChanged;frontiers+=fChanged;
        let forecastsEqual=true;const yByID=new Map(y.frontier.map(x=>[x.id,x.forecast]));for(const c of x.frontier){const d=yByID.get(c.id);if(!d)continue;common++;forecastsEqual&&=same(c.forecast,d);for(const law of["base_law","pre_residual_law","corrected_law"])maxLaw=Math.max(maxLaw,Math.abs(c.forecast[law].useful-d[law].useful));}
        full+=Number(!pChanged&&!fChanged&&forecastsEqual&&same(x.confidence,y.confidence)&&same(x.answer_certainty,y.answer_certainty));
      }
      const p=primary.comparisons.find(c=>c.arm===command.arm&&c.pair===command.pair&&c.k===k&&c.label===label&&c.key===field);assert.ok(p);assert.equal(p.changedPackets,packets);assert.equal(p.changedFrontiers,frontiers);assert.equal(p.commonLaws,common);assert.equal(p.strictEqualQueries,full);near(p.maxLawDelta,maxLaw);assert.equal(p.pass,full===96);checkedPhasePairs++;
    }
  }
  assert.equal(checkedRows,3072);assert.equal(checkedCells,32);assert.equal(checkedPhasePairs,48);
  return {verified:true,checkedRows,checkedCells,checkedPhasePairs,independentArithmetic:true,rawCanonicalVectorRecomputationIndependent:false,wholeGoalValidation:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url))){const root=path.dirname(fileURLToPath(import.meta.url)),r=await independent(root);fs.writeFileSync(path.join(root,"independent-results.json"),JSON.stringify(r,null,2)+"\n");console.log(JSON.stringify(r));}
