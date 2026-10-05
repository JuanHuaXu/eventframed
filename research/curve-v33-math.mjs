// Independent finite joint-model reconstruction, extracted into a NEW module
// so old sealed auditors remain untouched. Omission changes Beta integrals;
// simplex fitting uses edges/interior, not the Go seven-face KKT routine.
import crypto from 'node:crypto';
export const n=150,fields=['Brier','PriorityBrier','PacketUsefulness','PacketBrier','PacketBias'];
export const regimes=['independent','aligned','reversed','calibrated','curved','shifted_peak','alternating','phase_alternating','permuted_curved','baseline_matched','tree_matched','narrow_peak'];
export const check=(p,s)=>{if(!p)throw Error(s);};
export const near=(a,b,s,t=1e-10)=>check(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=t,`${s}: ${a} != ${b}`);
export const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const factorial=[0];for(let j=1;j<=201;j++)factorial[j]=factorial[j-1]+Math.log(j);
const beta=(s,f)=>factorial[s]+factorial[f]-factorial[s+f+1];
export function trees(lo=0,hi=8,depth=0,node=0){const p={nodes:Array(8).fill(0),prior:depth===3?1:.5};for(let i=lo;i<hi;i++)p.nodes[i]=node;if(depth===3)return[p];const out=[p],mid=(lo+hi)/2;for(const l of trees(lo,mid,depth+1,node*2+1))for(const r of trees(mid,hi,depth+1,node*2+2)){const q={nodes:Array(8).fill(0),prior:.5*l.prior*r.prior};for(let i=lo;i<mid;i++)q.nodes[i]=l.nodes[i];for(let i=mid;i<hi;i++)q.nodes[i]=r.nodes[i];out.push(q);}return out;}
const partitions=trees();check(partitions.length===26,'partition count');near(partitions.reduce((z,p)=>z+p.prior,0),1,'partition mass');
function norm(logs){const max=Math.max(...logs),v=logs.map(x=>Math.exp(x-max)),sum=v.reduce((z,x)=>z+x,0);return{w:v.map(x=>x/sum),z:max+Math.log(sum)};}
export function components(base){const out=[{prior:.1,p:base},{prior:.8,p:base.map(b=>(.9*b-.05)/.8)}];for(const a of [.1,.3,.5,.7,.9])for(const c of [-.8,-.4,0,.4,.8])out.push({prior:.004,p:base.map((_,i)=>Math.max(.02,Math.min(.98,a+c*i/149)))});return out;}
export function joint(base,coordinate,evidence,affine){
  const s=Array(15).fill(0),f=Array(15).fill(0);let baseLL=0;
  const affineLogs=affine.map(h=>{let v=Math.log(h.prior);for(const[i,y]of evidence)v+=Math.log(y?h.p[i]:1-h.p[i]);return v;});
  for(const[i,y]of evidence){baseLL+=Math.log(y?base[i]:1-base[i]);let node=Math.min(7,Math.floor(8*coordinate[i]))+7;while(true){(y?s:f)[node]++;if(node===0)break;node=Math.floor((node-1)/2);}}
  const partLogs=[Math.log(.99)+baseLL,...partitions.map(p=>{let v=Math.log(.01*p.prior);for(const node of new Set(p.nodes))v+=beta(s[node],f[node]);return v;})];
  const af=norm(affineLogs),pt=norm(partLogs),family=norm([Math.log(.98)+baseLL,Math.log(.01)+af.z,Math.log(.01)+pt.z]).w;
  const future=i=>{let q=[base[i],af.w.reduce((z,v,k)=>z+v*affine[k].p[i],0),pt.w[0]*base[i]];for(let k=0;k<26;k++){const node=partitions[k].nodes[Math.min(7,Math.floor(8*coordinate[i]))];q[2]+=pt.w[k+1]*(1+s[node])/(2+s[node]+f[node]);}if(evidence.has(i))q=q.map(v=>(2*v+Number(evidence.get(i)))/3);return q;};
  const loo=i=>{check(evidence.has(i),'LOO member observed');const y=evidence.get(i),removed=Number(y),a=norm(affineLogs.map((v,k)=>v-Math.log(y?affine[k].p[i]:1-affine[k].p[i]))),logs=partLogs.slice();logs[0]-=Math.log(y?base[i]:1-base[i]);for(let k=0;k<26;k++){const node=partitions[k].nodes[Math.min(7,Math.floor(8*coordinate[i]))];logs[k+1]+=beta(s[node]-removed,f[node]-(1-removed))-beta(s[node],f[node]);}const p=norm(logs);let part=p.w[0]*base[i];for(let k=0;k<26;k++){const node=partitions[k].nodes[Math.min(7,Math.floor(8*coordinate[i]))];part+=p.w[k+1]*(1+s[node]-removed)/(1+s[node]+f[node]);}return[base[i],a.w.reduce((z,v,k)=>z+v*affine[k].p[i],0),part];};
  return{family,future,loo};
}
export const gram0=()=>[[1,0,0],[0,1,0],[0,0,1]],dot=(a,b)=>a.reduce((z,v,k)=>z+v*b[k],0);
export function ridge(a,c){const options=[[1,0,0],[0,1,0],[0,0,1]],cost=w=>w.reduce((z,v,i)=>z-2*c[i]*v+w.reduce((x,u,j)=>x+v*a[i][j]*u,0),0);for(const[i,j]of [[0,1],[0,2],[1,2]]){const t=Math.max(0,Math.min(1,(c[i]-c[j]-a[i][j]+a[j][j])/(a[i][i]+a[j][j]-2*a[i][j]))),w=[0,0,0];w[i]=t;w[j]=1-t;options.push(w);}const aa=a[0][0]+a[2][2]-2*a[0][2],bb=a[0][1]-a[0][2]-a[1][2]+a[2][2],cc=a[1][1]+a[2][2]-2*a[1][2],u=c[0]-c[2]-a[0][2]+a[2][2],v=c[1]-c[2]-a[1][2]+a[2][2],det=aa*cc-bb*bb;check(det>0,'positive ridge');const x=(cc*u-bb*v)/det,y=(aa*v-bb*u)/det;if(x>=0&&y>=0&&x+y<=1)options.push([x,y,1-x-y]);options.sort((x,y)=>cost(x)-cost(y));return options[0];}
export function add(a,c,row,y){for(let i=0;i<3;i++){c[i]+=row[i]*Number(y);for(let j=0;j<3;j++)a[i][j]+=row[i]*row[j];}}
export function fit(state,evidence){const a=gram0(),c=[.98,.01,.01];for(const[i,y]of evidence)add(a,c,state.loo(i),y);return ridge(a,c);}
export function metrics(q,p){const Packet=Array.from({length:n},(_,i)=>i).sort((a,b)=>q[b]-q[a]||a-b).slice(0,10),risk=i=>(q[i]-p[i])**2+p[i]*(1-p[i]);return{Packet,Brier:q.reduce((z,_,i)=>z+risk(i)/150,0),PriorityBrier:q.reduce((z,_,i)=>z+(i<10?3:1)*risk(i)/170,0),PacketUsefulness:Packet.reduce((z,i)=>z+p[i]/10,0),PacketBrier:Packet.reduce((z,i)=>z+risk(i)/10,0),PacketBias:Packet.reduce((z,i)=>z+(q[i]-p[i])/10,0)};}
export const interval=a=>{check(a.length>=2,'interval needs trajectories');const mean=a.reduce((z,v)=>z+v,0)/a.length,se=Math.sqrt(a.reduce((z,v)=>z+(v-mean)**2,0)/(a.length-1)/a.length);return{mean,se,lower:mean-3.5*se,upper:mean+3.5*se};};
