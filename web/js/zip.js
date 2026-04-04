// Minimal ZIP file creator (STORE method, no compression)
// Works without any external dependencies

const ZipBuilder = {
    create(files) {
        // files: [{ name: string, data: Uint8Array }]
        const localHeaders = [];
        const centralHeaders = [];
        let offset = 0;

        for (const file of files) {
            const nameBytes = new TextEncoder().encode(file.name);
            const localHeader = this._localFileHeader(nameBytes, file.data);
            localHeaders.push(localHeader);
            centralHeaders.push(this._centralDirHeader(nameBytes, file.data, offset));
            offset += localHeader.byteLength + file.data.byteLength;
        }

        const centralDirOffset = offset;
        let centralDirSize = 0;
        for (const h of centralHeaders) centralDirSize += h.byteLength;

        const endRecord = this._endOfCentralDir(files.length, centralDirSize, centralDirOffset);

        // Assemble ZIP
        const totalSize = offset + centralDirSize + endRecord.byteLength;
        const zip = new Uint8Array(totalSize);
        let pos = 0;

        for (let i = 0; i < files.length; i++) {
            zip.set(new Uint8Array(localHeaders[i]), pos);
            pos += localHeaders[i].byteLength;
            zip.set(files[i].data, pos);
            pos += files[i].data.byteLength;
        }

        for (const h of centralHeaders) {
            zip.set(new Uint8Array(h), pos);
            pos += h.byteLength;
        }

        zip.set(new Uint8Array(endRecord), pos);

        return new Blob([zip], { type: 'application/zip' });
    },

    _localFileHeader(nameBytes, data) {
        const buf = new ArrayBuffer(30 + nameBytes.length);
        const view = new DataView(buf);
        const crc = this._crc32(data);

        view.setUint32(0, 0x04034b50, true);   // Local file header signature
        view.setUint16(4, 20, true);            // Version needed (2.0)
        view.setUint16(6, 0, true);             // General purpose bit flag
        view.setUint16(8, 0, true);             // Compression: STORE
        view.setUint16(10, 0, true);            // Mod time
        view.setUint16(12, 0, true);            // Mod date
        view.setUint32(14, crc, true);          // CRC-32
        view.setUint32(18, data.byteLength, true);  // Compressed size
        view.setUint32(22, data.byteLength, true);  // Uncompressed size
        view.setUint16(26, nameBytes.length, true);  // Filename length
        view.setUint16(28, 0, true);            // Extra field length

        new Uint8Array(buf).set(nameBytes, 30);
        return buf;
    },

    _centralDirHeader(nameBytes, data, localHeaderOffset) {
        const buf = new ArrayBuffer(46 + nameBytes.length);
        const view = new DataView(buf);
        const crc = this._crc32(data);

        view.setUint32(0, 0x02014b50, true);   // Central directory header signature
        view.setUint16(4, 20, true);            // Version made by
        view.setUint16(6, 20, true);            // Version needed
        view.setUint16(8, 0, true);             // General purpose bit flag
        view.setUint16(10, 0, true);            // Compression: STORE
        view.setUint16(12, 0, true);            // Mod time
        view.setUint16(14, 0, true);            // Mod date
        view.setUint32(16, crc, true);          // CRC-32
        view.setUint32(20, data.byteLength, true);  // Compressed size
        view.setUint32(24, data.byteLength, true);  // Uncompressed size
        view.setUint16(28, nameBytes.length, true);  // Filename length
        view.setUint16(30, 0, true);            // Extra field length
        view.setUint16(32, 0, true);            // File comment length
        view.setUint16(34, 0, true);            // Disk number start
        view.setUint16(36, 0, true);            // Internal file attributes
        view.setUint32(38, 0, true);            // External file attributes
        view.setUint32(42, localHeaderOffset, true); // Relative offset of local header

        new Uint8Array(buf).set(nameBytes, 46);
        return buf;
    },

    _endOfCentralDir(numFiles, centralDirSize, centralDirOffset) {
        const buf = new ArrayBuffer(22);
        const view = new DataView(buf);

        view.setUint32(0, 0x06054b50, true);   // End of central directory signature
        view.setUint16(4, 0, true);             // Disk number
        view.setUint16(6, 0, true);             // Disk with central dir
        view.setUint16(8, numFiles, true);      // Entries on this disk
        view.setUint16(10, numFiles, true);     // Total entries
        view.setUint32(12, centralDirSize, true);
        view.setUint32(16, centralDirOffset, true);
        view.setUint16(20, 0, true);            // Comment length

        return buf;
    },

    // CRC-32 lookup table
    _crcTable: null,

    _makeCrcTable() {
        const table = new Uint32Array(256);
        for (let n = 0; n < 256; n++) {
            let c = n;
            for (let k = 0; k < 8; k++) {
                c = (c & 1) ? (0xEDB88320 ^ (c >>> 1)) : (c >>> 1);
            }
            table[n] = c;
        }
        return table;
    },

    _crc32(data) {
        if (!this._crcTable) this._crcTable = this._makeCrcTable();
        let crc = 0xFFFFFFFF;
        for (let i = 0; i < data.byteLength; i++) {
            crc = this._crcTable[(crc ^ data[i]) & 0xFF] ^ (crc >>> 8);
        }
        return (crc ^ 0xFFFFFFFF) >>> 0;
    }
};
