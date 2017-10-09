// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

"use strict"

let getStdin = require('get-stdin');

let handler = require('./function/handler');

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

let isArray = (a) => {
    return (!!a) && (a.constructor === Array);
};

let isObject = (a) => {
    return (!!a) && (a.constructor === Object);
};
