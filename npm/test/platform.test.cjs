const test = require('node:test');
const assert = require('node:assert/strict');

const {
  resolvePlatformTarget,
  buildAssetName,
  buildArchiveBinaryPath,
  buildReleaseDownloadUrl,
} = require('../lib/platform.cjs');

test('resolvePlatformTarget maps darwin arm64', () => {
  assert.deepEqual(resolvePlatformTarget('darwin', 'arm64'), {
    platform: 'darwin',
    arch: 'arm64',
    goos: 'darwin',
    goarch: 'arm64',
    archiveExtension: 'tar.gz',
    binaryName: 'codex-sessions',
  });
});

test('resolvePlatformTarget maps windows x64', () => {
  assert.deepEqual(resolvePlatformTarget('win32', 'x64'), {
    platform: 'win32',
    arch: 'x64',
    goos: 'windows',
    goarch: 'amd64',
    archiveExtension: 'zip',
    binaryName: 'codex-sessions.exe',
  });
});

test('resolvePlatformTarget throws for unsupported targets', () => {
  assert.throws(() => resolvePlatformTarget('freebsd', 'x64'), /Unsupported platform/);
});

test('buildAssetName matches GitHub release archive naming', () => {
  const target = resolvePlatformTarget('linux', 'arm64');
  assert.equal(buildAssetName('0.2.0', target), 'codex-sessions_0.2.0_linux_arm64.tar.gz');
});

test('buildArchiveBinaryPath handles windows zip layout', () => {
  const target = resolvePlatformTarget('win32', 'x64');
  assert.equal(buildArchiveBinaryPath(target), 'codex-sessions.exe');
});

test('buildReleaseDownloadUrl points at the matching GitHub asset', () => {
  const target = resolvePlatformTarget('darwin', 'arm64');
  assert.equal(
    buildReleaseDownloadUrl('0.2.0', target),
    'https://github.com/tszaks/codex-sessions/releases/download/v0.2.0/codex-sessions_0.2.0_darwin_arm64.tar.gz'
  );
});
