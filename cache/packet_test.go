package cache

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/tile-fund/lod/str"
)

// TestEncode will test that a given tile and metadata encodes properly
func TestDecode(t *testing.T) {
	headers := map[string]string{
		"Content-Type":     "application/vnd.mapbox-vector-tile",
		"Content-Encoding": "gzip",
	}

	// encode test tile
	tile := testCache.Encode("7/37/47", testTile, headers)

	// decode tile and check for errors
	decodedTile, decodedHeaders, err := tile.Decode()
	if err != nil {
		t.Errorf(str.TCacheBadDecode, err.Error())
	}

	// test that tile data was decoded properly
	if !reflect.DeepEqual(decodedTile, tile) {
		t.Errorf(str.TCacheBadTileData)
	}

	// test that header data was decoded properly
	if !reflect.DeepEqual(decodedHeaders, headers) {
		t.Errorf(str.TCacheBadHeaderData)
	}
}
