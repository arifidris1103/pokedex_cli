package ui

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// DemoBackend is an offline stand-in backend with a handful of real
// Pokémon, so the UI can run (and be tested) before the real backend is done.
// Only sprites come from the network; set Offline to skip those too.
type DemoBackend struct {
	Offline bool
	Delay   time.Duration // fake network latency
	Rand    func() float64

	mu     sync.Mutex
	caught map[string]Pokemon
}

func NewDemoBackend() *DemoBackend {
	return &DemoBackend{Delay: 250 * time.Millisecond, Rand: rand.Float64, caught: map[string]Pokemon{}}
}

var demoAreas = []string{
	"canalave-city-area", "eterna-city-area", "pastoria-city-area", "sunyshore-city-area",
	"sinnoh-pokemon-league-area", "oreburgh-mine-1f", "oreburgh-mine-b1f", "valley-windworks-area",
	"eterna-forest-area", "fuego-ironworks-area", "mt-coronet-1f-route-207", "mt-coronet-2f",
	"mt-coronet-3f", "mt-coronet-exterior-snowfall", "mt-coronet-exterior-blizzard", "mt-coronet-4f",
	"mt-coronet-4f-small-room", "mt-coronet-5f", "mt-coronet-6f", "mt-coronet-1f-from-exterior",
	"great-marsh-area-1", "great-marsh-area-2", "great-marsh-area-3", "great-marsh-area-4",
	"great-marsh-area-5", "great-marsh-area-6", "solaceon-ruins-2f", "solaceon-ruins-1f",
	"solaceon-ruins-b1f-a", "solaceon-ruins-b1f-b", "lake-verity-before-galactic-intervention",
	"lake-acuity-area", "lake-valor-area", "iron-island-area", "old-chateau-entrance",
	"wayward-cave-1f", "victory-road-1f", "ravaged-path-area", "oreburgh-gate-1f", "stark-mountain-area",
}

var demoEncounters = map[string][]string{
	"canalave-city-area":  {"tentacool", "tentacruel", "staryu", "magikarp", "gyarados", "wingull", "pelipper", "shellos", "gastrodon", "finneon", "lumineon"},
	"eterna-city-area":    {"psyduck", "golduck", "magikarp", "gyarados", "bidoof"},
	"pastoria-city-area":  {"tentacool", "tentacruel", "magikarp", "gyarados", "shellos", "gastrodon", "wingull"},
	"sunyshore-city-area": {"tentacool", "staryu", "starmie", "magikarp", "wingull", "pelipper"},
	"oreburgh-mine-1f":    {"zubat", "geodude", "onix"},
	"oreburgh-mine-b1f":   {"zubat", "geodude", "onix"},
	"wayward-cave-1f":     {"zubat", "geodude", "onix", "machop"},
}

type demoMon struct {
	id, exp, h, w int
	types         []string
	stats         [6]int
}

var demoDex = map[string]demoMon{
	"tentacool":  {72, 67, 9, 455, []string{"water", "poison"}, [6]int{40, 40, 35, 50, 100, 70}},
	"tentacruel": {73, 180, 16, 550, []string{"water", "poison"}, [6]int{80, 70, 65, 80, 120, 100}},
	"staryu":     {120, 68, 8, 345, []string{"water"}, [6]int{30, 45, 55, 70, 55, 85}},
	"starmie":    {121, 182, 11, 800, []string{"water", "psychic"}, [6]int{60, 75, 85, 100, 85, 115}},
	"magikarp":   {129, 40, 9, 100, []string{"water"}, [6]int{20, 10, 55, 15, 20, 80}},
	"gyarados":   {130, 189, 65, 2350, []string{"water", "flying"}, [6]int{95, 125, 79, 60, 100, 81}},
	"wingull":    {278, 54, 6, 95, []string{"water", "flying"}, [6]int{40, 30, 30, 55, 30, 85}},
	"pelipper":   {279, 154, 12, 280, []string{"water", "flying"}, [6]int{60, 50, 100, 95, 70, 65}},
	"shellos":    {422, 65, 3, 63, []string{"water"}, [6]int{76, 48, 48, 57, 62, 34}},
	"gastrodon":  {423, 166, 9, 299, []string{"water", "ground"}, [6]int{111, 83, 68, 92, 82, 39}},
	"finneon":    {456, 66, 4, 70, []string{"water"}, [6]int{49, 49, 56, 49, 61, 66}},
	"lumineon":   {457, 161, 12, 240, []string{"water"}, [6]int{69, 69, 76, 69, 86, 91}},
	"psyduck":    {54, 64, 8, 196, []string{"water"}, [6]int{50, 52, 48, 65, 50, 55}},
	"golduck":    {55, 175, 17, 766, []string{"water"}, [6]int{80, 82, 78, 95, 80, 85}},
	"bidoof":     {399, 50, 5, 200, []string{"normal"}, [6]int{59, 45, 40, 35, 40, 31}},
	"zubat":      {41, 49, 8, 75, []string{"poison", "flying"}, [6]int{40, 45, 35, 30, 40, 55}},
	"geodude":    {74, 60, 4, 200, []string{"rock", "ground"}, [6]int{40, 80, 100, 30, 30, 20}},
	"onix":       {95, 77, 88, 2100, []string{"rock", "ground"}, [6]int{35, 45, 160, 30, 45, 70}},
	"machop":     {66, 61, 8, 195, []string{"fighting"}, [6]int{70, 80, 50, 35, 35, 35}},
}

