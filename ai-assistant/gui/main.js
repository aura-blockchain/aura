const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('node:path');
const { spawn } = require('node:child_process');
const keytar = require('keytar');

const isMac = process.platform === 'darwin';
const SERVICE_NAME = 'com.aura.assistant';

function createWindow() {
  const win = new BrowserWindow({
    width: 1100,
    height: 800,
    title: 'Aura Assistant Control Center',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  win.loadFile(path.join(__dirname, 'renderer', 'index.html'));
}

app.whenReady().then(() => {
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (!isMac) {
    app.quit();
  }
});

ipcMain.handle('assistant:openFile', async () => {
  const result = await dialog.showOpenDialog({
    properties: ['openFile'],
  });
  if (result.canceled || result.filePaths.length === 0) {
    return null;
  }
  return result.filePaths[0];
});

ipcMain.handle('assistant:runCommand', async (_event, payload) => {
  const { command, args = [], cwd, env = {} } = payload;
  if (!command) {
    throw new Error('command is required');
  }
  return runCommand(command, args, cwd, env);
});

ipcMain.handle('assistant:saveSecret', async (_event, { key, value }) => {
  if (!key || typeof value !== 'string') {
    throw new Error('key and value required');
  }
  await keytar.setPassword(SERVICE_NAME, key, value);
  return true;
});

ipcMain.handle('assistant:readSecret', async (_event, key) => {
  if (!key) {
    return null;
  }
  return keytar.getPassword(SERVICE_NAME, key);
});

ipcMain.handle('assistant:deleteSecret', async (_event, key) => {
  if (!key) {
    return false;
  }
  return keytar.deletePassword(SERVICE_NAME, key);
});

function runCommand(command, args, cwd, env) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: cwd || process.cwd(),
      env: { ...process.env, ...env },
      shell: false,
    });
    let stdout = '';
    let stderr = '';
    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });
    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });
    child.on('error', (err) => {
      reject(err);
    });
    child.on('close', (code) => {
      resolve({ stdout, stderr, code });
    });
  });
}
