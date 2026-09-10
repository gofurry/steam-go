package steamid_test

import (
	"errors"
	"math"
	"testing"

	"github.com/gofurry/steam-go/steamid"
)

func TestComponents(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		raw      uint64
		account  uint32
		instance uint32
		universe steamid.Universe
		kind     steamid.AccountType
	}{
		{76561197960278073, 12345, 1, 1, 1},
		{0, 0, 0, 0, 0},
		{0xfedcba9876543210, 0x76543210, 0xcba98, 0xfe, 0xd},
		{math.MaxUint64, math.MaxUint32, 0xfffff, 255, 15},
	} {
		id := steamid.ID(tt.raw)
		if id.Uint64() != tt.raw || id.AccountID() != tt.account || id.Instance() != tt.instance ||
			id.Universe() != tt.universe || id.AccountType() != tt.kind {
			t.Errorf("components of %d: account=%d instance=%d universe=%d type=%d", tt.raw, id.AccountID(), id.Instance(), id.Universe(), id.AccountType())
		}
	}
	if steamid.ID(math.MaxUint64).String() != "18446744073709551615" || steamid.ID(0).String() != "0" {
		t.Fatal("String must format invalid IDs without validation")
	}
}

func TestEnums(t *testing.T) {
	t.Parallel()
	for value, universe := range []steamid.Universe{steamid.UniverseInvalid, steamid.UniversePublic, steamid.UniverseBeta, steamid.UniverseInternal, steamid.UniverseDev} {
		if int(universe) != value {
			t.Errorf("universe = %d, want %d", universe, value)
		}
	}
	for value, kind := range []steamid.AccountType{
		steamid.AccountTypeInvalid, steamid.AccountTypeIndividual, steamid.AccountTypeMultiseat,
		steamid.AccountTypeGameServer, steamid.AccountTypeAnonGameServer, steamid.AccountTypePending,
		steamid.AccountTypeContentServer, steamid.AccountTypeClan, steamid.AccountTypeChat,
		steamid.AccountTypeConsoleUser, steamid.AccountTypeAnonUser,
	} {
		if int(kind) != value {
			t.Errorf("account type = %d, want %d", kind, value)
		}
	}
	if steamid.InstanceAll != 0 || steamid.InstanceDesktop != 1 || steamid.InstanceConsole != 2 || steamid.InstanceWeb != 4 {
		t.Fatal("unexpected instance constants")
	}
}

func TestNewAndValid(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		account  uint32
		universe steamid.Universe
		kind     steamid.AccountType
		instance uint32
		valid    bool
	}{
		{"individual", 12345, 1, 1, 1, true},
		{"individual all", 1, 1, 1, 0, true},
		{"individual console", 1, 1, 1, 2, true},
		{"individual combined", 1, 1, 1, 3, true},
		{"individual web", 1, 1, 1, 4, true},
		{"individual instance too large", 1, 1, 1, 5, false},
		{"individual zero", 0, 1, 1, 1, false},
		{"clan", 1, 1, 7, 0, true},
		{"clan zero", 0, 1, 7, 0, false},
		{"clan instance", 1, 1, 7, 1, false},
		{"server", 1, 1, 3, 0xfffff, true},
		{"server zero", 0, 1, 3, 1, false},
		{"multiseat zero", 0, 1, 2, 0xfffff, true},
		{"anon server zero", 0, 1, 4, 0, true},
		{"pending zero", 0, 1, 5, 0xfffff, true},
		{"content zero", 0, 1, 6, 0xfffff, true},
		{"chat zero", 0, 1, 8, 0xfffff, true},
		{"console zero", 0, 1, 9, 0xfffff, true},
		{"anon user zero", 0, 1, 10, 0xfffff, true},
		{"invalid universe", 1, 0, 1, 1, false},
		{"future universe", 1, 5, 1, 1, false},
		{"universe max", 1, 255, 1, 1, false},
		{"beta", 1, 2, 1, 1, true},
		{"internal", 1, 3, 1, 1, true},
		{"dev", math.MaxUint32, 4, 1, 1, true},
		{"invalid type", 1, 1, 0, 1, false},
		{"future type", 1, 1, 11, 1, false},
		{"packed type max", 1, 1, 15, 1, false},
		{"type overflow", 1, 1, 17, 1, false},
		{"type max", 1, 1, 255, 1, false},
		{"instance overflow", 1, 1, 8, 0x100000, false},
		{"instance max uint32", 1, 1, 8, math.MaxUint32, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			id, err := steamid.New(tt.account, tt.universe, tt.kind, tt.instance)
			if !tt.valid {
				if id != 0 || !errors.Is(err, steamid.ErrInvalidID) {
					t.Fatalf("New = %d, %v; want zero, ErrInvalidID", id, err)
				}
			} else {
				if err != nil || !id.Valid() || id.AccountID() != tt.account || id.Universe() != tt.universe || id.AccountType() != tt.kind || id.Instance() != tt.instance {
					t.Fatalf("New = %d, %v; lost components", id, err)
				}
				assertRoundTrips(t, id)
			}
			// Independently exercise Valid on raw casts within the packed widths.
			if tt.kind <= 15 && tt.instance <= 0xfffff {
				raw := uint64(tt.account) | uint64(tt.instance)<<32 | uint64(tt.kind)<<52 | uint64(tt.universe)<<56
				if steamid.ID(raw).Valid() != tt.valid {
					t.Fatalf("Valid on raw %d = %v, want %v", raw, steamid.ID(raw).Valid(), tt.valid)
				}
			}
		})
	}
}

