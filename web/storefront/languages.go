package storefront

import (
	"html"
	"regexp"
	"strings"
)

// LanguageTier describes whether Steam supports a language across the platform
// or only as game-provided content.
type LanguageTier string

const (
	// LanguageTierPlatform identifies a full Steam platform language.
	LanguageTierPlatform LanguageTier = "platform"
	// LanguageTierGameOnly identifies a language supported only by game content.
	LanguageTierGameOnly LanguageTier = "game_only"
	// LanguageTierUnknown identifies a language absent from the registry.
	LanguageTierUnknown LanguageTier = "unknown"
)

// LanguageDefinition describes one language known to Steam. Code is
// steam-go's stable canonical identifier; the Steam code fields preserve
// Valve-specific identifiers where Valve defines them.
type LanguageDefinition struct {
	Code         string       `json:"code"`
	SteamName    string       `json:"steam_name"`
	SteamAPICode string       `json:"steam_api_code,omitempty"`
	SteamWebCode string       `json:"steam_web_code,omitempty"`
	Tier         LanguageTier `json:"tier"`
}

// LanguageSupport is one language parsed from Storefront
// supported_languages text. Interface and Subtitles remain nil because that
// field does not reliably encode those columns.
type LanguageSupport struct {
	Code         string       `json:"code"`
	SteamName    string       `json:"steam_name"`
	SteamAPICode string       `json:"steam_api_code,omitempty"`
	SteamWebCode string       `json:"steam_web_code,omitempty"`
	Tier         LanguageTier `json:"tier"`
	Interface    *bool        `json:"interface,omitempty"`
	Subtitles    *bool        `json:"subtitles,omitempty"`
	FullAudio    *bool        `json:"full_audio,omitempty"`
	Known        bool         `json:"known"`
}

