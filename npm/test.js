const assert = require('assert');
const lib = require('./lib');

describe('Platform Suffix', function () {
    it('should return .exe for Windows_NT', function () {
        let result = lib.getSuffix('Windows_NT', '');
        assert.equal(result, '.exe', `Expected .exe, but got: ${result}`);
    });
    it('should return blank for Linux x64', function () {
        let result = lib.getSuffix('Linux', 'x64');
        assert.equal(result, '', `Expected '', but got: ${result}`);
    });
    it('should return -arm64 for Linux aarch64', function () {
        let result = lib.getSuffix('Linux', 'aarch64');
        assert.equal(result, '-arm64', `Expected '-arm64', but got: ${result}`);
    });
    it('should return -armhf for Linux armv61', function () {
        let result = lib.getSuffix('Linux', 'armv61');
        assert.equal(result, '-armhf', `Expected '-armhf', but got: ${result}`);
    });
    it('should return -armhf for Linux armv71', function () {
        let result = lib.getSuffix('Linux', 'armv71');
        assert.equal(result, '-armhf', `Expected '-armhf', but got: ${result}`);
    });
    it('should return -darwin for MacOS', function () {
        let result = lib.getSuffix('Darwin', '');
        assert.equal(result, '-darwin', `Expected '-darwin', but got: ${result}`);
    });
    it('should throw for unsupported platform', function () {
        assert.throws(() => {lib.getSuffix('BadType', 'BadArch')});
    });
});

describe('Binary Name', function() {
    it('should return faas-cli.exe for Windows_NT', function() {
        let result = lib.getBinaryName('Windows_NT', '');
        assert.equal(result, 'faas-cli.exe', `Expected 'faas-cli.exe', but got: ${result}`);
    });
    it('should return faas-cli for Linux', function() {
        let result = lib.getBinaryName('Linux', 'x64');
        assert.equal(result, 'faas-cli', `Expected 'faas-cli', but got: ${result}`);
    });
    it('should return faas-cli for Darwin', function() {
        let result = lib.getBinaryName('Darwin', '');
        assert.equal(result, 'faas-cli', `Expected 'faas-cli', but got: ${result}`);
    });
});

describe('Release Response', function() {
    it('should get a result from release', async function() {
        let result = await lib.getRelease();
        assert.ok(result, 'Releases response was empty');
    })
});