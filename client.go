package steam

import (
	"fmt"
	"net/http"
	"time"

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
	itraffic "github.com/GoFurry/steam-go/internal/traffic"
	"github.com/GoFurry/steam-go/internal/transport"
)

// Client is the root Steam Web API entrypoint.
type Client struct {
	API *API

	httpClients []*http.Client
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

	runtimes, err := buildTrafficRuntimes(cfg)
	if err != nil {
		return nil, err
	}
	executor, err := request.NewExecutor(
		cfg.baseURL,
		cfg.apiKeyProvider,
		cfg.accessTokenProvider,
		cfg.maxResponseBodyBytes,
		runtimes.defaultPolicy,
		runtimes.classPolicies,
	)
	if err != nil {
		return nil, err
	}

	client := &Client{
		httpClients: runtimes.httpClients,
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
	if c == nil {
		return
	}
	for _, httpClient := range c.httpClients {
		if httpClient == nil {
			continue
		}
		httpClient.CloseIdleConnections()
	}
}

type trafficRuntimeSet struct {
	defaultPolicy request.ExecutionPolicy
	classPolicies map[itraffic.Class]request.ExecutionPolicy
	httpClients   []*http.Client
}

type runtimePolicyConfig struct {
	proxySelector   ProxySelector
	cookieJar       http.CookieJar
	rateLimiter     transport.RateLimiterConfig
	hostControl     transport.RequestControlConfig
	sessionControl  transport.RequestControlConfig
	cacheTTL        time.Duration
	blockPolicy     *TrafficBlockPolicy
	trafficClass    itraffic.Class
	headerProfile   *HeaderProfile
	refererSelector RefererSelector
	transportHook   TransportHook
	retry           int
	retryBackoff    request.RetryBackoffConfig
}

func buildTrafficRuntimes(cfg clientConfig) (trafficRuntimeSet, error) {
	defaultRuntime, err := buildRuntime(cfg, runtimePolicyConfig{
		proxySelector:   cfg.proxySelector,
		cookieJar:       cfg.cookieJar,
		rateLimiter:     cfg.rateLimiter,
		hostControl:     transport.RequestControlConfig{},
		sessionControl:  transport.RequestControlConfig{},
		cacheTTL:        0,
		blockPolicy:     nil,
		trafficClass:    itraffic.ClassOfficialAPI,
		headerProfile:   nil,
		refererSelector: nil,
		transportHook:   nil,
		retry:           cfg.retry,
		retryBackoff:    cfg.retryBackoff,
	}, cfg.cookieJarConfigured)
	if err != nil {
		return trafficRuntimeSet{}, err
	}

	runtimes := trafficRuntimeSet{
		defaultPolicy: defaultRuntime.executionPolicy,
		classPolicies: make(map[itraffic.Class]request.ExecutionPolicy, len(cfg.trafficPolicies)),
		httpClients:   []*http.Client{defaultRuntime.httpClient},
	}

	for class, policy := range cfg.trafficPolicies {
		resolved := runtimePolicyConfig{
			proxySelector:   cfg.proxySelector,
			cookieJar:       cfg.cookieJar,
			rateLimiter:     cfg.rateLimiter,
			hostControl:     transport.RequestControlConfig{},
			sessionControl:  transport.RequestControlConfig{},
			cacheTTL:        0,
			blockPolicy:     nil,
			trafficClass:    itraffic.NormalizeClass(class),
			headerProfile:   nil,
			refererSelector: nil,
			transportHook:   nil,
			retry:           cfg.retry,
			retryBackoff:    cfg.retryBackoff,
		}
		cookieJarConfigured := cfg.cookieJarConfigured
		if policy.proxySelector != nil {
			resolved.proxySelector = policy.proxySelector
		}
		if policy.cookieJarProvided {
			resolved.cookieJar = policy.cookieJar
			cookieJarConfigured = true
		}
		if policy.rateLimiter != nil {
			resolved.rateLimiter = transport.RateLimiterConfig{
				Limit: policy.rateLimiter.Limit,
				Burst: policy.rateLimiter.Burst,
			}
		}
		if policy.retry != nil {
			resolved.retry = policy.retry.Retry
			resolved.retryBackoff = request.RetryBackoffConfig(policy.retry.Backoff)
		}
		if policy.hostControl != nil {
			resolved.hostControl = transport.RequestControlConfig{
				MaxConcurrent: policy.hostControl.MaxConcurrent,
			}
			if policy.hostControl.RateLimiter != nil {
				resolved.hostControl.RateLimiter = transport.RateLimiterConfig{
					Limit: policy.hostControl.RateLimiter.Limit,
					Burst: policy.hostControl.RateLimiter.Burst,
				}
			}
		}
		if policy.sessionControl != nil {
			resolved.sessionControl = transport.RequestControlConfig{
				MaxConcurrent: policy.sessionControl.MaxConcurrent,
			}
			if policy.sessionControl.RateLimiter != nil {
				resolved.sessionControl.RateLimiter = transport.RateLimiterConfig{
					Limit: policy.sessionControl.RateLimiter.Limit,
					Burst: policy.sessionControl.RateLimiter.Burst,
				}
			}
		}
		if policy.cache != nil {
			resolved.cacheTTL = policy.cache.TTL
		}
		if policy.blockPolicy != nil {
			resolved.blockPolicy = policy.blockPolicy
		}
		if policy.headerProfile != nil {
			resolved.headerProfile = cloneHeaderProfile(policy.headerProfile)
		}
		if policy.refererSelector != nil {
			resolved.refererSelector = policy.refererSelector
		}
		if policy.transportHook != nil {
			resolved.transportHook = policy.transportHook
		}

		runtime, err := buildRuntime(cfg, resolved, cookieJarConfigured)
		if err != nil {
			return trafficRuntimeSet{}, err
		}
		class = itraffic.NormalizeClass(class)
		runtimes.classPolicies[class] = runtime.executionPolicy
		runtimes.httpClients = append(runtimes.httpClients, runtime.httpClient)
	}

	return runtimes, nil
}

type builtRuntime struct {
	httpClient      *http.Client
	executionPolicy request.ExecutionPolicy
}

func buildRuntime(cfg clientConfig, policy runtimePolicyConfig, cookieJarConfigured bool) (builtRuntime, error) {
	httpClient, err := buildHTTPClient(cfg, policy.cookieJar, cookieJarConfigured)
	if err != nil {
		return builtRuntime{}, err
	}

	if proxyAwareHook, ok := policy.transportHook.(ProxyAwareTransportHook); ok {
		hookedClient, err := proxyAwareHook.WrapHTTPClientWithProxy(policy.trafficClass, cloneHTTPClient(httpClient), policy.proxySelector)
		if err != nil {
			return builtRuntime{}, err
		}
		if hookedClient == nil {
			return builtRuntime{}, fmt.Errorf("transport hook returned a nil http client")
		}
		httpClient = hookedClient
	} else {
		httpClient, err = WrapHTTPClientWithProxySelector(httpClient, policy.proxySelector)
		if err != nil {
			return builtRuntime{}, err
		}
		if policy.transportHook != nil {
			hookedClient, err := policy.transportHook.WrapHTTPClient(policy.trafficClass, cloneHTTPClient(httpClient))
			if err != nil {
				return builtRuntime{}, err
			}
			if hookedClient == nil {
				return builtRuntime{}, fmt.Errorf("transport hook returned a nil http client")
			}
			httpClient = hookedClient
		}
	}
	return builtRuntime{
		httpClient: httpClient,
		executionPolicy: request.ExecutionPolicy{
			Retry:          policy.retry,
			RetryBackoff:   policy.retryBackoff,
			CacheRuntime:   request.NewMemoryCacheRuntime(policy.cacheTTL, policy.cookieJar),
			BlockRuntime:   request.NewBlockRuntime(policy.trafficClass, request.BlockConfig{HTMLSniffBytes: blockSniffBytes(policy.blockPolicy)}),
			PrepareRequest: buildRequestPreparer(policy.headerProfile, policy.refererSelector),
			Transport: transport.New(httpClient, transport.ClientConfig{
				RateLimiter:    policy.rateLimiter,
				HostControl:    policy.hostControl,
				SessionControl: policy.sessionControl,
			}),
		},
	}, nil
}

func blockSniffBytes(policy *TrafficBlockPolicy) int {
	if policy == nil {
		return 0
	}
	return policy.HTMLSniffBytes
}

func buildHTTPClient(cfg clientConfig, jar http.CookieJar, cookieJarConfigured bool) (*http.Client, error) {
	if cfg.httpClient != nil {
		cloned := cloneHTTPClient(cfg.httpClient)
		cloned.Timeout = cfg.timeout
		if cookieJarConfigured {
			cloned.Jar = jar
		}
		return cloned, nil
	}

	return &http.Client{
		Timeout:   cfg.timeout,
		Transport: cloneDefaultTransport(),
		Jar:       jar,
	}, nil
}

func cloneHTTPClient(base *http.Client) *http.Client {
	if base == nil {
		return nil
	}

	cloned := *base
	if transport, ok := cloned.Transport.(*http.Transport); ok && transport != nil {
		cloned.Transport = transport.Clone()
	}
	return &cloned
}

func cloneDefaultTransport() http.RoundTripper {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok && transport != nil {
		return transport.Clone()
	}
	return http.DefaultTransport
}
