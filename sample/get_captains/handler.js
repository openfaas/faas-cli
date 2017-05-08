"use strict"

let fs = require('fs');
let cheerio = require('cheerio');
let Parser = require('./parser');
var request = require("request");

module.exports = (context, callback) => {
    createList((err, sortedCaptains) => {
        callback(null, sortedCaptains);
    });
}

let createList = (next) => {
    let parser = new Parser(cheerio);

    request.get(process.env.url, (err, res, text) => {
        if (err) {
            return next(err, {});
        }
        let captains = parser.parse(text);

        let valid = 0;
        let sorted = captains.sort((x, y) => {
            if (x.text > y.text) {
                return 1;
            } else if (x.text < y.text) {
                return -1;
            }
            return 0;
        });
        next(null, sorted);
    });
};