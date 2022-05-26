package helpers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/cache"
	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/tile"
)

// GetTile computes the tile from the request URL
func GetTile(ctx *fiber.Ctx) (*tile.Tile, error) {
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

	return &tile.Tile{
		X:    x,
		Y:    y,
		Zoom: zoom,
	}, nil
}

// BuildTileUrl will substitute URL tile params into the proxy tile URL
func BuildTileUrl(proxy config.Proxy, ctx *fiber.Ctx, tileOverride ...tile.Tile) (string, error) {
	var currentTile *tile.Tile
	var err error

	if len(tileOverride) == 0 || tileOverride == nil {
		currentTile, err = GetTile(ctx)
		if err != nil {
			return "", err
		}
	} else {
		currentTile = &tileOverride[0]
	}

	// replace XYZ values in the tile URL
	baseUrl := currentTile.InjectString(proxy.TileURL)

	// replace dynamic endpoint parameter in URL if configured
	if proxy.HasEndpointParam {
		endpoint := ctx.Params(str.ParamEndpoint)
		baseUrl = strings.ReplaceAll(baseUrl, str.EndpointTemplate, endpoint)
	}

	// fetch params from context for possible addition to URL
	paramsMap := GetParamsFromCtx(ctx)

	// if no query parameters, return baseUrl
	if paramsMap == nil {
		return baseUrl, nil
	}

	// parse baseURL to add URL parameters
	paramUrl, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	// replace params by name in the key template if any exist
	for param, val := range paramsMap {
		params.Add(param, val)
	}

	// set encoded params in URL
	paramUrl.RawQuery = params.Encode()

	// return generated URL with substitutions for query parameters
	return paramUrl.String(), nil
}

// BuildCacheKey will put together a cache key from the configured template
func BuildCacheKey(proxy config.Proxy, ctx *fiber.Ctx, tileOverride ...tile.Tile) (string, error) {
	var currentTile *tile.Tile
	var err error

	if len(tileOverride) == 0 || tileOverride == nil {
		currentTile, err = GetTile(ctx)
		if err != nil {
			return "", err
		}
	} else {
		currentTile = &tileOverride[0]
	}

	// replace XYZ values in the key template
	key := currentTile.InjectString(proxy.Cache.KeyTemplate)

	// replace dynamic endpoint parameter in cache key if configured
	if proxy.HasEndpointParam && strings.Contains(key, str.EndpointTemplate) {
		endpoint := ctx.Params(str.ParamEndpoint)
		key = strings.ReplaceAll(key, str.EndpointTemplate, endpoint)
	}

	// fetch params from context for possible substitution
	paramsMap := GetParamsFromCtx(ctx)
	if paramsMap == nil {
		return key, nil
	}

	// replace params by name in the key template if any exist
	for param, val := range paramsMap {
		key = strings.ReplaceAll(key, fmt.Sprintf("{%s}", param), val)
	}

	return key, nil
}

// FillParamsMap will populate a map local to the request context with configured
// parameter values if any are present in the request
func FillParamsMap(proxy config.Proxy, ctx *fiber.Ctx) {
	paramsMap := make(map[string]string)
	for _, param := range proxy.Params {
		if val := ctx.Query(param.Name, param.Default); val != "" {
			paramsMap[param.Name] = val
		}
	}

	if len(paramsMap) > 0 {
		ctx.Locals(str.LocalParams, paramsMap)
	}
}

// GetParamsFromCtx will attempt to fetch the params map from the request
// context locals if any parameters are present and valid
func GetParamsFromCtx(ctx *fiber.Ctx) map[string]string {
	if ctx.Locals(str.LocalParams) != nil {
		return ctx.Locals(str.LocalParams).(map[string]string)
	}
	return nil
}

// ProxyResponse is a container struct encapsulating data retrieved from the
// upstream tile server during an agent-proxied request
type ProxyResponse struct {
	Code int
	Body []byte
	Resp *fiber.Response
}

