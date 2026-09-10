package steamid_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/gofurry/steam-go/steamid"
)

func TestParseCommunityURL(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		input string
		err   error
	}{
		{"https://steamcommunity.com/profiles/76561197960278073", nil},
		{"http://steamcommunity.com/profiles/76561197960278073/", nil},
		{"https://www.steamcommunity.com/profiles/76561197960278073/?l=english#profile", nil},
		{" HTTPS://STEAMCOMMUNITY.COM:443/profiles/76561197960278073 \n", nil},
		{"https://steamcommunity.com/profiles/00076561197960278073", nil},
		{"https://steamcommunity.com/profiles/76561197960278073?next=https://example.org", nil},
		{"https://steamcommunity.com/id/example-user", steamid.ErrVanityReference},
		{"http://www.steamcommunity.com/id/example-user/?l=english#profile", steamid.ErrVanityReference},
		{"https://steamcommunity.com/id/12345", steamid.ErrVanityReference},
		{"", steamid.ErrInvalidFormat},
		{"steamcommunity.com/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"//steamcommunity.com/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https:steamcommunity.com/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"ftp://steamcommunity.com/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"steam://profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com.example.org/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https://example.org/steamcommunity.com/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com@evil.example/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https://user:offline-secret@steamcommunity.com/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com./profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com:bad/profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/profiles/76561197960278073/extra", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/profiles/76561197960278073//", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com//profiles/76561197960278073", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/profiles/", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/profiles/%2076561197960278073", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/profiles/76561197960278073%20", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/profiles/%zz", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/profiles/12345", steamid.ErrInvalidID},
		{"https://steamcommunity.com/profiles/0", steamid.ErrInvalidID},
		{"https://steamcommunity.com/profiles/103582791429533753", steamid.ErrInvalidID},
		{"https://steamcommunity.com/profiles/18446744073709551616", steamid.ErrInvalidID},
		{"https://steamcommunity.com/gid/103582791429533753", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/groups/example", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/sharedfiles/filedetails/?id=12345", steamid.ErrInvalidFormat},
		{"https://s.team/p/example", steamid.ErrInvalidFormat},
		{"https://example.org/id/example", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/id/", steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/id/example/extra", steamid.ErrInvalidFormat},
	} {
		t.Run(tt.input, func(t *testing.T) {
			id, err := steamid.ParseCommunityURL(tt.input)
			if !errors.Is(err, tt.err) {
				t.Fatalf("ParseCommunityURL = %d, %v; want %v", id, err, tt.err)
			}
			if err != nil {
				if id != 0 || strings.Contains(err.Error(), "offline-secret") {
					t.Fatal("failure must return zero without echoing URL credentials")
				}
				return
			}
			if id.Uint64() != 76561197960278073 {
				t.Fatalf("unexpected profile ID: %d", id)
			}
			assertRoundTrips(t, id)
		})
	}
}

func TestProfileURLRequiresIndividual(t *testing.T) {
	t.Parallel()
	for kind := steamid.AccountTypeIndividual; kind <= steamid.AccountTypeAnonUser; kind++ {
		id, err := steamid.New(12345, steamid.UniversePublic, kind, steamid.InstanceAll)
		if err != nil {
			t.Fatal(err)
		}
		got, err := steamid.ParseCommunityURL("https://steamcommunity.com/profiles/" + id.String())
		if kind == steamid.AccountTypeIndividual {
			if err != nil || got != id {
				t.Fatalf("Individual URL = %d, %v", got, err)
			}
		} else if got != 0 || !errors.Is(err, steamid.ErrInvalidID) {
			t.Fatalf("profile URL for type %d = %d, %v", kind, got, err)
		}
	}
}
