package vdf

import (
	"io"

	upstream "github.com/gofurry/vdf-go"
)

type Document = upstream.Document
type Node = upstream.Node

type Option = upstream.Option
type Config = upstream.Config
type ParseError = upstream.ParseError

type EncodeOption = upstream.EncodeOption
type EncodeConfig = upstream.EncodeConfig

func NewDocument(nodes ...*Node) *Document {
	return upstream.NewDocument(nodes...)
}

func NewNode(key string, children ...*Node) *Node {
	return upstream.NewNode(key, children...)
}

func NewValue(key, value string) *Node {
	return upstream.NewValue(key, value)
}

func DefaultConfig() Config {
	return upstream.DefaultConfig()
}

func WithEscapeSequences(enabled bool) Option {
	return upstream.WithEscapeSequences(enabled)
}

func WithMaxDepth(max int) Option {
	return upstream.WithMaxDepth(max)
}

func WithMaxTokenBytes(max int) Option {
	return upstream.WithMaxTokenBytes(max)
}

func WithMaxNodes(max int) Option {
	return upstream.WithMaxNodes(max)
}

func WithBareTokens(enabled bool) Option {
	return upstream.WithBareTokens(enabled)
}

func WithComments(enabled bool) Option {
	return upstream.WithComments(enabled)
}

func WithPreserveDirectives(enabled bool) Option {
	return upstream.WithPreserveDirectives(enabled)
}

func Parse(data []byte, opts ...Option) (*Document, error) {
	return upstream.Parse(data, opts...)
}

func ParseString(s string, opts ...Option) (*Document, error) {
	return upstream.ParseString(s, opts...)
}

func ParseReader(r io.Reader, opts ...Option) (*Document, error) {
	return upstream.ParseReader(r, opts...)
}

func ParseReaderLimit(r io.Reader, maxBytes int64, opts ...Option) (*Document, error) {
	return upstream.ParseReaderLimit(r, maxBytes, opts...)
}

func ParseFile(path string, opts ...Option) (*Document, error) {
	return upstream.ParseFile(path, opts...)
}

func DefaultEncodeConfig() EncodeConfig {
	return upstream.DefaultEncodeConfig()
}

func WithIndent(indent string) EncodeOption {
	return upstream.WithIndent(indent)
}

func WithQuoteKeys(enabled bool) EncodeOption {
	return upstream.WithQuoteKeys(enabled)
}

func WithQuoteValues(enabled bool) EncodeOption {
	return upstream.WithQuoteValues(enabled)
}

func WithSortKeys(enabled bool) EncodeOption {
	return upstream.WithSortKeys(enabled)
}

func Marshal(doc *Document, opts ...EncodeOption) ([]byte, error) {
	return upstream.Marshal(doc, opts...)
}

func MarshalString(doc *Document, opts ...EncodeOption) (string, error) {
	return upstream.MarshalString(doc, opts...)
}

func Write(w io.Writer, doc *Document, opts ...EncodeOption) error {
	return upstream.Write(w, doc, opts...)
}
