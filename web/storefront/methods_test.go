package storefront

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	itraffic "github.com/gofurry/steam-go/internal/traffic"
)

func TestGetAppDetailsBuildsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	official := &recordingTransport{statuses: []int{http.StatusOK}}
	store := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"550":{"success":true,"data":{"type":"game","name":"Left 4 Dead 2","steam_appid":550,"required_age":"0","is_free":false,"controller_support":"full","detailed_description":"details","about_the_game":"about","short_description":"coop","supported_languages":"English","header_image":"header.jpg","capsule_image":"capsule.jpg","capsule_imagev5":"capsulev5.jpg","website":"https://example.test","pc_requirements":{"minimum":"min","recommended":"rec"},"mac_requirements":[],"developers":["Valve"],"publishers":["Valve"],"price_overview":{"currency":"USD","initial":999,"final":499,"discount_percent":50,"initial_formatted":"$9.99","final_formatted":"$4.99"},"platforms":{"windows":true,"mac":true,"linux":true},"metacritic":{"score":89,"url":"https://metacritic.test"},"categories":[{"id":1,"description":"Multi-player"}],"genres":[{"id":"1","description":"Action"}],"screenshots":[{"id":1,"path_thumbnail":"thumb.jpg","path_full":"full.jpg"}],"movies":[{"id":2,"name":"Trailer","thumbnail":"movie.jpg","webm":{"480":"480.webm","max":"max.webm"},"mp4":{"480":"480.mp4","max":"max.mp4"},"dash_av1":"av1.mpd","dash_h264":"h264.mpd","hls_h264":"h264.m3u8","highlight":true}],"recommendations":{"total":10},"achievements":{"total":1,"highlighted":[{"icon":"a.jpg","localized_name":"Done","name":"DONE","path":"https://example.test/a.jpg"}]},"packages":[469],"package_groups":[{"name":"default"}],"release_date":{"coming_soon":false,"date":"16 Nov, 2009"},"support_info":{"url":"https://support.test","email":""},"background":"bg.jpg","background_raw":"bg_raw.jpg","content_descriptors":{"ids":[2],"notes":"notes"},"ratings":{"steam_germany":{"required_age":"18"},"usk":{"rating":"18"}}}}}`,
	}
	service := newTestService(t, official, store, 16*1024)

	resp, err := service.GetAppDetails(context.Background(), 550, &GetAppDetailsOptions{
		CountryCode: "US",
		Language:    "english",
		Filters:     []string{"name", "price_overview"},
	})
	if err != nil {
		t.Fatalf("GetAppDetails returned error: %v", err)
	}
	if got := resp["550"].Data.Name; got != "Left 4 Dead 2" {
		t.Fatalf("unexpected name: %q", got)
	}
	if got := resp["550"].Data.CapsuleImage; got != "capsule.jpg" {
		t.Fatalf("unexpected capsule image: %q", got)
	}
	if got := resp["550"].Data.RequiredAge.Int(); got != 0 {
		t.Fatalf("unexpected required age: %d", got)
	}
	if got := resp["550"].Data.Screenshots[0].PathFull; got != "full.jpg" {
		t.Fatalf("unexpected screenshot: %q", got)
	}
	if got := resp["550"].Data.Movies[0].DASHH264; got != "h264.mpd" {
		t.Fatalf("unexpected movie dash URL: %q", got)
	}
	if got := resp["550"].Data.Achievements.Highlighted[0].Path; got != "https://example.test/a.jpg" {
		t.Fatalf("unexpected achievement path: %q", got)
	}
	requiredAge, ok, err := resp["550"].Data.SteamGermanyRequiredAge()
	if err != nil {
		t.Fatalf("SteamGermanyRequiredAge returned error: %v", err)
	}
	if !ok || requiredAge != "18" {
		t.Fatalf("unexpected steam germany required age: %q ok=%v", requiredAge, ok)
	}

	official.mu.Lock()
	officialCount := len(official.requests)
	official.mu.Unlock()
	if officialCount != 0 {
		t.Fatalf("expected official transport to stay unused, got %d", officialCount)
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.requests) != 1 {
		t.Fatalf("expected 1 store request, got %d", len(store.requests))
	}
	req := store.requests[0]
	if req.path != "/api/appdetails" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	if got := req.query.Get("appids"); got != "550" {
		t.Fatalf("unexpected appids: %q", got)
	}
	if got := req.query.Get("cc"); got != "US" {
		t.Fatalf("unexpected country code: %q", got)
	}
	if got := req.query.Get("l"); got != "english" {
		t.Fatalf("unexpected language: %q", got)
	}
	if got := req.query.Get("filters"); got != "name,price_overview" {
		t.Fatalf("unexpected filters: %q", got)
	}
}

func TestGetPackageDetailsReturnsRawResponse(t *testing.T) {
	t.Parallel()

	store := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"469":{"success":true,"data":{"packageid":469,"name":"Valve Complete Pack","header_image":"header.jpg","apps":[{"id":550,"name":"Left 4 Dead 2"}],"price":{"currency":"USD","initial":9999,"final":4999,"discount_percent":50,"individual":12999},"platforms":{"windows":true,"mac":true,"linux":true}}}}`,
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, store, 1024)

	raw, err := service.GetPackageDetailsRaw(context.Background(), 469, &GetPackageDetailsOptions{
		CountryCode: "US",
		Language:    "english",
	})
	if err != nil {
		t.Fatalf("GetPackageDetailsRaw returned error: %v", err)
	}
	if !strings.Contains(string(raw), `"469"`) {
		t.Fatalf("unexpected raw body: %s", string(raw))
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	req := store.requests[0]
	if req.path != "/api/packagedetails" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	if got := req.query.Get("packageids"); got != "469" {
		t.Fatalf("unexpected packageids: %q", got)
	}
}

