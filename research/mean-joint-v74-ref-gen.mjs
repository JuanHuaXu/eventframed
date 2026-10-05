// Independent reference fork: DP priors, full dense transitions, latent-Y sums.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const source='internal/researchdynvarianceref/reference.go',raw=fs.readFileSync(source),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
let s=raw.toString().replaceAll('researchdynvarianceref','researchmeanjointref').replaceAll('[9]','[243]').replaceAll('[198]','[5346]').replaceAll('k < 9','k < 243');
function replace(a,b){assert.equal(s.split(a).length-1,1,a);s=s.replace(a,b)}
replace('prior   [3]vector\n\tmatrix  [3][22][22]float64','mean    [27]float64\n\tprior   [27][3]vector\n\tmatrix  [27][3][22][22]float64');
replace('family  [3]float64','family  [3]float64\n\tmeans   [27]float64');
function body(a,b,newText){const i=s.indexOf(a),j=s.indexOf(b,i);assert(i>=0&&j>i,a);s=s.slice(0,i)+newText+'\n'+s.slice(j)}
body('func New(', 'func rate(', `func New(base []float64, mode, family, means string, hazard float64) (*Reference,error) {
 if len(base)<2||len(base)>200||math.IsNaN(hazard)||math.IsInf(hazard,0)||hazard<0||hazard>1||(mode!="local"&&mode!="noise"&&mode!="individual"){return nil,errors.New("mean reference constructor")}
 r:=&Reference{base:append([]float64(nil),base...),mode:mode,tasks:make([]task,len(base))}
 switch family{case "learn":r.family=[3]float64{1./3,1./3,1./3};case "baseline":r.family[0]=1;case "current":r.family[1]=1;case "free":r.family[2]=1;default:return nil,errors.New("mean reference family")}
 switch means{case "learn":r.means[0],r.means[1]=.1,.8;for k:=2;k<27;k++{r.means[k]=.004};case "baseline":r.means[0]=1;default:return nil,errors.New("mean reference means")}
 for i,b:=range base{
  if math.IsNaN(b)||math.IsInf(b,0)||b<.25||b>.925+1e-12{return nil,errors.New("mean reference baseline")}
  task:=&r.tasks[i];task.mean[0],task.mean[1]=b,(.9*b-.05)/.8;k:=2
  for _,intercept:=range []float64{.1,.3,.5,.7,.9}{for _,slope:=range []float64{-.8,-.4,0,.4,.8}{task.mean[k]=math.Max(.02,math.Min(.98,intercept+slope*float64(i)/float64(len(base)-1)));k++}}
  for theta,p:=range task.mean{for a:=0;a<3;a++{
   task.prior[theta][a][0]=1
   if a>0{
    strength,spike:=1.,.8;if a==2{strength,spike=2,0}
    urn:=vector{1};for n:=0;n<20;n++{var next vector;for z:=0;z<=n;z++{next[z+1]+=urn[z]*(strength*p+float64(z))/(strength+float64(n));next[z]+=urn[z]*(strength*(1-p)+float64(n-z))/(strength+float64(n))};urn=next}
    task.prior[theta][a][0]=spike;for z:=0;z<21;z++{task.prior[theta][a][z+1]=(1-spike)*urn[z]}
   }
   for from:=0;from<22;from++{for to,mass:=range task.prior[theta][a]{task.matrix[theta][a][from][to]=hazard*mass;if from==to{task.matrix[theta][a][from][to]+=1-hazard}}}
   for h:=0;h<3;h++{task.latest.p[9*theta+3*a+h]=task.prior[theta][a]}
  }}
 }
 return r,nil
}`);
body('func (t *task) move(', 'func (r *Reference) rebuild(', `func (t *task) move(k int,p vector) (out vector) {
 theta,a:=k/9,(k/3)%3
 for from,q:=range p{for to:=0;to<22;to++{out[to]+=q*t.matrix[theta][a][from][to]}}
 return
}`);
s=s.replaceAll('a, h := k/3, k%3','a, h := (k/3)%3, k%3').replaceAll('t.prior[a], 0.','t.prior[k/9][a], 0.').replaceAll('p := t.prior[a]','p := t.prior[k/9][a]').replaceAll('t.move(a, p)','t.move(k, p)').replaceAll('t.move(k/3, p)','t.move(k, p)').replaceAll('r.tasks[i].move(k/3, p.p[k])','r.tasks[i].move(k, p.p[k])').replaceAll('t.matrix[a][from][to]','t.matrix[k/9][a][from][to]').replaceAll('rate(r.base[i],','rate(r.tasks[i].mean[k/9],').replaceAll('out[k/3]','out[(k/3)%3]');
body('func (r *Reference) weights(', 'func (r *Reference) memberWeights(', `func (r *Reference) weights(i int,replacement *belief) [243]float64 {
 var logs,prior [243]float64
 for theta,p:=range r.means{for a,q:=range r.family{
  if p*q==0{continue};k:=9*theta+3*a
  if r.mode=="noise"{for h,w:=range hp{prior[k+h]=p*q*w}}else if r.mode=="local"{prior[k]=p*q}
  for j:=range r.tasks{b:=&r.tasks[j].latest;if j==i&&replacement!=nil{b=replacement}
   if r.mode=="noise"{for h:=0;h<3;h++{logs[k+h]+=b.log[k+h]}}else{var l,q [243]float64;for h:=0;h<3;h++{l[h],q[h]=b.log[k+h],hp[h]};_,total:=probabilities(l,q);logs[k]+=total}
  }
 }}
 w,_:=probabilities(logs,prior);return w
}`);
body('func (r *Reference) memberWeights(', 'func (r *Reference) forecast(', `func (r *Reference) memberWeights(i int,w [243]float64,replacement *belief) [243]float64 {
 if r.mode=="noise"{return w};b:=&r.tasks[i].latest;if replacement!=nil{b=replacement}
 if r.mode=="individual"{var prior [243]float64;for theta,p:=range r.means{for a,q:=range r.family{for h,u:=range hp{prior[9*theta+3*a+h]=p*q*u}}};v,_:=probabilities(b.log,prior);return v}
 var out [243]float64
 for theta:=0;theta<27;theta++{for a:=0;a<3;a++{k:=9*theta+3*a;var l,q [243]float64;for h:=0;h<3;h++{l[h],q[h]=b.log[k+h],hp[h]};v,_:=probabilities(l,q);for h:=0;h<3;h++{out[k+h]=w[k]*v[h]}}}
 return out
}`);
// Reference may skip zero-prior hyperstates without changing the joint law.
replace('a, h := (k/3)%3, k%3\n\t\tp, log', 'a, h := (k/3)%3, k%3\n\t\tif r.means[k/9]*r.family[a]==0 {continue}\n\t\tp, log');
replace('a, h := (k/3)%3, k%3\n\t\tp :=', 'a, h := (k/3)%3, k%3\n\t\tif r.means[k/9]*r.family[a]==0 {continue}\n\t\tp :=');
replace('for k, p := range t.latest.p {\n\t\t\tt.latest', 'for k, p := range t.latest.p {\n\t\t\tif r.means[k/9]*r.family[(k/3)%3]==0 {continue}\n\t\t\tt.latest');
s=s.replace('"variance reference', '"mean reference');
const out='internal/researchmeanjointref/reference.go';fs.mkdirSync('internal/researchmeanjointref',{mode:0o700});fs.writeFileSync(out,s,{flag:'wx',mode:0o600});
fs.writeFileSync('research/mean-joint-v74-ref-generation.json',JSON.stringify({source,sourceSHA256:hash(raw),out,initialOutputSHA256:hash(s),independence:'imports no candidate; independent urn DP, full dense transitions, explicit latent-Y likelihood and posterior branch concentration',newStates:243,all27MeansRetained:true},null,2)+'\n',{flag:'wx',mode:0o600});
