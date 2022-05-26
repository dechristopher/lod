package tile

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/twpayne/go-geos"

	"github.com/dechristopher/lod/str"
)

// Tile represents a request for a single tile by layer class
type Tile struct {
	X    int
	Y    int
	Zoom int
}

// Get computes and returns the tile from the request URL
func Get(ctx *fiber.Ctx) (*Tile, error) {
	x, xErr := ctx.ParamsInt(str.ParamX)
	if xErr != nil {
		return nil, xErr
	}

	y, yErr := ctx.ParamsInt(str.ParamY)
	if yErr != nil {
		return nil, yErr
	}

	zoom, zErr := ctx.ParamsInt(str.ParamZ)
	if zErr != nil {
		return nil, zErr
	}

	return &Tile{
		X:    x,
		Y:    y,
		Zoom: zoom,
	}, nil
}

func (t Tile) String() string {
	return fmt.Sprintf("(Z:%d,X:%d,Y:%d)", t.Zoom, t.X, t.Y)
}

// XFloat returns the tile X value as a float64
func (t Tile) XFloat() float64 {
	return float64(t.X)
}

// YFloat returns the tile Y value as a float64
func (t Tile) YFloat() float64 {
	return float64(t.Y)
}

// ZoomFloat returns the tile Zoom level as a float64
func (t Tile) ZoomFloat() float64 {
	return float64(t.Zoom)
}

// InjectString fills the {x}, {y}, and {z} tokens in a template URL
//  or cache key with the provided tile values
func (t Tile) InjectString(base string) string {
	base = strings.ReplaceAll(base, "{x}", strconv.Itoa(t.X))
	base = strings.ReplaceAll(base, "{y}", strconv.Itoa(t.Y))
	return strings.ReplaceAll(base, "{z}", strconv.Itoa(t.Zoom))
}

// Children returns the four child tiles of a given tile
func (t Tile) Children() [4]Tile {
	return [4]Tile{
		{
			X:    2 * t.X,
			Y:    2 * t.Y,
			Zoom: t.Zoom + 1,
		},
		{
			X:    (2 * t.X) + 1,
			Y:    2 * t.Y,
			Zoom: t.Zoom + 1,
		},
		{
			X:    2 * t.X,
			Y:    (2 * t.Y) + 1,
			Zoom: t.Zoom + 1,
		},
		{
			X:    (2 * t.X) + 1,
			Y:    (2 * t.Y) + 1,
			Zoom: t.Zoom + 1,
		},
	}
}

// DeepChildren returns a list of all tiles that are descendant of the
// given tile up to the given zoom level
func (t Tile) DeepChildren(maxZoom int) []Tile {
	tileChan := make(chan Tile)
	tiles := make([]Tile, 0)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// append children to tiles list as we receive them
	go func() {
		for tile := range tileChan {
			tiles = append(tiles, tile)
		}
	}()

	// begin async deepening algorithm to compute children
	go deepChildrenImpl(maxZoom, t, tileChan, wg)

	wg.Wait()
	close(tileChan)

	return tiles
}

// deepChildrenImpl emits, to the given channel, all tiles that are descendant of
// the given tile up to a given zoom level, starting at the tile provided
func deepChildrenImpl(maxZoom int, tile Tile, tileChan chan Tile, wg *sync.WaitGroup) {
	defer wg.Done()

	tileChan <- tile

	if tile.Zoom >= maxZoom {
		return
	}

	children := tile.Children()
	wg.Add(4)

	for _, childTile := range children {
		go deepChildrenImpl(maxZoom, childTile, tileChan, wg)
	}
}

// Bounds calculates bounding box of the given tile based
// on the tile's X and Y value and zoom level
func (t Tile) Bounds() *geos.Bounds {
	// get northwest corner of current tile
	nwLat, nwLon := getCorner(t.XFloat(), t.YFloat(), t.ZoomFloat())
	// get northwest corner of tile southeast of current tile
	// which gives us the southeast corner of current tile
	seLat, seLon := getCorner(t.XFloat()+1, t.YFloat()+1, t.ZoomFloat())

	bounds := geos.NewBounds(nwLat, nwLon, seLat, seLon)
	return bounds
}

// DeepIntersect emits, to the given channel, all tiles that a given geometry
// intersects at all zoom levels, starting at the tile provided
func DeepIntersect(geometry *geos.Geom, tile Tile, tileChan chan Tile, wg *sync.WaitGroup) {
	defer wg.Done()

	box := tile.Bounds()
	boxPolygon := geos.NewGeomFromBounds(box)

	intersects := geometry.Intersects(boxPolygon)

	if !intersects {
		// fmt.Printf("[%d] does not intersect at tile %+v\n", tile.Zoom, tile)
		return
	}

	tileChan <- tile

	if tile.Zoom == 16 {
		return
	}

	children := tile.Children()
	wg.Add(4)

	for _, childTile := range children {
		go DeepIntersect(geometry, childTile, tileChan, wg)
	}
}

// getCorners calculates the NW corner of a given tile
func getCorner(x, y, zoom float64) (float64, float64) {
	n := math.Pow(2.0, zoom)
	lonDeg := x/n*360.0 - 180.0
	latDeg := math.Atan(math.Sinh(math.Pi*(1-2*y/n))) * (180 / math.Pi)
	return latDeg, lonDeg
}
