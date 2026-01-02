#!/usr/bin/env node
// Sync the shared grammar file to the vscode extension directory
const fs = require('fs');
const path = require('path');

const sourceFile = path.join(__dirname, '..', 'shared', 'ensurascript.tmLanguage.json');
const targetDir = path.join(__dirname, 'grammars');
const targetFile = path.join(targetDir, 'ensurascript.tmLanguage.json');

// Create grammars directory if it doesn't exist
if (!fs.existsSync(targetDir)) {
  fs.mkdirSync(targetDir, { recursive: true });
}

// Copy the file
fs.copyFileSync(sourceFile, targetFile);

console.log('✓ Grammar file synced from ../shared/ to grammars/');
