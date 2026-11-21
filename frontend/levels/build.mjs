import {build} from 'esbuild';
import {readFile, writeFile} from 'node:fs/promises';
const result = await build({entryPoints:['src/app.ts'], bundle:true, write:false, target:'es2020'});
const template = await readFile('template.html','utf8');
await writeFile('index.html', template.replace('/* BUNDLE */', () => result.outputFiles[0].text));
