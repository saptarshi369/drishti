import{d as Ie,s as p,b as R,a as x,f as g,c as Se}from"../chunks/BSuGuYH_.js";import{am as he,ar as a,as as r,ao as n,at as w,aq as we,g as e,ay as je,ap as $,s as u,an as F,au as V,aF as Le,ad as Te,av as Ae}from"../chunks/vgqs9OYJ.js";import{s as Fe,a as Re}from"../chunks/B35gKHx8.js";import{i as E}from"../chunks/CVO-F4Pu.js";import{e as fe,i as Be}from"../chunks/gF_0R78z.js";import{r as Ee}from"../chunks/CD8kYuFA.js";import{s as O}from"../chunks/BDJf1OCU.js";import{b as Oe,a as We}from"../chunks/BEiM3_EX.js";import{c as Ne,d as Ue}from"../chunks/fFSdtn5_.js";import{i as Ve}from"../chunks/Bgykq5k7.js";var He=g(`<div role="button" tabindex="0" class="ladder-row svelte-1skdls8" style="
    width:100%;
    display:grid;
    grid-template-columns:1.5fr 1fr 1fr 1.3fr;
    gap:14px;
    align-items:center;
    padding:12px 16px;
    border-bottom:1px solid var(--border-soft);
    background:transparent;
    color:var(--text);
    cursor:pointer;
    text-align:left;
  "><span style="display:flex;align-items:center;gap:9px;min-width:0;"><span> </span> <span style="font-weight:500;font-size:13px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;"> </span></span> <span style="font-size:12px;color:var(--text-dim);"> </span> <span style="font-size:12px;color:var(--text-dim);"> </span> <span style="display:flex;align-items:center;gap:8px;min-width:0;"><span style="font-size:12.5px;color:var(--text);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;"> </span> <span> </span> <button class="why-btn svelte-1skdls8" style="
        font-size:10.5px;
        padding:2px 7px;
        border-radius:5px;
        border:1px solid var(--border);
        background:var(--panel-2);
        color:var(--text-faint);
        cursor:pointer;
        white-space:nowrap;
        font:inherit;
      ">why?</button></span></div>`);function qe(X,o){he(o,!0);function D(f){switch(f){case"skill":return"✦";case"mcp":return"⬡";case"hook":return"⚓";case"agent":return"◉";default:return"▪"}}function L(f){return f==="active"?"var(--accent)":"var(--text-faint)"}function z(f){const P="font-size:11px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;";switch(f){case"active":return P+"background:var(--green-soft);color:var(--green);";case"overridden":return P+"background:var(--amber-soft);color:var(--amber);";default:return P+"background:var(--panel-2);color:var(--text-faint);"}}function H(f){return f.charAt(0).toUpperCase()+f.slice(1)}function y(f){return f?"Active":"—"}var b=He(),d=a(b),v=a(d),S=a(v,!0);r(v);var C=n(v,2),W=a(C,!0);r(C),r(d);var c=n(d,2),k=a(c,!0);r(c);var M=n(c,2),q=a(M,!0);r(M);var T=n(M,2),j=a(T),Y=a(j,!0);r(j);var G=n(j,2),ee=a(G,!0);r(G);var Z=n(G,2);r(T),r(b),w((f,P,te,re,J,N)=>{O(v,`width:16px;text-align:center;font-size:13px;color:${f??""};`),p(S,P),p(W,o.row.name),p(k,te),p(q,re),p(Y,o.row.winner_scope||"—"),O(G,J),p(ee,N)},[()=>L(o.row.effective_status),()=>D(o.row.category),()=>y(o.row.in_user),()=>y(o.row.in_project),()=>z(o.row.effective_status),()=>H(o.row.effective_status)]),R("click",b,function(...f){var P;(P=o.onClick)==null||P.apply(this,f)}),R("keydown",b,f=>(f.key==="Enter"||f.key===" ")&&o.onClick()),R("click",Z,f=>{f.stopPropagation(),o.onWhy()}),x(X,b),we()}Ie(["click","keydown"]);var Ge=g(`<div style="
      display:flex;
      align-items:center;
      gap:11px;
      padding:10px 14px;
      border-bottom:1px solid var(--border-soft);
    "><span style="
        font-family:'IBM Plex Mono',monospace;
        font-size:11px;
        color:var(--text-faint);
        width:18px;
        text-align:center;
        flex:none;
      "> </span> <span style="font-size:12.5px;color:var(--text);flex:1;"> </span> <span> </span></div>`),Je=g('<div style="padding:12px 14px;font-size:12.5px;color:var(--text-faint);">No trail steps recorded.</div>'),Ke=g(`<div style="border:1px solid var(--border);border-radius:10px;overflow:hidden;"><div style="
    padding:10px 14px;
    background:var(--amber-soft);
    border-bottom:1px solid var(--border-soft);
    font-size:12px;
    font-weight:600;
    color:var(--amber);
    display:flex;
    align-items:center;
    gap:7px;
  ">⤣ Override trail · why this resolved</div> <!> <!></div>`);function Qe(X,o){he(o,!0);function D(d){const v="font-size:10.5px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;";switch(d){case"wins":case"found":return v+"background:var(--green-soft);color:var(--green);";case"overridden":return v+"background:var(--amber-soft);color:var(--amber);";default:return v+"background:var(--panel-2);color:var(--text-faint);"}}function L(d){return d.scope?`[${d.scope}] ${d.reason}`:d.reason}var z=Ke(),H=n(a(z),2);fe(H,17,()=>o.trail,d=>d.step,(d,v)=>{var S=Ge(),C=a(S),W=a(C,!0);r(C);var c=n(C,2),k=a(c,!0);r(c);var M=n(c,2),q=a(M,!0);r(M),r(S),w((T,j)=>{p(W,e(v).step),p(k,T),O(M,j),p(q,e(v).decision)},[()=>L(e(v)),()=>D(e(v).decision)]),x(d,S)});var y=n(H,2);{var b=d=>{var v=Je();x(d,v)};E(y,d=>{o.trail.length===0&&d(b)})}r(z),x(X,z),we()}var De=g("<span> </span>"),Xe=g('<div style="font-size:12.5px;color:var(--text-faint);padding:8px 0;">Loading trail…</div>'),Ye=g('<div style="font-size:12.5px;color:var(--amber);padding:8px 0;"> </div>'),Ze=g(`<div><div style="font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-faint);margin-bottom:7px;">Source</div> <div style="
            font-family:'IBM Plex Mono',monospace;
            font-size:11.5px;
            color:var(--text-dim);
            padding:9px 12px;
            border-radius:8px;
            background:var(--bg);
            border:1px solid var(--border-soft);
            word-break:break-all;
          "> </div></div>`),$e=g(`<div style="
                display:flex;
                gap:10px;
                padding:8px 12px;
                border-bottom:1px solid var(--border-soft);
                font-size:11.5px;
              "><span style="
                  font-family:'IBM Plex Mono',monospace;
                  color:var(--text-faint);
                  flex:none;
                  min-width:110px;
                "> </span> <span style="
                  color:var(--text-dim);
                  word-break:break-all;
                  font-family:'IBM Plex Mono',monospace;
                "> </span></div>`),et=g(`<div><div style="font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-faint);margin-bottom:7px;">Definition · read-only</div> <div style="
            border-radius:8px;
            background:var(--bg);
            border:1px solid var(--border-soft);
            overflow:hidden;
          "></div></div>`),tt=g(`<div role="button" tabindex="-1" aria-label="Close detail drawer" style="position:absolute;inset:0;background:rgba(0,0,0,.32);z-index:40;"></div> <aside style="
      position:absolute;
      top:0;right:0;bottom:0;
      width:420px;
      z-index:41;
      background:var(--panel);
      border-left:1px solid var(--border);
      box-shadow:var(--shadow);
      overflow-y:auto;
      animation:hud-drawer .22s ease;
    "><div style="
      display:flex;
      align-items:flex-start;
      justify-content:space-between;
      padding:18px 20px;
      border-bottom:1px solid var(--border-soft);
      position:sticky;
      top:0;
      background:var(--panel);
      z-index:1;
    "><div style="display:flex;align-items:center;gap:11px;"><span style="
          width:34px;height:34px;
          border-radius:9px;
          background:var(--accent-soft);
          display:flex;align-items:center;justify-content:center;
          font-size:16px;
          color:var(--accent);
        "> </span> <div><div style="font-size:15px;font-weight:600;"> </div> <div style="font-size:11.5px;color:var(--text-faint);"> </div></div></div> <button class="close-btn svelte-9ocgkg" style="
          width:28px;height:28px;
          border:1px solid var(--border);
          border-radius:7px;
          background:transparent;
          color:var(--text-faint);
          font-size:14px;
          cursor:pointer;
        ">✕</button></div> <div style="padding:18px 20px;display:flex;flex-direction:column;gap:18px;"><div style="display:flex;gap:8px;flex-wrap:wrap;align-items:center;"><span> </span> <!> <!></div> <!> <!> <!></div></aside>`,1);function rt(X,o){he(o,!0);let D=F(null),L=F(null),z=F(!1);je(()=>{if(!o.row){u(D,null),u(L,null),u(z,!1);return}u(D,null),u(L,null),u(z,!0),Ne(o.row.id).then(c=>{u(D,c.trail,!0),u(z,!1)}).catch(c=>{u(L,c instanceof Error?c.message:"Failed to load trail",!0),u(z,!1)})});function H(c){switch(c){case"skill":return"✦";case"mcp":return"⬡";case"hook":return"⚓";case"agent":return"◉";default:return"▪"}}function y(c){const k="font-size:11.5px;padding:3px 9px;border-radius:6px;font-weight:600;";switch(c){case"active":return k+"background:var(--green-soft);color:var(--green);";case"overridden":return k+"background:var(--amber-soft);color:var(--amber);";default:return k+"background:var(--panel-2);color:var(--text-faint);"}}function b(){return"font-size:11.5px;padding:3px 9px;border-radius:6px;background:var(--panel-2);border:1px solid var(--border);color:var(--text-dim);"}function d(c){return Object.entries(c??{}).filter(([,k])=>k!=="")}let v=V(()=>o.row?d(o.row.attrs):[]);var S=Se(),C=$(S);{var W=c=>{var k=tt(),M=$(k),q=n(M,2),T=a(q),j=a(T),Y=a(j),G=a(Y,!0);r(Y);var ee=n(Y,2),Z=a(ee),f=a(Z,!0);r(Z);var P=n(Z,2),te=a(P,!0);r(P),r(ee),r(j);var re=n(j,2);r(T);var J=n(T,2),N=a(J),K=a(N),oe=a(K,!0);r(K);var ue=n(K,2);{var ne=i=>{var s=De(),h=a(s,!0);r(s),w(U=>{O(s,U),p(h,o.row.winner_scope)},[()=>b()]),x(i,s)};E(ue,i=>{o.row.winner_scope&&i(ne)})}var ke=n(ue,2);{var ge=i=>{var s=De(),h=a(s);r(s),w((U,de)=>{O(s,U),p(h,`~${de??""} tokens`)},[()=>b(),()=>o.row.est_context_tokens.toLocaleString()]),x(i,s)};E(ke,i=>{o.row.est_context_tokens>0&&i(ge)})}r(N);var be=n(N,2);{var ie=i=>{var s=Xe();x(i,s)},me=i=>{var s=Ye(),h=a(s);r(s),w(()=>p(h,`Could not load trail: ${e(L)??""}`)),x(i,s)},ye=i=>{Qe(i,{get trail(){return e(D)}})};E(be,i=>{e(z)?i(ie):e(L)?i(me,1):e(D)!==null&&i(ye,2)})}var ae=n(be,2);{var se=i=>{var s=Ze(),h=n(a(s),2),U=a(h,!0);r(h),r(s),w(()=>p(U,o.row.winner_path)),x(i,s)};E(ae,i=>{o.row.winner_path&&i(se)})}var ze=n(ae,2);{var le=i=>{var s=et(),h=n(a(s),2);fe(h,21,()=>e(v),Be,(U,de)=>{var pe=V(()=>Le(e(de),2));let Ce=()=>e(pe)[0],Me=()=>e(pe)[1];var t=$e(),l=a(t),m=a(l,!0);r(l);var _=n(l,2),B=a(_,!0);r(_),r(t),w(()=>{p(m,Ce()),p(B,Me())}),x(U,t)}),r(h),r(s),x(i,s)};E(ze,i=>{e(v).length>0&&i(le)})}r(J),r(q),w((i,s,h)=>{p(G,i),p(f,o.row.name),p(te,o.row.category),O(K,s),p(oe,h)},[()=>H(o.row.category),()=>y(o.row.effective_status),()=>o.row.effective_status.charAt(0).toUpperCase()+o.row.effective_status.slice(1)]),R("click",M,function(...i){var s;(s=o.onClose)==null||s.apply(this,i)}),R("keydown",M,i=>i.key==="Escape"&&o.onClose()),R("click",re,function(...i){var s;(s=o.onClose)==null||s.apply(this,i)}),x(c,k)};E(C,c=>{o.row&&c(W)})}x(X,S),we()}Ie(["click","keydown"]);var at=g("<span> </span>"),ot=g("<button> <!></button>"),nt=g(`<div style="
        display:grid;
        grid-template-columns:1.5fr 1fr 1fr 1.3fr;
        gap:14px;
        padding:13px 16px;
        border-bottom:1px solid var(--border-soft);
        align-items:center;
      "><span style="height:12px;border-radius:4px;background:var(--panel-2);width:60%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:40%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:40%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:50%;display:block;"></span></div>`),it=g(`<div style="padding:28px 20px;text-align:center;"><div style="font-size:13.5px;color:var(--text-dim);margin-bottom:10px;"> </div> <button style="
          font:inherit;font-size:12.5px;
          padding:7px 16px;
          border:1px solid var(--border);
          border-radius:7px;
          background:var(--panel-2);
          color:var(--text);
          cursor:pointer;
        ">Retry</button></div>`),st=g(` <button style="
            margin-left:8px;font:inherit;font-size:12.5px;
            padding:3px 10px;border:1px solid var(--border);
            border-radius:6px;background:var(--panel-2);
            color:var(--text-dim);cursor:pointer;
          ">Clear filter</button>`,1),lt=g(`<button style="
              margin-left:8px;font:inherit;font-size:12.5px;
              padding:3px 10px;border:1px solid var(--border);
              border-radius:6px;background:var(--panel-2);
              color:var(--text-dim);cursor:pointer;
            ">Show disabled</button>`),dt=g(" <!>",1),pt=g('<div style="padding:28px 20px;text-align:center;color:var(--text-faint);font-size:13px;"><!></div>'),vt=g(`<p style="margin:11px 4px 0;font-size:11.5px;color:var(--text-faint);">Precedence applied: <span style="font-family:'IBM Plex Mono',monospace;"> </span></p>`),ct=g(`<div style="margin-bottom:16px;"><h1 style="margin:0;font-size:21px;font-weight:600;letter-spacing:-.02em;">Harness Map</h1> <p style="margin:4px 0 0;font-size:13px;color:var(--text-faint);">What's <em style="font-style:normal;color:var(--text-dim);">active</em> — resolved across
    user → project scope, with override trails.</p></div> <div style="display:flex;gap:3px;border-bottom:1px solid var(--border);margin-bottom:0;overflow-x:auto;"></div> <div style="display:flex;align-items:center;gap:14px;padding:13px 2px;"><span style="font-size:12.5px;color:var(--text-dim);"><span style="color:var(--green);font-weight:600;"> </span> · <span style="color:var(--amber);"> </span> · <span style="color:var(--text-faint);"> </span></span> <span style="flex:1;"></span> <div style="
    display:flex;align-items:center;gap:7px;height:30px;
    padding:0 11px;border:1px solid var(--border);
    border-radius:7px;background:var(--panel);
    color:var(--text-faint);font-size:12.5px;
  ">🔎 <input type="text" placeholder="filter…" style="
        border:none;background:transparent;
        color:var(--text);font:inherit;font-size:12.5px;
        outline:none;width:140px;
      "/></div> <label style="display:flex;align-items:center;gap:7px;font-size:12.5px;color:var(--text-dim);cursor:pointer;"><input type="checkbox" style="position:absolute;opacity:0;width:0;height:0;"/> <span><span></span></span> show disabled</label></div> <div style="
  display:grid;
  grid-template-columns:1.5fr 1fr 1fr 1.3fr;
  gap:14px;
  padding:0 16px 9px;
  font-size:11px;
  text-transform:uppercase;
  letter-spacing:.05em;
  color:var(--text-faint);
"><span>Component</span> <span>User <span style="font-family:'IBM Plex Mono',monospace;text-transform:none;letter-spacing:0;">~/.claude</span></span> <span>Project <span style="font-family:'IBM Plex Mono',monospace;text-transform:none;letter-spacing:0;">.claude</span></span> <span>Effective</span></div> <div style="border:1px solid var(--border);border-radius:11px;background:var(--panel);overflow:hidden;"><!></div> <!> <!>`,1);function kt(X,o){he(o,!0);const D=()=>Re(Ve,"$inventoryVersion",L),[L,z]=Fe(),H=[{id:"skill",label:"Skills"},{id:"mcp",label:"MCP"},{id:"hook",label:"Hooks"},{id:"agent",label:"Agents"},{id:"memory",label:"Memory"},{id:"command",label:"Commands"},{id:"output-style",label:"Output styles"},{id:"plugin",label:"Plugins"}];let y=F("skill"),b=F(!1),d=F(""),v=F(Te([])),S=F(!0),C=F(null),W=F(null);async function c(){u(S,!0),u(C,null);try{const t=await Ue(e(y),e(b));u(v,t.items??[],!0)}catch(t){u(C,t instanceof Error?t.message:"Failed to load inventory",!0),u(v,[],!0)}finally{u(S,!1)}}je(()=>{e(y),e(b),c()}),je(()=>{D()>0&&c()});let k=V(()=>e(v).filter(t=>t.effective_status==="active").length),M=V(()=>e(v).filter(t=>t.effective_status==="overridden").length),q=V(()=>e(v).filter(t=>t.effective_status==="disabled"||t.effective_status==="shadowed").length),T=V(()=>e(d).trim()===""?e(v):e(v).filter(t=>t.name.toLowerCase().includes(e(d).toLowerCase())));function j(t){return t.id===e(y)?String(e(v).length):""}function Y(t){return t.id===e(y)?"var(--accent)":"transparent"}function G(t){return t.id===e(y)?"var(--text)":"var(--text-dim)"}function ee(t){return t.id===e(y)?"var(--accent-soft)":"var(--panel-2)"}function Z(t){return t.id===e(y)?"var(--accent)":"var(--text-faint)"}function f(t){u(W,t,!0)}function P(){u(W,null)}let te=V(()=>(()=>{switch(e(y)){case"skill":return"enterprise > user > project · deny beats allow · same-name skill beats command";case"agent":return"enterprise > project > user";case"mcp":return"local > project > user · disabled/enabled via settings";case"hook":return"hooks from all scopes merge — every matching hook runs";case"memory":return"memory files from all scopes merge into context · claudeMdExcludes hides files";case"command":return"enterprise > user > project · a same-name skill shadows the command";case"output-style":return"one active style (the outputStyle setting) · others available but not in effect";case"plugin":return"enabled/disabled via enabledPlugins · highest scope wins";default:return""}})());var re=ct(),J=n($(re),2);fe(J,21,()=>H,Be,(t,l)=>{var m=ot(),_=a(m),B=n(_);{var ve=A=>{var I=at(),ce=a(I,!0);r(I),w((Pe,xe,_e)=>{O(I,`
          font-size:10.5px;
          padding:1px 6px;
          border-radius:9px;
          background:${Pe??""};
          color:${xe??""};
        `),p(ce,_e)},[()=>ee(e(l)),()=>Z(e(l)),()=>j(e(l))]),x(A,I)},Q=V(()=>j(e(l)));E(B,A=>{e(Q)&&A(ve)})}r(m),w((A,I)=>{O(m,`
        font:inherit;
        font-size:13px;
        font-weight:500;
        padding:9px 14px;
        border:none;
        background:transparent;
        cursor:pointer;
        margin-bottom:-1px;
        display:flex;
        align-items:center;
        gap:7px;
        white-space:nowrap;
        border-bottom:2px solid ${A??""};
        color:${I??""};
      `),p(_,`${e(l).label??""} `)},[()=>Y(e(l)),()=>G(e(l))]),R("click",m,()=>u(y,e(l).id,!0)),x(t,m)}),r(J);var N=n(J,2),K=a(N),oe=a(K),ue=a(oe);r(oe);var ne=n(oe,2),ke=a(ne);r(ne);var ge=n(ne,2),be=a(ge);r(ge),r(K);var ie=n(K,4),me=n(a(ie));Ee(me),r(ie);var ye=n(ie,2),ae=a(ye);Ee(ae);var se=n(ae,2),ze=a(se);r(se),Ae(),r(ye),r(N);var le=n(N,4),i=a(le);{var s=t=>{var l=Se(),m=$(l);fe(m,16,()=>[1,2,3],Be,(_,B)=>{var ve=nt();x(_,ve)}),x(t,l)},h=t=>{var l=it(),m=a(l),_=a(m);r(m);var B=n(m,2);r(l),w(()=>p(_,`Could not load inventory: ${e(C)??""}`)),R("click",B,c),x(t,l)},U=t=>{var l=pt(),m=a(l);{var _=Q=>{var A=st(),I=$(A),ce=n(I);w(()=>p(I,`No ${e(y)??""}s match "${e(d)??""}". `)),R("click",ce,()=>u(d,"")),x(Q,A)},B=V(()=>e(d).trim()),ve=Q=>{var A=dt(),I=$(A),ce=n(I);{var Pe=xe=>{var _e=lt();R("click",_e,()=>u(b,!0)),x(xe,_e)};E(ce,xe=>{e(b)||xe(Pe)})}w(()=>p(I,`No ${e(y)??""}s found. `)),x(Q,A)};E(m,Q=>{e(B)?Q(_):Q(ve,-1)})}r(l),x(t,l)},de=t=>{var l=Se(),m=$(l);fe(m,17,()=>e(T),_=>_.id,(_,B)=>{qe(_,{get row(){return e(B)},onClick:()=>f(e(B)),onWhy:()=>f(e(B))})}),x(t,l)};E(i,t=>{e(S)?t(s):e(C)?t(h,1):e(T).length===0?t(U,2):t(de,-1)})}r(le);var pe=n(le,2);{var Ce=t=>{var l=vt(),m=n(a(l)),_=a(m,!0);r(m),r(l),w(()=>p(_,e(te))),x(t,l)};E(pe,t=>{e(te)&&t(Ce)})}var Me=n(pe,2);rt(Me,{get row(){return e(W)},onClose:P}),w(()=>{p(ue,`${e(k)??""} active`),p(ke,`${e(M)??""} overridden`),p(be,`${e(q)??""} disabled`),O(se,`
      width:30px;height:17px;border-radius:10px;
      background:${e(b)?"var(--accent)":"var(--border)"};
      position:relative;cursor:pointer;transition:.15s;flex:none;
    `),O(ze,`
        position:absolute;
        top:2px;
        left:${e(b)?"15px":"2px"};
        width:13px;height:13px;
        border-radius:50%;
        background:${e(b)?"white":"var(--text-faint)"};
        transition:.15s;
      `)}),Oe(me,()=>e(d),t=>u(d,t)),We(ae,()=>e(b),t=>u(b,t)),x(X,re),we(),z()}Ie(["click"]);export{kt as component};
