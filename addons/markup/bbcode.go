package markup

import (
	stdhtml "html"
	"regexp"
	"strings"
)

type tagReplacement struct {
	re          *regexp.Regexp
	replacement string
}

var simpleTagReplacements = []tagReplacement{
	{regexp.MustCompile(`(?is)\[b\](.*?)\[/b\]`), `<strong>$1</strong>`},
	{regexp.MustCompile(`(?is)\[i\](.*?)\[/i\]`), `<em>$1</em>`},
	{regexp.MustCompile(`(?is)\[u\](.*?)\[/u\]`), `<u>$1</u>`},
	{regexp.MustCompile(`(?is)\[s\](.*?)\[/s\]`), `<s>$1</s>`},
	{regexp.MustCompile(`(?is)\[h1\](.*?)\[/h1\]`), `<h1>$1</h1>`},
	{regexp.MustCompile(`(?is)\[h2\](.*?)\[/h2\]`), `<h2>$1</h2>`},
	{regexp.MustCompile(`(?is)\[h3\](.*?)\[/h3\]`), `<h3>$1</h3>`},
	{regexp.MustCompile(`(?is)\[p\](.*?)\[/p\]`), `<p>$1</p>`},
	{regexp.MustCompile(`(?is)\[quote\](.*?)\[/quote\]`), `<blockquote>$1</blockquote>`},
}

var (
	urlWithTargetRe = regexp.MustCompile(`(?is)\[url=([^\]]+)\](.*?)\[/url\]`)
	urlBareRe       = regexp.MustCompile(`(?is)\[url\](.*?)\[/url\]`)
	imgBareRe       = regexp.MustCompile(`(?is)\[img\](.*?)\[/img\]`)
	imgSrcRe        = regexp.MustCompile(`(?is)\[img\s+src="(.*?)"\]\s*\[/img\]`)
	videoRe         = regexp.MustCompile(`(?is)\[video\](.*?)\[/video\]`)
	youtubeRe       = regexp.MustCompile(`(?is)\[youtube\](.*?)\[/youtube\]`)
	codeRe          = regexp.MustCompile(`(?is)\[code\](.*?)\[/code\]`)
)

// SteamBBCodeToHTML converts common Steam BBCode into sanitized HTML by default.
func SteamBBCodeToHTML(input string, opts ...Option) (string, error) {
	cfg := resolveConfig(opts)
	html := convertSteamBBCode(input, cfg)
	if !cfg.sanitize {
		return html, nil
	}
	return sanitizeHTML(html, cfg), nil
}

func convertSteamBBCode(input string, cfg config) string {
	imageBase := strings.TrimRight(cfg.steamClanImageBase, "/") + "/"
	input = strings.ReplaceAll(input, "{STEAM_CLAN_IMAGE}/", imageBase)
	input = strings.ReplaceAll(input, "{STEAM_CLAN_IMAGE}", cfg.steamClanImageBase)
	input = normalizeNewlines(input)

	input = imgSrcRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := imgSrcRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		return buildImage(parts[1])
	})

	input = imgBareRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := imgBareRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		return buildImage(parts[1])
	})

	input = urlWithTargetRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := urlWithTargetRe.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		return buildLink(stripBBCodeQuote(parts[1]), parts[2], cfg)
	})

	input = urlBareRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := urlBareRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		href := strings.TrimSpace(parts[1])
		return buildLink(href, href, cfg)
	})

	input = videoRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := videoRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		src := strings.TrimSpace(parts[1])
		if src == "" {
			return ""
		}
		if !cfg.allowVideo {
			return buildLink(src, src, cfg)
		}
		return `<video src="` + stdhtml.EscapeString(src) + `" controls></video>`
	})

	input = youtubeRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := youtubeRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		id := strings.TrimSpace(parts[1])
		if id == "" {
			return ""
		}
		if !cfg.allowYouTubeEmbed {
			return buildLink("https://www.youtube.com/watch?v="+id, "YouTube: "+id, cfg)
		}
		return `<iframe src="https://www.youtube.com/embed/` + stdhtml.EscapeString(id) + `" frameborder="0" allowfullscreen></iframe>`
	})

	input = codeRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := codeRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		return `<pre><code>` + stdhtml.EscapeString(parts[1]) + `</code></pre>`
	})

	for i := 0; i < 4; i++ {
		next := input
		for _, item := range simpleTagReplacements {
			next = item.re.ReplaceAllString(next, item.replacement)
		}
		next = parseLists(next)
		if next == input {
			break
		}
		input = next
	}

	input = unescapeSteamText(input)
	return strings.ReplaceAll(input, "\n", "<br>")
}

func buildImage(src string) string {
	src = strings.TrimSpace(src)
	if src == "" {
		return ""
	}
	return `<img src="` + stdhtml.EscapeString(src) + `" />`
}

func buildLink(href string, label string, cfg config) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return stdhtml.EscapeString(label)
	}
	attrs := ` href="` + stdhtml.EscapeString(href) + `"`
	if cfg.linkTarget != "" {
		attrs += ` target="` + stdhtml.EscapeString(cfg.linkTarget) + `"`
	}
	if cfg.linkRel != "" {
		attrs += ` rel="` + stdhtml.EscapeString(cfg.linkRel) + `"`
	}
	return `<a` + attrs + `>` + label + `</a>`
}

func stripBBCodeQuote(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\'`, `'`)
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return strings.TrimSpace(value[1 : len(value)-1])
		}
	}
	return value
}

func parseLists(input string) string {
	input = parseListTag(input, "list", "ul")
	input = parseListTag(input, "olist", "ol")
	return input
}

func parseListTag(input, bbTag, htmlTag string) string {
	re := regexp.MustCompile(`(?is)\[` + bbTag + `\](.*?)\[/` + bbTag + `\]`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		items := splitListItems(parts[1])
		if len(items) == 0 {
			return ""
		}
		var builder strings.Builder
		builder.WriteString("<")
		builder.WriteString(htmlTag)
		builder.WriteString(">")
		for _, item := range items {
			builder.WriteString("<li>")
			builder.WriteString(parseLists(item))
			builder.WriteString("</li>")
		}
		builder.WriteString("</")
		builder.WriteString(htmlTag)
		builder.WriteString(">")
		return builder.String()
	})
}

func splitListItems(content string) []string {
	content = regexp.MustCompile(`(?i)\[/\*\]`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`(?i)\[\*\]`).ReplaceAllString(content, "\x00")
	parts := strings.Split(content, "\x00")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

func normalizeNewlines(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	return strings.ReplaceAll(input, "\r", "\n")
}

func unescapeSteamText(input string) string {
	input = strings.ReplaceAll(input, `\[`, `[`)
	input = strings.ReplaceAll(input, `\]`, `]`)
	input = strings.ReplaceAll(input, `\"`, `"`)
	return strings.ReplaceAll(input, `\'`, `'`)
}
