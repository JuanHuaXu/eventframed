import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import {fileURLToPath} from "node:url";
import {verify} from "./verify.mjs";
import {independent} from "./independent.mjs";

const root=path.dirname(fileURLToPath(import.meta.url)),read=n=>fs.readFileSync(path.join(root,n)),json=n=>JSON.parse(read(n));
const verification=await verify(root),separate=await independent(root);assert.deepEqual(separate,json("independent-results.json"));
const negatives=json("negative-results.json");assert.equal(negatives.verified,true);assert.equal(negatives.positivePass,true);assert.equal(negatives.historicalInputsMutated,false);assert.equal(negatives.negativeControls.length,4);assert.ok(negatives.negativeControls.every(c=>c.status===1&&c.rejected&&!c.setupError));
const units=read("evaluator-after.txt").toString();assert.match(units,/tests 9/);assert.match(units,/pass 9/);assert.match(units,/fail 0/);
const result={verification,independentArithmetic:separate,verifierNegativeControls:true,evaluatorUnitTests:9,allSevenWholeGoals:"OPEN",wholeGoalValidation:false,productionTouched:false,privateDataUsed:false};
fs.writeFileSync(path.join(root,"CHECKPOINT_VERIFICATION.json"),JSON.stringify(result,null,2)+"\n");
const omitted=new Set(json("LOCAL_ONLY.json").omittedFiles.map(x=>x.file));
const files=fs.readdirSync(root).filter(f=>f!=="SHA256SUMS"&&!omitted.has(f)&&fs.statSync(path.join(root,f)).isFile()).sort();
fs.writeFileSync(path.join(root,"SHA256SUMS"),files.map(f=>crypto.createHash("sha256").update(read(f)).digest("hex")+"  "+f).join("\n")+"\n");
console.log(JSON.stringify(result));
