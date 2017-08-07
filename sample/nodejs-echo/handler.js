"use strict"

module.exports = (context, callback) => {
    callback(undefined, {nodeVersion: process.version, input: context });
}
