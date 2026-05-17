package freeclaim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchPromotionsBuildsRequestAndParsesResults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/results/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("specials") != "1" || query.Get("maxprice") != "free" || query.Get("infinite") != "1" || query.Get("force_infinite") != "1" {
			t.Fatalf("unexpected free search query: %s", r.URL.RawQuery)
		}
		if query.Get("count") != "50" || query.Get("os") != defaultSearchOS || query.Get("snr") != defaultSearchSNR {
			t.Fatalf("unexpected default query values: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"success":1,"results_html":"<a class=\"search_result_row\" href=\"/app/10/\" data-ds-appid=\"10\"><span class=\"title\">Demo Game</span><div class=\"discount_block\" data-discount=\"100\" data-price-final=\"0\"><div class=\"discount_pct\">-100%</div><div class=\"discount_original_price\">$9.99</div><div class=\"discount_final_price\">Free</div></div><img src=\"https://cdn.example/app10.jpg\"/></a><a class=\"search_result_row\" href=\"/app/11/\" data-ds-appid=\"11\"><span class=\"title\">Paid Game</span><div class=\"discount_block\" data-discount=\"50\" data-price-final=\"499\"><div class=\"discount_original_price\">$9.99</div><div class=\"discount_final_price\">$4.99</div></div></a>","total_count":2}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client())
	promotions, err := client.SearchPromotions(context.Background(), nil)
	if err != nil {
		t.Fatalf("SearchPromotions returned error: %v", err)
	}
	if len(promotions) != 1 {
		t.Fatalf("expected 1 promotion, got %d", len(promotions))
	}
	promotion := promotions[0]
	if promotion.AppID != 10 || promotion.Title != "Demo Game" {
		t.Fatalf("unexpected promotion: %#v", promotion)
	}
	if promotion.StoreURL != server.URL+"/app/10/" {
		t.Fatalf("unexpected store url: %s", promotion.StoreURL)
	}
	if promotion.FinalPrice != "Free" || promotion.OriginalPrice != "$9.99" || promotion.Discount != "-100%" {
		t.Fatalf("unexpected price fields: %#v", promotion)
	}
}
