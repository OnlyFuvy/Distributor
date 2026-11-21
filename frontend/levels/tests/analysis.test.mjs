import test from 'node:test';
import assert from 'node:assert/strict';
import {analyze, db, csv, demo} from '../src/analysis.ts';
test('stereo channels of opposite polarity keep their energy', () => {
  const [row] = analyze([Float32Array.of(1, -1), Float32Array.of(-1, 1)], 2, 1);
  assert.equal(row.rms, 1); assert.equal(row.peak, 1);
});
test('partial final window uses actual sample count and duration', () => {
  const rows = analyze([Float32Array.of(0, 0, 0.5)], 2, 1);
  assert.equal(rows[1].end, 1.5); assert.equal(rows[1].rms, 0.5);
  assert.match(csv(rows), /1.0000,1.5000,-6.02,-6.02/);
});
test('silence, demo and invalid input', () => {
  assert.equal(db(0), '-∞'); assert.equal(analyze([demo()], 8000, .5)[15].peak, 0);
  assert.throws(() => analyze([], 1, 1)); assert.throws(() => analyze([demo()], 1, 0));
});
