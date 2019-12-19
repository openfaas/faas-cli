const axios = require('axios');
const pkg = require('./package.json');
const fs = require('fs');
const { promisify } = require('util');
const exec = promisify(require('child_process').exec);

module.exports = {
    getSuffix(type, arch) {
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
    },
    getBinaryName(type, arch) {
        let suffix = this.getSuffix(type, arch);
        let extname = suffix === '.exe' ? '.exe' : '';
        return `faas-cli${extname}`;
    },
    getRelease() {
        return new Promise(async (resolve, reject)=> {
            let latestURL = "https://github.com/openfaas/faas-cli/releases/latest"
            let location = ""
            try {
                await axios({url:latestURL, maxRedirects:0});
            }catch (e){
                location = e.response.headers.location;
            }

            return resolve(location);
        })
    },
    download(url, dest) {
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
    },
    async cmd(...args) {
        let { stdout, stderr } = await exec(...args);
        if (stdout) console.log(stdout);
        if (stderr) console.error(stderr);
    }
}