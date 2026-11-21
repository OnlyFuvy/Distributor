import { analyze, csv, db, demo } from './analysis';
import type { WindowLevel } from './analysis';

function element<T extends HTMLElement>(id: string): T {
  const item = document.getElementById(id);
  if (!item) throw new Error('Нет элемента ' + id);
  return item as T;
}

let signal = { channels: [demo()], rate: 8000, name: 'Демо · тестовый тон' };
let revision = 0;
let rows: WindowLevel[] = [];
const interval = element<HTMLSelectElement>('interval');
const cursor = element<HTMLInputElement>('cursor');
const file = element<HTMLInputElement>('file');
const chart = element<HTMLCanvasElement>('chart');
const status = element('status');

function draw(): void {
  const ctx = chart.getContext('2d');
  if (!ctx || !rows.length) return;
  const w = chart.clientWidth, h = chart.clientHeight;
  const scale = window.devicePixelRatio || 1;
  chart.width = Math.round(w * scale);
  chart.height = Math.round(h * scale);
  ctx.scale(scale, scale);
  const y = (v: number) => h - 8 - (h - 22) * (1 + Math.max(-60, Math.min(0, 20 * Math.log10(v))) / 60);
  const x = (i: number) => 46 + (w - 54) * i / Math.max(1, rows.length - 1);
  ctx.font = '11px monospace';
  for (const tick of [0, -20, -40, -60]) {
    const pos = y(10 ** (tick / 20));
    ctx.fillStyle = '#a4afbe'; ctx.fillText(String(tick), 4, pos + 3);
    ctx.fillRect(38, pos, w - 46, 0.5);
  }
  for (const [key, color] of [['peak', '#fbbb70'], ['rms', '#71d6db']] as const) {
    ctx.beginPath(); ctx.strokeStyle = color; ctx.lineWidth = 2;
    rows.forEach((row, i) => i ? ctx.lineTo(x(i), y(row[key])) : ctx.moveTo(x(i), y(row[key])));
    ctx.stroke();
  }
  const index = Number(cursor.value);
  ctx.fillStyle = '#ffffff'; ctx.fillRect(x(index), 5, 1, h - 10);
  const row = rows[index];
  element('detail').textContent = row.start.toFixed(2) + '–' + row.end.toFixed(2) +
    ' с · пик ' + db(row.peak) + ' dBFS · RMS ' + db(row.rms) + ' dBFS';
}

function refresh(): void {
  rows = analyze(signal.channels, signal.rate, Number(interval.value));
  cursor.max = String(rows.length - 1); cursor.value = '0';
  element('track').textContent = signal.name;
  element('length').textContent = rows[rows.length - 1].end.toFixed(2) + ' с';
  element('peak').textContent = db(rows.reduce((max, row) => Math.max(max, row.peak), 0)) + ' dBFS';
  element('count').textContent = String(rows.length);
  draw();
}

file.onchange = async () => {
  const selected = file.files?.[0];
  if (!selected) return;
  const token = ++revision;
  file.value = '';
  if (selected.size > 24 * 1024 * 1024) { status.textContent = 'Выберите файл до 24 МиБ.'; return; }
  status.textContent = 'Декодирование…';
  let context: AudioContext | undefined;
  try {
    context = new AudioContext();
    const audio = await context.decodeAudioData(await selected.arrayBuffer());
    if (token !== revision) return;
    if (audio.duration > 300) throw new Error('Выберите запись до 5 минут.');
    signal = {channels: Array.from({length: audio.numberOfChannels}, (_, i) => audio.getChannelData(i)),
      rate: audio.sampleRate, name: selected.name};
    refresh(); status.textContent = '';
  } catch (err) {
    if (token === revision) status.textContent = err instanceof Error && err.message.includes('5 минут')
      ? err.message : 'Браузер не смог прочитать файл. Попробуйте WAV или MP3.';
  } finally { if (context) await context.close(); }
};
element('demo').onclick = () => {
  revision++; signal = {channels: [demo()], rate: 8000, name: 'Демо · тестовый тон'};
  status.textContent = ''; refresh();
};
interval.onchange = refresh; cursor.oninput = draw;
element('export').onclick = () => {
  const url = URL.createObjectURL(new Blob([csv(rows)], {type: 'text/csv;charset=utf-8'}));
  const link = document.createElement('a'); link.href = url; link.download = 'levels.csv'; link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
};
window.addEventListener('resize', draw);
refresh();