var languageDefinitions = []LanguageDefinition{
	{Code: "ar", SteamName: "Arabic", SteamAPICode: "arabic", SteamWebCode: "ar", Tier: LanguageTierPlatform},
	{Code: "bg", SteamName: "Bulgarian", SteamAPICode: "bulgarian", SteamWebCode: "bg", Tier: LanguageTierPlatform},
	{Code: "zh-Hans", SteamName: "Simplified Chinese", SteamAPICode: "schinese", SteamWebCode: "zh-CN", Tier: LanguageTierPlatform},
	{Code: "zh-Hant", SteamName: "Traditional Chinese", SteamAPICode: "tchinese", SteamWebCode: "zh-TW", Tier: LanguageTierPlatform},
	{Code: "cs", SteamName: "Czech", SteamAPICode: "czech", SteamWebCode: "cs", Tier: LanguageTierPlatform},
	{Code: "da", SteamName: "Danish", SteamAPICode: "danish", SteamWebCode: "da", Tier: LanguageTierPlatform},
	{Code: "nl", SteamName: "Dutch", SteamAPICode: "dutch", SteamWebCode: "nl", Tier: LanguageTierPlatform},
	{Code: "en", SteamName: "English", SteamAPICode: "english", SteamWebCode: "en", Tier: LanguageTierPlatform},
	{Code: "fi", SteamName: "Finnish", SteamAPICode: "finnish", SteamWebCode: "fi", Tier: LanguageTierPlatform},
	{Code: "fr", SteamName: "French", SteamAPICode: "french", SteamWebCode: "fr", Tier: LanguageTierPlatform},
	{Code: "de", SteamName: "German", SteamAPICode: "german", SteamWebCode: "de", Tier: LanguageTierPlatform},
	{Code: "el", SteamName: "Greek", SteamAPICode: "greek", SteamWebCode: "el", Tier: LanguageTierPlatform},
	{Code: "hu", SteamName: "Hungarian", SteamAPICode: "hungarian", SteamWebCode: "hu", Tier: LanguageTierPlatform},
	{Code: "id", SteamName: "Indonesian", SteamAPICode: "indonesian", SteamWebCode: "id", Tier: LanguageTierPlatform},
	{Code: "it", SteamName: "Italian", SteamAPICode: "italian", SteamWebCode: "it", Tier: LanguageTierPlatform},
	{Code: "ja", SteamName: "Japanese", SteamAPICode: "japanese", SteamWebCode: "ja", Tier: LanguageTierPlatform},
	{Code: "ko", SteamName: "Korean", SteamAPICode: "koreana", SteamWebCode: "ko", Tier: LanguageTierPlatform},
	{Code: "ms", SteamName: "Malay", SteamAPICode: "malay", SteamWebCode: "ms", Tier: LanguageTierPlatform},
	{Code: "no", SteamName: "Norwegian", SteamAPICode: "norwegian", SteamWebCode: "no", Tier: LanguageTierPlatform},
	{Code: "pl", SteamName: "Polish", SteamAPICode: "polish", SteamWebCode: "pl", Tier: LanguageTierPlatform},
	{Code: "pt", SteamName: "Portuguese - Portugal", SteamAPICode: "portuguese", SteamWebCode: "pt", Tier: LanguageTierPlatform},
	{Code: "pt-BR", SteamName: "Portuguese - Brazil", SteamAPICode: "brazilian", SteamWebCode: "pt-BR", Tier: LanguageTierPlatform},
	{Code: "ro", SteamName: "Romanian", SteamAPICode: "romanian", SteamWebCode: "ro", Tier: LanguageTierPlatform},
	{Code: "ru", SteamName: "Russian", SteamAPICode: "russian", SteamWebCode: "ru", Tier: LanguageTierPlatform},
	{Code: "es", SteamName: "Spanish - Spain", SteamAPICode: "spanish", SteamWebCode: "es", Tier: LanguageTierPlatform},
	{Code: "es-419", SteamName: "Spanish - Latin America", SteamAPICode: "latam", SteamWebCode: "es-419", Tier: LanguageTierPlatform},
	{Code: "sv", SteamName: "Swedish", SteamAPICode: "swedish", SteamWebCode: "sv", Tier: LanguageTierPlatform},
	{Code: "th", SteamName: "Thai", SteamAPICode: "thai", SteamWebCode: "th", Tier: LanguageTierPlatform},
	{Code: "tr", SteamName: "Turkish", SteamAPICode: "turkish", SteamWebCode: "tr", Tier: LanguageTierPlatform},
	{Code: "uk", SteamName: "Ukrainian", SteamAPICode: "ukrainian", SteamWebCode: "uk", Tier: LanguageTierPlatform},
	{Code: "vi", SteamName: "Vietnamese", SteamAPICode: "vietnamese", SteamWebCode: "vi", Tier: LanguageTierPlatform},

	{Code: "af", SteamName: "Afrikaans", Tier: LanguageTierGameOnly},
	{Code: "sq", SteamName: "Albanian", Tier: LanguageTierGameOnly},
	{Code: "am", SteamName: "Amharic", Tier: LanguageTierGameOnly},
	{Code: "hy", SteamName: "Armenian", Tier: LanguageTierGameOnly},
	{Code: "as", SteamName: "Assamese", Tier: LanguageTierGameOnly},
	{Code: "az", SteamName: "Azerbaijani", Tier: LanguageTierGameOnly},
	{Code: "bn", SteamName: "Bangla", Tier: LanguageTierGameOnly},
	{Code: "eu", SteamName: "Basque", Tier: LanguageTierGameOnly},
	{Code: "be", SteamName: "Belarusian", Tier: LanguageTierGameOnly},
	{Code: "bs", SteamName: "Bosnian", Tier: LanguageTierGameOnly},
	{Code: "ca", SteamName: "Catalan", Tier: LanguageTierGameOnly},
	{Code: "chr", SteamName: "Cherokee", Tier: LanguageTierGameOnly},
	{Code: "hr", SteamName: "Croatian", Tier: LanguageTierGameOnly},
	{Code: "prs", SteamName: "Dari", Tier: LanguageTierGameOnly},
	{Code: "et", SteamName: "Estonian", Tier: LanguageTierGameOnly},
	{Code: "fil", SteamName: "Filipino", Tier: LanguageTierGameOnly},
	{Code: "gl", SteamName: "Galician", Tier: LanguageTierGameOnly},
	{Code: "ka", SteamName: "Georgian", Tier: LanguageTierGameOnly},
	{Code: "gu", SteamName: "Gujarati", Tier: LanguageTierGameOnly},
	{Code: "pa-Guru", SteamName: "Punjabi (Gurmukhi)", Tier: LanguageTierGameOnly},
	{Code: "ha", SteamName: "Hausa", Tier: LanguageTierGameOnly},
	{Code: "he", SteamName: "Hebrew", Tier: LanguageTierGameOnly},
	{Code: "hi", SteamName: "Hindi", Tier: LanguageTierGameOnly},
	{Code: "is", SteamName: "Icelandic", Tier: LanguageTierGameOnly},
	{Code: "ig", SteamName: "Igbo", Tier: LanguageTierGameOnly},
	{Code: "ga", SteamName: "Irish", Tier: LanguageTierGameOnly},
	{Code: "kn", SteamName: "Kannada", Tier: LanguageTierGameOnly},
	{Code: "kk", SteamName: "Kazakh", Tier: LanguageTierGameOnly},
	{Code: "km", SteamName: "Khmer", Tier: LanguageTierGameOnly},
	{Code: "quc", SteamName: "K'iche'", Tier: LanguageTierGameOnly},
	{Code: "rw", SteamName: "Kinyarwanda", Tier: LanguageTierGameOnly},
	{Code: "kok", SteamName: "Konkani", Tier: LanguageTierGameOnly},
	{Code: "ky", SteamName: "Kyrgyz", Tier: LanguageTierGameOnly},
	{Code: "lv", SteamName: "Latvian", Tier: LanguageTierGameOnly},
	{Code: "lt", SteamName: "Lithuanian", Tier: LanguageTierGameOnly},
	{Code: "lb", SteamName: "Luxembourgish", Tier: LanguageTierGameOnly},
	{Code: "mk", SteamName: "Macedonian", Tier: LanguageTierGameOnly},
	{Code: "ml", SteamName: "Malayalam", Tier: LanguageTierGameOnly},
	{Code: "mt", SteamName: "Maltese", Tier: LanguageTierGameOnly},
	{Code: "mi", SteamName: "Maori", Tier: LanguageTierGameOnly},
	{Code: "mr", SteamName: "Marathi", Tier: LanguageTierGameOnly},
	{Code: "mn", SteamName: "Mongolian", Tier: LanguageTierGameOnly},
	{Code: "ne", SteamName: "Nepali", Tier: LanguageTierGameOnly},
	{Code: "or", SteamName: "Odia", Tier: LanguageTierGameOnly},
	{Code: "fa", SteamName: "Persian", Tier: LanguageTierGameOnly},
	{Code: "qu", SteamName: "Quechua", Tier: LanguageTierGameOnly},
	{Code: "sco", SteamName: "Scots", Tier: LanguageTierGameOnly},
	{Code: "sr", SteamName: "Serbian", Tier: LanguageTierGameOnly},
	{Code: "pa-Arab", SteamName: "Punjabi (Shahmukhi)", Tier: LanguageTierGameOnly},
	{Code: "sd", SteamName: "Sindhi", Tier: LanguageTierGameOnly},
	{Code: "si", SteamName: "Sinhala", Tier: LanguageTierGameOnly},
	{Code: "sk", SteamName: "Slovak", Tier: LanguageTierGameOnly},
	{Code: "sl", SteamName: "Slovenian", Tier: LanguageTierGameOnly},
	{Code: "ckb", SteamName: "Sorani", Tier: LanguageTierGameOnly},
	{Code: "st", SteamName: "Sotho", Tier: LanguageTierGameOnly},
	{Code: "sw", SteamName: "Swahili", Tier: LanguageTierGameOnly},
	{Code: "tg", SteamName: "Tajik", Tier: LanguageTierGameOnly},
	{Code: "ta", SteamName: "Tamil", Tier: LanguageTierGameOnly},
	{Code: "tt", SteamName: "Tatar", Tier: LanguageTierGameOnly},
	{Code: "te", SteamName: "Telugu", Tier: LanguageTierGameOnly},
	{Code: "ti", SteamName: "Tigrinya", Tier: LanguageTierGameOnly},
	{Code: "tn", SteamName: "Tswana", Tier: LanguageTierGameOnly},
	{Code: "tk", SteamName: "Turkmen", Tier: LanguageTierGameOnly},
	{Code: "ur", SteamName: "Urdu", Tier: LanguageTierGameOnly},
	{Code: "ug", SteamName: "Uyghur", Tier: LanguageTierGameOnly},
	{Code: "uz", SteamName: "Uzbek", Tier: LanguageTierGameOnly},
	{Code: "ca-valencia", SteamName: "Valencian", Tier: LanguageTierGameOnly},
	{Code: "cy", SteamName: "Welsh", Tier: LanguageTierGameOnly},
	{Code: "wo", SteamName: "Wolof", Tier: LanguageTierGameOnly},
	{Code: "xh", SteamName: "Xhosa", Tier: LanguageTierGameOnly},
	{Code: "yo", SteamName: "Yoruba", Tier: LanguageTierGameOnly},
	{Code: "zu", SteamName: "Zulu", Tier: LanguageTierGameOnly},
}

