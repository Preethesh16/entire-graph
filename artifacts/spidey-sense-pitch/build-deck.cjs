const pptxgen = require('pptxgenjs');
const path = require('path');

const pptx = new pptxgen();
pptx.layout = 'LAYOUT_WIDE';
pptx.author = 'HardCoders';
pptx.company = 'HardCoders';
pptx.subject = 'Spidey Sense — Graph-powered engineering coordination';
pptx.title = 'Spidey Sense: Graph is evidence, not an oracle';
pptx.lang = 'en-US';
pptx.theme = {
  headFontFace: 'Aptos Display', bodyFontFace: 'Aptos', lang: 'en-US',
};
pptx.defineSlideMaster({
  title: 'SIGNAL',
  background: { color: '050706' },
  objects: [
    { line: { x: 0, y: 0.42, w: 13.333, h: 0, line: { color: '202522', width: 0.6 } } },
    { text: { text: 'SPIDEY SENSE  /  GRAPH INTELLIGENCE', options: { x: 0.55, y: 0.13, w: 4.2, h: 0.18, fontFace: 'Aptos', fontSize: 5.8, charSpacing: 2.4, color: '748079', bold: true, margin: 0 } } },
    { text: { text: 'EVIDENCE, NOT ORACLE', options: { x: 10.6, y: 0.13, w: 2.18, h: 0.18, fontFace: 'Aptos', fontSize: 5.8, charSpacing: 2.1, color: '70D6AF', bold: true, align: 'right', margin: 0 } } },
  ],
  slideNumber: { x: 12.82, y: 7.12, w: 0.22, h: 0.14, color: '626A66', fontFace: 'Aptos', fontSize: 5.5, align: 'right', margin: 0 },
});

const C = { bg: '050706', panel: '0A0D0B', ink: 'F3F5F4', muted: '8D9691', line: '29302C', green: '70D6AF', cyan: '87C7FF', amber: 'E8C273', red: 'FF5E76', violet: 'BBA0FF' };
const A = (name) => path.join(__dirname, 'assets', name);

function slide() { return pptx.addSlide('SIGNAL'); }
function tx(s, text, x, y, w, h, size = 18, color = C.ink, opts = {}) {
  s.addText(text, { x, y, w, h, fontFace: opts.fontFace || 'Aptos', fontSize: size, color, margin: 0, breakLine: false, valign: opts.valign || 'mid', bold: !!opts.bold, charSpacing: opts.charSpacing || 0, align: opts.align || 'left', fit: 'shrink', ...opts });
}
function eyebrow(s, text, x = 0.65, y = 0.7, w = 5.8) { tx(s, text.toUpperCase(), x, y, w, 0.22, 6.5, C.green, { bold: true, charSpacing: 2.8 }); }
function title(s, text, sub) {
  tx(s, text, 0.65, 1.03, 8.9, 0.84, 30, C.ink, { bold: false, breakLine: true, valign: 'top' });
  if (sub) tx(s, sub, 0.68, 1.9, 8.8, 0.48, 11, C.muted, { valign: 'top' });
}
function footer(s, text = 'HardCoders  •  Entire Graph Buildathon') { tx(s, text, 0.6, 7.08, 5.8, 0.14, 5.5, '626A66', { charSpacing: 1.3 }); }
function line(s, x, y, w, h, color = C.line, width = 1, dash = 'solid') { s.addShape(pptx.ShapeType.line, { x, y, w, h, line: { color, width, dashType: dash } }); }
function panel(s, x, y, w, h, color = C.panel, border = C.line) { s.addShape(pptx.ShapeType.rect, { x, y, w, h, fill: { color }, line: { color: border, width: 0.8 } }); }
function pill(s, text, x, y, w, color) { s.addShape(pptx.ShapeType.roundRect, { x, y, w, h: 0.3, rectRadius: 0.04, fill: { color, transparency: 88 }, line: { color, transparency: 38, width: 0.7 } }); tx(s, text.toUpperCase(), x + 0.08, y + 0.03, w - 0.16, 0.2, 6.2, color, { bold: true, charSpacing: 1.1, align: 'center' }); }
function image(s, file, x, y, w, h) { s.addShape(pptx.ShapeType.rect, { x: x - 0.035, y: y - 0.035, w: w + 0.07, h: h + 0.07, fill: { color: C.panel }, line: { color: '303834', width: 0.8 } }); s.addImage({ path: A(file), x, y, w, h }); }
function metric(s, value, label, x, y, w, accent = C.green) { panel(s, x, y, w, 1.15); tx(s, value, x + 0.18, y + 0.12, w - 0.36, 0.55, 25, C.ink, { bold: false }); tx(s, label.toUpperCase(), x + 0.2, y + 0.75, w - 0.4, 0.18, 5.8, accent, { bold: true, charSpacing: 1.8 }); }
function web(s, cx, cy, r) {
  const spokes = 8;
  for (let i = 0; i < spokes; i++) {
    const a = Math.PI * 2 * i / spokes;
    line(s, cx, cy, Math.cos(a) * r, Math.sin(a) * r, '2A302D', 0.6);
  }
  for (const ring of [0.3, 0.55, 0.8, 1]) {
    s.addShape(pptx.ShapeType.ellipse, { x: cx - r * ring, y: cy - r * ring, w: r * 2 * ring, h: r * 2 * ring, fill: { color: C.bg, transparency: 100 }, line: { color: '252C28', transparency: 25, width: 0.5 } });
  }
}
function node(s, x, y, n, label, color = C.cyan) {
  s.addShape(pptx.ShapeType.ellipse, { x, y, w: 0.62, h: 0.62, fill: { color, transparency: 91 }, line: { color, width: 1.1 } });
  tx(s, n, x, y + 0.16, 0.62, 0.2, 8, color, { bold: true, align: 'center' });
  tx(s, label, x - 0.35, y + 0.72, 1.32, 0.28, 7, C.muted, { align: 'center', valign: 'top' });
}

