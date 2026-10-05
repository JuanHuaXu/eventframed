import {readFileSync, mkdirSync, writeFileSync} from 'node:fs';
let source = readFileSync('cmd/research-task-atomic-durable/main.go', 'utf8');
function replace(old, next) {
  if (source.split(old).length !== 2) throw new Error(`ambiguous/missing anchor ${old}`);
  source = source.replace(old, next);
}
replace('"github.com/JuanHuaXu/eventframed/internal/researchadmission"', '"github.com/JuanHuaXu/eventframed/internal/researchadmission"\n"github.com/JuanHuaXu/eventframed/internal/researchbatch"');
replace('attempts, stale atomic.Int64', 'attempts, stale atomic.Int64\nqueue *researchbatch.Queue[libravdbstore.ResearchEventWrite, store.PutResult]');
replace('func (s *monitoredStore) PutBayesianJournal', `func (s *monitoredStore) Put(ctx context.Context, event model.Event, vector []float32, digest string) (store.PutResult, error) {
 if s.queue == nil { return s.Store.Put(ctx,event,vector,digest) }
 return s.queue.Put(ctx,libravdbstore.ResearchEventWrite{Event:event,Vector:vector,Digest:digest})
}

func (s *monitoredStore) PutBayesianJournal`);
replace('DatabasePath                     string', 'BatchSizes []int\nDatabasePath                     string');
replace('\tcall := func(i int, due time.Time) sample {', `	var batchSizes []int
 memory.queue, err = researchbatch.New(64,16,100*time.Millisecond,func(c context.Context, values []libravdbstore.ResearchEventWrite) ([]store.PutResult,error) {
   batchSizes=append(batchSizes,len(values))
   var results []store.PutResult
   invoke := func(c context.Context) error { var err error; results,err=db.PutResearchEventBatch(c,values); return err }
   var err error
   if lease { err=permits.Write(c,invoke) } else { err=invoke(c) }
   return results,err
 })
 if err != nil { panic(err) }
 defer memory.queue.Close()
 call := func(i int, due time.Time) sample {`);
replace(`			if lease {
				err = permits.Write(c, invoke)
			} else {
				err = invoke(c)
			}`, `			// Admission is inside the shared commit, not outside each CaptureTurn.
			err = invoke(c)`);
replace('\twriters.Wait()\n', '\twriters.Wait()\n memory.queue.Close()\n out.BatchSizes=batchSizes\n');
replace('"cmd/research-task-atomic-durable/main.go"', '"cmd/research-task-batch-durable/main.go", "internal/researchbatch/queue.go", "internal/researchbatch/queue_test.go", "internal/store/libravdbstore/research_event_batch.go", "research/public-task-pilot/BATCH_QUEUE_PROTOCOL.md", "research/public-task-pilot/build-batch-runner.mjs"');
mkdirSync('cmd/research-task-batch-durable', {recursive:true});
writeFileSync('cmd/research-task-batch-durable/main.go', source, {flag:'wx', mode:0o600});
