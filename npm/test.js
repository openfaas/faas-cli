const assert = require('assert');
const lib = require('./lib');

const greenCheck = "\x1b[92m\u2714\x1b[39m";
const redNoSign = "\x1b[91m\u2718\x1b[39m";

const tests = {
    SuffixTests: {
        Windows() {
            let type = 'Windows_NT';
            let arch = '';
            let expect = '.exe';
            let result = lib.getSuffix(type, arch);
            assert.ok(result == expect, `Expected [${expect}], but got [${result}]`);
        },
        LinuxX64() {
            let type = 'Linux';
            let arch = 'x64';
            let expect = '';
            let result = lib.getSuffix(type, arch);
            assert.ok(result == expect, `Expected [${expect}], but got [${result}]`);
        },
        LinuxArm64() {
            let type = 'Linux';
            let arch = 'aarch64';
            let expect = '-arm64';
            let result = lib.getSuffix(type, arch);
            assert.ok(result == expect, `Expected [${expect}], but got [${result}]`);
        },
        LinuxArmhf6() {
            let type = 'Linux';
            let arch = 'armv61';
            let expect = '-armhf';
            let result = lib.getSuffix(type, arch);
            assert.ok(result == expect, `Expected [${expect}], but got [${result}]`);
        },
        LinuxArmhf7() {
            let type = 'Linux';
            let arch = 'armv71';
            let expect = '-armhf';
            let result = lib.getSuffix(type, arch);
            assert.ok(result == expect, `Expected [${expect}], but got [${result}]`);
        },
        MacOS() {
            let type = 'Darwin';
            let arch = '';
            let expect = '-darwin';
            let result = lib.getSuffix(type, arch);
            assert.ok(result == expect, `Expected [${expect}], but got [${result}]`);
        },
        Unsupported() {
            let type = 'BadType';
            let arch = 'BadArch';
            assert.throws(() => lib.getSuffix(type, arch), `Should throw error on unexpected type (${type}) and arch ${arch}`);
        }
    },
    BinaryNameTests: {
        Windows() {
            let type = 'Windows_NT';
            let arch = '';
            let expected = 'faas-cli.exe'
            let binary = lib.getBinaryName(type, arch);
            assert.ok(binary == expected, `Expected [${expected}], but got [${binary}]`);
        },
        Linux() {
            let type = 'Linux';
            let arch = 'x64';
            let expected = 'faas-cli'
            let binary = lib.getBinaryName(type, arch);
            assert.ok(binary == expected, `Expected [${expected}], but got [${binary}]`);
        },
        MacOS() {
            let type = 'Darwin';
            let arch = '';
            let expected = 'faas-cli'
            let binary = lib.getBinaryName(type, arch);
            assert.ok(binary == expected, `Expected [${expected}], but got [${binary}]`);
        }
    },
    ReleasesTests: {
        async GetReleases() {
            let result = await lib.getReleases();
            assert.ok(result, 'Releases response was empty');
            assert.ok(result.length > 1);            
        }
    }
}

for (var suite in tests) {
    if (tests.hasOwnProperty(suite)) {
        console.log('Test Suite:', suite);
        let suiteTests = tests[suite];

        for (var test in suiteTests) {
            let testFunc = suiteTests[test];
            process.stdout.write(`\t${test} `);
            try {
                testFunc();
                console.log(greenCheck);
            } catch (error) {
                console.log(redNoSign, error.message);
            }
        }
    }
}