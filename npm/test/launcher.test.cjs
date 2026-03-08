const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');

const { getInstalledBinaryPath } = require('../lib/launcher.cjs');

test('getInstalledBinaryPath points at the package binary cache', () => {
  const packageRoot = '/tmp/codex-sessions-package';
  assert.equal(
    getInstalledBinaryPath(packageRoot, false),
    path.join(packageRoot, 'npm', 'bin', 'codex-sessions-bin')
  );
});

test('getInstalledBinaryPath adds .exe on windows', () => {
  const packageRoot = '/tmp/codex-sessions-package';
  assert.equal(
    getInstalledBinaryPath(packageRoot, true),
    path.join(packageRoot, 'npm', 'bin', 'codex-sessions-bin.exe')
  );
});
