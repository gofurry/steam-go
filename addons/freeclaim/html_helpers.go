package freeclaim

import (
	"strings"

	"golang.org/x/net/html"
)

func visitNodes(node *html.Node, visit func(*html.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		visitNodes(child, visit)
	}
}

func firstNode(node *html.Node, predicate func(*html.Node) bool) *html.Node {
	var match *html.Node
	visitNodes(node, func(current *html.Node) {
		if match != nil || !predicate(current) {
			return
		}
		match = current
	})
	return match
}

func nodeAttr(node *html.Node, key string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func hasClass(node *html.Node, className string) bool {
	classes := strings.Fields(nodeAttr(node, "class"))
	for _, class := range classes {
		if class == className {
			return true
		}
	}
	return false
}

func textContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	var builder strings.Builder
	visitNodes(node, func(current *html.Node) {
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
		}
	})
	return strings.TrimSpace(builder.String())
}
