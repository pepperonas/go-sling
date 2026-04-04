// WebSocket client for signaling

const WS = {
    socket: null,
    myId: null,
    myName: null,
    peers: [],
    handlers: {},
    reconnectDelay: 1000,
    maxReconnectDelay: 30000,

    connect() {
        const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
        const url = `${proto}//${location.host}/ws`;

        this.socket = new WebSocket(url);

        this.socket.onopen = () => {
            this.reconnectDelay = 1000;
            const device = detectDevice();
            // Send join message with device info and a device-specific name
            this.send({
                type: 'join',
                payload: {
                    name: generateDeviceName(device),
                    browser: detectBrowser(),
                    os: device,
                }
            });
        };

        this.socket.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                this._handleMessage(msg);
            } catch (e) {
                console.error('WS parse error:', e);
            }
        };

        this.socket.onclose = () => {
            setTimeout(() => {
                this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
                this.connect();
            }, this.reconnectDelay);
        };

        this.socket.onerror = () => {
            // onclose will fire after this
        };
    },

    send(msg) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(msg));
        }
    },

    on(type, handler) {
        if (!this.handlers[type]) this.handlers[type] = [];
        this.handlers[type].push(handler);
    },

    _handleMessage(msg) {
        switch (msg.type) {
            case 'welcome':
                this.myId = msg.payload.id;
                this.myName = msg.payload.name;
                break;

            case 'peer-list':
                this.peers = msg.peers || [];
                UI.updatePeerCount(this.peers.length);
                UI.renderPeerList(this.peers, this.myId);
                break;

            case 'offer':
            case 'answer':
            case 'ice-candidate':
            case 'transfer-request':
            case 'transfer-accept':
            case 'transfer-reject':
                break;

            default:
                console.log('Unknown WS message:', msg.type);
        }

        // Call registered handlers
        const handlers = this.handlers[msg.type] || [];
        handlers.forEach(h => h(msg));
    },

    isConnected() {
        return this.socket && this.socket.readyState === WebSocket.OPEN;
    }
};
