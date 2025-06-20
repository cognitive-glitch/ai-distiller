#!/usr/bin/env node

const os = require('os');
const path = require('path');
const fs = require('fs');
const https = require('https');
const { execSync } = require('child_process');
const zlib = require('zlib');
const tar = require('tar');

// This should match the AI Distiller release version
const VERSION = '1.0.0';

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
    platform: platformMap[platform],
    arch: archMap[arch],
    ext: platform === 'win32' ? 'zip' : 'tar.gz'
  };
  
  if (!mapped.platform || !mapped.arch) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }
  
  return mapped;
}

function downloadFile(url, dest, isStream = false) {
  return new Promise((resolve, reject) => {
    console.log(`Downloading from ${url}...`);
    
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        return downloadFile(response.headers.location, dest, isStream)
          .then(resolve)
          .catch(reject);
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Download failed with status code: ${response.statusCode}`));
        return;
      }
      
      if (isStream) {
        resolve(response);
      } else {
        const file = fs.createWriteStream(dest);
        response.pipe(file);
        file.on('finish', () => {
          file.close(resolve);
        });
        file.on('error', (err) => {
          fs.unlink(dest, () => {}); // Delete the file on error
          reject(err);
        });
      }
    }).on('error', reject);
  });
}

async function extractArchive(archivePath, destDir, platform) {
  console.log('Extracting archive...');
  
  if (!fs.existsSync(destDir)) {
    fs.mkdirSync(destDir, { recursive: true });
  }
  
  if (platform === 'win32') {
    // For Windows, use built-in PowerShell or fallback to manual extraction
    try {
      // Try PowerShell first (available on all modern Windows)
      execSync(`powershell -command "Expand-Archive -Path '${archivePath}' -DestinationPath '${destDir}' -Force"`, { 
        stdio: 'pipe' 
      });
    } catch (e) {
      // Fallback to Node.js unzip implementation
      console.log('PowerShell extraction failed, using fallback method...');
      // For simplicity, we'll require unzip to be installed
      try {
        execSync(`unzip -o "${archivePath}" -d "${destDir}"`, { stdio: 'inherit' });
      } catch (e2) {
        throw new Error('Failed to extract ZIP file. Please ensure unzip is installed or extract manually.');
      }
    }
  } else {
    // For Unix systems, use tar
    await tar.x({
      file: archivePath,
      cwd: destDir
    });
  }
}

async function install() {
  try {
    const { platform, arch, ext } = getPlatformInfo();
    const archiveName = `aid-${platform}-${arch}-v${VERSION}.${ext}`;
    const url = `https://github.com/janreges/ai-distiller/releases/download/v${VERSION}/${archiveName}`;
    
    console.log(`Installing AI Distiller MCP for ${platform}/${arch}...`);
    console.log(`Version: ${VERSION}`);
    
    const binDir = path.join(__dirname, '..', 'bin');
    const tempFile = path.join(binDir, archiveName);
    const binaryName = platform === 'windows' ? 'aid.exe' : 'aid';
    const binaryPath = path.join(binDir, binaryName);
    
    // Create bin directory
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    // Check if binary already exists and is valid
    if (fs.existsSync(binaryPath)) {
      try {
        // Try to get version to verify it works
        execSync(`"${binaryPath}" --version`, { stdio: 'pipe' });
        console.log('AI Distiller binary already installed and working.');
        return;
      } catch (e) {
        console.log('Existing binary not working, re-downloading...');
        fs.unlinkSync(binaryPath);
      }
    }
    
    try {
      // Download archive
      await downloadFile(url, tempFile);
      console.log('Download complete.');
      
      // Extract archive
      await extractArchive(tempFile, binDir, platform);
      console.log('Extraction complete.');
      
      // Verify binary exists
      if (!fs.existsSync(binaryPath)) {
        throw new Error(`Binary not found after extraction. Expected at: ${binaryPath}`);
      }
      
      // Set executable permissions on Unix
      if (platform !== 'windows') {
        fs.chmodSync(binaryPath, 0o755);
      }
      
      // Verify it works
      try {
        const version = execSync(`"${binaryPath}" --version`, { 
          stdio: 'pipe',
          encoding: 'utf8'
        }).trim();
        console.log(`AI Distiller installed successfully: ${version}`);
      } catch (e) {
        console.warn('Warning: Could not verify binary version, but installation completed.');
      }
      
    } finally {
      // Clean up archive file
      if (fs.existsSync(tempFile)) {
        fs.unlinkSync(tempFile);
      }
    }
    
  } catch (error) {
    console.error('Installation failed:', error.message);
    console.error('\nYou can manually download the binary from:');
    console.error(`https://github.com/janreges/ai-distiller/releases/tag/v${VERSION}`);
    console.error('\nAnd place it in:', path.join(__dirname, '..', 'bin'));
    
    // Exit with error code to fail npm install
    process.exit(1);
  }
}

// Only run if called directly (not required as module)
if (require.main === module) {
  install();
}