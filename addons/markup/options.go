package markup

const defaultSteamClanImageBase = "https://clan.fastly.steamstatic.com/images/"

type config struct {
	sanitize           bool
	allowYouTubeEmbed  bool
	allowVideo         bool
	steamClanImageBase string
	linkRel            string
	linkTarget         string
}

// Option configures Steam markup conversion and cleaning.
type Option func(*config)

func defaultConfig() config {
	return config{
		sanitize:           true,
		allowYouTubeEmbed:  false,
		allowVideo:         true,
		steamClanImageBase: defaultSteamClanImageBase,
		linkRel:            "nofollow noopener noreferrer",
		linkTarget:         "_blank",
	}
}

// WithSanitize toggles HTML sanitization. It is enabled by default.
func WithSanitize(enabled bool) Option {
	return func(cfg *config) {
		cfg.sanitize = enabled
	}
}

// WithYouTubeEmbeds toggles YouTube iframe output for [youtube] tags.
func WithYouTubeEmbeds(enabled bool) Option {
	return func(cfg *config) {
		cfg.allowYouTubeEmbed = enabled
	}
}

// WithVideoTags toggles video output for [video] tags.
func WithVideoTags(enabled bool) Option {
	return func(cfg *config) {
		cfg.allowVideo = enabled
	}
}

// WithSteamClanImageBase overrides the replacement base for {STEAM_CLAN_IMAGE}.
func WithSteamClanImageBase(base string) Option {
	return func(cfg *config) {
		if base != "" {
			cfg.steamClanImageBase = base
		}
	}
}

// WithLinkRel sets the rel attribute generated for BBCode links.
func WithLinkRel(rel string) Option {
	return func(cfg *config) {
		cfg.linkRel = rel
	}
}

// WithLinkTarget sets the target attribute generated for BBCode links.
func WithLinkTarget(target string) Option {
	return func(cfg *config) {
		cfg.linkTarget = target
	}
}

func resolveConfig(opts []Option) config {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}