// 1 — opening
{
  const s = slide();
  web(s, 10.85, 3.55, 2.2);
  s.addShape(pptx.ShapeType.ellipse, { x: 10.67, y: 3.37, w: 0.36, h: 0.36, fill: { color: C.green }, line: { color: C.green } });
  eyebrow(s, 'HardCoders presents');
  tx(s, 'Sense the collision', 0.65, 1.35, 7.3, 0.72, 35, C.ink);
  tx(s, 'before it lands.', 0.65, 2.05, 7.3, 0.72, 35, C.muted);
  tx(s, 'Spidey Sense combines live team presence with truthful repository intelligence—so agents coordinate before changes collide.', 0.68, 3.05, 6.3, 1.1, 15, C.ink, { valign: 'top' });
  pill(s, 'Browser-first', 0.68, 4.55, 1.45, C.cyan); pill(s, 'No-egress graph', 2.28, 4.55, 1.72, C.green); pill(s, 'Verification built in', 4.16, 4.55, 1.95, C.amber);
  tx(s, 'GRAPH IS EVIDENCE, NOT AN ORACLE', 0.68, 5.55, 6.8, 0.36, 9, C.red, { bold: true, charSpacing: 2.3 });
  footer(s, 'HardCoders  •  Spidey Sense  •  Production pitch');
}

// 2 — problem
{
  const s = slide(); eyebrow(s, 'The coordination blind spot'); title(s, 'Agents move fast. Context does not.', 'Modern teams have three fragmented truths—and collisions happen in the gaps.');
  const cards = [
    ['01', 'Human intent', 'Who owns the mission? What outcome is expected?', C.cyan],
    ['02', 'Agent activity', 'Who is online? Which task is active or blocked?', C.green],
    ['03', 'Code structure', 'Which files and symbols are actually connected?', C.amber],
  ];
  cards.forEach(([n, h, body, color], i) => { const x = 0.68 + i * 4.12; panel(s, x, 2.72, 3.72, 2.25); tx(s, n, x + 0.22, 2.92, 0.55, 0.36, 12, color, { bold: true }); tx(s, h, x + 0.22, 3.36, 3.05, 0.35, 18, C.ink); tx(s, body, x + 0.22, 3.93, 3.1, 0.7, 10.5, C.muted, { valign: 'top' }); });
  line(s, 1.25, 5.55, 10.8, 0, C.red, 1.2, 'dash');
  tx(s, 'A static dependency graph alone cannot see runtime dispatch, generated code, reflection—or team intent.', 1.2, 5.75, 10.9, 0.62, 17, C.ink, { align: 'center', bold: true });
  footer(s);
}

// 3 — loop
{
  const s = slide(); eyebrow(s, 'The product'); title(s, 'One coordination loop. Four decisive moments.', 'Every surface shares the same live plan and the same evidence contract.');
  const xs = [1.08, 4.03, 6.98, 9.93];
  const items = [['01','CONNECT','Invite teammates in-browser'],['02','ASSIGN','Declare mission + target'],['03','MAP','Project Graph evidence'],['04','VERIFY','Inspect source + run tests']];
  xs.forEach((x,i) => { if(i<3) line(s,x+1.0,3.45,1.95,0,C.line,1.4,'dash'); node(s,x,3.12,items[i][0],items[i][1], [C.cyan,C.green,C.violet,C.amber][i]); tx(s,items[i][2],x-0.45,4.28,1.55,0.55,9,C.muted,{align:'center',valign:'top'}); });
  panel(s, 1.05, 5.25, 11.2, 0.85, '080B09', '314039');
  tx(s, 'The output is not “trust the graph.” It is: here is the evidence class, the caveat, and the safest next verification.', 1.35, 5.46, 10.6, 0.4, 15, C.ink, { align: 'center' });
  footer(s);
}

