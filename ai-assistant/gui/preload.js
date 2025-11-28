const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('assistantBridge', {
  runCommand: (command, args = [], options = {}) =>
    ipcRenderer.invoke('assistant:runCommand', { command, args, ...options }),
  openFileDialog: () => ipcRenderer.invoke('assistant:openFile'),
  saveSecret: (key, value) => ipcRenderer.invoke('assistant:saveSecret', { key, value }),
  readSecret: (key) => ipcRenderer.invoke('assistant:readSecret', key),
  deleteSecret: (key) => ipcRenderer.invoke('assistant:deleteSecret', key),
});
