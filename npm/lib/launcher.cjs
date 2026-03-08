const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

function getInstalledBinaryPath(packageRoot, isWindows) {
  return path.join(packageRoot, 'npm', 'bin', isWindows ? 'codex-sessions-bin.exe' : 'codex-sessions-bin');
}

function launch(args) {
  const packageRoot = path.resolve(__dirname, '..', '..');
  const binaryPath = getInstalledBinaryPath(packageRoot, process.platform === 'win32');

  if (!fs.existsSync(binaryPath)) {
    console.error(
      [
        'codex-sessions binary is not installed.',
        'Try reinstalling the package:',
        '  npm install -g @tszaks/codex-sessions',
      ].join('\n')
    );
    process.exit(1);
  }

  const result = spawnSync(binaryPath, args, {
    stdio: 'inherit',
  });

  if (result.error) {
    throw result.error;
  }

  process.exit(result.status === null ? 1 : result.status);
}

module.exports = {
  getInstalledBinaryPath,
  launch,
};