func TestGetAppReviewsAppliesDefaultsAndTrafficClass(t *testing.T) {
	t.Parallel()

	official := &recordingTransport{statuses: []int{http.StatusOK}}
	store := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"success":1,"query_summary":{"num_reviews":1,"review_score":8,"review_score_desc":"Very Positive","total_positive":10,"total_negative":2,"total_reviews":12},"cursor":"AoIIPw%3D%3D","reviews":[{"recommendationid":"1","author":{"steamid":"2","playtime_forever":120},"review":"good","timestamp_created":1700000000,"timestamp_updated":1700000100,"voted_up":true,"votes_up":1,"votes_funny":0,"weighted_vote_score":0.5,"steam_purchase":true,"received_for_free":false,"written_during_early_access":false,"primarily_steam_deck":false}]}`,
	}
	service := newTestService(t, official, store, 1024)

	resp, err := service.GetAppReviews(context.Background(), 550, nil)
	if err != nil {
		t.Fatalf("GetAppReviews returned error: %v", err)
	}
	if got := resp.QuerySummary.TotalReviews; got != 12 {
		t.Fatalf("unexpected total reviews: %d", got)
	}
	if got := resp.Reviews[0].WeightedVoteScore.Float64(); got != 0.5 {
		t.Fatalf("unexpected weighted vote score: %v", got)
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	req := store.requests[0]
	if req.path != "/appreviews/550" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	assertQuery(t, req.query, "json", "1")
	assertQuery(t, req.query, "filter", "recent")
	assertQuery(t, req.query, "language", "all")
	assertQuery(t, req.query, "cursor", "*")
	assertQuery(t, req.query, "review_type", "all")
	assertQuery(t, req.query, "purchase_type", "all")
	assertQuery(t, req.query, "num_per_page", "100")
	assertQuery(t, req.query, "day_range", "365")
}

func TestStorefrontValidationAndErrors(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 16)

	if _, err := service.GetAppDetails(context.Background(), 0, nil); err == nil {
		t.Fatal("expected validation error for app id")
	}
	if _, err := service.GetAppReviews(context.Background(), 550, &GetAppReviewsOptions{NumPerPage: 101}); err == nil {
		t.Fatal("expected validation error for num per page")
	}

	httpErrService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusInternalServerError}}, 1024)
	_, err := httpErrService.GetAppReviews(context.Background(), 550, nil)
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindHTTPStatus {
		t.Fatalf("expected http status error, got %v", err)
	}

	decodeErrService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{`,
	}, 1024)
	_, err = decodeErrService.GetAppReviews(context.Background(), 550, nil)
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindDecode {
		t.Fatalf("expected decode error, got %v", err)
	}

	bodyLimitService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: strings.Repeat("a", 32),
	}, 8)
	_, err = bodyLimitService.GetPackageDetailsRaw(context.Background(), 469, nil)
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindTransport {
		t.Fatalf("expected body limit transport error, got %v", err)
	}
}

