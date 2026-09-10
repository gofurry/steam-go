package steamid_test

import (
	"errors"
	"math"
	"testing"

	"github.com/gofurry/steam-go/steamid"
)

func TestParse(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		input string
		want  uint64
		err   error
	}{
		{"12345", 76561197960278073, nil},
		{" \t00012345\n", 76561197960278073, nil},
		{"1", 76561197960265729, nil},
		{"4294967295", 76561202255233023, nil},
		{"76561197960278073", 76561197960278073, nil},
		{"STEAM_0:1:6172", 76561197960278073, nil},
		{"steam_1:1:6172", 76561197960278073, nil},
		{"[U:1:12345]", 76561197960278073, nil},
		{"[U:1:12345:2]", 76561202255245369, nil},
		{"", 0, steamid.ErrInvalidFormat},
		{" \n", 0, steamid.ErrInvalidFormat},
		{"0", 0, steamid.ErrInvalidID},
		{"0000", 0, steamid.ErrInvalidID},
		{"4294967296", 0, steamid.ErrInvalidID},
		{"76561197960265728", 0, steamid.ErrInvalidID},
		{"18446744073709551615", 0, steamid.ErrInvalidID},
		{"18446744073709551616", 0, steamid.ErrInvalidID},
		{"+12345", 0, steamid.ErrInvalidFormat},
		{"-1", 0, steamid.ErrInvalidFormat},
		{"1.0", 0, steamid.ErrInvalidFormat},
		{"1e3", 0, steamid.ErrInvalidFormat},
		{"0x12345", 0, steamid.ErrInvalidFormat},
		{"１２３４５", 0, steamid.ErrInvalidFormat},
		{"123 45", 0, steamid.ErrInvalidFormat},
		{"12345\x00", 0, steamid.ErrInvalidFormat},
		{"not-a-steamid", 0, steamid.ErrInvalidFormat},
		{"https://steamcommunity.com/profiles/76561197960278073", 0, steamid.ErrInvalidFormat},
		{"steam://friends/add/12345", 0, steamid.ErrInvalidFormat},
	} {
		t.Run(tt.input, func(t *testing.T) {
			got, err := steamid.Parse(tt.input)
			if got.Uint64() != tt.want || !errors.Is(err, tt.err) {
				t.Fatalf("Parse(%q) = %d, %v; want %d, %v", tt.input, got, err, tt.want, tt.err)
			}
			if err == nil {
				assertRoundTrips(t, got)
			}
		})
	}
}

func TestParseSteamID64(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"76561197960278073", " 00076561197960278073\n"} {
		got, err := steamid.ParseSteamID64(input)
		if err != nil || got.String() != "76561197960278073" {
			t.Errorf("ParseSteamID64(%q) = %d, %v", input, got, err)
		}
	}
	for _, input := range []string{"", "+76561197960278073", "-76561197960278073", "76561197960278073.0", "0x110000100003039", "[U:1:12345]", "STEAM_1:1:6172"} {
		if id, err := steamid.ParseSteamID64(input); id != 0 || !errors.Is(err, steamid.ErrInvalidFormat) {
			t.Errorf("ParseSteamID64(%q) = %d, %v; want ErrInvalidFormat", input, id, err)
		}
	}
	for _, input := range []string{"0", "12345", "76561197960265728", "18446744073709551615", "18446744073709551616"} {
		if id, err := steamid.ParseSteamID64(input); id != 0 || !errors.Is(err, steamid.ErrInvalidID) {
			t.Errorf("ParseSteamID64(%q) = %d, %v; want ErrInvalidID", input, id, err)
		}
	}
}

func TestSteam2(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		input     string
		canonical string
		universe  steamid.Universe
		account   uint32
	}{
		{"STEAM_0:1:6172", "STEAM_1:1:6172", 1, 12345},
		{" \tStEaM_1:1:6172\n", "STEAM_1:1:6172", 1, 12345},
		{"STEAM_2:0:6172", "STEAM_2:0:6172", 2, 12344},
		{"STEAM_3:1:0", "STEAM_3:1:0", 3, 1},
		{"STEAM_4:1:2147483647", "STEAM_4:1:2147483647", 4, math.MaxUint32},
		{"STEAM_1:0:2147483647", "STEAM_1:0:2147483647", 1, math.MaxUint32 - 1},
	} {
		t.Run(tt.input, func(t *testing.T) {
			id, err := steamid.ParseSteam2(tt.input)
			if err != nil || id.AccountID() != tt.account || id.Universe() != tt.universe || id.Instance() != 1 || id.AccountType() != 1 {
				t.Fatalf("ParseSteam2 = %d, %v", id, err)
			}
			text, err := id.Steam2()
			if err != nil || text != tt.canonical {
				t.Fatalf("Steam2 = %q, %v; want %q", text, err, tt.canonical)
			}
			assertRoundTrips(t, id)
		})
	}
	for _, tt := range []struct {
		input string
		err   error
	}{
		{"STEAM_x:y:z", steamid.ErrInvalidFormat},
		{"STEAM_1:1", steamid.ErrInvalidFormat},
		{"STEAM_1:1:2:3", steamid.ErrInvalidFormat},
		{"STEAM_1:00:1", steamid.ErrInvalidFormat},
		{"STEAM_1:1:-1", steamid.ErrInvalidFormat},
		{"STEAM_1:1:+1", steamid.ErrInvalidFormat},
		{"STEAM_1:1:1x", steamid.ErrInvalidFormat},
		{"prefixSTEAM_1:1:1", steamid.ErrInvalidFormat},
		{"STEAM_1: 1:1", steamid.ErrInvalidFormat},
		{"STEAM_1:2:1", steamid.ErrInvalidID},
		{"STEAM_0:0:0", steamid.ErrInvalidID},
		{"STEAM_1:0:0", steamid.ErrInvalidID},
		{"STEAM_5:1:1", steamid.ErrInvalidID},
		{"STEAM_256:1:1", steamid.ErrInvalidID},
		{"STEAM_1:0:2147483648", steamid.ErrInvalidID},
		{"STEAM_1:1:2147483648", steamid.ErrInvalidID},
		{"STEAM_1:1:4294967296", steamid.ErrInvalidID},
	} {
		if id, err := steamid.ParseSteam2(tt.input); id != 0 || !errors.Is(err, tt.err) {
			t.Errorf("ParseSteam2(%q) = %d, %v; want %v", tt.input, id, err, tt.err)
		}
	}
}

