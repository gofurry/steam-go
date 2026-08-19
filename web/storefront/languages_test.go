package storefront

import (
	"reflect"
	"testing"
)

func TestLookupLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		code  string
		tier  LanguageTier
	}{
		{input: "English", code: "en", tier: LanguageTierPlatform},
		{input: " english ", code: "en", tier: LanguageTierPlatform},
		{input: "EN", code: "en", tier: LanguageTierPlatform},
		{input: "Simplified Chinese", code: "zh-Hans", tier: LanguageTierPlatform},
		{input: "Chinese (Simplified)", code: "zh-Hans", tier: LanguageTierPlatform},
		{input: "schinese", code: "zh-Hans", tier: LanguageTierPlatform},
		{input: "zh-CN", code: "zh-Hans", tier: LanguageTierPlatform},
		{input: "zh-Hans", code: "zh-Hans", tier: LanguageTierPlatform},
		{input: "Chinese (Traditional)", code: "zh-Hant", tier: LanguageTierPlatform},
		{input: "tchinese", code: "zh-Hant", tier: LanguageTierPlatform},
		{input: "zh-TW", code: "zh-Hant", tier: LanguageTierPlatform},
		{input: "Japanese", code: "ja", tier: LanguageTierPlatform},
		{input: "koreana", code: "ko", tier: LanguageTierPlatform},
		{input: "Portuguese", code: "pt", tier: LanguageTierPlatform},
		{input: "Portuguese-Portugal", code: "pt", tier: LanguageTierPlatform},
		{input: "pt-BR", code: "pt-BR", tier: LanguageTierPlatform},
		{input: "Portuguese - Brazil", code: "pt-BR", tier: LanguageTierPlatform},
		{input: "Spanish-Spain", code: "es", tier: LanguageTierPlatform},
		{input: "latam", code: "es-419", tier: LanguageTierPlatform},
		{input: "Punjabi (Gurmukhi)", code: "pa-Guru", tier: LanguageTierGameOnly},
		{input: "pa-Arab", code: "pa-Arab", tier: LanguageTierGameOnly},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, ok := LookupLanguage(tt.input)
			if !ok || got.Code != tt.code || got.Tier != tt.tier {
				t.Fatalf("LookupLanguage(%q) = %#v, %v", tt.input, got, ok)
			}
		})
	}
	if _, ok := LookupLanguage("Example Future Language"); ok {
		t.Fatal("unknown language unexpectedly resolved")
	}
}

func TestLanguageDefinitionsCompleteAndCopied(t *testing.T) {
	t.Parallel()
	definitions := LanguageDefinitions()
	platform, gameOnly := 0, 0
	for _, definition := range definitions {
		switch definition.Tier {
		case LanguageTierPlatform:
			platform++
		case LanguageTierGameOnly:
			gameOnly++
		default:
			t.Fatalf("unexpected tier in registry: %#v", definition)
		}
		if got, ok := LookupLanguage(definition.Code); !ok || got != definition {
			t.Fatalf("registry entry is not lookup-addressable: %#v", definition)
		}
	}
	if platform != 31 || gameOnly != 72 {
		t.Fatalf("registry tiers = platform %d, game-only %d", platform, gameOnly)
	}
	definitions[0].Code = "changed"
	if got := LanguageDefinitions()[0].Code; got == "changed" {
		t.Fatal("LanguageDefinitions exposed registry backing storage")
	}
}

func TestParseSupportedLanguages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want []LanguageSupport
	}{
		{name: "empty", raw: "  ", want: []LanguageSupport{}},
		{name: "plain no footer", raw: "English, Japanese", want: []LanguageSupport{
			{Code: "en", SteamName: "English", SteamAPICode: "english", SteamWebCode: "en", Tier: LanguageTierPlatform, Known: true},
			{Code: "ja", SteamName: "Japanese", SteamAPICode: "japanese", SteamWebCode: "ja", Tier: LanguageTierPlatform, Known: true},
		}},
		{name: "current full audio", raw: "English<strong>*</strong>, Japanese, Korean<strong>*</strong><br><strong>*</strong>languages with full audio support", want: []LanguageSupport{
			languageSupportForTest("en", true), languageSupportForTest("ja", false), languageSupportForTest("ko", true),
		}},
		{name: "literal marker with footer", raw: "English*, Japanese<br>*languages with full audio support", want: []LanguageSupport{
			languageSupportForTest("en", true), languageSupportForTest("ja", false),
		}},
		{name: "marker without footer stays unknown state", raw: "English<strong>*</strong>", want: []LanguageSupport{
			{Code: "en", SteamName: "English", SteamAPICode: "english", SteamWebCode: "en", Tier: LanguageTierPlatform, Known: true},
		}},
		{name: "unknown preserved and marked", raw: "<b>Example &amp; Future Language</b><strong>*</strong>, English<br><strong>*</strong>languages with full audio support", want: []LanguageSupport{
			{SteamName: "Example & Future Language", Tier: LanguageTierUnknown, FullAudio: boolPointer(true), Known: false},
			languageSupportForTest("en", false),
		}},
		{name: "duplicates stable and true wins", raw: "Japanese, english, ENGLISH<strong>*</strong>, Spanish - Latin America, Portuguese - Portugal, Portuguese - Brazil<br><strong>*</strong>languages with full audio support", want: []LanguageSupport{
			languageSupportForTest("ja", false), languageSupportForTest("en", true), languageSupportForTest("es-419", false), languageSupportForTest("pt", false), languageSupportForTest("pt-BR", false),
		}},
		{name: "game only", raw: "Afrikaans, Valencian", want: []LanguageSupport{
			{Code: "af", SteamName: "Afrikaans", Tier: LanguageTierGameOnly, Known: true},
			{Code: "ca-valencia", SteamName: "Valencian", Tier: LanguageTierGameOnly, Known: true},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseSupportedLanguages(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseSupportedLanguages() =\n%#v\nwant\n%#v", got, tt.want)
			}
		})
	}
}

func TestLanguageTierConstants(t *testing.T) {
	t.Parallel()
	if LanguageTierPlatform != "platform" || LanguageTierGameOnly != "game_only" || LanguageTierUnknown != "unknown" {
		t.Fatalf("unexpected language tier constants: %q %q %q", LanguageTierPlatform, LanguageTierGameOnly, LanguageTierUnknown)
	}
}

func languageSupportForTest(code string, fullAudio bool) LanguageSupport {
	definition, ok := LookupLanguage(code)
	if !ok {
		panic("test language not registered: " + code)
	}
	return LanguageSupport{
		Code: definition.Code, SteamName: definition.SteamName,
		SteamAPICode: definition.SteamAPICode, SteamWebCode: definition.SteamWebCode,
		Tier: definition.Tier, FullAudio: boolPointer(fullAudio), Known: true,
	}
}

func boolPointer(value bool) *bool {
	return &value
}
