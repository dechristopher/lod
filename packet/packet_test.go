package packet

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/dechristopher/lod/str"
)

// TestEncode will test that a given tile and metadata encodes properly
func TestDecode(t *testing.T) {
	// encode test tile
	tile := testCache.Encode(testTile, testHeaders)

	// decode tile and check for errors
	decodedTile, decodedHeaders, err := tile.Decode()
	if err != nil {
		t.Errorf(str.TCacheBadDecode, err.Error())
	}

	// test that tile data was decoded properly
	if !reflect.DeepEqual(decodedTile, testTile) {
		t.Errorf(str.TCacheBadTileData)
	}

	// test that header data was decoded properly
	if !reflect.DeepEqual(decodedHeaders, testHeaders) {
		t.Errorf(str.TCacheBadHeaderData)
	}
}

// BenchmarkDecode will benchmark a standard tile and metadata decode
func BenchmarkDecode(b *testing.B) {
	// encode test tile
	tile := testCache.Encode(testTile, testHeaders)

	for i := 0; i < b.N; i++ {
		// decode tile and check for errors
		_, _, err := tile.Decode()
		if err != nil {
			b.Errorf(str.TCacheBadDecode, err.Error())
		}
	}
}
