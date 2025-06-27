/* constants ----------------------------------------------------------- */
const HOST = 'https://pb.launay.one';
const FILES = HOST + '/api/files';
const API = HOST +
  '/api/collections/cards/records?filter=(type="hero")&expand=powers';

/* helpers -------------------------------------------------------------- */
const fileUrl = r => `${FILES}/${r.collectionId}/${r.id}/${r.cover}`;   // PB file route :contentReference[oaicite:0]{index=0}
const deck = document.getElementById('deck');
const overlay = document.getElementById('overlay');
const infoBox = document.getElementById('info');

let openCard = null;
let stored = null;                // remembers the card’s origin {top,left,w,h}

/* fetch + render ------------------------------------------------------- */
(async () => {
  const { items } = await fetch(API).then(r => r.json());
  items.forEach(makeCard);
})().catch(console.error);

/* build one DOM card --------------------------------------------------- */
function makeCard(rec) {
  const c = document.createElement('div');
  c.className = 'card';
  c.style.backgroundImage = `url("${fileUrl(rec)}")`;

  const lbl = document.createElement('div');
  lbl.className = 'label';
  lbl.textContent = `${rec.name} • ${rec.strength}`;
  c.appendChild(lbl);
  deck.appendChild(c);

  c.onpointerenter = () => gsap.to(c, { y: -20, duration: .25, overwrite: 'auto' });        // lift :contentReference[oaicite:1]{index=1}
  c.onpointerleave = () => { if (c !== openCard) gsap.to(c, { y: 0, duration: .25 }); };
  c.onclick = () => open(rec, c);
}

/* open modal ----------------------------------------------------------- */
function open(rec, card) {
  if (openCard) close();          // only one at a time
  openCard = card;

  overlay.style.display = 'block';

  /* remember current rect so we can fly back later */
  const r = card.getBoundingClientRect();                                       // :contentReference[oaicite:2]{index=2}
  stored = { top: r.top, left: r.left, w: r.width, h: r.height };

  const gap = 24, boxW = 320, scale = 1.3;
  const dstL = (innerWidth - (r.width * scale + gap + boxW)) / 2;
  const dstT = (innerHeight - r.height * scale) / 2;

  const cardZ = parseInt(getComputedStyle(overlay).zIndex) + 1;                 // keep above backdrop :contentReference[oaicite:3]{index=3}
  gsap.set(card, {
    position: 'fixed', top: r.top, left: r.left, width: r.width, height: r.height,
    margin: 0, zIndex: cardZ
  });

  gsap.to(card, {
    top: dstT, left: dstL, scale, rotateY: 360, duration: .6, ease: 'power2.out'      // spin :contentReference[oaicite:4]{index=4}
  });

  /* info panel -------------------------------------------------------- */
  infoBox.style.display = 'block';
  infoBox.style.top = dstT + 'px';
  infoBox.style.left = dstL + r.width * scale + gap + 'px';

  const list = rec.expand.powers
    .map(p => `<li><strong>${p.name}</strong> (cost ${p.cost})<br>${p.description}</li>`)
    .join('');
  infoBox.innerHTML =
    `<h2>${rec.name}</h2>
     <p><em>Strength:</em> ${rec.strength}</p>
     <p><em>Powers:</em></p><ul>${list}</ul>`;
}

/* close on backdrop click – includes exit animation ------------------- */
overlay.onclick = close;

function close() {
  if (!openCard) return;

  /* slide card back to remembered coords, then clean up */
  gsap.to(openCard, {
    top: stored.top, left: stored.left, scale: 1, rotateY: 0,
    duration: .5, ease: 'power2.in', onComplete: () => {
      openCard.removeAttribute('style');
      openCard = stored = null;
    }
  });

  gsap.delayedCall(.2, () => { overlay.style.display = 'none'; infoBox.style.display = 'none'; });
}
