'use strict';

const os = require('os');
const path = require('path');
const del = require('delete');
const mkdirp = require('mkdirp');

const lib = require('./lib');

async function main() {
  const type = os.type();
  const arch = os.arch();
  let suffix = lib.getSuffix(type, arch);
  let name = `faas-cli${suffix}`;
  let extname = suffix === '.exe' ? '.exe' : '';
  let dest = path.join(__dirname, `bin/faas-cli${extname}`);
  mkdirp.sync(path.dirname(dest));
  del.sync(dest, { force: true });

  try {
    let releases = await lib.getReleases();
    let release = releases.find(release => release.name === name);
    let url = release && release.browser_download_url;

    console.log(`Downloading package ${url} to ${dest}`);
    await lib.download(url, dest);
  } catch (error) {
    throw new Error(`Download failed! ${error.message}`);
  }

  // Don't use `chmod` on Windows
  if (suffix !== '.exe') {
    await lib.cmd(`chmod +x ${dest}`);
  }
  console.log('Download complete.');
}

main()
  .then(() => process.exit())
  .catch(err => {
    console.error(err.message);
    process.exit(1);
  });