func TestNewIndividual(t *testing.T) {
	t.Parallel()
	for _, account := range []uint32{0, 1, 12345, math.MaxUint32} {
		id, err := steamid.NewIndividual(account)
		if account == 0 {
			if id != 0 || !errors.Is(err, steamid.ErrInvalidID) {
				t.Fatalf("NewIndividual(0) = %d, %v", id, err)
			}
			continue
		}
		if err != nil || id.Uint64() != 76561197960265728+uint64(account) {
			t.Fatalf("NewIndividual(%d) = %d, %v", account, id, err)
		}
	}
}

func assertRoundTrips(t testing.TB, id steamid.ID) {
	t.Helper()
	if !id.Valid() {
		t.Fatalf("invalid successful parse: %d", id)
	}
	if got, err := steamid.ParseSteamID64(id.String()); err != nil || got != id {
		t.Fatalf("SteamID64 round-trip of %d = %d, %v", id, got, err)
	}
	if got, err := steamid.Parse(id.String()); err != nil || got != id {
		t.Fatalf("generic round-trip of %d = %d, %v", id, got, err)
	}
	if got, err := steamid.New(id.AccountID(), id.Universe(), id.AccountType(), id.Instance()); err != nil || got != id {
		t.Fatalf("component round-trip of %d = %d, %v", id, got, err)
	}
	steam2, err := id.Steam2()
	if id.AccountType() == steamid.AccountTypeIndividual && id.Instance() == steamid.InstanceDesktop {
		if err != nil {
			t.Fatalf("Steam2 conversion of %d: %v", id, err)
		}
		if got, err := steamid.ParseSteam2(steam2); err != nil || got != id {
			t.Fatalf("Steam2 round-trip %q: %d, %v", steam2, got, err)
		}
	} else if steam2 != "" || !errors.Is(err, steamid.ErrUnsupportedConversion) {
		t.Fatalf("Steam2 of unrepresentable %d = %q, %v", id, steam2, err)
	}
	steam3, err := id.Steam3()
	if id.AccountType() == steamid.AccountTypeConsoleUser {
		if steam3 != "" || !errors.Is(err, steamid.ErrUnsupportedConversion) {
			t.Fatalf("Steam3 of ConsoleUser = %q, %v", steam3, err)
		}
	} else {
		if err != nil {
			t.Fatalf("Steam3 conversion of %d: %v", id, err)
		}
		if got, err := steamid.ParseSteam3(steam3); err != nil || got != id {
			t.Fatalf("Steam3 round-trip %q of %d: %d, %v", steam3, id, got, err)
		}
	}
}

func TestInvalidConversions(t *testing.T) {
	t.Parallel()
	for _, id := range []steamid.ID{0, 12345, math.MaxUint64, 76561197960265728} {
		for _, convert := range []func() (string, error){id.Steam2, id.Steam3} {
			if text, err := convert(); text != "" || !errors.Is(err, steamid.ErrInvalidID) {
				t.Fatalf("invalid conversion of %d = %q, %v", id, text, err)
			}
		}
	}
}
