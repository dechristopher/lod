package cache

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/karlseguin/ccache/v2"
	"github.com/tile-fund/lod/str"
)

//go:embed data/test.pbf
var testTile []byte

var (
	testCache = &Cache{
		internal: ccache.New(ccache.Configure().
			MaxSize(5000),
		),
	}
)

// TestEncode will test that a given tile and metadata encodes properly
func TestEncode(t *testing.T) {
	headers := map[string]string{
		"Content-Type":     "application/vnd.mapbox-vector-tile",
		"Content-Encoding": "gzip",
	}

	// encode test tile
	tile := testCache.Encode("7/37/47", testTile, headers)

	// test the header counter was encoded properly
	if tile.LenHeaders() != len(headers) {
		t.Errorf(str.TCacheEncodeHeaders, tile.LenHeaders(), len(headers))
	}

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

// TestEncode will test that a given tile without metadata encodes properly
func TestEncodeNoHeaders(t *testing.T) {
	headers := map[string]string{}

	// encode test tile
	tile := testCache.Encode("7/37/47", testTile, headers)

	// test the header counter was encoded properly
	if tile.LenHeaders() != len(headers) {
		t.Errorf(str.TCacheEncodeHeaders, tile.LenHeaders(), len(headers))
	}

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
