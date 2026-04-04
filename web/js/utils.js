// Utility functions

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function formatSpeed(bytesPerSec) {
    return formatBytes(bytesPerSec) + '/s';
}

function formatETA(seconds) {
    if (!isFinite(seconds) || seconds < 0) return '--';
    if (seconds < 60) return Math.ceil(seconds) + 's';
    if (seconds < 3600) return Math.floor(seconds / 60) + 'm ' + Math.ceil(seconds % 60) + 's';
    return Math.floor(seconds / 3600) + 'h ' + Math.floor((seconds % 3600) / 60) + 'm';
}

function formatTime(timestamp) {
    const d = new Date(timestamp);
    const now = new Date();
    const diff = (now - d) / 1000;

    if (diff < 60) return 'just now';
    if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
    if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
    return d.toLocaleDateString();
}

function formatCountdown(expiresAt) {
    const remaining = new Date(expiresAt) - new Date();
    if (remaining <= 0) return 'expired';
    const hours = Math.floor(remaining / 3600000);
    const mins = Math.floor((remaining % 3600000) / 60000);
    return hours + 'h ' + mins + 'm';
}

function getFileIcon(name, isDir) {
    if (isDir) return '📁';
    const ext = name.split('.').pop().toLowerCase();
    const icons = {
        pdf: '📄', doc: '📝', docx: '📝', txt: '📝',
        jpg: '🖼️', jpeg: '🖼️', png: '🖼️', gif: '🖼️', svg: '🖼️', webp: '🖼️',
        mp4: '🎬', mov: '🎬', avi: '🎬', mkv: '🎬', webm: '🎬',
        mp3: '🎵', wav: '🎵', flac: '🎵', ogg: '🎵',
        zip: '📦', tar: '📦', gz: '📦', rar: '📦', '7z': '📦',
        js: '📜', ts: '📜', py: '📜', go: '📜', rs: '📜',
        html: '🌐', css: '🎨', json: '📋', yaml: '📋', yml: '📋',
    };
    return icons[ext] || '📄';
}

function getDeviceIcon(device) {
    if (!device) return '💻';
    const d = device.toLowerCase();
    if (d.includes('iphone')) return '📱';
    if (d.includes('ipad')) return '📲';
    if (d.includes('android phone')) return '📱';
    if (d.includes('android tablet')) return '📲';
    if (d.includes('macbook')) return '💻';
    if (d.includes('imac') || d.includes('mac desktop')) return '🖥️';
    if (d.includes('windows laptop')) return '💻';
    if (d.includes('windows desktop')) return '🖥️';
    if (d.includes('linux')) return '🐧';
    if (d.includes('raspi') || d.includes('raspberry')) return '🍓';
    return '💻';
}

// Keep for backward compat in peer avatar
function getOSIcon(os) {
    return getDeviceIcon(os);
}

function detectBrowser() {
    const ua = navigator.userAgent;
    if (ua.includes('Firefox')) return 'Firefox';
    if (ua.includes('Edg')) return 'Edge';
    if (ua.includes('OPR') || ua.includes('Opera')) return 'Opera';
    if (ua.includes('Brave')) return 'Brave';
    if (ua.includes('Chrome')) return 'Chrome';
    if (ua.includes('Safari')) return 'Safari';
    return 'Browser';
}

function detectDevice() {
    const ua = navigator.userAgent;

    // Mobile devices
    if (/iPhone/.test(ua)) return 'iPhone';
    if (/iPad/.test(ua)) return 'iPad';
    if (/Android/.test(ua)) {
        return /Mobile/.test(ua) ? 'Android Phone' : 'Android Tablet';
    }

    // Desktop
    if (/Macintosh/.test(ua)) {
        // Check for MacBook vs iMac via screen size heuristic
        const w = screen.width;
        return w <= 1680 ? 'MacBook' : 'Mac Desktop';
    }
    if (/Windows/.test(ua)) {
        // Laptops tend to have touch or smaller screens
        const hasTouch = navigator.maxTouchPoints > 0;
        return hasTouch ? 'Windows Laptop' : 'Windows Desktop';
    }
    if (/Linux/.test(ua)) {
        if (/Raspberry/.test(ua)) return 'Raspberry Pi';
        return 'Linux';
    }

    return 'Unknown';
}

function detectOS() {
    return detectDevice();
}

// Creative name parts by device type
const _deviceAdjectives = {
    'iPhone':          ['Zippy', 'Swift', 'Nimble', 'Sleek', 'Spark'],
    'iPad':            ['Mighty', 'Grand', 'Bright', 'Vivid', 'Prime'],
    'Android Phone':   ['Turbo', 'Flash', 'Blitz', 'Rapid', 'Bolt'],
    'Android Tablet':  ['Atlas', 'Titan', 'Nova', 'Orbit', 'Cosmo'],
    'MacBook':         ['Lunar', 'Stellar', 'Cosmic', 'Nebula', 'Astro'],
    'Mac Desktop':     ['Thunder', 'Storm', 'Forge', 'Titan', 'Peak'],
    'Windows Laptop':  ['Pixel', 'Cyber', 'Neon', 'Prism', 'Volt'],
    'Windows Desktop': ['Granite', 'Steel', 'Iron', 'Chrome', 'Apex'],
    'Linux':           ['Kernel', 'Root', 'Daemon', 'Cron', 'Shell'],
    'Raspberry Pi':    ['Berry', 'Tiny', 'Micro', 'Nano', 'Pico'],
};

const _deviceNouns = {
    'iPhone':          ['Pocket', 'Dart', 'Comet', 'Arrow', 'Flare'],
    'iPad':            ['Canvas', 'Shield', 'Slate', 'Prism', 'Deck'],
    'Android Phone':   ['Droid', 'Spark', 'Pulse', 'Wave', 'Beam'],
    'Android Tablet':  ['Pad', 'Grid', 'Pane', 'Slab', 'Matrix'],
    'MacBook':         ['Book', 'Wing', 'Rider', 'Craft', 'Glider'],
    'Mac Desktop':     ['Tower', 'Hub', 'Core', 'Base', 'Vault'],
    'Windows Laptop':  ['Flip', 'Node', 'Link', 'Gate', 'Port'],
    'Windows Desktop': ['Rig', 'Desk', 'Mill', 'Bunker', 'Forge'],
    'Linux':           ['Box', 'Node', 'Stack', 'Lab', 'Den'],
    'Raspberry Pi':    ['Pi', 'Chip', 'Dot', 'Seed', 'Bit'],
};

function generateDeviceName(device) {
    const adjs = _deviceAdjectives[device] || ['Smart', 'Cool', 'Fast', 'Bold', 'Keen'];
    const nouns = _deviceNouns[device] || ['Device', 'Node', 'Unit', 'Peer', 'Link'];
    const a = adjs[Math.floor(Math.random() * adjs.length)];
    const n = nouns[Math.floor(Math.random() * nouns.length)];
    const suffix = Math.random().toString(36).substring(2, 5);
    return a + '-' + n + '-' + suffix;
}

function generateChecksum(data) {
    return crypto.subtle.digest('SHA-256', data).then(hash => {
        return Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2, '0')).join('');
    });
}
