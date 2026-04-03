// UI management and notifications

const UI = {
    toastContainer: null,

    init() {
        this.toastContainer = document.getElementById('toast-container');
    },

    // Escape HTML to prevent XSS
    esc(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    },

    toast(message, type = 'info', duration = 4000) {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;

        const icons = { success: '✓', error: '✗', info: 'ℹ', warning: '⚠' };
        const iconEl = document.createElement('span');
        iconEl.textContent = icons[type] || 'ℹ';
        const msgEl = document.createElement('span');
        msgEl.textContent = message;

        toast.appendChild(iconEl);
        toast.appendChild(msgEl);
        this.toastContainer.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'slideOut 0.3s ease forwards';
            setTimeout(() => toast.remove(), 300);
        }, duration);
    },

    notify(title, body) {
        if (Notification.permission === 'granted') {
            new Notification(title, { body, icon: '/assets/favicon.svg' });
        }
    },

    requestNotificationPermission() {
        if ('Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }
    },

    switchTab(tabName) {
        document.querySelectorAll('.tab').forEach(t => t.classList.toggle('active', t.dataset.tab === tabName));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.toggle('active', c.id === 'tab-' + tabName));
    },

    showModal(title, message, onConfirm) {
        const overlay = document.getElementById('modal-overlay');
        overlay.querySelector('h2').textContent = title;
        overlay.querySelector('p').textContent = message;
        overlay.classList.add('active');

        const confirmBtn = overlay.querySelector('.btn-danger');
        const cancelBtn = overlay.querySelector('.btn-secondary');

        const cleanup = () => {
            overlay.classList.remove('active');
            confirmBtn.removeEventListener('click', handleConfirm);
            cancelBtn.removeEventListener('click', cleanup);
        };

        const handleConfirm = () => { cleanup(); onConfirm(); };

        confirmBtn.addEventListener('click', handleConfirm);
        cancelBtn.addEventListener('click', cleanup);
    },

    updatePeerCount(count) {
        document.getElementById('peer-count-num').textContent = count;
    },

    _createPeerItem(p, myId) {
        const li = document.createElement('li');
        li.className = 'peer-item' + (p.id === myId ? ' self' : '');
        li.dataset.peerId = p.id;

        const avatar = document.createElement('div');
        avatar.className = 'peer-avatar';
        avatar.textContent = getOSIcon(p.os);

        const info = document.createElement('div');
        info.className = 'peer-info';

        const name = document.createElement('div');
        name.className = 'peer-name';
        name.textContent = p.name + (p.id === myId ? ' (You)' : '');

        const meta = document.createElement('div');
        meta.className = 'peer-meta';
        meta.textContent = (p.browser || '') + ' · ' + (p.os || '');

        info.appendChild(name);
        info.appendChild(meta);
        li.appendChild(avatar);
        li.appendChild(info);

        return li;
    },

    renderPeerList(peers, myId) {
        const list = document.getElementById('peer-list');
        const select = document.getElementById('target-peer');

        list.innerHTML = '';

        if (peers.length === 0) {
            const li = document.createElement('li');
            li.className = 'no-peers';
            li.textContent = 'No other devices connected';
            list.appendChild(li);
            select.innerHTML = '';
            const opt = document.createElement('option');
            opt.value = '';
            opt.textContent = 'No peers available';
            select.appendChild(opt);
            return;
        }

        peers.forEach(p => list.appendChild(this._createPeerItem(p, myId)));

        select.innerHTML = '';
        const otherPeers = peers.filter(p => p.id !== myId);
        if (otherPeers.length === 0) {
            const opt = document.createElement('option');
            opt.value = '';
            opt.textContent = 'No other peers';
            select.appendChild(opt);
        } else {
            const placeholder = document.createElement('option');
            placeholder.value = '';
            placeholder.textContent = 'Select a peer...';
            select.appendChild(placeholder);
            otherPeers.forEach(p => {
                const opt = document.createElement('option');
                opt.value = p.id;
                opt.textContent = p.name;
                select.appendChild(opt);
            });
        }
    },

    _createFileRow(f) {
        const row = document.createElement('div');
        row.className = 'file-row';
        row.dataset.fileId = f.id;

        const icon = document.createElement('span');
        icon.className = 'file-icon';
        icon.textContent = getFileIcon(f.name, f.isDir);

        const info = document.createElement('div');
        info.className = 'file-info';

        const name = document.createElement('div');
        name.className = 'file-name';
        name.textContent = f.name;

        const meta = document.createElement('div');
        meta.className = 'file-meta';

        const size = document.createElement('span');
        size.textContent = formatBytes(f.size);
        const time = document.createElement('span');
        time.textContent = formatTime(f.uploadedAt);
        const expiry = document.createElement('span');
        expiry.textContent = 'Expires: ' + formatCountdown(f.expiresAt);

        meta.appendChild(size);
        meta.appendChild(time);
        meta.appendChild(expiry);
        info.appendChild(name);
        info.appendChild(meta);

        const actions = document.createElement('div');
        actions.className = 'file-actions';

        const dlBtn = document.createElement('button');
        dlBtn.className = 'btn-icon';
        dlBtn.title = 'Download';
        dlBtn.textContent = '⬇️';
        dlBtn.addEventListener('click', () => App.downloadFile(f.id));

        const delBtn = document.createElement('button');
        delBtn.className = 'btn-icon';
        delBtn.title = 'Delete';
        delBtn.textContent = '🗑️';
        delBtn.addEventListener('click', () => App.deleteFile(f.id));

        actions.appendChild(dlBtn);
        actions.appendChild(delBtn);

        row.appendChild(icon);
        row.appendChild(info);
        row.appendChild(actions);

        return row;
    },

    renderFiles(files) {
        const container = document.getElementById('files-list');
        container.innerHTML = '';

        if (!files || files.length === 0) {
            const empty = document.createElement('div');
            empty.className = 'empty-state';

            const emIcon = document.createElement('div');
            emIcon.className = 'empty-icon';
            emIcon.textContent = '📂';
            const p1 = document.createElement('p');
            p1.textContent = 'No files uploaded yet';
            const p2 = document.createElement('p');
            p2.textContent = 'Drop files here or use the Send tab for P2P transfer';

            empty.appendChild(emIcon);
            empty.appendChild(p1);
            empty.appendChild(p2);
            container.appendChild(empty);
            return;
        }

        files.forEach(f => container.appendChild(this._createFileRow(f)));
    },

    renderStagedFiles(files) {
        const container = document.getElementById('staged-files');
        const list = document.getElementById('staged-list');

        if (files.length === 0) {
            container.classList.remove('has-files');
            return;
        }

        container.classList.add('has-files');
        document.getElementById('staged-count').textContent = files.length + ' file' + (files.length !== 1 ? 's' : '');

        list.innerHTML = '';
        let totalSize = 0;
        files.forEach((f, i) => {
            totalSize += f.size;
            const li = document.createElement('li');
            li.className = 'staged-item';

            const icon = document.createElement('span');
            icon.className = 'file-icon';
            icon.textContent = getFileIcon(f.name, false);

            const name = document.createElement('span');
            name.textContent = f.webkitRelativePath || f.name;

            const size = document.createElement('span');
            size.className = 'file-size';
            size.textContent = formatBytes(f.size);

            const removeBtn = document.createElement('button');
            removeBtn.className = 'btn-icon';
            removeBtn.title = 'Remove';
            removeBtn.textContent = '✕';
            removeBtn.addEventListener('click', () => App.removeStagedFile(i));

            li.appendChild(icon);
            li.appendChild(name);
            li.appendChild(size);
            li.appendChild(removeBtn);
            list.appendChild(li);
        });

        document.getElementById('staged-total').textContent = formatBytes(totalSize);
    },

    addTransfer(id, name, peerName, direction) {
        const container = document.getElementById('transfers');
        const item = document.createElement('div');
        item.className = 'transfer-item';
        item.id = 'transfer-' + id;

        const header = document.createElement('div');
        header.className = 'transfer-header';

        const tName = document.createElement('span');
        tName.className = 'transfer-name';
        tName.textContent = (direction === 'send' ? '⬆️ ' : '⬇️ ') + name;

        const tPeer = document.createElement('span');
        tPeer.className = 'transfer-peer';
        tPeer.textContent = (direction === 'send' ? 'to ' : 'from ') + peerName;

        header.appendChild(tName);
        header.appendChild(tPeer);

        const progressBar = document.createElement('div');
        progressBar.className = 'progress-bar';
        const progressFill = document.createElement('div');
        progressFill.className = 'progress-fill';
        progressFill.style.width = '0%';
        progressBar.appendChild(progressFill);

        const stats = document.createElement('div');
        stats.className = 'transfer-stats';
        const pct = document.createElement('span');
        pct.className = 'transfer-progress';
        pct.textContent = '0%';
        const spd = document.createElement('span');
        spd.className = 'transfer-speed';
        spd.textContent = '--';
        const eta = document.createElement('span');
        eta.className = 'transfer-eta';
        eta.textContent = '--';
        stats.appendChild(pct);
        stats.appendChild(spd);
        stats.appendChild(eta);

        const actions = document.createElement('div');
        actions.className = 'transfer-actions';
        const cancelBtn = document.createElement('button');
        cancelBtn.className = 'btn btn-secondary';
        cancelBtn.style.cssText = 'padding:6px 12px;font-size:12px';
        cancelBtn.textContent = 'Cancel';
        cancelBtn.addEventListener('click', () => App.cancelTransfer(id));
        actions.appendChild(cancelBtn);

        item.appendChild(header);
        item.appendChild(progressBar);
        item.appendChild(stats);
        item.appendChild(actions);

        container.insertBefore(item, container.firstChild);
    },

    updateTransfer(id, percent, speed, eta) {
        const el = document.getElementById('transfer-' + id);
        if (!el) return;
        el.querySelector('.progress-fill').style.width = percent + '%';
        el.querySelector('.transfer-progress').textContent = Math.round(percent) + '%';
        el.querySelector('.transfer-speed').textContent = speed ? formatSpeed(speed) : '--';
        el.querySelector('.transfer-eta').textContent = eta ? formatETA(eta) : '--';
    },

    completeTransfer(id) {
        const el = document.getElementById('transfer-' + id);
        if (!el) return;
        el.querySelector('.progress-fill').style.width = '100%';
        el.querySelector('.progress-fill').classList.add('complete');
        el.querySelector('.transfer-progress').textContent = 'Complete';
        el.querySelector('.transfer-speed').textContent = '';
        el.querySelector('.transfer-eta').textContent = '';
        el.querySelector('.transfer-actions').innerHTML = '';
    },

    updateStatus(status) {
        document.getElementById('status-uptime').textContent = status.uptime || '--';
        document.getElementById('status-peers').textContent = status.connectedPeers || 0;
        document.getElementById('status-files').textContent = status.fileCount || 0;
        document.getElementById('status-storage').textContent = formatBytes(status.storageUsed || 0);
        document.getElementById('status-memory').textContent = formatBytes(status.memAlloc || 0);
    }
};
