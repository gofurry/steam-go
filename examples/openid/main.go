package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	steam "github.com/GoFurry/steam-go"
	"github.com/GoFurry/steam-go/addons/openid"
)

const stateCookieName = "steam_openid_state"

func main() {
	var (
		proxyRaw = flag.String("proxy", "", "optional HTTP proxy URL for OpenID verification requests, e.g. http://127.0.0.1:7897")
		listen   = flag.String("listen", ":8080", "listen address")
		baseURL  = flag.String("base-url", "http://localhost:8080", "public base URL used for realm and callback")
		timeout  = flag.Duration("timeout", 20*time.Second, "verification request timeout")
	)
	flag.Parse()

	base, err := url.Parse(*baseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		log.Fatal("invalid --base-url")
	}

	selector, err := steam.NewStaticProxySelector(*proxyRaw)
	if err != nil {
		log.Fatal(err)
	}

	httpClient, err := steam.NewHTTPClientWithProxySelector(selector, *timeout)
	if err != nil {
		log.Fatal(err)
	}

	verifier, err := openid.NewVerifier(openid.Config{
		Realm:    *baseURL,
		ReturnTo: strings.TrimRight(*baseURL, "/") + "/callback",
	},
		openid.WithHTTPClient(httpClient),
		openid.WithTimeout(*timeout),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(
			w,
			"Steam OpenID example\n\nVisit %s/login to start.\nThis example stores state in a cookie and verifies it on callback.\n",
			strings.TrimRight(*baseURL, "/"),
		)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state, err := randomState()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     stateCookieName,
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   strings.EqualFold(base.Scheme, "https"),
			MaxAge:   300,
		})
		loginURL, err := verifier.LoginURL(state)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, loginURL, http.StatusFound)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		identity, err := verifier.VerifyRequest(context.Background(), r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie(stateCookieName)
		if err != nil {
			http.Error(w, "missing state cookie", http.StatusBadRequest)
			return
		}
		if cookie.Value == "" || cookie.Value != identity.State {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     stateCookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   strings.EqualFold(base.Scheme, "https"),
			MaxAge:   -1,
		})

		_, _ = fmt.Fprintf(
			w,
			"Steam login verified.\nSteamID: %s\nState: %s\nClaimedID: %s\n\nState cookie matched and was cleared.\n",
			identity.SteamID,
			identity.State,
			identity.ClaimedID,
		)
	})

	if *proxyRaw == "" {
		log.Printf("Steam OpenID example running in direct mode on %s", *listen)
	} else {
		log.Printf("Steam OpenID example running with proxy=%s on %s", *proxyRaw, *listen)
	}
	log.Printf("Open %s/login to start the Steam OpenID flow", strings.TrimRight(*baseURL, "/"))
	log.Fatal(http.ListenAndServe(*listen, mux))
}

func randomState() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}
