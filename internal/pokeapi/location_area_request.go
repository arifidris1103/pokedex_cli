package pokeapi

// ListLocationAreas returns one page of location areas. Pass nil for the
// first page, or a Next/Previous URL from an earlier page.
func (c *Client) ListLocationAreas(pageURL *string) (LocationAreaResp, error) {
	url := baseURL + "/location-area"
	if pageURL != nil {
		url = *pageURL
	}
	return get[LocationAreaResp](c, url)
}

// GetLocationArea returns one location area, including the Pokémon found there.
func (c *Client) GetLocationArea(name string) (LocationArea, error) {
	return get[LocationArea](c, baseURL+"/location-area/"+name)
}
