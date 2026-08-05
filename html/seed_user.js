#!/usr/bin/env node
// Create or update a user in the SQLite user store.
// Usage: node seed_user.js --username admin
// You'll be prompted for a password (input hidden, never logged).
const readline = require('readline');
const auth = require('./auth_db');

function parseArgs() {
  const args = process.argv.slice(2);
  const idx = args.indexOf('--username');
  if (idx === -1 || !args[idx + 1]) {
    console.error('Usage: node seed_user.js --username <name>');
    process.exit(1);
  }
  return args[idx + 1];
}

function readHiddenLine(prompt) {
  return new Promise((resolve) => {
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
    const stdin = process.stdin;
    process.stdout.write(prompt);

    let value = '';
    const onData = (char) => {
      char = char.toString();
      if (char === '\n' || char === '\r' || char === '') {
        stdin.removeListener('data', onData);
        stdin.setRawMode && stdin.setRawMode(false);
        process.stdout.write('\n');
        rl.close();
        resolve(value);
        return;
      }
      if (char === '') {
        process.exit(1);
      }
      if (char === '') {
        value = value.slice(0, -1);
        return;
      }
      value += char;
    };

    if (stdin.isTTY) stdin.setRawMode(true);
    stdin.resume();
    stdin.setEncoding('utf8');
    stdin.on('data', onData);
  });
}

async function main() {
  const username = parseArgs();
  const password = await readHiddenLine('Password: ');
  const confirm = await readHiddenLine('Confirm password: ');

  if (password !== confirm) {
    console.error('Passwords do not match.');
    process.exit(1);
  }
  if (password.length < 8) {
    console.error('Password must be at least 8 characters.');
    process.exit(1);
  }

  auth.upsertUser(username, password);
  console.log(`User '${username}' created/updated.`);
  process.exit(0);
}

main();
