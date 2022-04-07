package packet

import "fmt"

// ErrTilePacketValidate is an error struct for errors
// encountered during TilePacket validation
type ErrTilePacketValidate struct {
	Key string
}

// Error returns the string representation of ErrTilePacketValidate
func (e ErrTilePacketValidate) Error() string {
	return fmt.Sprintf("cache: failed to validate tile packet at %s, invalidating", e.Key)
}
