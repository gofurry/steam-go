package websession

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"time"
)

const webCookieSnapshotVersion = 1

// WebCookieSnapshot is an explicit JSON-serializable snapshot of web cookies.
//
// Values in this snapshot are credentials. Store the encoded data only in a
// caller-controlled location with appropriate local permissions.
type WebCookieSnapshot struct {
	Version   int                     `json:"version"`
	SteamID   string                  `json:"steamid,omitempty"`
	SessionID string                  `json:"sessionid,omitempty"`
	Domains   []string                `json:"domains,omitempty"`
	SavedAt   time.Time               `json:"saved_at,omitempty"`
	Cookies   []WebCookieSnapshotItem `json:"cookies,omitempty"`
}

// WebCookieSnapshotItem stores one cookie value for one domain.
type WebCookieSnapshotItem struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

// ExportWebCookieSnapshot converts a WebCookieResult into a JSON-serializable snapshot.
func ExportWebCookieSnapshot(result *WebCookieResult) (WebCookieSnapshot, error) {
	if result == nil {
		return WebCookieSnapshot{}, &Error{Code: ErrorCodeRequestBuild, Op: "export_web_cookie_snapshot", Message: "web cookie result must not be nil"}
	}
	if result.Jar == nil {
		return WebCookieSnapshot{}, &Error{Code: ErrorCodeRequestBuild, Op: "export_web_cookie_snapshot", Message: "cookie jar must not be nil"}
	}

	domains := normalizeSnapshotDomains(result.Domains)
	cookies := make([]WebCookieSnapshotItem, 0)
	for _, domain := range domains {
		cookieURL := snapshotDomainURL(domain)
		for _, cookie := range result.Jar.Cookies(cookieURL) {
			if cookie == nil || strings.TrimSpace(cookie.Name) == "" {
				continue
			}
			cookies = append(cookies, WebCookieSnapshotItem{
				Domain: domain,
				Name:   cookie.Name,
				Value:  cookie.Value,
			})
		}
	}
	sort.Slice(cookies, func(i, j int) bool {
		if cookies[i].Domain != cookies[j].Domain {
			return cookies[i].Domain < cookies[j].Domain
		}
		return cookies[i].Name < cookies[j].Name
	})

	return WebCookieSnapshot{
		Version:   webCookieSnapshotVersion,
		SteamID:   strings.TrimSpace(result.SteamID),
		SessionID: strings.TrimSpace(result.SessionID),
		Domains:   domains,
		SavedAt:   time.Now().UTC(),
		Cookies:   cookies,
	}, nil
}

// ImportWebCookieSnapshot restores a WebCookieResult from a snapshot.
func ImportWebCookieSnapshot(snapshot WebCookieSnapshot) (*WebCookieResult, error) {
	if snapshot.Version != webCookieSnapshotVersion {
		return nil, &Error{Code: ErrorCodeDecode, Op: "import_web_cookie_snapshot", Message: fmt.Sprintf("unsupported snapshot version %d", snapshot.Version)}
	}
	domains := normalizeSnapshotDomains(snapshot.Domains)
	if len(domains) == 0 && len(snapshot.Cookies) == 0 {
		return nil, &Error{Code: ErrorCodeDecode, Op: "import_web_cookie_snapshot", Message: "snapshot has no domains or cookies"}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, &Error{Code: ErrorCodeConfig, Op: "import_web_cookie_snapshot", Message: "create cookie jar failed", Err: err}
	}

	domainSet := make(map[string]struct{}, len(domains)+len(snapshot.Cookies))
	for _, domain := range domains {
		domainSet[domain] = struct{}{}
	}
	for _, item := range snapshot.Cookies {
		domain := normalizeSnapshotDomain(item.Domain)
		name := strings.TrimSpace(item.Name)
		if domain == "" || name == "" {
			return nil, &Error{Code: ErrorCodeDecode, Op: "import_web_cookie_snapshot", Message: "snapshot cookie domain and name must not be empty"}
		}
		cookieURL := snapshotDomainURL(domain)
		jar.SetCookies(cookieURL, []*http.Cookie{{
			Name:  name,
			Value: item.Value,
			Path:  "/",
		}})
		domainSet[domain] = struct{}{}
	}

	domains = make([]string, 0, len(domainSet))
	for domain := range domainSet {
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	return &WebCookieResult{
		Jar:       jar,
		SessionID: strings.TrimSpace(snapshot.SessionID),
		SteamID:   strings.TrimSpace(snapshot.SteamID),
		Domains:   domains,
	}, nil
}

// SaveWebCookieResultJSON writes one WebCookieResult snapshot as indented JSON.
func SaveWebCookieResultJSON(w io.Writer, result *WebCookieResult) error {
	if w == nil {
		return &Error{Code: ErrorCodeRequestBuild, Op: "save_web_cookie_result_json", Message: "writer must not be nil"}
	}
	snapshot, err := ExportWebCookieSnapshot(result)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		return &Error{Code: ErrorCodeDecode, Op: "save_web_cookie_result_json", Message: "encode cookie snapshot failed", Err: err}
	}
	return nil
}

// LoadWebCookieResultJSON reads one WebCookieResult snapshot from JSON.
func LoadWebCookieResultJSON(r io.Reader) (*WebCookieResult, error) {
	if r == nil {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "load_web_cookie_result_json", Message: "reader must not be nil"}
	}
	var snapshot WebCookieSnapshot
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, &Error{Code: ErrorCodeDecode, Op: "load_web_cookie_result_json", Message: "decode cookie snapshot failed", Err: err}
	}
	return ImportWebCookieSnapshot(snapshot)
}

func normalizeSnapshotDomains(domains []string) []string {
	seen := make(map[string]struct{}, len(domains))
	out := make([]string, 0, len(domains))
	for _, domain := range domains {
		normalized := normalizeSnapshotDomain(domain)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func normalizeSnapshotDomain(domain string) string {
	domain = strings.Trim(strings.ToLower(strings.TrimSpace(domain)), ".")
	if domain == "" || strings.ContainsAny(domain, "/?#@") {
		return ""
	}
	if parsed, err := url.Parse("//" + domain); err == nil {
		return strings.ToLower(strings.Trim(parsed.Hostname(), "."))
	}
	return ""
}

func snapshotDomainURL(domain string) *url.URL {
	return &url.URL{Scheme: "https", Host: domain, Path: "/"}
}