const fullAudioMarker = "__steam_go_full_audio__"

var (
	strongAsteriskPattern = regexp.MustCompile(`(?is)<strong[^>]*>\s*\*\s*</strong>`)
	breakTagPattern       = regexp.MustCompile(`(?is)<br\s*/?>`)
	htmlTagPattern        = regexp.MustCompile(`(?is)<[^>]+>`)
	footerPattern         = regexp.MustCompile(`(?i)(?:` + fullAudioMarker + `|\*)?\s*languages\s+with\s+full\s+audio\s+support`)
	hyphenSpacingPattern  = regexp.MustCompile(`\s*-\s*`)
)

var languageLookup = buildLanguageLookup()

// LanguageDefinitions returns a copy of the complete Steam language registry.
func LanguageDefinitions() []LanguageDefinition {
	return append([]LanguageDefinition(nil), languageDefinitions...)
}

// LookupLanguage resolves a canonical code, Steam API code, Steam Web code,
// Storefront name, or supported official-name alias without fuzzy matching.
func LookupLanguage(value string) (LanguageDefinition, bool) {
	definition, ok := languageLookup[normalizeLanguageKey(value)]
	return definition, ok
}

// ParseSupportedLanguages parses Storefront supported_languages HTML while
// preserving unknown languages and the optional full-audio marker semantics.
func ParseSupportedLanguages(raw string) []LanguageSupport {
	if strings.TrimSpace(raw) == "" {
		return []LanguageSupport{}
	}

	cleaned := strongAsteriskPattern.ReplaceAllString(raw, fullAudioMarker)
	cleaned = breakTagPattern.ReplaceAllString(cleaned, "\n")
	cleaned = htmlTagPattern.ReplaceAllString(cleaned, "")
	cleaned = html.UnescapeString(cleaned)

	hasFullAudioFooter := false
	if location := footerPattern.FindStringIndex(cleaned); location != nil {
		hasFullAudioFooter = true
		cleaned = cleaned[:location[0]]
	}
	cleaned = strings.ReplaceAll(cleaned, "\n", ",")

	result := make([]LanguageSupport, 0)
	seen := make(map[string]int)
	for _, entry := range strings.Split(cleaned, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		marked := strings.Contains(entry, fullAudioMarker)
		entry = strings.ReplaceAll(entry, fullAudioMarker, "")
		if strings.HasSuffix(strings.TrimSpace(entry), "*") {
			marked = true
			entry = strings.TrimSuffix(strings.TrimSpace(entry), "*")
		}
		name := strings.TrimSpace(entry)
		if name == "" {
			continue
		}

		definition, known := LookupLanguage(name)
		support := LanguageSupport{
			SteamName: name,
			Tier:      LanguageTierUnknown,
			Known:     known,
		}
		key := "unknown:" + normalizeLanguageKey(name)
		if known {
			support.Code = definition.Code
			support.SteamName = definition.SteamName
			support.SteamAPICode = definition.SteamAPICode
			support.SteamWebCode = definition.SteamWebCode
			support.Tier = definition.Tier
			key = "known:" + strings.ToLower(definition.Code)
		}
		if hasFullAudioFooter {
			value := marked
			support.FullAudio = &value
		}

		if index, ok := seen[key]; ok {
			if support.FullAudio != nil && *support.FullAudio {
				value := true
				result[index].FullAudio = &value
			}
			continue
		}
		seen[key] = len(result)
		result = append(result, support)
	}
	return result
}

func buildLanguageLookup() map[string]LanguageDefinition {
	lookup := make(map[string]LanguageDefinition, len(languageDefinitions)*4)
	for _, definition := range languageDefinitions {
		for _, value := range []string{definition.Code, definition.SteamName, definition.SteamAPICode, definition.SteamWebCode} {
			if key := normalizeLanguageKey(value); key != "" {
				lookup[key] = definition
			}
		}
	}
	aliases := map[string]string{
		"Chinese (Simplified)":  "zh-Hans",
		"Chinese (Traditional)": "zh-Hant",
		"Portuguese":            "pt",
		"Portuguese-Portugal":   "pt",
		"Portuguese-Brazil":     "pt-BR",
		"Spanish-Spain":         "es",
		"Spanish-Latin America": "es-419",
	}
	for alias, canonical := range aliases {
		lookup[normalizeLanguageKey(alias)] = lookup[normalizeLanguageKey(canonical)]
	}
	return lookup
}

func normalizeLanguageKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Join(strings.Fields(value), " ")
	return hyphenSpacingPattern.ReplaceAllString(value, "-")
}
