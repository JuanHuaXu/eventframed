import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),m=JSON.parse(fs.readFileSync(path.join(out,'manifest.json')));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'SOURCE_PINS.json')))throw Error('preserve frozen pins');
const files={};
function walk(root){for(const e of fs.readdirSync(root,{withFileTypes:true})){if(e.name==='.git')continue;const p=path.join(root,e.name);if(e.isDirectory())walk(p);else if(/\.(go|s|c|h|mod|sum)$/.test(e.name))files[p]=hash(fs.readFileSync(p));}}
for(const arm of['control','candidate'])walk(m.paths[arm]);walk(path.join(path.dirname(m.paths.control),'lexer'));
for(const f of['PROTOCOL.md','prepare.mjs','command.mjs','run.mjs','database_test.go.txt','preflight_test.go.txt','key_capacity_test.go.txt','profile_test.go.txt','ownership_test.go.txt','cold_test.go.txt'])files[path.join(out,f)]=hash(fs.readFileSync(path.join(out,f)));
for(const arm of['control','candidate'])fs.copyFileSync(path.join(m.paths[arm],'internal/index/interfaces.go'),path.join(out,'interfaces-'+arm+'.go.txt'));
fs.copyFileSync(path.join(m.paths.candidate,'internal/index/research_overlay.go'),path.join(out,'overlay.go.txt'));
fs.writeFileSync(path.join(out,'SOURCE_PINS.json'),JSON.stringify({files,go:execFileSync('go',['version']).toString().trim(),prospective:true,scope:'all copied dependency and lexer Go/assembly/C/module sources; stdlib and external module source trees not individually hashed',workers:10,wholeGoalValidation:false},null,2)+'\n');
console.log(JSON.stringify({frozen:true,files:Object.keys(files).length}));
