// Main application logic

const App = {
    stagedFiles: [],
    statusInterval: null,
    filesInterval: null,

    async start() {
        // Check if auth is required
        const authRequired = await Auth.checkRequired();

        if (authRequired) {
            document.getElementById('login-screen').style.display = 'flex';
            Auth.setupLoginForm();

            // Try accessing API to check if already authenticated
            try {
                const resp = await fetch('/api/status');
                if (resp.ok) {
                    document.getElementById('login-screen').style.display = 'none';
                    document.querySelector('.app').classList.add('active');
                    this.init();
                }
            } catch {}
        } else {
            document.querySelector('.app').classList.add('active');
            this.init();
        }
    },

    init() {
        UI.init();
        UI.requestNotificationPermission();
        this.setupTheme();
        this.setupTabs();
        this.setupDropZone();
        this.setupWebSocket();
        this.setupWebRTC();
        this.startPolling();
        this.loadFiles();
    },

    setupTheme() {
        const saved = localStorage.getItem('gosling-theme') || 'dark';
        document.documentElement.setAttribute('data-theme', saved);
        document.getElementById('theme-toggle').addEventListener('click', () => {
            const current = document.documentElement.getAttribute('data-theme');
            const next = current === 'dark' ? 'light' : 'dark';
            document.documentElement.setAttribute('data-theme', next);
            localStorage.setItem('gosling-theme', next);
            document.getElementById('theme-toggle').textContent = next === 'dark' ? '☀️' : '🌙';
        });
        document.getElementById('theme-toggle').textContent = saved === 'dark' ? '☀️' : '🌙';
    },

    setupTabs() {
        document.querySelectorAll('.tab').forEach(tab => {
            tab.addEventListener('click', () => UI.switchTab(tab.dataset.tab));
        });
    },

    setupDropZone() {
        const sendZone = document.getElementById('drop-zone-send');
        const filesZone = document.getElementById('drop-zone-files');

        [sendZone, filesZone].forEach(zone => {
            if (!zone) return;

            zone.addEventListener('dragover', (e) => {
                e.preventDefault();
                zone.classList.add('dragover');
            });

            zone.addEventListener('dragleave', () => {
                zone.classList.remove('dragover');
            });

            zone.addEventListener('drop', (e) => {
                e.preventDefault();
                zone.classList.remove('dragover');

                const items = e.dataTransfer.items;
                if (items) {
                    this._handleDropItems(items, zone.id === 'drop-zone-files');
                } else {
                    const files = e.dataTransfer.files;
                    if (zone.id === 'drop-zone-files') {
                        this.uploadToServer(files);
                    } else {
                        this.stageFiles(files);
                    }
                }
            });

            zone.addEventListener('click', () => {
                const input = zone.querySelector('input[type="file"]');
                if (input) input.click();
            });
        });

        // File input handlers
        const sendInput = document.getElementById('file-input-send');
        if (sendInput) {
            sendInput.addEventListener('change', (e) => {
                this.stageFiles(e.target.files);
                e.target.value = '';
            });
        }

        const filesInput = document.getElementById('file-input-files');
        if (filesInput) {
            filesInput.addEventListener('change', (e) => {
                this.uploadToServer(e.target.files);
                e.target.value = '';
            });
        }

        // Send button
        document.getElementById('btn-send-p2p').addEventListener('click', () => this.sendP2P());
        document.getElementById('btn-clear-staged').addEventListener('click', () => {
            this.stagedFiles = [];
            UI.renderStagedFiles(this.stagedFiles);
        });
    },

    async _handleDropItems(items, uploadToServer) {
        const files = [];

        const readEntry = (entry) => {
            return new Promise((resolve) => {
                if (entry.isFile) {
                    entry.file(f => {
                        // Preserve relative path
                        Object.defineProperty(f, 'webkitRelativePath', {
                            value: entry.fullPath.substring(1), // Remove leading /
                            writable: false,
                        });
                        files.push(f);
                        resolve();
                    });
                } else if (entry.isDirectory) {
                    const reader = entry.createReader();
                    reader.readEntries(async (entries) => {
                        for (const e of entries) {
                            await readEntry(e);
                        }
                        resolve();
                    });
                } else {
                    resolve();
                }
            });
        };

        const promises = [];
        for (let i = 0; i < items.length; i++) {
            const entry = items[i].webkitGetAsEntry ? items[i].webkitGetAsEntry() : null;
            if (entry) {
                promises.push(readEntry(entry));
            } else if (items[i].kind === 'file') {
                files.push(items[i].getAsFile());
            }
        }

        await Promise.all(promises);

        if (uploadToServer) {
            this.uploadToServer(files);
        } else {
            this.stageFiles(files);
        }
    },

    stageFiles(files) {
        const fileList = Array.from(files);
        this.stagedFiles.push(...fileList);
        UI.renderStagedFiles(this.stagedFiles);
        UI.toast(fileList.length + ' file(s) staged for P2P transfer', 'info');
    },

    removeStagedFile(index) {
        this.stagedFiles.splice(index, 1);
        UI.renderStagedFiles(this.stagedFiles);
    },

    async sendP2P() {
        if (this.stagedFiles.length === 0) {
            UI.toast('No files staged for transfer', 'warning');
            return;
        }

        const targetPeer = document.getElementById('target-peer').value;
        if (!targetPeer) {
            UI.toast('Select a target peer first', 'warning');
            return;
        }

        try {
            await Transfer.sendFiles(targetPeer, this.stagedFiles);
            this.stagedFiles = [];
            UI.renderStagedFiles(this.stagedFiles);
        } catch (err) {
            UI.toast('Transfer failed: ' + err.message, 'error');
        }
    },

    async uploadToServer(files) {
        const formData = new FormData();
        const fileList = Array.isArray(files) ? files : Array.from(files);
        fileList.forEach(f => formData.append('files', f, f.webkitRelativePath || f.name));

        try {
            UI.toast('Uploading ' + fileList.length + ' file(s)...', 'info');
            const resp = await fetch('/api/upload', { method: 'POST', body: formData });

            if (!resp.ok) {
                const err = await resp.json();
                throw new Error(err.error || 'Upload failed');
            }

            UI.toast('Upload complete!', 'success');
            this.loadFiles();
        } catch (err) {
            UI.toast('Upload failed: ' + err.message, 'error');
        }
    },

    async loadFiles() {
        try {
            const resp = await fetch('/api/files');
            if (!resp.ok) return;
            const data = await resp.json();
            UI.renderFiles(data.files);
        } catch {}
    },

    downloadFile(id) {
        window.open('/api/download/' + encodeURIComponent(id), '_blank');
    },

    deleteFile(id) {
        UI.showModal('Delete File', 'Are you sure you want to delete this file?', async () => {
            try {
                const resp = await fetch('/api/files/' + encodeURIComponent(id), { method: 'DELETE' });
                if (resp.ok) {
                    UI.toast('File deleted', 'success');
                    this.loadFiles();
                }
            } catch (err) {
                UI.toast('Delete failed', 'error');
            }
        });
    },

    cancelTransfer(transferId) {
        Transfer.cancelTransfer(transferId);
    },

    setupWebSocket() {
        WS.connect();
    },

    setupWebRTC() {
        // Handle incoming signaling messages
        WS.on('offer', async (msg) => {
            await WebRTCManager.handleOffer(msg.from, msg.payload.sdp);
        });

        WS.on('answer', async (msg) => {
            await WebRTCManager.handleAnswer(msg.from, msg.payload.sdp);
        });

        WS.on('ice-candidate', async (msg) => {
            await WebRTCManager.handleIceCandidate(msg.from, msg.payload.candidate);
        });

        WS.on('transfer-request', (msg) => {
            const peerName = WS.peers.find(p => p.id === msg.from)?.name || msg.from;
            UI.toast(peerName + ' wants to send you files', 'info');
        });
    },

    startPolling() {
        // Status polling
        const pollStatus = async () => {
            try {
                const resp = await fetch('/api/status');
                if (resp.ok) {
                    const data = await resp.json();
                    UI.updateStatus(data);
                }
            } catch {}
        };

        pollStatus();
        this.statusInterval = setInterval(pollStatus, 5000);

        // Files polling
        this.filesInterval = setInterval(() => this.loadFiles(), 10000);
    }
};

// Start app when DOM is ready
document.addEventListener('DOMContentLoaded', () => App.start());
