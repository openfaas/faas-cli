const os = require('os');
const path = require('path');
const fs = require('fs');
const installer = require('./install.js');
const lib = require('./lib');

const greenCheck = "\x1b[92m\u2714\x1b[39m";
const redNoSign = "\x1b[91m\u2718\x1b[39m";

async function runInstaller() {
    try {
        await installer.install()
    } catch (error) {
        throw new Error(`Installation process failed with error: ${error.message}`);
    }
}

async function verifyInstallation() {
    let binaryName = lib.getBinaryName(os.type(), os.arch());
    let dest = path.join(__dirname, `bin/${binaryName}`);
    if (!fs.existsSync(dest)) {
        throw new Error(`File was not found in ${dest}`);
    }

    await lib.cmd('faas-cli version');
}

(async function () {
    try {
        await runInstaller();
        await verifyInstallation();
        console.log(greenCheck, 'Installation test succeeded!');   
    } catch (error) {
        console.log(redNoSign, `Installation test failed with error: ${error.message}`);
        process.exit(1);
    }
})();