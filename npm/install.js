'use strict';

const os = require('os');
const path = require('path');
const del = require('delete');
const mkdirp = require('mkdirp');

const lib = require('./lib');

module.exports.install = async () => {
  const type = os.type();
  const arch = os.arch();
  let binaryName = lib.getBinaryName(type, arch);
  let dest = path.join(__dirname, "bin", binaryName);
  mkdirp.sync(path.dirname(dest));
  del.sync(dest, { force: true });

  try {

    let releaseURL = await lib.getRelease()
    let downloadURL = releaseURL.replace("tag", "download") +"/"+ binaryName;

    let url = downloadURL;

    console.log(`Downloading package ${url} to ${dest}`);
    await lib.download(url, dest);
  } catch (error) {
    throw new Error(`Download failed! ${error.message}`);
  }

  // Don't use `chmod` on Windows
  if (binaryName.endsWith('.exe')) {
    await lib.cmd(`chmod +x ${dest}`);
  }
  console.log('Download complete.');
}


module.exports.init = () => {
  this.install()
    .then(() => process.exit())
    .catch(err => {
      console.error(err.message);
      process.exit(1);
    });
}
