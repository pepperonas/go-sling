// File transfer protocol over DataChannel

const MSG_FILE_META = 0x01;
const MSG_FILE_CHUNK = 0x02;
const MSG_FILE_DONE = 0x03;
const MSG_TRANSFER_DONE = 0x04;
const MSG_ACK = 0x05;
const MSG_ERROR = 0x06;
const MSG_PAUSE = 0x07;
const MSG_RESUME = 0x08;
const MSG_CANCEL = 0x09;

const CHUNK_SIZE = 64 * 1024; // 64KB
const ACK_INTERVAL = 16;      // ACK every 16 chunks (1MB)

const Transfer = {
    activeSends: {},    // transferId -> send state
    activeReceives: {}, // peerId -> receive state
    transferCounter: 0,

    async sendFiles(peerId, files) {
        const transferId = 'tx-' + (++this.transferCounter);
        const peerName = WS.peers.find(p => p.id === peerId)?.name || peerId;

        const totalSize = Array.from(files).reduce((s, f) => s + f.size, 0);
        const totalFiles = files.length;

        UI.addTransfer(transferId, totalFiles + ' file(s)', peerName, 'send');

        // Request transfer
        WS.send({
            type: 'transfer-request',
            to: peerId,
            payload: { transferId, totalFiles, totalSize },
        });

        // Wait for acceptance or initiate directly
        const conn = await WebRTCManager.initiateTransfer(peerId);

        const state = {
            id: transferId,
            peerId,
            files: Array.from(files),
            totalSize,
            sentBytes: 0,
            currentFile: 0,
            paused: false,
            cancelled: false,
            startTime: Date.now(),
            dc: null,
        };
        this.activeSends[transferId] = state;

        // Wait for DataChannel to open
        WebRTCManager.onDataChannelOpen(peerId, async (dc) => {
            state.dc = dc;
            await this._sendAllFiles(state);
        });
    },

    async _sendAllFiles(state) {
        const { files, dc } = state;

        for (let i = 0; i < files.length && !state.cancelled; i++) {
            state.currentFile = i;
            const file = files[i];

            // Send file metadata
            const meta = {
                type: MSG_FILE_META,
                fileIndex: i,
                name: file.webkitRelativePath || file.name,
                size: file.size,
                totalFiles: files.length,
            };
            state.dc.send(JSON.stringify(meta));

            // Send chunks
            await this._sendFileChunks(state, file, i);

            if (state.cancelled) break;

            // Send file done with checksum
            const checksum = await this._computeChecksum(file);
            state.dc.send(JSON.stringify({
                type: MSG_FILE_DONE,
                fileIndex: i,
                checksum,
            }));
        }

        if (!state.cancelled) {
            state.dc.send(JSON.stringify({ type: MSG_TRANSFER_DONE }));
            UI.completeTransfer(state.id);
            UI.toast('Transfer complete!', 'success');
            UI.notify('go-sling', 'File transfer complete');
        }

        delete this.activeSends[state.id];
    },

    async _sendFileChunks(state, file, fileIndex) {
        const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
        let chunkIndex = 0;
        let offset = 0;

        while (offset < file.size && !state.cancelled) {
            if (state.paused) {
                await new Promise(resolve => {
                    state._resume = resolve;
                });
            }

            const end = Math.min(offset + CHUNK_SIZE, file.size);
            const chunk = await file.slice(offset, end).arrayBuffer();

            // Wait for bufferedAmount to be low
            while (state.dc.bufferedAmount > 1024 * 1024) { // 1MB buffer limit
                await new Promise(r => setTimeout(r, 10));
            }

            // Create binary message: [type(1)] [fileIndex(4)] [chunkIndex(4)] [data]
            const header = new ArrayBuffer(9);
            const view = new DataView(header);
            view.setUint8(0, MSG_FILE_CHUNK);
            view.setUint32(1, fileIndex);
            view.setUint32(5, chunkIndex);

            const message = new Uint8Array(9 + chunk.byteLength);
            message.set(new Uint8Array(header), 0);
            message.set(new Uint8Array(chunk), 9);

            state.dc.send(message.buffer);

            offset = end;
            chunkIndex++;
            state.sentBytes += chunk.byteLength;

            // Update progress
            const percent = (state.sentBytes / state.totalSize) * 100;
            const elapsed = (Date.now() - state.startTime) / 1000;
            const speed = state.sentBytes / elapsed;
            const remaining = (state.totalSize - state.sentBytes) / speed;

            UI.updateTransfer(state.id, percent, speed, remaining);
        }
    },

    async _computeChecksum(file) {
        // crypto.subtle is only available over HTTPS/localhost
        if (!crypto.subtle) return '';
        try {
            const buffer = await file.arrayBuffer();
            const hash = await crypto.subtle.digest('SHA-256', buffer);
            return Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2, '0')).join('');
        } catch {
            return '';
        }
    },

    // Receiving side
    handleIncoming(peerId, data) {
        if (typeof data === 'string') {
            const msg = JSON.parse(data);
            this._handleControlMessage(peerId, msg);
            return;
        }

        // Binary data = file chunk
        const view = new DataView(data);
        const type = view.getUint8(0);

        if (type === MSG_FILE_CHUNK) {
            const fileIndex = view.getUint32(1);
            const chunkIndex = view.getUint32(5);
            const chunkData = new Uint8Array(data, 9);

            this._handleChunk(peerId, fileIndex, chunkIndex, chunkData);
        }
    },

    _handleControlMessage(peerId, msg) {
        const type = Number(msg.type);
        switch (type) {
            case MSG_FILE_META: {
                if (!this.activeReceives[peerId]) {
                    const peerName = WS.peers.find(p => p.id === peerId)?.name || peerId;
                    this.activeReceives[peerId] = {
                        peerId,
                        transferId: 'rx-' + (++this.transferCounter),
                        files: {},
                        totalFiles: msg.totalFiles,
                        totalSize: 0,
                        receivedBytes: 0,
                        startTime: Date.now(),
                        peerName,
                    };
                    UI.addTransfer(
                        this.activeReceives[peerId].transferId,
                        msg.totalFiles + ' file(s)',
                        peerName,
                        'receive'
                    );
                }

                const rx = this.activeReceives[peerId];
                rx.totalSize += msg.size;
                rx.files[msg.fileIndex] = {
                    name: msg.name,
                    size: msg.size,
                    chunks: [],
                    received: 0,
                };
                break;
            }

            case MSG_FILE_DONE: {
                const rx = this.activeReceives[peerId];
                if (!rx) return;

                const file = rx.files[msg.fileIndex];
                if (!file) return;

                // Assemble file
                const blob = new Blob(file.chunks);

                // Verify checksum for small files (< 100MB), only if crypto.subtle available (HTTPS)
                if (crypto.subtle && blob.size < 100 * 1024 * 1024 && msg.checksum) {
                    blob.arrayBuffer().then(buf => {
                        crypto.subtle.digest('SHA-256', buf).then(hash => {
                            const checksum = Array.from(new Uint8Array(hash))
                                .map(b => b.toString(16).padStart(2, '0')).join('');
                            if (checksum !== msg.checksum) {
                                UI.toast('Checksum mismatch for ' + file.name, 'warning');
                            }
                        });
                    }).catch(() => {});
                }

                // Store blob for ZIP bundling at transfer end
                if (!rx.completedFiles) rx.completedFiles = [];
                rx.completedFiles.push({ name: file.name, blob });
                break;
            }

            case MSG_TRANSFER_DONE: {
                const rx = this.activeReceives[peerId];
                if (rx) {
                    UI.completeTransfer(rx.transferId);
                    this._bundleAndOffer(rx);
                    UI.notify('go-sling', 'File transfer received');
                    delete this.activeReceives[peerId];
                }
                break;
            }

            case MSG_CANCEL: {
                const rx = this.activeReceives[peerId];
                if (rx) {
                    UI.toast('Transfer cancelled by peer', 'warning');
                    delete this.activeReceives[peerId];
                }
                break;
            }
        }
    },

    _handleChunk(peerId, fileIndex, chunkIndex, data) {
        const rx = this.activeReceives[peerId];
        if (!rx || !rx.files[fileIndex]) return;

        const file = rx.files[fileIndex];
        file.chunks[chunkIndex] = data;
        file.received += data.byteLength;
        rx.receivedBytes += data.byteLength;

        // Update progress
        const percent = (rx.receivedBytes / rx.totalSize) * 100;
        const elapsed = (Date.now() - rx.startTime) / 1000;
        const speed = rx.receivedBytes / elapsed;
        const remaining = (rx.totalSize - rx.receivedBytes) / speed;

        UI.updateTransfer(rx.transferId, percent, speed, remaining);
    },

    async _bundleAndOffer(rx) {
        const files = rx.completedFiles || [];
        if (files.length === 0) return;

        const container = document.getElementById('transfers');
        const item = document.createElement('div');
        item.className = 'transfer-item';
        item.style.borderLeft = '3px solid var(--success)';

        const header = document.createElement('div');
        header.className = 'transfer-header';

        if (files.length === 1) {
            const f = files[0];
            const url = URL.createObjectURL(f.blob);
            const fileName = f.name.split('/').pop();

            const label = document.createElement('span');
            label.className = 'transfer-name';
            label.textContent = '✅ ' + fileName + ' (' + formatBytes(f.blob.size) + ')';
            header.appendChild(label);
            item.appendChild(header);

            const btn = document.createElement('a');
            btn.href = url;
            btn.download = fileName;
            btn.className = 'btn btn-primary';
            btn.style.cssText = 'margin-top:8px;display:inline-flex;text-decoration:none';
            btn.textContent = '💾 Save File';
            item.appendChild(btn);

            container.insertBefore(item, container.firstChild);
            UI.toast('File ready — tap Save', 'success', 6000);
        } else {
            const label = document.createElement('span');
            label.className = 'transfer-name';
            label.textContent = '✅ ' + files.length + ' files received from ' + rx.peerName;
            header.appendChild(label);
            item.appendChild(header);

            const status = document.createElement('div');
            status.textContent = 'Creating ZIP...';
            status.style.cssText = 'font-size:13px;color:var(--text-secondary);margin-top:8px';
            item.appendChild(status);
            container.insertBefore(item, container.firstChild);

            const zipFiles = [];
            for (const f of files) {
                const buf = await f.blob.arrayBuffer();
                zipFiles.push({ name: f.name, data: new Uint8Array(buf) });
            }

            const zipBlob = ZipBuilder.create(zipFiles);
            const url = URL.createObjectURL(zipBlob);

            // Derive ZIP name from common directory prefix
            const firstPath = files[0].name;
            const rootDir = firstPath.includes('/') ? firstPath.split('/')[0] : 'transfer';
            const zipName = rootDir + '.zip';

            status.textContent = files.length + ' files — ' + formatBytes(zipBlob.size);

            const btn = document.createElement('a');
            btn.href = url;
            btn.download = zipName;
            btn.className = 'btn btn-primary';
            btn.style.cssText = 'margin-top:8px;display:inline-flex;text-decoration:none';
            btn.textContent = '💾 Save ZIP (' + formatBytes(zipBlob.size) + ')';
            item.appendChild(btn);

            UI.toast(files.length + ' files ready — tap Save ZIP', 'success', 6000);
        }
    },

    cancelTransfer(transferId) {
        // Check sends
        const send = this.activeSends[transferId];
        if (send) {
            send.cancelled = true;
            if (send.dc) {
                send.dc.send(JSON.stringify({ type: MSG_CANCEL }));
            }
            delete this.activeSends[transferId];
            UI.toast('Transfer cancelled', 'info');
            return;
        }

        // Check receives
        for (const [peerId, rx] of Object.entries(this.activeReceives)) {
            if (rx.transferId === transferId) {
                const dc = WebRTCManager.getDataChannel(peerId);
                if (dc) dc.send(JSON.stringify({ type: MSG_CANCEL }));
                delete this.activeReceives[peerId];
                UI.toast('Transfer cancelled', 'info');
                return;
            }
        }
    },

    pauseTransfer(transferId) {
        const send = this.activeSends[transferId];
        if (send) send.paused = true;
    },

    resumeTransfer(transferId) {
        const send = this.activeSends[transferId];
        if (send) {
            send.paused = false;
            if (send._resume) send._resume();
        }
    },
};
