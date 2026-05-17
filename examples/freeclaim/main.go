package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/addons/freeclaim"
	"github.com/gofurry/steam-go/addons/websession"
	"github.com/gofurry/steam-go/examples/internal/secretinput"
)

func main() {
	var (
		proxyRaw    = flag.String("proxy", "", "optional HTTP proxy URL for Steam requests, e.g. http://127.0.0.1:7897")
		timeout     = flag.Duration("timeout", 15*time.Second, "per-request timeout")
		searchQuery = flag.String("search-query", "", "optional search query passed to the Store promotion search")
		searchCount = flag.Int("search-count", 10, "maximum number of promotions to print")
		appID       = flag.Uint("app-id", 0, "optional app id for package resolution or claim")
		packageID   = flag.Uint("package-id", 0, "optional package id for claim; if omitted, one free package will be auto-selected only when exactly one match exists")
		countryCode = flag.String("country-code", "us", "country code used for Storefront app details")
		language    = flag.String("language", "english", "language used for Storefront app details")
		claim       = flag.Bool("claim", false, "actually send one addfreelicense request; default behavior is read-only search and package resolution")
	)
	flag.Parse()

	if *searchCount < 1 {
		log.Fatal("-search-count must be greater than zero")
	}

	selector, err := steam.NewStaticProxySelector(*proxyRaw)
	if err != nil {
		log.Fatal(err)
	}

	sdk, err := steam.NewClient(
		steam.WithTimeout(*timeout),
		steam.WithProxySelector(selector),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer sdk.Close()

	freeclaimClient, err := freeclaim.NewClientFromSteamClient(
		sdk,
		freeclaim.WithTimeout(*timeout),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	promotions, err := freeclaimClient.SearchPromotions(ctx, &freeclaim.SearchPromotionsOptions{
		Query: *searchQuery,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("promotions_found=%d\n", len(promotions))
	for index, promotion := range promotions {
		if index >= *searchCount {
			break
		}
		fmt.Printf(
			"- appid=%d title=%q original=%q final=%q discount=%q url=%s\n",
			promotion.AppID,
			promotion.Title,
			promotion.OriginalPrice,
			promotion.FinalPrice,
			promotion.Discount,
			promotion.StoreURL,
		)
	}

	if *appID == 0 {
		if *claim {
			log.Fatal("-claim requires -app-id and a refresh token")
		}
		return
	}

	packages, err := freeclaimClient.ResolveFreePackages(ctx, uint32(*appID), &freeclaim.ResolveFreePackagesOptions{
		CountryCode: *countryCode,
		Language:    *language,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("free_packages_for_app_%d=%d\n", *appID, len(packages))
	for _, pkg := range packages {
		fmt.Printf("- packageid=%d title=%q option=%q\n", pkg.PackageID, pkg.Title, pkg.OptionText)
	}

	if !*claim {
		return
	}

	refreshToken, err := resolveRefreshToken(secretinput.DefaultResolver())
	if err != nil {
		log.Fatal(err)
	}

	selectedPackageID, err := resolveClaimPackageID(uint32(*packageID), packages)
	if err != nil {
		log.Fatal(err)
	}

	sessionClient, err := websession.NewClientFromSteamClient(
		sdk,
		websession.WithTimeout(*timeout),
	)
	if err != nil {
		log.Fatal(err)
	}
	cookies, err := sessionClient.RefreshTokenToWebCookies(ctx, refreshToken)
	if err != nil {
		log.Fatal(err)
	}
	if err := sessionClient.ValidateWebCookies(ctx, cookies); err != nil {
		log.Fatal(err)
	}

	result, err := freeclaimClient.ClaimPackage(ctx, cookies, freeclaim.ClaimPackageRequest{
		AppID:     uint32(*appID),
		PackageID: selectedPackageID,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(
		"claim_result status=%s appid=%d packageid=%d owned=%t message=%q\n",
		result.Status,
		result.AppID,
		result.PackageID,
		result.Owned,
		result.Message,
	)
}

func resolveClaimPackageID(explicit uint32, packages []freeclaim.FreePackage) (uint32, error) {
	if explicit != 0 {
		return explicit, nil
	}
	if len(packages) == 1 {
		return packages[0].PackageID, nil
	}
	if len(packages) == 0 {
		return 0, fmt.Errorf("no free package candidates were found for the selected app")
	}
	return 0, fmt.Errorf("multiple free package candidates found; rerun with -package-id")
}
