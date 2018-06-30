#!/bin/sh

for f in faas-cli*; do shasum -a 256 $f > $f.sha256; done