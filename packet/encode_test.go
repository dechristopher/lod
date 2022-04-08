package packet

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/dechristopher/lod/str"
)

//go:embed data/test.pbf
var testTile []byte

var (
	testHeaders = map[string]string{
		"Content-Type":     "application/vnd.mapbox-vector-tile",
		"Content-Encoding": "gzip",
	}
)

// TestEncode will test that a given tile and metadata encodes properly
func TestEncode(t *testing.T) {
	// encode test tile
	tile := Encode(testTile, testHeaders)

	// test the header counter was encoded properly
	if tile.LenHeaders() != len(testHeaders) {
		t.Errorf(str.TCacheEncodeHeaders, tile.LenHeaders(), len(testHeaders))
	}

	// test that header data was encoded properly
	if !reflect.DeepEqual(tile.Headers(), testHeaders) {
		t.Errorf(str.TCacheBadHeaderData)
	}

	// test that tile data was encoded properly
	if !reflect.DeepEqual(tile.TileData(), testTile) {
		t.Errorf(str.TCacheBadTileData)
	}

	// ensure checksum is computed properly and matches stored checksum
	if !tile.Validate() {
		t.Errorf(str.TCacheBadValidation)
	}
}

// TestEncode will test that a given tile without metadata encodes properly
func TestEncodeNoHeaders(t *testing.T) {
	headers := map[string]string{}

	// encode test tile
	tile := Encode(testTile, headers)

	// test the header counter was encoded properly
	if tile.LenHeaders() != len(headers) {
		t.Errorf(str.TCacheEncodeHeaders, tile.LenHeaders(), len(headers))
	}

	// test that header data was encoded properly
	if !reflect.DeepEqual(tile.Headers(), headers) {
		t.Errorf(str.TCacheBadHeaderData)
	}

	// test that tile data was encoded properly
	if !reflect.DeepEqual(tile.TileData(), testTile) {
		t.Errorf(str.TCacheBadTileData)
	}

	// ensure checksum is computed properly and matches stored checksum
	if !tile.Validate() {
		t.Errorf(str.TCacheBadValidation)
	}
}

// BenchmarkEncode will benchmark a standard tile and metadata encode
func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// encode test tile N times
		_ = Encode(testTile, testHeaders)
	}
}