// FetchUpstream will fetch and return relevant data from the configured
// upstream tileserver
func FetchUpstream(tileUrl string, p config.Proxy) func() (interface{}, error) {
	return func() (interface{}, error) {
		// configure proxy agent
		agent := fiber.AcquireAgent()

		req := agent.Request()
		req.Header.SetMethod(fiber.MethodGet)

		// set agent request URL
		req.SetRequestURI(tileUrl)

		// inject headers to upstream request if any are configured
		for _, header := range p.AddHeaders {
			req.Header.Add(header.Name, header.Value)
		}

		// parse agent request to find issues before making it
		if err := agent.Parse(); err != nil {
			panic(err)
		}

		// placeholder response for extracting headers from agent proxy request
		resp := fiber.AcquireResponse()
		agent.SetResponse(resp)

		// copy agent response, so we can transport its contents elsewhere while
		// returning the agent and its request pool to the fiber memory pool
		returnResponse := fiber.Response{}
		resp.CopyTo(&returnResponse)

		// make agent-proxied request
		code, body, errs := agent.Bytes()

		// immediately release response instance back to memory pool
		fiber.ReleaseResponse(resp)

		// return quickly if any issues arose
		if len(errs) > 0 {
			return nil, errs[0]
		}

		return ProxyResponse{
			Code: code,
			Body: body,
			Resp: &returnResponse,
		}, nil
	}
}

// ProcessResponsePayload is used by the proxy handler and some administrative
// cache endpoints to cleanly make calls to the ProcessResponse helper function
type ProcessResponsePayload struct {
	Ctx       *fiber.Ctx
	Cache     *cache.Cache
	Proxy     config.Proxy
	CacheKey  string
	Response  ProxyResponse
	WriteData bool
}

// ProcessResponse will cache fetched tile data, wrangle headers, and return the
// tile body in the provided fiber request context
func ProcessResponse(payload ProcessResponsePayload) error {
	// release response allocation back to memory pool for reuse
	// defer fiber.ReleaseResponse(payload.Response.Resp)

	// make sure a common 2XX response is received with relevant data, otherwise
	// we complain and throw a 500 due to misconfiguration of the proxy
	if payload.Response.Code == fiber.StatusNoContent || (len(payload.Response.Body) > 0 && payload.Response.Code == fiber.StatusOK) {
		// copy tile data into separate slice, so we don't lose the reference
		tileData := make([]byte, len(payload.Response.Body))
		copy(tileData, payload.Response.Body)

		headers := map[string]string{}
		// Store configured headers into the tile cache for this tile
		payload.Proxy.DoPullHeaders(payload.Response.Resp, headers)

		// write data to parent fiber request context if write mode is specified
		if payload.WriteData {
			// Delete headers from the final response that are on the DeleteHeaders list
			// if we got them from the tileserver. This can be used to prevent leaking
			// internals of the tileserver if you don't control what it returns
			payload.Proxy.DoDeleteHeaders(payload.Ctx)

			// set 204 Status No Content if upstream tileserver returned no/empty tile
			if payload.Response.Code == fiber.StatusNoContent {
				payload.Ctx.Status(fiber.StatusNoContent)
			}

			// write agent proxied response body to the response
			_, err := payload.Ctx.Write(payload.Response.Body)
			if err != nil {
				return err
			}
		}

		// spin off a routine to cache the tile without blocking the response
		go payload.Cache.EncodeSet(payload.CacheKey, tileData, headers)
	} else {
		// return an error to the parent fiber request if running in write mode
		if payload.WriteData {
			payload.Ctx.Locals(str.LocalCacheStatus, ":err-u")
			// Send internal server error response with empty body if upstream
			// fails to respond or responds with a non-200 status code
			return payload.Ctx.Status(fiber.StatusInternalServerError).SendString("")
		}

		//return generic error for non-proxy callers
		return ErrInvalidStatusCode{
			StatusCode: payload.Response.Code,
			CacheKey:   payload.CacheKey,
		}
	}

	return nil
}
