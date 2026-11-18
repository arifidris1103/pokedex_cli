// Package pokeapi is a small client for https://pokeapi.co with a response cache.
package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arifidris1103/pokedex_cli/pokecache"
)

const baseURL = "https://pokeapi.co/api/v2"

type Client struct {
	cache      pokecache.Cache
	httpClient http.Client
}

func NewClient(cacheInterval time.Duration) Client {
	return Client{
		cache: pokecache.NewCache(cacheInterval),
		httpClient: http.Client{
			Timeout: time.Minute,
		},
	}
}

// NamedAPIResource is PokéAPI's {"name": ..., "url": ...} pair, used everywhere.
type NamedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// get fetches url (or reads it from the cache) and decodes the JSON into T.
// Every request in this package goes through here.
func get[T any](c *Client, url string) (T, error) {
	var zero T

	dat, ok := c.cache.Get(url)
	if !ok {
		resp, err := c.httpClient.Get(url)
		if err != nil {
			return zero, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return zero, fmt.Errorf("GET %s: %s", url, resp.Status)
		}
		dat, err = io.ReadAll(resp.Body)
		if err != nil {
			return zero, err
		}
		c.cache.Add(url, dat)
	}

	var out T
	if err := json.Unmarshal(dat, &out); err != nil {
		return zero, err
	}
	return out, nil
}
