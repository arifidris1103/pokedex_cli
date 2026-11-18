package pokeapi

// GetPokemon returns one Pokémon by name (or Pokédex number).
func (c *Client) GetPokemon(name string) (Pokemon, error) {
	return get[Pokemon](c, baseURL+"/pokemon/"+name)
}
