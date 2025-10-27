#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const readline = require('readline');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

function question(prompt) {
  return new Promise((resolve) => {
    rl.question(prompt, resolve);
  });
}

function exec(command, options = {}) {
  console.log(`> ${command}`);
  return execSync(command, { stdio: 'inherit', ...options });
}

async function main() {
  try {
    console.log('AI Distiller MCP Release Script');
    console.log('================================\n');

    // Check if we're in the right directory
    const packageJsonPath = path.join(process.cwd(), 'package.json');
    if (!fs.existsSync(packageJsonPath)) {
      throw new Error('package.json not found. Please run this script from the mcp-npm directory.');
    }

    const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
    const currentVersion = packageJson.version;

    console.log(`Current version: ${currentVersion}`);

    // Ask for new version
    const newVersion = await question('Enter new version (or press Enter to keep current): ');
    const version = newVersion.trim() || currentVersion;

    if (version !== currentVersion) {
      // Update package.json
      packageJson.version = version;
      fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 2) + '\n');
      console.log(`Updated package.json to version ${version}`);

      // Update postinstall.js
      const postinstallPath = path.join(process.cwd(), 'scripts', 'postinstall.js');
      let postinstallContent = fs.readFileSync(postinstallPath, 'utf8');
      postinstallContent = postinstallContent.replace(
        /const VERSION = '[^']+'/,
        `const VERSION = '${version}'`
      );
      fs.writeFileSync(postinstallPath, postinstallContent);
      console.log(`Updated postinstall.js to version ${version}`);
    }

    // Check for uncommitted changes
    try {
      execSync('git diff-index --quiet HEAD --', { stdio: 'pipe' });
    } catch (e) {
      const proceed = await question('\nYou have uncommitted changes. Continue anyway? (y/N): ');
      if (proceed.toLowerCase() !== 'y') {
        console.log('Aborted.');
        process.exit(1);
      }
    }

    // Run npm install to ensure dependencies are up to date
    console.log('\nInstalling dependencies...');
    exec('npm install');

    // Create test package
    console.log('\nCreating test package...');
    exec('npm pack');

    const tarballName = `cognitive-ai-distiller-mcp-${version}.tgz`;
    console.log(`\nCreated ${tarballName}`);

    // Test installation
    const testInstall = await question('\nTest installation locally? (Y/n): ');
    if (testInstall.toLowerCase() !== 'n') {
      const testDir = path.join('/tmp', `test-aid-mcp-${Date.now()}`);
      fs.mkdirSync(testDir, { recursive: true });

      console.log(`\nTesting in ${testDir}...`);
      process.chdir(testDir);

      try {
        exec(`npm install ${path.join(path.dirname(packageJsonPath), tarballName)}`);
        console.log('\nInstallation test passed!');

        // Try to run the binary
        const testRun = await question('\nTest running the MCP server? (Y/n): ');
        if (testRun.toLowerCase() !== 'n') {
          exec('npx aid-mcp --version', { stdio: 'pipe' });
          console.log('Binary test passed!');
        }
      } catch (e) {
        console.error('Test failed:', e.message);
        const continueAnyway = await question('\nTest failed. Continue with publish? (y/N): ');
        if (continueAnyway.toLowerCase() !== 'y') {
          process.exit(1);
        }
      }

      process.chdir(path.dirname(packageJsonPath));
    }

    // Dry run
    console.log('\nRunning npm publish dry run...');
    exec('npm publish --dry-run');

    // Confirm publish
    const confirmPublish = await question('\nReady to publish to npm? (y/N): ');
    if (confirmPublish.toLowerCase() !== 'y') {
      console.log('Aborted.');
      process.exit(0);
    }

    // Publish
    console.log('\nPublishing to npm...');
    exec('npm publish --access public');

    console.log('\n✅ Successfully published to npm!');

    // Git operations
    const doGit = await question('\nCommit and tag this release? (Y/n): ');
    if (doGit.toLowerCase() !== 'n') {
      exec('git add .');
      exec(`git commit -m "chore: release AI Distiller MCP v${version}"`);
      exec(`git tag mcp-v${version}`);

      const push = await question('\nPush to git? (Y/n): ');
      if (push.toLowerCase() !== 'n') {
        exec('git push');
        exec('git push --tags');
      }
    }

    // Cleanup
    if (fs.existsSync(tarballName)) {
      fs.unlinkSync(tarballName);
    }

    console.log('\n🎉 Release complete!');
    console.log(`\nUsers can now install with: npm install @cognitive/ai-distiller-mcp`);

  } catch (error) {
    console.error('\n❌ Error:', error.message);
    process.exit(1);
  } finally {
    rl.close();
  }
}

main();