var demoEvolutions = [][]EvoStage{
	{{"tentacool", ""}, {"tentacruel", "Lv. 30"}},
	{{"staryu", ""}, {"starmie", "Water Stone"}},
	{{"magikarp", ""}, {"gyarados", "Lv. 20"}},
	{{"wingull", ""}, {"pelipper", "Lv. 25"}},
	{{"shellos", ""}, {"gastrodon", "Lv. 30"}},
	{{"finneon", ""}, {"lumineon", "Lv. 31"}},
	{{"psyduck", ""}, {"golduck", "Lv. 33"}},
	{{"bidoof", ""}, {"bibarel", "Lv. 15"}},
	{{"zubat", ""}, {"golbat", "Lv. 22"}, {"crobat", "Friendship"}},
	{{"geodude", ""}, {"graveler", "Lv. 25"}, {"golem", "Trade"}},
	{{"onix", ""}, {"steelix", "Trade w/ Metal Coat"}},
	{{"machop", ""}, {"machoke", "Lv. 28"}, {"machamp", "Trade"}},
}

const demoPageSize = 20

func (d *DemoBackend) wait() { time.Sleep(d.Delay) }

func (d *DemoBackend) ListLocationAreas(pageURL *string) (LocationPage, error) {
	d.wait()
	offset := 0
	if pageURL != nil {
		n, err := strconv.Atoi(*pageURL)
		if err != nil {
			return LocationPage{}, fmt.Errorf("bad page %q", *pageURL)
		}
		offset = n
	}
	end := min(len(demoAreas), offset+demoPageSize)
	page := LocationPage{Areas: demoAreas[offset:end]}
	if end < len(demoAreas) {
		next := strconv.Itoa(end)
		page.Next = &next
	}
	if offset > 0 {
		prev := strconv.Itoa(max(0, offset-demoPageSize))
		page.Prev = &prev
	}
	return page, nil
}

func (d *DemoBackend) ExploreArea(area string) ([]string, error) {
	d.wait()
	if names, ok := demoEncounters[area]; ok {
		return names, nil
	}
	return []string{"zubat", "geodude", "machop", "bidoof"}, nil
}

func (d *DemoBackend) GetPokemon(name string) (Pokemon, error) {
	d.wait()
	mon, ok := demoDex[name]
	if !ok {
		return Pokemon{}, fmt.Errorf("no pokemon named %q", name)
	}
	names := []string{"hp", "attack", "defense", "special-attack", "special-defense", "speed"}
	stats := make([]Stat, len(names))
	for i, n := range names {
		stats[i] = Stat{Name: n, Value: mon.stats[i]}
	}
	return Pokemon{
		ID: mon.id, Name: name, BaseExperience: mon.exp, Height: mon.h, Weight: mon.w,
		Types: mon.types, Stats: stats,
		SpriteURL: fmt.Sprintf("https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/%d.png", mon.id),
	}, nil
}

// Catch uses the classic rule: the higher the base experience, the harder.
func (d *DemoBackend) Catch(name string) (bool, error) {
	p, err := d.GetPokemon(name)
	if err != nil {
		return false, err
	}
	if d.Rand() > 50.0/float64(p.BaseExperience+10) {
		return false, nil
	}
	d.mu.Lock()
	d.caught[name] = p
	d.mu.Unlock()
	return true, nil
}

func (d *DemoBackend) Inspect(name string) (Pokemon, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	p, ok := d.caught[name]
	return p, ok
}

func (d *DemoBackend) Pokedex() []Pokemon {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]Pokemon, 0, len(d.caught))
	for _, p := range d.caught {
		out = append(out, p)
	}
	return out
}

func (d *DemoBackend) Evolutions(name string) ([]EvoStage, error) {
	for _, line := range demoEvolutions {
		for _, s := range line {
			if s.Name == name {
				return line, nil
			}
		}
	}
	return []EvoStage{{Name: name}}, nil
}

func (d *DemoBackend) Sprite(url string) ([]byte, error) {
	if d.Offline {
		return nil, ErrNotImplemented
	}
	client := http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sprite: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}
