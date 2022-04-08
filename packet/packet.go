package packet

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/pkg/errors"
)

// TilePacket is a custom binary data type for storing tile metadata alongside
// the tile data itself. Keeping it bundled up in bytes allows us to store it
// pretty much anywhere.
// |----------------------------------------------------------------------|
// | Checksum | Tile Data Size | Count | H1K Size | H1K | ... | Tile Data |
// |----------------------------------------------------------------------|
// | 32 bytes |     uint32     | uint8 |  uint16  | <-N | ... |  N bytes  |
// |----------------------------------------------------------------------|
type TilePacket []byte

// FromBytes wraps tile data from the cache and validates the
// contents, returning a TilePacket for additional processing
func FromBytes(data []byte, cacheKey string) (*TilePacket, error) {
	tile := TilePacket(data)
	valid := tile.Validate()

	if !valid {
		return nil, ErrTilePacketValidate{Key: cacheKey}
	}

	return &tile, nil
}

// Raw returns the TilePacket as a byte array
func (t TilePacket) Raw() []byte {
	return t
}

// Decode a TilePacket back into raw tile data and corresponding metadata
func (t TilePacket) Decode() ([]byte, map[string]string, error) {
	// ensure that the stored data is valid
	if !t.Validate() {
		return nil, nil, errors.New("checksum match failed, tile packet data corrupted")
	}

	return t.TileData(), t.Headers(), nil
}

// Validate the tile packet against the stored checksum
func (t TilePacket) Validate() bool {
	// guard against malformed or empty packets
	if len(t) < sha256.Size {
		return false
	}

	// compute checksum on packet data starting after stored checksum
	checksum := sha256.Sum256(t[32:])

	var storedChecksum [sha256.Size]byte
	for i := range t[:32] {
		storedChecksum[i] = t[i]
	}

	return checksum == storedChecksum
}

// TileData returns the raw tile data from the TilePacket
func (t TilePacket) TileData() []byte {
	// use tile data size to calculate the offset from the start of the packet that the tile data begins
	// this is necessary since header metadata can be variable in both count and length
	return t[(len(t) - t.TileDataSize()):]
}

// TileDataSize returns the raw tile data size in bytes from the TilePacket
func (t TilePacket) TileDataSize() int {
	return int(binary.LittleEndian.Uint32(t[32:36]))
}

// Headers returns a KV map of HTTP headers from the TilePacket
func (t TilePacket) Headers() map[string]string {
	// starting index for first byte of first header's key in the TilePacket structure
	offset := 37

	headers := make(map[string]string)

	for count := 0; count < t.LenHeaders(); count++ {
		keyLen := binary.LittleEndian.Uint16(t[offset : offset+2])
		offset += 2

		key := string(t[offset : offset+int(keyLen)])
		offset += int(keyLen)

		valLen := binary.LittleEndian.Uint16(t[offset : offset+2])
		offset += 2

		val := string(t[offset : offset+int(valLen)])
		offset += int(valLen)

		headers[key] = val
	}

	return headers
}

// LenHeaders returns the number of headers encoded in the TilePacket.
// Reads the count byte (byte index 36) and returns the value as an integer
func (t TilePacket) LenHeaders() int {
	return int(t[36:37][0])
}
