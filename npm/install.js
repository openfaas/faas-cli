'use strict';

const fs = require('fs');
const os = require('os');
const path = require('path');
const del = require('delete');
const axios = require('axios');
const mkdirp = require('mkdirp');
const { promisify } = require('util');
const exec = promisify(require('child_process').exec);

const pkg = require('./package.json');

main()
  .then(() => process.exit())
  .catch(err => {
    console.error(err);
    process.exit(1);
  });

async function main() {
  let suffix = getSuffix();
  let name = `faas-cli${suffix}`;
  let extname = suffix === '.exe' ? '.exe' : '';
  let dest = path.join(__dirname, `bin/faas-cli${extname}`);
  mkdirp.sync(path.dirname(dest));
  del.sync(dest, { force: true });

  let releases = await getReleases();
  let release = releases.find(release => release.name === name);
  let url = release && release.browser_download_url;

  console.log(`Downloading package ${url} to ${dest}`);
  await download(url, dest);

  // Don't use `chmod` on Windows
  if (suffix !== '.exe') {
    await cmd(`chmod +x ${dest}`);
  }
  console.log('Download complete.');
}

function download(url, dest) {
  return new Promise(async (resolve, reject) => {
    let ws = fs.createWriteStream(dest);
    let res = await axios({ url, responseType: 'stream' });
    res.data
      .pipe(ws)
      .on('error', reject)
      .on('finish', () => {
        resolve();
      });
  });
}

function getReleases() {
  return axios.get(`https://api.github.com/repos/openfaas/faas-cli/releases/tags/${pkg.version}`)
    .then(res => res.data.assets);
}

function getSuffix() {
  const type = os.type();
  const arch = os.arch();

  if (type === 'Windows_NT') {
    return '.exe';
  }

  if (type === 'Linux') {
    if (arch === 'x64') {
      return '';
    }

    if (arch === 'aarch64') {
      return '-arm64';
    }

    if (arch === 'armv61' || arch === 'armv71') {
      return '-armhf'
    }
  }

  if (type === 'Darwin') {
    return '-darwin';
  }

  throw new Error(`Unsupported platform: ${type} ${arch}`);
}

async function cmd(...args) {
  let { stdout, stderr } = await exec(...args);
  if (stdout) console.log(stdout);
  if (stderr) console.error(stderr);
}
