const path = require('node:path');

const REPO_OWNER = 'tszaks';
const REPO_NAME = 'codex-sessions';

function resolvePlatformTarget(platform, arch) {
  const matrix = {
    darwin: {
      arm64: { goos: 'darwin', goarch: 'arm64', archiveExtension: 'tar.gz', binaryName: 'codex-sessions' },
      x64: { goos: 'darwin', goarch: 'amd64', archiveExtension: 'tar.gz', binaryName: 'codex-sessions' },
    },
    linux: {
      arm64: { goos: 'linux', goarch: 'arm64', archiveExtension: 'tar.gz', binaryName: 'codex-sessions' },
      x64: { goos: 'linux', goarch: 'amd64', archiveExtension: 'tar.gz', binaryName: 'codex-sessions' },
    },
    win32: {
      x64: { goos: 'windows', goarch: 'amd64', archiveExtension: 'zip', binaryName: 'codex-sessions.exe' },
    },
  };

  const byPlatform = matrix[platform];
  const target = byPlatform && byPlatform[arch];
  if (!target) {
    throw new Error(`Unsupported platform for codex-sessions npm install: ${platform} ${arch}`);
  }

  return {
    platform,
    arch,
    ...target,
  };
}

function buildAssetName(version, target) {
  return `codex-sessions_${version}_${target.goos}_${target.goarch}.${target.archiveExtension}`;
}

function buildArchiveBinaryPath(target) {
  return target.binaryName;
}

function buildReleaseDownloadUrl(version, target) {
  return `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${version}/${buildAssetName(version, target)}`;
}

function buildInstalledBinaryPath(packageRoot, target) {
  return path.join(packageRoot, 'npm', 'bin', target.binaryName === 'codex-sessions.exe' ? 'codex-sessions-bin.exe' : 'codex-sessions-bin');
}

module.exports = {
  REPO_OWNER,
  REPO_NAME,
  resolvePlatformTarget,
  buildAssetName,
  buildArchiveBinaryPath,
  buildReleaseDownloadUrl,
  buildInstalledBinaryPath,
};
