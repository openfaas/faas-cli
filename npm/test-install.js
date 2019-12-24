const os = require('os');
const path = require('path');
const fs = require('fs');
const installer = require('./install.js');
const lib = require('./lib');
const assert = require('assert');

describe('E2E Install', function () {
    it('should install binary', async function() {
        this.timeout(5000);
        await installer.install();
    
        let binaryName = lib.getBinaryName(os.type(), os.arch());
        let dest = path.join(__dirname, 'bin', binaryName);
    
        assert.ok(fs.existsSync(dest), 'Installed executable was not found');
    });
});