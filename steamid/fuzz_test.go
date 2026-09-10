package steamid_test

import (
	"errors"
	"math"
	"testing"

	"github.com/gofurry/steam-go/steamid"
)

func FuzzParse(f *testing.F) {
	for _, input := range []string{
		"", "0", "12345", "76561197960278073", "4294967295", "18446744073709551616",
		"STEAM_0:1:6172", "STEAM_4:1:2147483647", "STEAM_1:2:3", "[U:1:12345]",
		"[U:1:12345:2]", "[U:1:12345(2)]", "[U:1:12345()]", "[g:1:12345]", "[G:1:12345:0]",
		"[M:1:0:1048575]", "[A:1:12345:123]", "[P:2:0]", "[C:3:0:123]", "[a:4:0]",
		"[T:1:12345:131072]", "[c:1:12345:3]", "[L:1:12345:524289]", "[T:1:1:1048576]",
		"https://steamcommunity.com/profiles/76561197960278073/?l=english#profile",
		"https://steamcommunity.com/id/example", "https://steamcommunity.com.example.org/profiles/1",
		"https://steamcommunity.com/profiles/%20", "[U:x:y", "\xff\x00",
	} {
		f.Add(input)
	}
	parsers := []func(string) (steamid.ID, error){
		steamid.Parse, steamid.ParseSteamID64, steamid.ParseSteam2, steamid.ParseSteam3, steamid.ParseCommunityURL,
	}
	f.Fuzz(func(t *testing.T, input string) {
		for _, parse := range parsers {
			id, err := parse(input)
			if err == nil {
				assertRoundTrips(t, id)
			} else if id != 0 || (!errors.Is(err, steamid.ErrInvalidFormat) && !errors.Is(err, steamid.ErrInvalidID) && !errors.Is(err, steamid.ErrVanityReference)) {
				t.Fatalf("unclassified parser failure: %d, %v", id, err)
			}
		}
	})
}

func FuzzIDRoundTrip(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(math.MaxUint64))
	for universe := steamid.UniversePublic; universe <= steamid.UniverseDev; universe++ {
		for kind := steamid.AccountTypeIndividual; kind <= steamid.AccountTypeAnonUser; kind++ {
			for _, instance := range []uint32{0, 1, 2, 3, 4, 123, 1 << 17, 1 << 18, 1 << 19, (1 << 19) | (1 << 18) | 7, 0xfffff} {
				if id, err := steamid.New(math.MaxUint32, universe, kind, instance); err == nil {
					f.Add(id.Uint64())
				}
			}
		}
	}
	f.Fuzz(func(t *testing.T, raw uint64) {
		id := steamid.ID(raw)
		if id.Valid() {
			assertRoundTrips(t, id)
		} else {
			if _, err := steamid.ParseSteamID64(id.String()); !errors.Is(err, steamid.ErrInvalidID) {
				t.Fatalf("invalid ID %d parsed: %v", raw, err)
			}
			for _, convert := range []func() (string, error){id.Steam2, id.Steam3} {
				if text, err := convert(); text != "" || !errors.Is(err, steamid.ErrInvalidID) {
					t.Fatalf("invalid ID conversion %d: %q, %v", raw, text, err)
				}
			}
		}
	})
}
