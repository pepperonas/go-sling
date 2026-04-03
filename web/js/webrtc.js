// WebRTC connection management

const WebRTCManager = {
    connections: {}, // peerId -> { pc, dc, state }

    config: {
        iceServers: [], // LAN only, no STUN/TURN needed
        iceTransportPolicy: 'all',
    },

    createConnection(peerId) {
        const pc = new RTCPeerConnection(this.config);
        const conn = { pc, dc: null, state: 'new' };
        this.connections[peerId] = conn;

        pc.onicecandidate = (event) => {
            if (event.candidate) {
                // Only use host candidates (LAN only)
                if (event.candidate.candidate.includes('host') || event.candidate.candidate.includes('typ host')) {
                    WS.send({
                        type: 'ice-candidate',
                        to: peerId,
                        payload: { candidate: event.candidate.toJSON() },
                    });
                }
            }
        };

        pc.onconnectionstatechange = () => {
            conn.state = pc.connectionState;
            if (pc.connectionState === 'failed' || pc.connectionState === 'disconnected') {
                UI.toast('P2P connection lost with peer', 'error');
                this.closeConnection(peerId);
            }
        };

        return conn;
    },

    async initiateTransfer(peerId) {
        const conn = this.createConnection(peerId);
        const dc = conn.pc.createDataChannel('file-transfer', {
            ordered: true,
        });

        conn.dc = dc;
        this._setupDataChannel(dc, peerId);

        const offer = await conn.pc.createOffer();
        await conn.pc.setLocalDescription(offer);

        WS.send({
            type: 'offer',
            to: peerId,
            payload: { sdp: conn.pc.localDescription.toJSON() },
        });

        return conn;
    },

    async handleOffer(fromId, sdp) {
        const conn = this.createConnection(fromId);

        conn.pc.ondatachannel = (event) => {
            conn.dc = event.channel;
            this._setupDataChannel(event.channel, fromId);
        };

        await conn.pc.setRemoteDescription(new RTCSessionDescription(sdp));
        const answer = await conn.pc.createAnswer();
        await conn.pc.setLocalDescription(answer);

        WS.send({
            type: 'answer',
            to: fromId,
            payload: { sdp: conn.pc.localDescription.toJSON() },
        });
    },

    async handleAnswer(fromId, sdp) {
        const conn = this.connections[fromId];
        if (conn) {
            await conn.pc.setRemoteDescription(new RTCSessionDescription(sdp));
        }
    },

    async handleIceCandidate(fromId, candidate) {
        const conn = this.connections[fromId];
        if (conn) {
            await conn.pc.addIceCandidate(new RTCIceCandidate(candidate));
        }
    },

    _setupDataChannel(dc, peerId) {
        dc.binaryType = 'arraybuffer';

        dc.onopen = () => {
            const handlers = this._onOpenHandlers[peerId];
            if (handlers) {
                handlers.forEach(h => h(dc));
                delete this._onOpenHandlers[peerId];
            }
        };

        dc.onmessage = (event) => {
            Transfer.handleIncoming(peerId, event.data);
        };

        dc.onclose = () => {
            console.log('DataChannel closed with', peerId);
        };

        dc.onerror = (err) => {
            console.error('DataChannel error:', err);
        };
    },

    _onOpenHandlers: {},

    onDataChannelOpen(peerId, handler) {
        if (!this._onOpenHandlers[peerId]) this._onOpenHandlers[peerId] = [];
        this._onOpenHandlers[peerId].push(handler);

        // If already open, call immediately
        const conn = this.connections[peerId];
        if (conn && conn.dc && conn.dc.readyState === 'open') {
            handler(conn.dc);
        }
    },

    getDataChannel(peerId) {
        const conn = this.connections[peerId];
        return conn ? conn.dc : null;
    },

    closeConnection(peerId) {
        const conn = this.connections[peerId];
        if (conn) {
            if (conn.dc) conn.dc.close();
            conn.pc.close();
            delete this.connections[peerId];
        }
    },

    closeAll() {
        Object.keys(this.connections).forEach(id => this.closeConnection(id));
    }
};