func TestFlexibleFloat64SupportsStringAndNumberPayloads(t *testing.T) {
	t.Parallel()

	var numeric FlexibleFloat64
	if err := numeric.UnmarshalJSON([]byte("0.5")); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if got := numeric.Float64(); got != 0.5 {
		t.Fatalf("unexpected numeric value: %v", got)
	}

	var quoted FlexibleFloat64
	if err := quoted.UnmarshalJSON([]byte(`"0.75"`)); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if got := quoted.Float64(); got != 0.75 {
		t.Fatalf("unexpected quoted value: %v", got)
	}
}

func TestFlexibleIntSupportsStringAndNumberPayloads(t *testing.T) {
	t.Parallel()

	var numeric FlexibleInt
	if err := numeric.UnmarshalJSON([]byte("18")); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if got := numeric.Int(); got != 18 {
		t.Fatalf("unexpected numeric value: %v", got)
	}

	var quoted FlexibleInt
	if err := quoted.UnmarshalJSON([]byte(`"16"`)); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if got := quoted.Int(); got != 16 {
		t.Fatalf("unexpected quoted value: %v", got)
	}
}

func newTestService(t *testing.T, officialTransport, storeTransport *recordingTransport, maxBodyBytes int64) *Service {
	t.Helper()

	executor, err := request.NewExecutor(
		"https://store.steampowered.com",
		nil,
		nil,
		maxBodyBytes,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    officialTransport,
		},
		map[itraffic.Class]request.ExecutionPolicy{
			itraffic.ClassPublicStorePage: {
				Retry:        0,
				RetryBackoff: request.DefaultRetryBackoffConfig(),
				Transport:    storeTransport,
			},
		},
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}
	return NewService(executor)
}

type recordingTransport struct {
	mu             sync.Mutex
	requests       []capturedRequest
	statuses       []int
	responseBody   string
	responseBodies []string
}

type capturedRequest struct {
	path  string
	query url.Values
}

func (t *recordingTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	clonedQuery := make(url.Values, len(req.URL.Query()))
	for key, values := range req.URL.Query() {
		copied := make([]string, len(values))
		copy(copied, values)
		clonedQuery[key] = copied
	}

	t.requests = append(t.requests, capturedRequest{
		path:  req.URL.Path,
		query: clonedQuery,
	})

	status := http.StatusOK
	if len(t.statuses) > 0 {
		status = t.statuses[0]
		t.statuses = t.statuses[1:]
	}
	body := t.responseBody
	if len(t.responseBodies) > 0 {
		body = t.responseBodies[0]
		t.responseBodies = t.responseBodies[1:]
	}
	if body == "" {
		body = "ok"
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func assertQuery(t *testing.T, query url.Values, key, want string) {
	t.Helper()
	if got := query.Get(key); got != want {
		t.Fatalf("unexpected %s: got %q want %q", key, got, want)
	}
}
