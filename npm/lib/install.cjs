const fs = require('node:fs');
const path = require('node:path');
const https = require('node:https');
const os = require('node:os');

const AdmZip = require('adm-zip');
const tar = require('tar');

const {
  resolvePlatformTarget,
  buildArchiveBinaryPath,
  buildInstalledBinaryPath,
  buildReleaseDownloadUrl,
} = require('./platform.cjs');

function downloadFile(url, outputPath) {
  return new Promise((resolve, reject) => {
    const request = https.get(url, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        downloadFile(response.headers.location, outputPath).then(resolve, reject);
        return;
      }

      if (response.statusCode !== 200) {
        response.resume();
        reject(new Error(`Download failed with status ${response.statusCode} for ${url}`));
        return;
      }

      const file = fs.createWriteStream(outputPath);
      response.pipe(file);

      file.on('finish', () => {
        file.close(resolve);
      });

      file.on('error', (error) => {
        file.close(() => reject(error));
      });
    });

    request.on('error', reject);
  });
}

async function extractArchive(archivePath, target, installPath) {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'codex-sessions-'));

  try {
    if (target.archiveExtension === 'zip') {
      const zip = new AdmZip(archivePath);
      zip.extractAllTo(tempDir, true);
    } else {
      await tar.x({
        file: archivePath,
        cwd: tempDir,
      });
    }

    const extractedBinary = path.join(tempDir, buildArchiveBinaryPath(target));
    if (!fs.existsSync(extractedBinary)) {
      throw new Error(`Expected binary not found in archive: ${buildArchiveBinaryPath(target)}`);
    }

    fs.mkdirSync(path.dirname(installPath), { recursive: true });
    fs.copyFileSync(extractedBinary, installPath);

    if (target.goos !== 'windows') {
      fs.chmodSync(installPath, 0o755);
    }
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

async function install() {
  if (process.env.CODEX_SESSIONS_SKIP_INSTALL === '1') {
    return;
  }

  const packageRoot = path.resolve(__dirname, '..', '..');
  const version = require(path.join(packageRoot, 'package.json')).version;
  const target = resolvePlatformTarget(process.platform, process.arch);
  const downloadUrl = buildReleaseDownloadUrl(version, target);
  const archivePath = path.join(os.tmpdir(), path.basename(downloadUrl));
  const installPath = buildInstalledBinaryPath(packageRoot, target);

  try {
    await downloadFile(downloadUrl, archivePath);
    await extractArchive(archivePath, target, installPath);
    console.log(`Installed codex-sessions ${version} for ${target.goos}/${target.goarch}`);
  } catch (error) {
    console.error(`Failed to install codex-sessions binary from ${downloadUrl}`);
    console.error(error instanceof Error ? error.message : String(error));
    process.exit(1);
  } finally {
    fs.rmSync(archivePath, { force: true });
  }
}

install();
