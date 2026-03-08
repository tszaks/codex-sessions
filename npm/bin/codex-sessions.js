#!/usr/bin/env node

const { launch } = require('../lib/launcher.cjs');

launch(process.argv.slice(2));
