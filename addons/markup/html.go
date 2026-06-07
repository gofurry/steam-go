package markup

import (
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	xhtml "golang.org/x/net/html"
)

// CleanSteamContent converts common Steam BBCode into HTML and sanitizes the result.
func CleanSteamContent(input string, opts ...Option) (string, error) {
	cfg := resolveConfig(opts)
	html := convertSteamBBCode(input, cfg)
	if !cfg.sanitize {
		return html, nil
	}
	return sanitizeHTML(html, cfg), nil
}

// CleanHTML sanitizes an HTML fragment using the Steam markup policy.
func CleanHTML(input string, opts ...Option) (string, error) {
	cfg := resolveConfig(opts)
	if !cfg.sanitize {
		return input, nil
	}
	return sanitizeHTML(input, cfg), nil
}

// PlainText converts Steam markup or HTML content into plain text.
func PlainText(input string, opts ...Option) (string, error) {
	cleaned, err := CleanSteamContent(input, opts...)
	if err != nil {
		return "", err
	}
	return htmlToText(cleaned), nil
}

// Summary converts Steam markup or HTML content into a plain text summary.
func Summary(input string, maxRunes int, opts ...Option) (string, error) {
	text, err := PlainText(input, opts...)
	if err != nil {
		return "", err
	}
	text = strings.Join(strings.Fields(text), " ")
	if maxRunes <= 0 || utf8.RuneCountInString(text) <= maxRunes {
		return text, nil
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:maxRunes])), nil
}

func sanitizeHTML(input string, cfg config) string {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("h1", "h2", "h3", "video", "source")
	policy.AllowAttrs("controls", "poster", "src").OnElements("video")
	policy.AllowAttrs("src", "type").OnElements("source")
	policy.AllowAttrs("target", "rel").OnElements("a")
	if cfg.allowYouTubeEmbed {
		policy.AllowElements("iframe")
		policy.AllowAttrs("src", "frameborder", "allow", "allowfullscreen").OnElements("iframe")
	}
	return policy.Sanitize(input)
}

func htmlToText(input string) string {
	root, err := xhtml.Parse(strings.NewReader("<div>" + input + "</div>"))
	if err != nil {
		return strings.Join(strings.Fields(input), " ")
	}
	var builder strings.Builder
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node.Type == xhtml.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" {
				if builder.Len() > 0 {
					builder.WriteByte(' ')
				}
				builder.WriteString(text)
			}
		}
		if node.Type == xhtml.ElementNode {
			switch node.Data {
			case "br", "p", "li", "h1", "h2", "h3", "blockquote":
				if builder.Len() > 0 {
					builder.WriteByte(' ')
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return strings.Join(strings.Fields(builder.String()), " ")
}