func TestSteam3(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		input     string
		canonical string
		kind      steamid.AccountType
		instance  uint32
	}{
		{"[U:1:12345]", "[U:1:12345]", 1, 1},
		{" [U:1:12345:1]\n", "[U:1:12345]", 1, 1},
		{"[U:1:12345:0]", "[U:1:12345:0]", 1, 0},
		{"[U:1:12345:2]", "[U:1:12345:2]", 1, 2},
		{"[U:1:12345(2)]", "[U:1:12345:2]", 1, 2},
		{"[U:1:12345:3]", "[U:1:12345:3]", 1, 3},
		{"[U:1:12345:4]", "[U:1:12345:4]", 1, 4},
		{"[M:1:12345]", "[M:1:12345:1]", 2, 1},
		{"[M:1:12345:123]", "[M:1:12345:123]", 2, 123},
		{"[G:1:12345]", "[G:1:12345]", 3, 1},
		{"[G:1:12345:0]", "[G:1:12345:0]", 3, 0},
		{"[A:1:12345]", "[A:1:12345:1]", 4, 1},
		{"[A:1:12345:123]", "[A:1:12345:123]", 4, 123},
		{"[P:1:12345]", "[P:1:12345]", 5, 1},
		{"[P:1:12345:4]", "[P:1:12345:4]", 5, 4},
		{"[C:1:12345]", "[C:1:12345]", 6, 1},
		{"[C:1:12345:0]", "[C:1:12345:0]", 6, 0},
		{"[g:1:12345]", "[g:1:12345]", 7, 0},
		{"[g:1:12345:0]", "[g:1:12345]", 7, 0},
		{"[T:1:12345]", "[T:1:12345]", 8, 0},
		{"[T:1:12345:123]", "[T:1:12345:123]", 8, 123},
		{"[c:1:12345]", "[c:1:12345]", 8, 524288},
		{"[c:1:12345:0]", "[c:1:12345]", 8, 524288},
		{"[c:1:12345:3]", "[c:1:12345:524291]", 8, 524291},
		{"[L:1:12345]", "[L:1:12345]", 8, 262144},
		{"[L:1:12345:7]", "[L:1:12345:262151]", 8, 262151},
		{"[T:1:12345:786432]", "[c:1:12345:786432]", 8, 786432},
		{"[L:1:12345:524288]", "[c:1:12345:786432]", 8, 786432},
		{"[T:1:12345:131072]", "[T:1:12345:131072]", 8, 131072},
		{"[T:1:12345:1048575]", "[c:1:12345:1048575]", 8, 1048575},
		{"[a:1:12345]", "[a:1:12345]", 10, 1},
		{"[a:1:12345:0]", "[a:1:12345:0]", 10, 0},
	} {
		t.Run(tt.input, func(t *testing.T) {
			id, err := steamid.ParseSteam3(tt.input)
			if err != nil || id.AccountID() != 12345 || id.Universe() != 1 || id.AccountType() != tt.kind || id.Instance() != tt.instance {
				t.Fatalf("ParseSteam3 = %d, %v; instance %d, type %d", id, err, id.Instance(), id.AccountType())
			}
			text, err := id.Steam3()
			if err != nil || text != tt.canonical {
				t.Fatalf("Steam3 = %q, %v; want %q", text, err, tt.canonical)
			}
			assertRoundTrips(t, id)
		})
	}
}

func TestSteam3Invalid(t *testing.T) {
	t.Parallel()
	for _, input := range []string{
		"", "[]", "U:1:12345", "[U:1:12345", "[U:1:12345]]", "[UU:1:12345]", "[u:1:12345]", "[I:1:12345]",
		"[i:1:12345]", "[X:1:12345]", "[9:1:12345]", "[l:1:12345]", "[U:x:y]", "[U:1:1:2:3]",
		"[U:1:-1]", "[U:1:+1]", "[U:1:1.0]", "[U:1:1:]", "[U:1:1: 2]", "[U:1:1:0x2]",
		"[U:1:1()]", "[U:1:1(2]", "[U:1:1(2):3]", "[U:1:1(2)(3)]", "[U:1:1(2))]", "[U:1:1)2(]",
	} {
		if id, err := steamid.ParseSteam3(input); id != 0 || !errors.Is(err, steamid.ErrInvalidFormat) {
			t.Errorf("ParseSteam3(%q) = %d, %v; want ErrInvalidFormat", input, id, err)
		}
	}
	for _, input := range []string{
		"[U:0:1]", "[U:5:1]", "[U:256:1]", "[U:1:0]", "[g:1:0]", "[g:1:1:1]", "[G:1:0]",
		"[U:1:4294967296]", "[U:1:1:5]", "[A:1:1:1048576]", "[c:1:1:1048576]", "[L:1:1:4294967296]",
	} {
		if id, err := steamid.ParseSteam3(input); id != 0 || !errors.Is(err, steamid.ErrInvalidID) {
			t.Errorf("ParseSteam3(%q) = %d, %v; want ErrInvalidID", input, id, err)
		}
	}
}
