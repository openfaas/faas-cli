#!/usr/bin/env node

const os = require('os');
const path = require('path');
const { spawnSync } = require('child_process');

const extname = os.type() === 'Windows_NT' ? '.exe' : '';
const bin = path.join(__dirname, `bin/faas-cli${extname}`);

spawnSync(bin, process.argv.slice(2), { cwd: process.cwd(), stdio: 'inherit' });
