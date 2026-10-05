import fs from 'node:fs';
let s=fs.readFileSync('internal/researchindex/partition_serving.go','utf8');
s=s.replaceAll('PartitionServing','AsyncPartitionServing').replaceAll('PartitionLease','AsyncPartitionLease');
function change(a,b){if(s.split(a).length!==2)throw Error(`ambiguous: ${a}`);s=s.replace(a,b);}
change('type AsyncPartitionServing struct {','type AsyncPartitionServing struct {\n retirement *retirementQueue\n closeDone chan struct{}');
change('return s, nil','s.retirement=newRetirementQueue(maxRetired)\n s.closeDone=make(chan struct{})\n return s, nil');
const start=s.indexOf('func (s *AsyncPartitionServing) closeRetired('),end=s.indexOf('func (l *AsyncPartitionLease) Release()',start);
if(start<0||end<0)throw Error('cleanup anchors');
s=s.slice(0,start)+`func (s *AsyncPartitionServing) closeRetired(h *indexHandle) error {
 return s.retirement.submit(h.index.Close,func(err error) {
  s.mu.Lock();if err==nil {delete(s.retired,h)};s.mu.Unlock()
 })
}
`+s.slice(end);
const closing=s.indexOf('func (s *AsyncPartitionServing) Close() error {');if(closing<0)throw Error('close anchor');
s=s.slice(0,closing)+`// Close refuses active users, then stops new admissions, drains cleanup, and
// closes current graphs. Concurrent Close callers join the same completion.
func (s *AsyncPartitionServing) Close() error {
 s.mu.Lock()
 if s.closed {done:=s.closeDone;s.mu.Unlock();<-done;s.mu.Lock();err:=s.closeErr;s.mu.Unlock();return err}
 if s.building||s.writes>0||s.leases>0 {s.mu.Unlock();return ErrServingBusy}
 s.closed=true;s.mu.Unlock()
 err:=s.retirement.close()
 for _,h:=range s.current {err=errors.Join(err,h.index.Close())}
 s.mu.Lock()
 if len(s.retired)>0 {err=errors.Join(err,errors.New("retired resource reclamation unproven"))}
 s.closeErr=err;close(s.closeDone);s.mu.Unlock();return err
}
`;
fs.writeFileSync('internal/researchindex/partition_async_serving.go',s,{flag:'wx'});
