package steam

import (
	"net/http"

	"github.com/GoFurry/steam-go/api/playerservice"
	"github.com/GoFurry/steam-go/api/steamnews"
	"github.com/GoFurry/steam-go/api/steamuser"
	"github.com/GoFurry/steam-go/api/steamuserstats"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/transport"
)

// Client is the root Steam Web API entrypoint.
type Client struct {
	SteamUser      *steamuser.Service
	PlayerService  *playerservice.Service
	SteamNews      *steamnews.Service
	SteamUserStats *steamuserstats.Service

	httpClient *http.Client
}

// NewClient builds a Steam Web API client using functional options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultClientConfig()
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	httpClient := buildHTTPClient(cfg)
	rt := transport.New(httpClient, cfg.rateLimit, cfg.logger)
	executor, err := request.NewExecutor(cfg.baseURL, cfg.apiKeyProvider, cfg.accessTokenProvider, cfg.retry, rt, cfg.logger)
	if err != nil {
		return nil, err
	}

	client := &Client{
		httpClient: httpClient,
	}
	client.SteamUser = steamuser.NewService(executor)
	client.PlayerService = playerservice.NewService(executor)
	client.SteamNews = steamnews.NewService(executor)
	client.SteamUserStats = steamuserstats.NewService(executor)
	return client, nil
}

// Close releases idle HTTP connections held by the SDK client.
func (c *Client) Close() {
	if c == nil || c.httpClient == nil {
		return
	}
	c.httpClient.CloseIdleConnections()
}

func buildHTTPClient(cfg clientConfig) *http.Client {
	if cfg.httpClient != nil {
		cloned := *cfg.httpClient
		cloned.Timeout = cfg.timeout
		cloned.Transport = transport.WrapRoundTripper(cloned.Transport, cfg.proxySelector)
		return &cloned
	}

	return &http.Client{
		Timeout:   cfg.timeout,
		Transport: transport.WrapRoundTripper(nil, cfg.proxySelector),
	}
}
