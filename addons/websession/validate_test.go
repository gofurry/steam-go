package websession

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

func TestValidateStoreSessionIgnoresLoginWordWithoutRedirect(t *testing.T) {
	t.Parallel()

	storeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/account/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("account login history"))
	}))
	defer storeServer.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	client := newTestClient(
		t,
		newTestAuthService(t, &sequenceAuthTransport{}),
		WithStoreBaseURL(storeServer.URL),
	)

	if err := client.ValidateStoreSession(context.Background(), jar); err != nil {
		t.Fatalf("ValidateStoreSession returned error: %v", err)
	}
}
