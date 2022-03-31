package cache

import (
	"crypto/sha256"
	"encoding/binary"
)

// Encode tile data and metadata into a TilePacket
func (c *Cache) Encode(tile []byte, headers map[string]string) TilePacket {
	// final tile packet
	var packet TilePacket

	// insert empty checksum for later, prevents us from appending the whole
	// packet to the computed checksum when we can just set the bytes
	checksum := make([]byte, sha256.Size)
	packet = append(packet, checksum...)

	tileDataSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(tileDataSize, uint32(len(tile)))

	// add tile data size (uint32) after empty checksum
	packet = append(packet, tileDataSize...)

	headerCount := uint8(len(headers))

	// add header count after tile data length
	packet = append(packet, headerCount)

	// append all header keys and values with their lengths
	for key, val := range headers {
		keyBytes := []byte(key)
		valBytes := []byte(val)

		keyBytesSize := make([]byte, 2)
		binary.LittleEndian.PutUint16(keyBytesSize, uint16(len(keyBytes)))

		valBytesSize := make([]byte, 2)
		binary.LittleEndian.PutUint16(valBytesSize, uint16(len(valBytes)))

		// append uint16 - size of header key in bytes
		packet = append(packet, keyBytesSize...)
		// append header key bytes
		packet = append(packet, keyBytes...)
		// append uint16 - size of header value in bytes
		packet = append(packet, valBytesSize...)
		// append header value bytes
		packet = append(packet, valBytes...)
	}

	// append tile to packet after end of metadata
	packet = append(packet, tile...)

	// compute checksum of packet up to now minus the space reserved for the checksum
	computedChecksum := sha256.Sum256(packet[32:])

	// set checksum value in packet
	for i := range computedChecksum {
		packet[i] = computedChecksum[i]
	}

	return packet
}
