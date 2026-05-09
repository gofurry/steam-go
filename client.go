package steam

import (
	"net/http"

	"github.com/GoFurry/steam-go/api/accountcartservice"
	"github.com/GoFurry/steam-go/api/billingservice"
	"github.com/GoFurry/steam-go/api/communityservice"
	"github.com/GoFurry/steam-go/api/familygroupsservice"
	"github.com/GoFurry/steam-go/api/loyaltyrewardsservice"
	"github.com/GoFurry/steam-go/api/mobilenotificationservice"
	"github.com/GoFurry/steam-go/api/newsservice"
	"github.com/GoFurry/steam-go/api/playerservice"
	"github.com/GoFurry/steam-go/api/questservice"
	"github.com/GoFurry/steam-go/api/salefeatureservice"
	"github.com/GoFurry/steam-go/api/steamapps"
	"github.com/GoFurry/steam-go/api/steamchartsservice"
	"github.com/GoFurry/steam-go/api/steamdirectory"
	"github.com/GoFurry/steam-go/api/steamnews"
	"github.com/GoFurry/steam-go/api/steamnotificationservice"
	"github.com/GoFurry/steam-go/api/steamuser"
	"github.com/GoFurry/steam-go/api/steamuseroauth"
	"github.com/GoFurry/steam-go/api/steamuserstats"
	"github.com/GoFurry/steam-go/api/steamwebapiutil"
	"github.com/GoFurry/steam-go/api/storebrowseservice"
	"github.com/GoFurry/steam-go/api/storecatalogservice"
	"github.com/GoFurry/steam-go/api/storepreferencesservice"
	"github.com/GoFurry/steam-go/api/storeservice"
	"github.com/GoFurry/steam-go/api/storetopsellersservice"
	"github.com/GoFurry/steam-go/api/useraccountservice"
	"github.com/GoFurry/steam-go/api/userreviewsservice"
	"github.com/GoFurry/steam-go/api/userstorevisitservice"
	"github.com/GoFurry/steam-go/api/wishlistservice"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/transport"
)

// Client is the root Steam Web API entrypoint.
type Client struct {
	API *API

	httpClient *http.Client
}

// API groups all typed Steam Web API services under one stable entrypoint.
type API struct {
	AccountCartService        *accountcartservice.Service
	BillingService            *billingservice.Service
	CommunityService          *communityservice.Service
	FamilyGroupsService       *familygroupsservice.Service
	LoyaltyRewardsService     *loyaltyrewardsservice.Service
	MobileNotificationService *mobilenotificationservice.Service
	NewsService               *newsservice.Service
	QuestService              *questservice.Service
	SaleFeatureService        *salefeatureservice.Service
	StoreBrowseService        *storebrowseservice.Service
	StoreCatalogService       *storecatalogservice.Service
	StorePreferencesService   *storepreferencesservice.Service
	StoreService              *storeservice.Service
	StoreTopSellersService    *storetopsellersservice.Service
	SteamDirectory            *steamdirectory.Service
	SteamApps                 *steamapps.Service
	SteamChartsService        *steamchartsservice.Service
	SteamNotificationService  *steamnotificationservice.Service
	SteamUser                 *steamuser.Service
	SteamUserOAuth            *steamuseroauth.Service
	SteamWebAPIUtil           *steamwebapiutil.Service
	UserAccountService        *useraccountservice.Service
	UserReviewsService        *userreviewsservice.Service
	UserStoreVisitService     *userstorevisitservice.Service
	WishlistService           *wishlistservice.Service
	PlayerService             *playerservice.Service
	SteamNews                 *steamnews.Service
	SteamUserStats            *steamuserstats.Service
}

// NewClient builds a Steam Web API client using functional options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultClientConfig()
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	httpClient, err := buildHTTPClient(cfg)
	if err != nil {
		return nil, err
	}
	rt := transport.New(httpClient, cfg.rateLimit)
	executor, err := request.NewExecutor(
		cfg.baseURL,
		cfg.apiKeyProvider,
		cfg.accessTokenProvider,
		cfg.retry,
		cfg.maxResponseBodyBytes,
		rt,
	)
	if err != nil {
		return nil, err
	}

	client := &Client{
		httpClient: httpClient,
	}
	client.API = &API{
		AccountCartService:        accountcartservice.NewService(executor),
		BillingService:            billingservice.NewService(executor),
		CommunityService:          communityservice.NewService(executor),
		FamilyGroupsService:       familygroupsservice.NewService(executor),
		LoyaltyRewardsService:     loyaltyrewardsservice.NewService(executor),
		MobileNotificationService: mobilenotificationservice.NewService(executor),
		NewsService:               newsservice.NewService(executor),
		QuestService:              questservice.NewService(executor),
		SaleFeatureService:        salefeatureservice.NewService(executor),
		StoreBrowseService:        storebrowseservice.NewService(executor),
		StoreCatalogService:       storecatalogservice.NewService(executor),
		StorePreferencesService:   storepreferencesservice.NewService(executor),
		StoreService:              storeservice.NewService(executor),
		StoreTopSellersService:    storetopsellersservice.NewService(executor),
		SteamDirectory:            steamdirectory.NewService(executor),
		SteamApps:                 steamapps.NewService(executor),
		SteamChartsService:        steamchartsservice.NewService(executor),
		SteamNotificationService:  steamnotificationservice.NewService(executor),
		SteamUser:                 steamuser.NewService(executor),
		SteamUserOAuth:            steamuseroauth.NewService(executor),
		SteamWebAPIUtil:           steamwebapiutil.NewService(executor),
		UserAccountService:        useraccountservice.NewService(executor),
		UserReviewsService:        userreviewsservice.NewService(executor),
		UserStoreVisitService:     userstorevisitservice.NewService(executor),
		WishlistService:           wishlistservice.NewService(executor),
		PlayerService:             playerservice.NewService(executor),
		SteamNews:                 steamnews.NewService(executor),
		SteamUserStats:            steamuserstats.NewService(executor),
	}
	return client, nil
}

// Close releases idle HTTP connections held by the SDK client.
func (c *Client) Close() {
	if c == nil || c.httpClient == nil {
		return
	}
	c.httpClient.CloseIdleConnections()
}

func buildHTTPClient(cfg clientConfig) (*http.Client, error) {
	if cfg.httpClient != nil {
		cloned := *cfg.httpClient
		cloned.Timeout = cfg.timeout
		rt, err := transport.WrapRoundTripper(cloned.Transport, cfg.proxySelector)
		if err != nil {
			return nil, err
		}
		cloned.Transport = rt
		return &cloned, nil
	}

	rt, err := transport.WrapRoundTripper(nil, cfg.proxySelector)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Timeout:   cfg.timeout,
		Transport: rt,
	}, nil
}
