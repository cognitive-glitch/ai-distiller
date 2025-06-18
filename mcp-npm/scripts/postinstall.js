#!/usr/bin/env node

const os = require('os');
const path = require('path');
const fs = require('fs');
const https = require('https');
const { execSync } = require('child_process');

const VERSION = require('../package.json').version;

function getPlatformInfo() {
  const platform = os.platform();
  const arch = os.arch();
  
  const platformMap = {
    'darwin': 'darwin',
    'linux': 'linux',
    'win32': 'windows'
  };
  
  const archMap = {
    'x64': 'amd64',
    'arm64': 'arm64'
  };
  
  const mapped = {
    platform: platformMap[platform] || platform,
    arch: archMap[arch] || arch,
    ext: platform === 'win32' ? 'zip' : 'tar.gz'
  };
  
  return mapped;
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        https.get(response.headers.location, (response) => {
          response.pipe(file);
          file.on('finish', () => {
            file.close(resolve);
          });
        }).on('error', reject);
      } else {
        response.pipe(file);
        file.on('finish', () => {
          file.close(resolve);
        });
      }
    }).on('error', reject);
  });
}

async function install() {
  const { platform, arch, ext } = getPlatformInfo();
  const filename = `aid-${platform}-${arch}-v${VERSION}.${ext}`;
  const url = `https://github.com/janreges/ai-distiller/releases/download/v${VERSION}/${filename}`;
  
  console.log(`Downloading AI Distiller for ${platform}/${arch}...`);
  console.log(`URL: ${url}`);
  
  const binDir = path.join(__dirname, '..', 'bin');
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }
  
  const tempFile = path.join(binDir, filename);
  
  try {
    // Download
    await downloadFile(url, tempFile);
    console.log('Download complete.');
    
    // Extract
    console.log('Extracting...');
    if (ext === 'zip') {
      execSync(`unzip -o "${tempFile}" -d "${binDir}"`, { stdio: 'inherit' });
    } else {
      execSync(`tar -xzf "${tempFile}" -C "${binDir}"`, { stdio: 'inherit' });
    }
    
    // Make executable
    const binaryName = platform === 'windows' ? 'aid.exe' : 'aid';
    const binaryPath = path.join(binDir, binaryName);
    if (fs.existsSync(binaryPath) && platform !== 'windows') {
      fs.chmodSync(binaryPath, 0o755);
    }
    
    // Clean up
    fs.unlinkSync(tempFile);
    
    console.log('AI Distiller MCP server installed successfully!');
  } catch (error) {
    console.error('Installation failed:', error.message);
    process.exit(1);
  }
}

install();