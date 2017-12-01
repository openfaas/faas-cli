// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

"use strict"

const getStdin = require('get-stdin');

const handler = require('./function/handler');

getStdin().then(val => {
    handler(val, (err, res) => {
        if (err) {
            return console.error(err);
        }
        if(isArray(res) || isObject(res)) {
            console.log(JSON.stringify(res));
        } else {
            process.stdout.write(res);
        }
    });
}).catch(e => {
    console.error(e.stack);
});

const isArray = (a) => {
    return (!!a) && (a.constructor === Array);
};

const isObject = (a) => {
    return (!!a) && (a.constructor === Object);
};
