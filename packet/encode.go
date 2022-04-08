package packet

import (
	"crypto/sha256"
	"encoding/binary"
)

// Encode tile data and metadata into a TilePacket
func Encode(tile []byte, headers map[string]string) TilePacket {
	// final tile packet
	var tilePacket TilePacket

	// insert empty checksum for later, prevents us from appending the whole
	// packet to the computed checksum when we can just set the bytes
	checksum := make([]byte, sha256.Size)
	tilePacket = append(tilePacket, checksum...)

	tileDataSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(tileDataSize, uint32(len(tile)))

	// add tile data size (uint32) after empty checksum
	tilePacket = append(tilePacket, tileDataSize...)

	headerCount := uint8(len(headers))

	// add header count after tile data length
	tilePacket = append(tilePacket, headerCount)

	// append all header keys and values with their lengths
	for key, val := range headers {
		keyBytes := []byte(key)
		valBytes := []byte(val)

		keyBytesSize := make([]byte, 2)
		binary.LittleEndian.PutUint16(keyBytesSize, uint16(len(keyBytes)))

		valBytesSize := make([]byte, 2)
		binary.LittleEndian.PutUint16(valBytesSize, uint16(len(valBytes)))

		// append uint16 - size of header key in bytes
		tilePacket = append(tilePacket, keyBytesSize...)
		// append header key bytes
		tilePacket = append(tilePacket, keyBytes...)
		// append uint16 - size of header value in bytes
		tilePacket = append(tilePacket, valBytesSize...)
		// append header value bytes
		tilePacket = append(tilePacket, valBytes...)
	}

	// append tile to packet after end of metadata
	tilePacket = append(tilePacket, tile...)

	// compute checksum of packet up to now minus the space reserved for the checksum
	computedChecksum := sha256.Sum256(tilePacket[32:])

	// set checksum value in packet
	for i := range computedChecksum {
		tilePacket[i] = computedChecksum[i]
	}

	return tilePacket
}
