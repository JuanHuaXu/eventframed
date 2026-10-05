// Preserve the actual failed test before its fixture-only correction.
import fs from 'node:fs';
import crypto from 'node:crypto';
const root='research/retention-v62-preflight';fs.mkdirSync(root,{recursive:true,mode:0o700});
const src='internal/researchwindowbank/bank_test.go',dest=root+'/failed-bank-test.go';fs.copyFileSync(src,dest,fs.constants.COPYFILE_EXCL);fs.chmodSync(dest,0o600);
fs.writeFileSync(root+'/failure.json',JSON.stringify({command:'go test ./internal/researchwindowbank ./internal/researchwindowbankref -count=1 -v -timeout=10m',exitCode:1,
 finding:'TestBankOldGradingDoesNotResurrectWindow: vacuous old grading test',stage:'development before freeze/quality experiment',
 failedSourceSHA256:crypto.createHash('sha256').update(fs.readFileSync(dest)).digest('hex'),
 rootCause:'first-ever issue has identical baseline-preserving priors and joint law in all windows; grading cannot change their relative weights',
 codeInferenceBug:false,repair:'warm all windows with different retained evidence before issuing the old forecast; keep no-resurrection and non-vacuity requirements'},null,2)+'\n',{flag:'wx',mode:0o600});
