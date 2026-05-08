package newsservice

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/GoFurry/steam-go/internal/endpoint"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// ConvertHTMLToBBCodeOptions controls optional query parameters for ConvertHTMLToBBCode.
type ConvertHTMLToBBCodeOptions struct {
	PreserveNewlines *bool
}

// ConvertHTMLToBBCode converts supported HTML fragments into Steam-flavored BBCode.
func (s *Service) ConvertHTMLToBBCode(ctx context.Context, content string, opts *ConvertHTMLToBBCodeOptions) (ConvertHTMLToBBCodeResponse, error) {
	body, err := s.ConvertHTMLToBBCodeRaw(ctx, content, opts)
	if err != nil {
		return ConvertHTMLToBBCodeResponse{}, err
	}
	return response.DecodeJSON[ConvertHTMLToBBCodeResponse](body)
}

// ConvertHTMLToBBCodeRaw returns the raw JSON response body.
func (s *Service) ConvertHTMLToBBCodeRaw(ctx context.Context, content string, opts *ConvertHTMLToBBCodeOptions) ([]byte, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "content is required", nil, nil)
	}

	query := url.Values{}
	query.Set("content", trimmed)
	if opts != nil && opts.PreserveNewlines != nil {
		query.Set("preserve_newlines", boolString(*opts.PreserveNewlines))
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.NewsServiceConvertHTMLToBBCode,
		Query:  query,
	})
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