// 4 — onboarding
{
  const s = slide(); eyebrow(s, 'Connection-first experience'); title(s, 'Real teammates. Zero invented identity.', 'Create and join entirely through the website; browser capability boundaries stay explicit.');
  image(s, 'create.png', 0.68, 2.45, 5.82, 3.27); image(s, 'join.png', 6.82, 2.45, 5.82, 3.27);
  pill(s, 'Private invite fragment', 0.85, 5.98, 1.95, C.green); pill(s, 'Scoped credentials', 3.02, 5.98, 1.65, C.cyan); pill(s, 'No terminal onboarding', 4.88, 5.98, 1.95, C.violet);
  tx(s, 'Names and roles are entered by people. Provider details are detected—never fabricated.', 7.05, 5.9, 5.35, 0.48, 11, C.ink, { bold: true, align: 'right' });
  footer(s);
}

// 5 — assignment
{
  const s = slide(); eyebrow(s, 'Dynamic orchestration'); title(s, 'Presence becomes actionable.', 'Authenticated members enter the plan by stable server-issued identity—not fragile display-name matching.');
  image(s, 'plan.png', 0.68, 2.38, 8.2, 4.61);
  const facts = [['LIVE','Connected teammates become assignable immediately.',C.green],['PRECISE','Mission, owner, state, file, and symbol stay editable.',C.cyan],['DYNAMIC','No fixed people, paths, providers, or mission count.',C.violet]];
  facts.forEach(([tag,body,color],i)=>{const y=2.58+i*1.28; pill(s,tag,9.3,y,1.15,color); tx(s,body,9.35,y+0.42,3.25,0.62,10.5,C.ink,{valign:'top'});});
  footer(s);
}

// 6 — graph
{
  const s = slide(); eyebrow(s, 'Graph-powered experience'); title(s, 'A spatial map with an honesty layer.', 'Mission nodes, source evidence, relation strands, and risk signals share one accessible SVG scene.');
  image(s, 'graph.png', 0.68, 2.28, 12.0, 5.02);
  pill(s, 'Confirmed', 8.72, 1.34, 1.15, C.green); pill(s, 'Heuristic', 10.0, 1.34, 1.15, C.amber); pill(s, 'Incomplete', 11.28, 1.34, 1.15, C.red);
  footer(s, 'Real HardCoders_ repository evidence  •  Selectable nodes  •  Keyboard accessible');
}

// 7 — evidence policy
{
  const s = slide(); eyebrow(s, 'Track 2 requirement'); title(s, 'Confidence changes the decision.', 'The product never turns an incomplete static relationship into a certain runtime claim.');
  const rows = [
    ['CONFIRMED','Exact / package / import-bound structural evidence','May support BLOCK when direct','Inspect source before merge',C.green],
    ['HEURISTIC','Inferred, pattern, co-change, framework or test evidence','REVIEW','Validate call sites + focused tests',C.amber],
    ['INCOMPLETE','Partial snapshot, parser failure, dropped support','REVIEW — never certainty','Runtime registration + source/test verification',C.red],
  ];
  rows.forEach((r,i)=>{const y=2.55+i*1.18; panel(s,0.7,y,11.95,0.95,'080B09',i===2?'54212A':C.line); pill(s,r[0],0.9,y+0.32,1.45,r[4]); tx(s,r[1],2.65,y+0.15,3.45,0.6,10,C.ink,{valign:'mid'}); tx(s,r[2],6.28,y+0.15,2.2,0.6,10,r[4],{bold:true}); tx(s,r[3],8.72,y+0.15,3.55,0.6,9.5,C.muted);});
  tx(s, 'ABSENCE OF EVIDENCE ≠ EVIDENCE OF INDEPENDENCE', 1.0, 6.25, 11.3, 0.4, 13, C.red, { bold: true, align: 'center', charSpacing: 1.6 });
  footer(s);
}

