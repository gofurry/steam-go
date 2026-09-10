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

func TestParseCommunityURLPercentEncoding(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		path  string
		extra string
		err   error
	}{
		{"encoded profiles", "/%70rofiles/76561197960278073", "", steamid.ErrInvalidFormat},
		{"encoded id", "/%69d/example", "", steamid.ErrInvalidFormat},
		{"encoded digit", "/profiles/%376561197960278073", "", steamid.ErrInvalidFormat},
		{"encoded numeric id", "/profiles/%37%36%35%36%31%31%39%37%39%36%30%32%37%38%30%37%33", "", steamid.ErrInvalidFormat},
		{"encoded separator", "/profiles%2f76561197960278073", "", steamid.ErrInvalidFormat},
		{"encoded trailing slash", "/profiles/76561197960278073%2F", "", steamid.ErrInvalidFormat},
		{"encoded vanity", "/id/%65xample", "", steamid.ErrInvalidFormat},
		{"canonical escaping", "/id/example%20name", "", steamid.ErrInvalidFormat},
		{"escaped percent", "/id/example%25name", "", steamid.ErrInvalidFormat},
		{"query encoding", "/profiles/76561197960278073", "?next=%2F%70rofiles%2F%37&token=offline-sensitive-value", nil},
		{"fragment encoding", "/profiles/76561197960278073", "#%69d%2Fexample%20name", nil},
		{"query and fragment", "/profiles/76561197960278073/", "?q=%E4%B8%AD#%23%25", nil},
		{"vanity query and fragment", "/id/example", "?q=%70#%69d", steamid.ErrVanityReference},
		{"encoded path with sensitive query", "/%70rofiles/76561197960278073", "?token=offline-sensitive-value", steamid.ErrInvalidFormat},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := "https://steamcommunity.com" + tt.path + tt.extra
			id, err := steamid.ParseCommunityURL(input)
			if !errors.Is(err, tt.err) {
				t.Fatalf("ParseCommunityURL = %d, %v; want %v", id, err, tt.err)
			}
			if err != nil {
				if id != 0 || strings.Contains(err.Error(), input) || strings.Contains(err.Error(), "offline-sensitive-value") {
					t.Fatal("failure must return zero without echoing the URL or sensitive content")
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
