package gameserversservice

import "testing"

func TestFilterStringEmpty(t *testing.T) {
	t.Parallel()

	if got := NewFilter().String(); got != "" {
		t.Fatalf("empty filter = %q, want empty string", got)
	}
}

func TestFilterBuilderString(t *testing.T) {
	t.Parallel()

	filter := NewFilter().
		AppID(730).
		NotAppID(440).
		Secure().
		GameDir(" csgo ").
		Map(" de_dust2 ").
		Linux().
		NotEmpty().
		NotFull().
		NoPlayers().
		Proxy().
		White().
		Dedicated().
		NoPassword().
		Password().
		GameType(" coop ", "versus").
		GameData("tag1", " tag2 ").
		GameDataOr("or1", " or2 ").
		NameMatch(" *surf* ").
		VersionMatch(" 1.39.* ").
		CollapseAddrHash().
		GameAddr(" 1.2.3.4:27015 ").
		Type(ServerTypeDedicated).
		Raw("custom", "value")

	want := `\appid\730\napp\440\secure\1\gamedir\csgo\map\de_dust2\linux\1\empty\1\full\1\noplayers\1\proxy\1\white\1\dedicated\1\password\0\password\1\gametype\coop,versus\gamedata\tag1,tag2\gamedataor\or1,or2\name_match\*surf*\version_match\1.39.*\collapse_addr_hash\1\gameaddr\1.2.3.4:27015\type\d\custom\value`
	if got := filter.String(); got != want {
		t.Fatalf("filter = %q, want %q", got, want)
	}
}

func TestFilterBuilderIsImmutable(t *testing.T) {
	t.Parallel()

	base := NewFilter().AppID(730)
	secure := base.Secure()
	notFull := base.NotFull()

	if got, want := base.String(), `\appid\730`; got != want {
		t.Fatalf("base filter = %q, want %q", got, want)
	}
	if got, want := secure.String(), `\appid\730\secure\1`; got != want {
		t.Fatalf("secure filter = %q, want %q", got, want)
	}
	if got, want := notFull.String(), `\appid\730\full\1`; got != want {
		t.Fatalf("notFull filter = %q, want %q", got, want)
	}
}
