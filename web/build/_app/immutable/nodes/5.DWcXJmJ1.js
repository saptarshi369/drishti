import{d as De,s as v,b as V,a as x,f as g,c as Be}from"../chunks/Cy1FvEMy.js";import{am as ze,as as a,at as r,ap as n,au as k,ar as Ce,g as e,ao as Ie,aq as $,s as u,an as A,av as U,aF as Fe,d as Le,aw as Re}from"../chunks/D3FMhXU-.js";import{s as Oe,a as Te}from"../chunks/BsQxJJNp.js";import{i as j}from"../chunks/COYe2co9.js";import{e as be,i as Ee}from"../chunks/-AORdRbl.js";import{r as Ae}from"../chunks/DHUQKWZM.js";import{s as F}from"../chunks/DDWtzUIH.js";import{b as We,a as Ne}from"../chunks/CrdmV5Wy.js";import{g as Ue}from"../chunks/D2u2xyT8.js";import{p as He}from"../chunks/BozALIez.js";import{d as qe,e as Ge}from"../chunks/BFHe-9W2.js";import{i as Je,r as Ke}from"../chunks/CKZpVgm7.js";var Qe=g(`<div role="button" tabindex="0" class="ladder-row svelte-1skdls8" style="
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
      ">why?</button></span></div>`);function Xe(X,o){ze(o,!0);function B(f){switch(f){case"skill":return"✦";case"mcp":return"⬡";case"hook":return"⚓";case"agent":return"◉";default:return"▪"}}function I(f){return f==="active"?"var(--accent)":"var(--text-faint)"}function z(f){const P="font-size:11px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;";switch(f){case"active":return P+"background:var(--green-soft);color:var(--green);";case"overridden":return P+"background:var(--amber-soft);color:var(--amber);";default:return P+"background:var(--panel-2);color:var(--text-faint);"}}function H(f){return f.charAt(0).toUpperCase()+f.slice(1)}function R(f){return f?"Active":"—"}var M=Qe(),l=a(M),p=a(l),y=a(p,!0);r(p);var m=n(p,2),O=a(m,!0);r(m),r(l);var c=n(l,2),_=a(c,!0);r(c);var h=n(c,2),q=a(h,!0);r(h);var L=n(h,2),E=a(L),G=a(E,!0);r(E);var W=n(E,2),ee=a(W,!0);r(W);var Y=n(W,2);r(L),r(M),k((f,P,te,ae,Z,J)=>{F(p,`width:16px;text-align:center;font-size:13px;color:${f??""};`),v(y,P),v(O,o.row.name),v(_,te),v(q,ae),v(G,o.row.winner_scope||"—"),F(W,Z),v(ee,J)},[()=>I(o.row.effective_status),()=>B(o.row.category),()=>R(o.row.in_user),()=>R(o.row.in_project),()=>z(o.row.effective_status),()=>H(o.row.effective_status)]),V("click",M,function(...f){var P;(P=o.onClick)==null||P.apply(this,f)}),V("keydown",M,f=>(f.key==="Enter"||f.key===" ")&&o.onClick()),V("click",Y,f=>{f.stopPropagation(),o.onWhy()}),x(X,M),Ce()}De(["click","keydown"]);var Ye=g(`<div style="
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
      "> </span> <span style="font-size:12.5px;color:var(--text);flex:1;"> </span> <span> </span></div>`),Ze=g('<div style="padding:12px 14px;font-size:12.5px;color:var(--text-faint);">No trail steps recorded.</div>'),$e=g(`<div style="border:1px solid var(--border);border-radius:10px;overflow:hidden;"><div style="
    padding:10px 14px;
    background:var(--amber-soft);
    border-bottom:1px solid var(--border-soft);
    font-size:12px;
    font-weight:600;
    color:var(--amber);
    display:flex;
    align-items:center;
    gap:7px;
  ">⤣ Override trail · why this resolved</div> <!> <!></div>`);function et(X,o){ze(o,!0);function B(l){const p="font-size:10.5px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;";switch(l){case"wins":case"found":return p+"background:var(--green-soft);color:var(--green);";case"overridden":return p+"background:var(--amber-soft);color:var(--amber);";default:return p+"background:var(--panel-2);color:var(--text-faint);"}}function I(l){return l.scope?`[${l.scope}] ${l.reason}`:l.reason}var z=$e(),H=n(a(z),2);be(H,17,()=>o.trail,l=>l.step,(l,p)=>{var y=Ye(),m=a(y),O=a(m,!0);r(m);var c=n(m,2),_=a(c,!0);r(c);var h=n(c,2),q=a(h,!0);r(h),r(y),k((L,E)=>{v(O,e(p).step),v(_,L),F(h,E),v(q,e(p).decision)},[()=>I(e(p)),()=>B(e(p).decision)]),x(l,y)});var R=n(H,2);{var M=l=>{var p=Ze();x(l,p)};j(R,l=>{o.trail.length===0&&l(M)})}r(z),x(X,z),Ce()}var Ve=g("<span> </span>"),tt=g('<div style="font-size:12.5px;color:var(--text-faint);padding:8px 0;">Loading trail…</div>'),rt=g('<div style="font-size:12.5px;color:var(--amber);padding:8px 0;"> </div>'),at=g(`<div><div style="font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-faint);margin-bottom:7px;">Source</div> <div style="
            font-family:'IBM Plex Mono',monospace;
            font-size:11.5px;
            color:var(--text-dim);
            padding:9px 12px;
            border-radius:8px;
            background:var(--bg);
            border:1px solid var(--border-soft);
            word-break:break-all;
          "> </div></div>`),ot=g(`<div style="
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
                "> </span></div>`),nt=g(`<div><div style="font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-faint);margin-bottom:7px;">Definition · read-only</div> <div style="
            border-radius:8px;
            background:var(--bg);
            border:1px solid var(--border-soft);
            overflow:hidden;
          "></div></div>`),it=g(`<div role="button" tabindex="-1" aria-label="Close detail drawer" style="position:absolute;inset:0;background:rgba(0,0,0,.32);z-index:40;"></div> <aside style="
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
        ">✕</button></div> <div style="padding:18px 20px;display:flex;flex-direction:column;gap:18px;"><div style="display:flex;gap:8px;flex-wrap:wrap;align-items:center;"><span> </span> <!> <!></div> <!> <!> <!></div></aside>`,1);function st(X,o){ze(o,!0);let B=A(null),I=A(null),z=A(!1);Ie(()=>{if(!o.row){u(B,null),u(I,null),u(z,!1);return}u(B,null),u(I,null),u(z,!0),qe(o.row.id).then(c=>{u(B,c.trail,!0),u(z,!1)}).catch(c=>{u(I,c instanceof Error?c.message:"Failed to load trail",!0),u(z,!1)})});function H(c){switch(c){case"skill":return"✦";case"mcp":return"⬡";case"hook":return"⚓";case"agent":return"◉";default:return"▪"}}function R(c){const _="font-size:11.5px;padding:3px 9px;border-radius:6px;font-weight:600;";switch(c){case"active":return _+"background:var(--green-soft);color:var(--green);";case"overridden":return _+"background:var(--amber-soft);color:var(--amber);";default:return _+"background:var(--panel-2);color:var(--text-faint);"}}function M(){return"font-size:11.5px;padding:3px 9px;border-radius:6px;background:var(--panel-2);border:1px solid var(--border);color:var(--text-dim);"}function l(c){return Object.entries(c??{}).filter(([,_])=>_!=="")}let p=U(()=>o.row?l(o.row.attrs):[]);var y=Be(),m=$(y);{var O=c=>{var _=it(),h=$(_),q=n(h,2),L=a(q),E=a(L),G=a(E),W=a(G,!0);r(G);var ee=n(G,2),Y=a(ee),f=a(Y,!0);r(Y);var P=n(Y,2),te=a(P,!0);r(P),r(ee),r(E);var ae=n(E,2);r(L);var Z=n(L,2),J=a(Z),K=a(J),oe=a(K,!0);r(K);var re=n(K,2);{var ne=i=>{var s=Ve(),w=a(s,!0);r(s),k(N=>{F(s,N),v(w,o.row.winner_scope)},[()=>M()]),x(i,s)};j(re,i=>{o.row.winner_scope&&i(ne)})}var Me=n(re,2);{var ie=i=>{var s=Ve(),w=a(s);r(s),k((N,ve)=>{F(s,N),v(w,`~${ve??""} tokens`)},[()=>M(),()=>o.row.est_context_tokens.toLocaleString()]),x(i,s)};j(Me,i=>{o.row.est_context_tokens>0&&i(ie)})}r(J);var me=n(J,2);{var ye=i=>{var s=tt();x(i,s)},Pe=i=>{var s=rt(),w=a(s);r(s),k(()=>v(w,`Could not load trail: ${e(I)??""}`)),x(i,s)},se=i=>{et(i,{get trail(){return e(B)}})};j(me,i=>{e(z)?i(ye):e(I)?i(Pe,1):e(B)!==null&&i(se,2)})}var le=n(me,2);{var _e=i=>{var s=at(),w=n(a(s),2),N=a(w,!0);r(w),r(s),k(()=>v(N,o.row.winner_path)),x(i,s)};j(le,i=>{o.row.winner_path&&i(_e)})}var de=n(le,2);{var pe=i=>{var s=nt(),w=n(a(s),2);be(w,21,()=>e(p),Ee,(N,ve)=>{var he=U(()=>Fe(e(ve),2));let Se=()=>e(he)[0],we=()=>e(he)[1];var ce=ot(),xe=a(ce),t=a(xe,!0);r(xe);var d=n(xe,2),b=a(d,!0);r(d),r(ce),k(()=>{v(t,Se()),v(b,we())}),x(N,ce)}),r(w),r(s),x(i,s)};j(de,i=>{e(p).length>0&&i(pe)})}r(Z),r(q),k((i,s,w)=>{v(W,i),v(f,o.row.name),v(te,o.row.category),F(K,s),v(oe,w)},[()=>H(o.row.category),()=>R(o.row.effective_status),()=>o.row.effective_status.charAt(0).toUpperCase()+o.row.effective_status.slice(1)]),V("click",h,function(...i){var s;(s=o.onClose)==null||s.apply(this,i)}),V("keydown",h,i=>i.key==="Escape"&&o.onClose()),V("click",ae,function(...i){var s;(s=o.onClose)==null||s.apply(this,i)}),x(c,_)};j(m,c=>{o.row&&c(O)})}x(X,y),Ce()}De(["click","keydown"]);var lt=g("<span> </span>"),dt=g("<button> <!></button>"),pt=g(`<div style="
        display:grid;
        grid-template-columns:1.5fr 1fr 1fr 1.3fr;
        gap:14px;
        padding:13px 16px;
        border-bottom:1px solid var(--border-soft);
        align-items:center;
      "><span style="height:12px;border-radius:4px;background:var(--panel-2);width:60%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:40%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:40%;display:block;"></span> <span style="height:12px;border-radius:4px;background:var(--panel-2);width:50%;display:block;"></span></div>`),vt=g(`<div style="padding:28px 20px;text-align:center;"><div style="font-size:13.5px;color:var(--text-dim);margin-bottom:10px;"> </div> <button style="
          font:inherit;font-size:12.5px;
          padding:7px 16px;
          border:1px solid var(--border);
          border-radius:7px;
          background:var(--panel-2);
          color:var(--text);
          cursor:pointer;
        ">Retry</button></div>`),ct=g(` <button style="
            margin-left:8px;font:inherit;font-size:12.5px;
            padding:3px 10px;border:1px solid var(--border);
            border-radius:6px;background:var(--panel-2);
            color:var(--text-dim);cursor:pointer;
          ">Clear filter</button>`,1),xt=g(`<button style="
              margin-left:8px;font:inherit;font-size:12.5px;
              padding:3px 10px;border:1px solid var(--border);
              border-radius:6px;background:var(--panel-2);
              color:var(--text-dim);cursor:pointer;
            ">Show disabled</button>`),ft=g(" <!>",1),ut=g('<div style="padding:28px 20px;text-align:center;color:var(--text-faint);font-size:13px;"><!></div>'),gt=g(`<p style="margin:11px 4px 0;font-size:11.5px;color:var(--text-faint);">Precedence applied: <span style="font-family:'IBM Plex Mono',monospace;"> </span></p>`),bt=g(`<div style="margin-bottom:16px;"><h1 style="margin:0;font-size:21px;font-weight:600;letter-spacing:-.02em;">Harness Map</h1> <p style="margin:4px 0 0;font-size:13px;color:var(--text-faint);">What's <em style="font-style:normal;color:var(--text-dim);">active</em> — resolved across
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
"><span>Component</span> <span>User <span style="font-family:'IBM Plex Mono',monospace;text-transform:none;letter-spacing:0;">~/.claude</span></span> <span>Project <span style="font-family:'IBM Plex Mono',monospace;text-transform:none;letter-spacing:0;">.claude</span></span> <span>Effective</span></div> <div style="border:1px solid var(--border);border-radius:11px;background:var(--panel);overflow:hidden;"><!></div> <!> <!>`,1);function Bt(X,o){ze(o,!0);const B=()=>Te(Je,"$inventoryVersion",z),I=()=>Te(Ke,"$rootVersion",z),[z,H]=Oe(),R=[{id:"skill",label:"Skills"},{id:"mcp",label:"MCP"},{id:"hook",label:"Hooks"},{id:"agent",label:"Agents"},{id:"memory",label:"Memory"},{id:"command",label:"Commands"},{id:"output-style",label:"Output styles"},{id:"plugin",label:"Plugins"}],M=Ue(He).url.searchParams.get("cat");let l=A(Le(R.some(t=>t.id===M)?M:"skill")),p=A(!1),y=A(""),m=A(Le([])),O=A(!0),c=A(null),_=A(null);async function h(){u(O,!0),u(c,null);try{const t=await Ge(e(l),e(p));u(m,t.items??[],!0)}catch(t){u(c,t instanceof Error?t.message:"Failed to load inventory",!0),u(m,[],!0)}finally{u(O,!1)}}Ie(()=>{e(l),e(p),h()}),Ie(()=>{B()+I()>0&&h()});let q=U(()=>e(m).filter(t=>t.effective_status==="active").length),L=U(()=>e(m).filter(t=>t.effective_status==="overridden").length),E=U(()=>e(m).filter(t=>t.effective_status==="disabled"||t.effective_status==="shadowed").length),G=U(()=>e(y).trim()===""?e(m):e(m).filter(t=>t.name.toLowerCase().includes(e(y).toLowerCase())));function W(t){return t.id===e(l)?String(e(m).length):""}function ee(t){return t.id===e(l)?"var(--accent)":"transparent"}function Y(t){return t.id===e(l)?"var(--text)":"var(--text-dim)"}function f(t){return t.id===e(l)?"var(--accent-soft)":"var(--panel-2)"}function P(t){return t.id===e(l)?"var(--accent)":"var(--text-faint)"}function te(t){u(_,t,!0)}function ae(){u(_,null)}let Z=U(()=>(()=>{switch(e(l)){case"skill":return"enterprise > user > project · deny beats allow · same-name skill beats command";case"agent":return"enterprise > project > user";case"mcp":return"local > project > user · disabled/enabled via settings";case"hook":return"hooks from all scopes merge — every matching hook runs";case"memory":return"memory files from all scopes merge into context · claudeMdExcludes hides files";case"command":return"enterprise > user > project · a same-name skill shadows the command";case"output-style":return"one active style (the outputStyle setting) · others available but not in effect";case"plugin":return"enabled/disabled via enabledPlugins · highest scope wins";default:return""}})());var J=bt(),K=n($(J),2);be(K,21,()=>R,Ee,(t,d)=>{var b=dt(),C=a(b),T=n(C);{var fe=D=>{var S=lt(),ue=a(S,!0);r(S),k((je,ge,ke)=>{F(S,`
          font-size:10.5px;
          padding:1px 6px;
          border-radius:9px;
          background:${je??""};
          color:${ge??""};
        `),v(ue,ke)},[()=>f(e(d)),()=>P(e(d)),()=>W(e(d))]),x(D,S)},Q=U(()=>W(e(d)));j(T,D=>{e(Q)&&D(fe)})}r(b),k((D,S)=>{F(b,`
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
        color:${S??""};
      `),v(C,`${e(d).label??""} `)},[()=>ee(e(d)),()=>Y(e(d))]),V("click",b,()=>u(l,e(d).id,!0)),x(t,b)}),r(K);var oe=n(K,2),re=a(oe),ne=a(re),Me=a(ne);r(ne);var ie=n(ne,2),me=a(ie);r(ie);var ye=n(ie,2),Pe=a(ye);r(ye),r(re);var se=n(re,4),le=n(a(se));Ae(le),r(se);var _e=n(se,2),de=a(_e);Ae(de);var pe=n(de,2),i=a(pe);r(pe),Re(),r(_e),r(oe);var s=n(oe,4),w=a(s);{var N=t=>{var d=Be(),b=$(d);be(b,16,()=>[1,2,3],Ee,(C,T)=>{var fe=pt();x(C,fe)}),x(t,d)},ve=t=>{var d=vt(),b=a(d),C=a(b);r(b);var T=n(b,2);r(d),k(()=>v(C,`Could not load inventory: ${e(c)??""}`)),V("click",T,h),x(t,d)},he=t=>{var d=ut(),b=a(d);{var C=Q=>{var D=ct(),S=$(D),ue=n(S);k(()=>v(S,`No ${e(l)??""}s match "${e(y)??""}". `)),V("click",ue,()=>u(y,"")),x(Q,D)},T=U(()=>e(y).trim()),fe=Q=>{var D=ft(),S=$(D),ue=n(S);{var je=ge=>{var ke=xt();V("click",ke,()=>u(p,!0)),x(ge,ke)};j(ue,ge=>{e(p)||ge(je)})}k(()=>v(S,`No ${e(l)??""}s found. `)),x(Q,D)};j(b,Q=>{e(T)?Q(C):Q(fe,-1)})}r(d),x(t,d)},Se=t=>{var d=Be(),b=$(d);be(b,17,()=>e(G),C=>C.id,(C,T)=>{Xe(C,{get row(){return e(T)},onClick:()=>te(e(T)),onWhy:()=>te(e(T))})}),x(t,d)};j(w,t=>{e(O)?t(N):e(c)?t(ve,1):e(G).length===0?t(he,2):t(Se,-1)})}r(s);var we=n(s,2);{var ce=t=>{var d=gt(),b=n(a(d)),C=a(b,!0);r(b),r(d),k(()=>v(C,e(Z))),x(t,d)};j(we,t=>{e(Z)&&t(ce)})}var xe=n(we,2);st(xe,{get row(){return e(_)},onClose:ae}),k(()=>{v(Me,`${e(q)??""} active`),v(me,`${e(L)??""} overridden`),v(Pe,`${e(E)??""} disabled`),F(pe,`
      width:30px;height:17px;border-radius:10px;
      background:${e(p)?"var(--accent)":"var(--border)"};
      position:relative;cursor:pointer;transition:.15s;flex:none;
    `),F(i,`
        position:absolute;
        top:2px;
        left:${e(p)?"15px":"2px"};
        width:13px;height:13px;
        border-radius:50%;
        background:${e(p)?"white":"var(--text-faint)"};
        transition:.15s;
      `)}),We(le,()=>e(y),t=>u(y,t)),Ne(de,()=>e(p),t=>u(p,t)),x(X,J),Ce(),H()}De(["click"]);export{Bt as component};
