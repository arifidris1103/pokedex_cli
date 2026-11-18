package pokeapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A fake PokéAPI that counts requests, so we can check the cache.
func fakeAPI(t *testing.T) (*httptest.Server, *int) {
	t.Helper()
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path == "/pokemon/missingno" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `{"id": 25, "name": "pikachu", "base_experience": 112,
			"types": [{"slot": 1, "type": {"name": "electric"}}]}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func TestGetDecodesAndCaches(t *testing.T) {
	srv, hits := fakeAPI(t)
	c := NewClient(time.Minute)

	for range 3 {
		p, err := get[Pokemon](&c, srv.URL+"/pokemon/pikachu")
		if err != nil {
			t.Fatal(err)
		}
		if p.ID != 25 || p.Name != "pikachu" || p.Types[0].Type.Name != "electric" {
			t.Fatalf("decoded wrong: %+v", p)
		}
	}
	if *hits != 1 {
		t.Fatalf("server hit %d times, want 1 (the rest should come from the cache)", *hits)
	}
}

func TestGetBadStatus(t *testing.T) {
	srv, _ := fakeAPI(t)
	c := NewClient(time.Minute)

	if _, err := get[Pokemon](&c, srv.URL+"/pokemon/missingno"); err == nil {
		t.Fatal("expected an error for a 404")
	}
}
