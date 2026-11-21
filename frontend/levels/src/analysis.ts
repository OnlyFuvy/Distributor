export interface WindowLevel {
  start: number;
  end: number;
  peak: number;
  rms: number;
}

export function analyze(channels: readonly Float32Array[], rate: number, seconds: number): WindowLevel[] {
  const frames = channels[0]?.length ?? 0;
  if (!frames || channels.some(channel => channel.length !== frames)
      || !Number.isFinite(rate) || rate <= 0 || !Number.isFinite(seconds) || seconds <= 0) {
    throw new Error('Недопустимые параметры аудио');
  }
  const width = Math.max(1, Math.round(rate * seconds));
  const rows: WindowLevel[] = [];
  for (let start = 0; start < frames; start += width) {
    const end = Math.min(frames, start + width);
    let peak = 0, squares = 0;
    for (const channel of channels) {
      for (let i = start; i < end; i++) {
        const sample = channel[i];
        peak = Math.max(peak, Math.abs(sample));
        squares += sample * sample;
      }
    }
    rows.push({ start: start / rate, end: end / rate, peak,
      rms: Math.sqrt(squares / ((end - start) * channels.length)) });
  }
  return rows;
}

export function db(value: number): string {
  return value > 0 ? (20 * Math.log10(value)).toFixed(2) : '-∞';
}

export function csv(rows: readonly WindowLevel[]): string {
  return ['start_s,end_s,peak_dbfs,rms_dbfs', ...rows.map(row =>
    [row.start.toFixed(4), row.end.toFixed(4), db(row.peak), db(row.rms)].join(',')
  )].join('\r\n');
}

export function demo(rate = 8000): Float32Array {
  return Float32Array.from({length: rate * 12}, (_, i) => {
    const t = i / rate;
    const amplitude = t < 3 ? 0.13 : t < 7 ? 0.75 : t < 9 ? 0 : 0.4;
    return amplitude * Math.sin(2 * Math.PI * 220 * t);
  });
}
