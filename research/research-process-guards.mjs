import path from 'node:path';import{execFileSync}from'node:child_process';
export function processes(){return execFileSync('ps',['-axo','pid,ppid,command'],{encoding:'utf8'}).split('\n')}
function parse(line){const m=line.match(/^\s*(\d+)\s+(\d+)\s+(\S+)\s*(.*)$/);return m&&{pid:Number(m[1]),parent:Number(m[2]),exe:path.basename(m[3]),args:m[4],line}}
function tests(lines,names){return lines.map(parse).filter(p=>p&&(p.exe==='go'&&p.args.startsWith('test ')||['researchswitch.test','researchdispersion.test'].includes(p.exe))&&names.some(n=>new RegExp('(?:-run|-test\\.run)(?:=|\\s+)\\^?'+n+'\\$?(?:\\s|$)').test(p.args))).map(p=>p.line)}
export function v43TimedProcesses(lines=processes()){return tests(lines,['TestSwitchStudyV43','TestSwitchFixtureV43'])}
export function v43AuditProcesses(lines=processes()){return tests(lines,['TestSwitchStudyAuditV43'])}
export function v43NormalRunners(lines=processes()){return lines.map(parse).filter(p=>p&&p.exe==='node'&&p.args==='research/switch-v43-study.mjs normal').map(p=>p.line)}
