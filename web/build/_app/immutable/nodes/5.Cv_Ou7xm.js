import{d as Ee,s as p,b as F,a as x,f as m,c as je}from"../chunks/BAR84ptJ.js";import{am as he,as as a,at as r,ap as n,au as k,ar as we,f as e,ao as Be,aq as $,s as u,an as V,av as N,aF as Ae,ad as Ve,aw as Fe}from"../chunks/B-MoM-Cy.js";import{s as Re,a as De}from"../chunks/DdXUsnLU.js";import{i as B}from"../chunks/BrMA2RRU.js";import{e as ge,i as Ie}from"../chunks/moMkP6cn.js";import{r as Le}from"../chunks/CXZR5lLA.js";import{s as R}from"../chunks/c3QbN61_.js";import{b as Oe,a as We}from"../chunks/DIWnwjxA.js";import{d as Ne,e as Ue}from"../chunks/BFHe-9W2.js";import{i as He,r as qe}from"../chunks/CJRuSGCi.js";var Ge=m(`<div role="button" tabindex="0" class="ladder-row svelte-1skdls8" style="
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
      ">why?</button></span></div>`);function Je(X,o){he(o,!0);function I(f){switch(f){case"skill":return"✦";case"mcp":return"⬡";case"hook":return"⚓";case"agent":return"◉";default:return"▪"}}function E(f){return f==="active"?"var(--accent)":"var(--text-faint)"}function z(f){const C="font-size:11px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;";switch(f){case"active":return C+"background:var(--green-soft);color:var(--green);";case"overridden":return C+"background:var(--amber-soft);color:var(--amber);";default:return C+"background:var(--panel-2);color:var(--text-faint);"}}function U(f){return f.charAt(0).toUpperCase()+f.slice(1)}function H(f){return f?"Active":"—"}var g=Ge(),l=a(g),v=a(l),y=a(v,!0);r(v);var P=n(v,2),L=a(P,!0);r(P),r(l);var c=n(l,2),_=a(c,!0);r(c);var M=n(c,2),q=a(M,!0);r(M);var T=n(M,2),S=a(T),G=a(S,!0);r(S);var J=n(S,2),ee=a(J,!0);r(J);var Y=n(J,2);r(T),r(g),k((f,C,ae,te,Z,O)=>{R(v,`width:16px;text-align:center;font-size:13px;color:${f??""};`),p(y,C),p(L,o.row.name),p(_,ae),p(q,te),p(G,o.row.winner_scope||"—"),R(J,Z),p(ee,O)},[()=>E(o.row.effective_status),()=>I(o.row.category),()=>H(o.row.in_user),()=>H(o.row.in_project),()=>z(o.row.effective_status),()=>U(o.row.effective_status)]),F("click",g,function(...f){var C;(C=o.onClick)==null||C.apply(this,f)}),F("keydown",g,f=>(f.key==="Enter"||f.key===" ")&&o.onClick()),F("click",Y,f=>{f.stopPropagation(),o.onWhy()}),x(X,g),we()}Ee(["click","keydown"]);var Ke=m(`<div style="
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
      "> </span> <span style="font-size:12.5px;color:var(--text);flex:1;"> </span> <span> </span></div>`),Qe=m('<div style="padding:12px 14px;font-size:12.5px;color:var(--text-faint);">No trail steps recorded.</div>'),Xe=m(`<div style="border:1px solid var(--border);border-radius:10px;overflow:hidden;"><div style="
    padding:10px 14px;
    background:var(--amber-soft);
    border-bottom:1px solid var(--border-soft);
    font-size:12px;
    font-weight:600;
    color:var(--amber);
    display:flex;
    align-items:center;
    gap:7px;
  ">⤣ Override trail · why this resolved</div> <!> <!></div>`);function Ye(X,o){he(o,!0);function I(l){const v="font-size:10.5px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;";switch(l){case"wins":case"found":return v+"background:var(--green-soft);color:var(--green);";case"overridden":return v+"background:var(--amber-soft);color:var(--amber);";default:return v+"background:var(--panel-2);color:var(--text-faint);"}}function E(l){return l.scope?`[${l.scope}] ${l.reason}`:l.reason}var z=Xe(),U=n(a(z),2);ge(U,17,()=>o.trail,l=>l.step,(l,v)=>{var y=Ke(),P=a(y),L=a(P,!0);r(P);var c=n(P,2),_=a(c,!0);r(c);var M=n(c,2),q=a(M,!0);r(M),r(y),k((T,S)=>{p(L,e(v).step),p(_,T),R(M,S),p(q,e(v).decision)},[()=>E(e(v)),()=>I(e(v).decision)]),x(l,y)});var H=n(U,2);{var g=l=>{var v=Qe();x(l,v)};B(H,l=>{o.trail.length===0&&l(g)})}r(z),x(X,z),we()}var Te=m("<span> </span>"),Ze=m('<div style="font-size:12.5px;color:var(--text-faint);padding:8px 0;">Loading trail…</div>'),$e=m('<div style="font-size:12.5px;color:var(--amber);padding:8px 0;"> </div>'),et=m(`<div><div style="font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-faint);margin-bottom:7px;">Source</div> <div style="
            font-family:'IBM Plex Mono',monospace;
            font-size:11.5px;
            color:var(--text-dim);
            padding:9px 12px;
            border-radius:8px;
            background:var(--bg);
            border:1px solid var(--border-soft);
            word-break:break-all;
          "> </div></div>`),tt=m(`<div style="
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
                "> </span></div>`),rt=m(`<div><div style="font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-faint);margin-bottom:7px;">Definition · read-only</div> <div style="
            border-radius:8px;
            background:var(--bg);
            border:1px solid var(--border-soft);
            overflow:hidden;
          "></div></div>`),at=m(`<div role="button" tabindex="-1" aria-label="Close detail drawer" style="position:absolute;inset:0;background:rgba(0,0,0,.32);z-index:40;"></div> <aside style="
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
        ">✕</button></div> <div style="padding:18px 20px;display:flex;flex-direction:column;gap:18px;"><div style="display:flex;gap:8px;flex-wrap:wrap;align-items:center;"><span> </span> <!> <!></div> <!> <!> <!></div></aside>`,1);function ot(X,o){he(o,!0);let I=V(null),E=V(null),z=V(!1);Be(()=>{if(!o.row){u(I,null),u(E,null),u(z,!1);return}u(I,null),u(E,null),u(z,!0),Ne(o.row.id).then(c=>{u(I,c.trail,!0),u(z,!1)}).catch(c=>{u(E,c instanceof Error?c.message:"Failed to load trail",!0),u(z,!1)})});function U(c){switch(c){case"skill":return"✦";case"mcp":return"⬡";case"hook":return"⚓";case"agent":return"◉";default:return"▪"}}function H(c){const _="font-size:11.5px;padding:3px 9px;border-radius:6px;font-weight:600;";switch(c){case"active":return _+"background:var(--green-soft);color:var(--green);";case"overridden":return _+"background:var(--amber-soft);color:var(--amber);";default:return _+"background:var(--panel-2);color:var(--text-faint);"}}function g(){return"font-size:11.5px;padding:3px 9px;border-radius:6px;background:var(--panel-2);border:1px solid var(--border);color:var(--text-dim);"}function l(c){return Object.entries(c??{}).filter(([,_])=>_!=="")}let v=N(()=>o.row?l(o.row.attrs):[]);var y=je(),P=$(y);{var L=c=>{var _=at(),M=$(_),q=n(M,2),T=a(q),S=a(T),G=a(S),J=a(G,!0);r(G);var ee=n(G,2),Y=a(ee),f=a(Y,!0);r(Y);var C=n(Y,2),ae=a(C,!0);r(C),r(ee),r(S);var te=n(S,2);r(T);var Z=n(T,2),O=a(Z),K=a(O),oe=a(K,!0);r(K);var re=n(K,2);{var ke=i=>{var s=Te(),w=a(s,!0);r(s),k(W=>{R(s,W),p(w,o.row.winner_scope)},[()=>g()]),x(i,s)};B(re,i=>{o.row.winner_scope&&i(ke)})}var ne=n(re,2);{var ze=i=>{var s=Te(),w=a(s);r(s),k((W,ve)=>{R(s,W),p(w,`~${ve??""} tokens`)},[()=>g(),()=>o.row.est_context_tokens.toLocaleString()]),x(i,s)};B(ne,i=>{o.row.est_context_tokens>0&&i(ze)})}r(O);var ie=n(O,2);{var Ce=i=>{var s=Ze();x(i,s)},se=i=>{var s=$e(),w=a(s);r(s),k(()=>p(w,`Could not load trail: ${e(E)??""}`)),x(i,s)},be=i=>{Ye(i,{get trail(){return e(I)}})};B(ie,i=>{e(z)?i(Ce):e(E)?i(se,1):e(I)!==null&&i(be,2)})}var le=n(ie,2);{var de=i=>{var s=et(),w=n(a(s),2),W=a(w,!0);r(w),r(s),k(()=>p(W,o.row.winner_path)),x(i,s)};B(le,i=>{o.row.winner_path&&i(de)})}var pe=n(le,2);{var Me=i=>{var s=rt(),w=n(a(s),2);ge(w,21,()=>e(v),Ie,(W,ve)=>{var me=N(()=>Ae(e(ve),2));let ye=()=>e(me)[0],Pe=()=>e(me)[1];var ce=tt(),t=a(ce),d=a(t,!0);r(t);var b=n(t,2),h=a(b,!0);r(b),r(ce),k(()=>{p(d,ye()),p(h,Pe())}),x(W,ce)}),r(w),r(s),x(i,s)};B(pe,i=>{e(v).length>0&&i(Me)})}r(Z),r(q),k((i,s,w)=>{p(J,i),p(f,o.row.name),p(ae,o.row.category),R(K,s),p(oe,w)},[()=>U(o.row.category),()=>H(o.row.effective_status),()=>o.row.effective_status.charAt(0).toUpperCase()+o.row.effective_status.slice(1)]),F("click",M,function(...i){var s;(s=o.onClose)==null||s.apply(this,i)}),F("keydown",M,i=>i.key==="Escape"&&o.onClose()),F("click",te,function(...i){var s;(s=o.onClose)==null||s.apply(this,i)}),x(c,_)};B(P,c=>{o.row&&c(L)})}x(X,y),we()}Ee(["click","keydown"]);var nt=m("<span> </span>"),it=m("<button> <!></button>"),st=m(`<div style="
        display:grid;
        grid-template-columns:1.5fr 1fr 1fr 1.3fr;
        gap:14px;
        padding:13px 16px;
        border-bottom:1px solid var(--border-soft);
        align-items:center;
      "><span style="height:12px;border-radius:4px;background:var(--panel-2);width:60%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:40%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:40%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:50%;display:block;"></span></div>`),lt=m(`<div style="padding:28px 20px;text-align:center;"><div style="font-size:13.5px;color:var(--text-dim);margin-bottom:10px;"> </div> <button style="
          font:inherit;font-size:12.5px;
          padding:7px 16px;
          border:1px solid var(--border);
          border-radius:7px;
          background:var(--panel-2);
          color:var(--text);
          cursor:pointer;
        ">Retry</button></div>`),dt=m(` <button style="
            margin-left:8px;font:inherit;font-size:12.5px;
            padding:3px 10px;border:1px solid var(--border);
            border-radius:6px;background:var(--panel-2);
            color:var(--text-dim);cursor:pointer;
          ">Clear filter</button>`,1),pt=m(`<button style="
              margin-left:8px;font:inherit;font-size:12.5px;
              padding:3px 10px;border:1px solid var(--border);
              border-radius:6px;background:var(--panel-2);
              color:var(--text-dim);cursor:pointer;
            ">Show disabled</button>`),vt=m(" <!>",1),ct=m('<div style="padding:28px 20px;text-align:center;color:var(--text-faint);font-size:13px;"><!></div>'),xt=m(`<p style="margin:11px 4px 0;font-size:11.5px;color:var(--text-faint);">Precedence applied: <span style="font-family:'IBM Plex Mono',monospace;"> </span></p>`),ft=m(`<div style="margin-bottom:16px;"><h1 style="margin:0;font-size:21px;font-weight:600;letter-spacing:-.02em;">Harness Map</h1> <p style="margin:4px 0 0;font-size:13px;color:var(--text-faint);">What's <em style="font-style:normal;color:var(--text-dim);">active</em> — resolved across
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
"><span>Component</span> <span>User <span style="font-family:'IBM Plex Mono',monospace;text-transform:none;letter-spacing:0;">~/.claude</span></span> <span>Project <span style="font-family:'IBM Plex Mono',monospace;text-transform:none;letter-spacing:0;">.claude</span></span> <span>Effective</span></div> <div style="border:1px solid var(--border);border-radius:11px;background:var(--panel);overflow:hidden;"><!></div> <!> <!>`,1);function Ct(X,o){he(o,!0);const I=()=>De(He,"$inventoryVersion",z),E=()=>De(qe,"$rootVersion",z),[z,U]=Re(),H=[{id:"skill",label:"Skills"},{id:"mcp",label:"MCP"},{id:"hook",label:"Hooks"},{id:"agent",label:"Agents"},{id:"memory",label:"Memory"},{id:"command",label:"Commands"},{id:"output-style",label:"Output styles"},{id:"plugin",label:"Plugins"}];let g=V("skill"),l=V(!1),v=V(""),y=V(Ve([])),P=V(!0),L=V(null),c=V(null);async function _(){u(P,!0),u(L,null);try{const t=await Ue(e(g),e(l));u(y,t.items??[],!0)}catch(t){u(L,t instanceof Error?t.message:"Failed to load inventory",!0),u(y,[],!0)}finally{u(P,!1)}}Be(()=>{e(g),e(l),_()}),Be(()=>{I()+E()>0&&_()});let M=N(()=>e(y).filter(t=>t.effective_status==="active").length),q=N(()=>e(y).filter(t=>t.effective_status==="overridden").length),T=N(()=>e(y).filter(t=>t.effective_status==="disabled"||t.effective_status==="shadowed").length),S=N(()=>e(v).trim()===""?e(y):e(y).filter(t=>t.name.toLowerCase().includes(e(v).toLowerCase())));function G(t){return t.id===e(g)?String(e(y).length):""}function J(t){return t.id===e(g)?"var(--accent)":"transparent"}function ee(t){return t.id===e(g)?"var(--text)":"var(--text-dim)"}function Y(t){return t.id===e(g)?"var(--accent-soft)":"var(--panel-2)"}function f(t){return t.id===e(g)?"var(--accent)":"var(--text-faint)"}function C(t){u(c,t,!0)}function ae(){u(c,null)}let te=N(()=>(()=>{switch(e(g)){case"skill":return"enterprise > user > project · deny beats allow · same-name skill beats command";case"agent":return"enterprise > project > user";case"mcp":return"local > project > user · disabled/enabled via settings";case"hook":return"hooks from all scopes merge — every matching hook runs";case"memory":return"memory files from all scopes merge into context · claudeMdExcludes hides files";case"command":return"enterprise > user > project · a same-name skill shadows the command";case"output-style":return"one active style (the outputStyle setting) · others available but not in effect";case"plugin":return"enabled/disabled via enabledPlugins · highest scope wins";default:return""}})());var Z=ft(),O=n($(Z),2);ge(O,21,()=>H,Ie,(t,d)=>{var b=it(),h=a(b),A=n(h);{var xe=D=>{var j=nt(),fe=a(j,!0);r(j),k((Se,ue,_e)=>{R(j,`
          font-size:10.5px;
          padding:1px 6px;
          border-radius:9px;
          background:${Se??""};
          color:${ue??""};
        `),p(fe,_e)},[()=>Y(e(d)),()=>f(e(d)),()=>G(e(d))]),x(D,j)},Q=N(()=>G(e(d)));B(A,D=>{e(Q)&&D(xe)})}r(b),k((D,j)=>{R(b,`
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
        border-bottom:2px solid ${D??""};
        color:${j??""};
      `),p(h,`${e(d).label??""} `)},[()=>J(e(d)),()=>ee(e(d))]),F("click",b,()=>u(g,e(d).id,!0)),x(t,b)}),r(O);var K=n(O,2),oe=a(K),re=a(oe),ke=a(re);r(re);var ne=n(re,2),ze=a(ne);r(ne);var ie=n(ne,2),Ce=a(ie);r(ie),r(oe);var se=n(oe,4),be=n(a(se));Le(be),r(se);var le=n(se,2),de=a(le);Le(de);var pe=n(de,2),Me=a(pe);r(pe),Fe(),r(le),r(K);var i=n(K,4),s=a(i);{var w=t=>{var d=je(),b=$(d);ge(b,16,()=>[1,2,3],Ie,(h,A)=>{var xe=st();x(h,xe)}),x(t,d)},W=t=>{var d=lt(),b=a(d),h=a(b);r(b);var A=n(b,2);r(d),k(()=>p(h,`Could not load inventory: ${e(L)??""}`)),F("click",A,_),x(t,d)},ve=t=>{var d=ct(),b=a(d);{var h=Q=>{var D=dt(),j=$(D),fe=n(j);k(()=>p(j,`No ${e(g)??""}s match "${e(v)??""}". `)),F("click",fe,()=>u(v,"")),x(Q,D)},A=N(()=>e(v).trim()),xe=Q=>{var D=vt(),j=$(D),fe=n(j);{var Se=ue=>{var _e=pt();F("click",_e,()=>u(l,!0)),x(ue,_e)};B(fe,ue=>{e(l)||ue(Se)})}k(()=>p(j,`No ${e(g)??""}s found. `)),x(Q,D)};B(b,Q=>{e(A)?Q(h):Q(xe,-1)})}r(d),x(t,d)},me=t=>{var d=je(),b=$(d);ge(b,17,()=>e(S),h=>h.id,(h,A)=>{Je(h,{get row(){return e(A)},onClick:()=>C(e(A)),onWhy:()=>C(e(A))})}),x(t,d)};B(s,t=>{e(P)?t(w):e(L)?t(W,1):e(S).length===0?t(ve,2):t(me,-1)})}r(i);var ye=n(i,2);{var Pe=t=>{var d=xt(),b=n(a(d)),h=a(b,!0);r(b),r(d),k(()=>p(h,e(te))),x(t,d)};B(ye,t=>{e(te)&&t(Pe)})}var ce=n(ye,2);ot(ce,{get row(){return e(c)},onClose:ae}),k(()=>{p(ke,`${e(M)??""} active`),p(ze,`${e(q)??""} overridden`),p(Ce,`${e(T)??""} disabled`),R(pe,`
      width:30px;height:17px;border-radius:10px;
      background:${e(l)?"var(--accent)":"var(--border)"};
      position:relative;cursor:pointer;transition:.15s;flex:none;
    `),R(Me,`
        position:absolute;
        top:2px;
        left:${e(l)?"15px":"2px"};
        width:13px;height:13px;
        border-radius:50%;
        background:${e(l)?"white":"var(--text-faint)"};
        transition:.15s;
      `)}),Oe(be,()=>e(v),t=>u(v,t)),We(de,()=>e(l),t=>u(l,t)),x(X,Z),we(),U()}Ee(["click"]);export{Ct as component};
