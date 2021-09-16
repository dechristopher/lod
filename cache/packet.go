package cache

// TilePacket is a custom binary data type for storing tile metadata alongside
// the tile data itself. Keeping it bundled up in bytes allows us to store it
// pretty much anywhere.
// |----------------------------------------------------------------------|
// | Checksum | Tile Data Size | Count | H1K Size | H1K | ... | Tile Data |
// |----------------------------------------------------------------------|
// |  uint32  |     uint32     | uint8 |  uint16  | <-N | ... |  N bytes  |
// |----------------------------------------------------------------------|
type TilePacket []byte

// Decode a TilePacket back into raw tile data and corresponding metadata
func (t *TilePacket) Decode() ([]byte, map[string]string, error) {
	return nil, nil, nil
}

// RawData returns the raw tile data from the TilePacket
func (t *TilePacket) RawData() []byte {
	return nil
}

// DataSize returns the raw tile data size in bytes from the TilePacket
func (t *TilePacket) DataSize() int {
	return 0
}

// Headers returns a KV map of HTTP headers from the TilePacket
func (t *TilePacket) Headers() map[string]string {
	return nil
}