// 8 — proof
{
  const s = slide(); eyebrow(s, 'Running proof, not synthetic claims'); title(s, 'Mapped against HardCoders_.', 'The live production build analyzes the Anchor repository and preserves every detected limitation.');
  metric(s,'4,631','Indexed nodes',0.7,2.55,2.75,C.cyan); metric(s,'14,850','Graph relations',3.65,2.55,2.75,C.green); metric(s,'01','Analysis warning',6.6,2.55,2.75,C.amber); metric(s,'02','Partial failures',9.55,2.55,2.75,C.red);
  image(s, 'evidence.png', 0.7, 4.02, 7.15, 2.82);
  panel(s,8.18,4.02,4.12,2.82,'080B09','54212A');
  tx(s,'Why the red matters',8.48,4.34,3.4,0.35,17,C.ink,{bold:true});
  tx(s,'SQL parser failures and dynamic behavior lower confidence globally. Spidey Sense surfaces that state, downgrades claims, and points to source/test verification.',8.48,4.95,3.35,1.2,11,C.muted,{valign:'top'});
  pill(s,'Truth over theater',8.48,6.25,1.65,C.red);
  footer(s);
}

// 9 — architecture/security
{
  const s = slide(); eyebrow(s, 'Production architecture'); title(s, 'Shared presence. Local code intelligence.', 'The boundary is deliberate: team metadata can travel; repository analysis remains no-egress.');
  const boxes = [
    [0.75,'BROWSER','Invite • identity • mission • blocker',C.cyan],
    [3.8,'SHARED COORDINATOR','Scoped tokens • heartbeat • SSE',C.green],
    [7.15,'LOCAL ENTIRE GRAPH','Repository snapshot • risk engine',C.violet],
    [10.35,'VERIFICATION','Source inspection • focused tests',C.amber],
  ];
  boxes.forEach((b,i)=>{panel(s,b[0],2.65,i===1?2.7:2.25,1.55,'080B09',b[3]);tx(s,b[1],b[0]+0.16,2.86,(i===1?2.38:1.93),0.28,9,b[3],{bold:true,charSpacing:1});tx(s,b[2],b[0]+0.16,3.38,(i===1?2.38:1.93),0.55,8.5,C.muted,{valign:'top'});if(i<3)line(s,b[0]+(i===1?2.7:2.25),3.42,(i===0?.8:i===1?.65:.95),0,C.line,1.2,'dash');});
  panel(s,0.75,4.75,11.85,1.25,'090B0A','303A35');
  tx(s,'TRANSMITTED',1.02,4.98,1.3,0.25,7,C.green,{bold:true,charSpacing:1.4});tx(s,'Name • role • mission • blocker • bounded activity',2.35,4.94,4.4,0.35,11,C.ink);
  tx(s,'NEVER COLLECTED',1.02,5.48,1.3,0.25,7,C.red,{bold:true,charSpacing:1.2});tx(s,'Prompts • reasoning • terminal • file contents • secrets',2.35,5.44,4.8,0.35,11,C.ink);
  tx(s,'127.0.0.1 stays local. Cross-machine deployments require a reachable HTTPS origin or an explicitly controlled LAN binding.',7.45,4.95,4.55,0.7,9.5,C.muted,{align:'right',valign:'top'});
  footer(s);
}

// 10 — close
{
  const s = slide();
  web(s, 10.9, 3.55, 2.15);
  s.addShape(pptx.ShapeType.ellipse, { x: 10.72, y: 3.37, w: 0.36, h: 0.36, fill: { color: C.red }, line: { color: C.red } });
  eyebrow(s, 'The ask');
  tx(s,'Don’t trust the map.',0.68,1.42,7.3,0.65,31,C.ink);
  tx(s,'Use it to verify faster.',0.68,2.08,7.3,0.65,31,C.green);
  tx(s,'Judge the complete loop live:',0.7,3.1,4.0,0.3,12,C.muted,{bold:true});
  const steps=['Create a room','Join from another browser','Assign a live teammate','Inspect Graph evidence','Refresh and verify'];
  steps.forEach((step,i)=>{tx(s,String(i+1).padStart(2,'0'),0.72,3.68+i*.48,.38,.22,8,[C.cyan,C.green,C.violet,C.amber,C.red][i],{bold:true});tx(s,step,1.22,3.65+i*.48,3.6,.26,11,C.ink);});
  panel(s,6.6,5.58,5.75,0.78,'080B09','34433C');
  tx(s,'github.com/Preethesh16/entire-graph',6.85,5.78,5.25,0.26,12,C.ink,{bold:true,align:'center'});
  tx(s,'SPIDEY SENSE  /  SEE THE COLLISION BEFORE IT LANDS',6.6,6.62,5.75,0.25,7,C.red,{bold:true,charSpacing:1.4,align:'center'});
  footer(s,'HardCoders  •  Preethesh + team  •  Built with Entire Graph');
}

pptx.writeFile({ fileName: path.join(__dirname, 'Spidey-Sense-Pitch-Deck.pptx') });
