package daemon

import (
	"errors"
	"os"
	"testing"
	"time"

	"cookey/internal/config"
	"cookey/internal/models"
)

func TestDecodeSessionPayloadDirect(t *testing.T) {
	session, err := decodeSessionPayload([]byte(`{"cookies":[],"origins":[{"origin":"https://example.com","localStorage":[]}]}`))
	if err != nil {
		t.Fatalf("decodeSessionPayload() error = %v", err)
	}

	if len(session.Origins) != 1 || session.Origins[0].Origin != "https://example.com" {
		t.Fatalf("unexpected session: %+v", session)
	}
}

func TestDecodeSessionPayloadWrapped(t *testing.T) {
	session, err := decodeSessionPayload([]byte(`{"payload":{"cookies":[],"origins":[{"origin":"https://example.com","localStorage":[]}]}}`))
	if err != nil {
		t.Fatalf("decodeSessionPayload() error = %v", err)
	}

	if len(session.Origins) != 1 || session.Origins[0].Origin != "https://example.com" {
		t.Fatalf("unexpected session: %+v", session)
	}
}

func TestDecodeSessionPayloadJSONString(t *testing.T) {
	session, err := decodeSessionPayload([]byte(`"{\"cookies\":[],\"origins\":[{\"origin\":\"https://example.com\",\"localStorage\":[]}]}"`))
	if err != nil {
		t.Fatalf("decodeSessionPayload() error = %v", err)
	}

	if len(session.Origins) != 1 || session.Origins[0].Origin != "https://example.com" {
		t.Fatalf("unexpected session: %+v", session)
	}
}

func TestDecodeSessionPayloadEmpty(t *testing.T) {
	_, err := decodeSessionPayload([]byte(`   `))
	if err == nil {
		t.Fatal("decodeSessionPayload() error = nil, want non-nil")
	}
	if !errors.Is(err, ErrEmptySessionPayload) {
		t.Fatalf("decodeSessionPayload() error = %v, want ErrEmptySessionPayload", err)
	}
}

func TestMergeRefreshSessionFromSeedPreservesMissingState(t *testing.T) {
	root := t.TempDir()
	paths := config.NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", dir, err)
		}
	}

	previous := models.SessionFile{
		Cookies: []models.BrowserCookie{{
			Name:     "session",
			Value:    "old",
			Domain:   "example.com",
			Path:     "/",
			Expires:  -1,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Lax",
		}},
		Origins: []models.OriginState{{
			Origin: "https://example.com",
			LocalStorage: []models.OriginStorageItem{{
				Name:  "token",
				Value: "old-token",
			}},
		}},
		DeviceInfo: &models.DeviceInfo{
			DeviceID:       "device-1",
			APNEnvironment: "sandbox",
			APNToken:       "token-old",
			PublicKey:      "public-old",
		},
		Metadata: &models.SessionMetadata{TargetURL: "https://example.com"},
	}
	if err := config.WriteSession(previous, "rid-seed", paths); err != nil {
		t.Fatalf("WriteSession() error = %v", err)
	}

	merged := mergeRefreshSessionFromSeed(models.SessionFile{}, models.LoginManifest{
		RequestType: "refresh",
		TargetURL:   "https://example.com",
	}, paths)

	if len(merged.Cookies) != 1 || merged.Cookies[0].Value != "old" {
		t.Fatalf("merged cookies = %+v, want seeded cookie", merged.Cookies)
	}
	if len(merged.Origins) != 1 || len(merged.Origins[0].LocalStorage) != 1 || merged.Origins[0].LocalStorage[0].Value != "old-token" {
		t.Fatalf("merged origins = %+v, want seeded local storage", merged.Origins)
	}
	if merged.DeviceInfo == nil || merged.DeviceInfo.APNToken != "token-old" {
		t.Fatalf("merged device info = %+v, want seeded device info", merged.DeviceInfo)
	}
}

func TestMergeRefreshSessionFromSeedOverlaysNewState(t *testing.T) {
	root := t.TempDir()
	paths := config.NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", dir, err)
		}
	}

	previous := models.SessionFile{
		Cookies: []models.BrowserCookie{
			{Name: "session", Value: "old", Domain: "example.com", Path: "/", Expires: -1, SameSite: "Lax"},
			{Name: "keep", Value: "persist", Domain: "example.com", Path: "/", Expires: -1, SameSite: "Lax"},
		},
		Origins: []models.OriginState{{
			Origin: "https://example.com",
			LocalStorage: []models.OriginStorageItem{
				{Name: "token", Value: "old-token"},
				{Name: "keep", Value: "persist"},
			},
		}},
		Metadata: &models.SessionMetadata{TargetURL: "https://example.com"},
	}
	if err := config.WriteSession(previous, "rid-seed", paths); err != nil {
		t.Fatalf("WriteSession() error = %v", err)
	}

	merged := mergeRefreshSessionFromSeed(models.SessionFile{
		Cookies: []models.BrowserCookie{{
			Name:     "session",
			Value:    "new",
			Domain:   "example.com",
			Path:     "/",
			Expires:  -1,
			SameSite: "Lax",
		}},
		Origins: []models.OriginState{{
			Origin: "https://example.com",
			LocalStorage: []models.OriginStorageItem{{
				Name:  "token",
				Value: "new-token",
			}},
		}},
	}, models.LoginManifest{
		RequestType: "refresh",
		TargetURL:   "https://example.com",
	}, paths)

	if len(merged.Cookies) != 2 {
		t.Fatalf("merged cookies len = %d, want 2", len(merged.Cookies))
	}
	if merged.Cookies[0].Name != "session" || merged.Cookies[0].Value != "new" {
		t.Fatalf("merged cookies[0] = %+v, want refreshed session cookie", merged.Cookies[0])
	}
	if merged.Cookies[1].Name != "keep" || merged.Cookies[1].Value != "persist" {
		t.Fatalf("merged cookies[1] = %+v, want preserved cookie", merged.Cookies[1])
	}
	if len(merged.Origins) != 1 || len(merged.Origins[0].LocalStorage) != 2 {
		t.Fatalf("merged origins = %+v, want merged storage", merged.Origins)
	}
	if merged.Origins[0].LocalStorage[0].Name != "token" || merged.Origins[0].LocalStorage[0].Value != "new-token" {
		t.Fatalf("merged localStorage[0] = %+v, want refreshed token", merged.Origins[0].LocalStorage[0])
	}
	if merged.Origins[0].LocalStorage[1].Name != "keep" || merged.Origins[0].LocalStorage[1].Value != "persist" {
		t.Fatalf("merged localStorage[1] = %+v, want preserved item", merged.Origins[0].LocalStorage[1])
	}
}

func TestMergeRefreshSessionFromSeedSkipsNormalLogin(t *testing.T) {
	merged := mergeRefreshSessionFromSeed(models.SessionFile{}, models.LoginManifest{
		RequestType: "login",
		TargetURL:   "https://example.com",
	}, config.AppPaths{})

	if len(merged.Cookies) != 0 || len(merged.Origins) != 0 || merged.DeviceInfo != nil {
		t.Fatalf("merged session = %+v, want unchanged empty session", merged)
	}
}

func TestMergeRefreshSessionFromSeedFallsBackToOlderDeviceInfo(t *testing.T) {
	root := t.TempDir()
	paths := config.NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", dir, err)
		}
	}

	if err := config.WriteSession(models.SessionFile{
		Cookies:    []models.BrowserCookie{},
		Origins:    []models.OriginState{},
		DeviceInfo: &models.DeviceInfo{DeviceID: "device-1", APNEnvironment: "sandbox", APNToken: "token-old", PublicKey: "public-old"},
		Metadata:   &models.SessionMetadata{TargetURL: "https://example.com"},
	}, "rid-with-device-info", paths); err != nil {
		t.Fatalf("WriteSession(rid-with-device-info) error = %v", err)
	}

	olderTime := time.Now().Add(-1 * time.Minute)
	if err := os.Chtimes(paths.SessionPath("rid-with-device-info"), olderTime, olderTime); err != nil {
		t.Fatalf("Chtimes(rid-with-device-info) error = %v", err)
	}

	if err := config.WriteSession(models.SessionFile{
		Cookies:  []models.BrowserCookie{{Name: "session", Value: "latest", Domain: "example.com", Path: "/", Expires: -1, SameSite: "Lax"}},
		Origins:  []models.OriginState{},
		Metadata: &models.SessionMetadata{TargetURL: "https://example.com"},
	}, "rid-latest-no-device-info", paths); err != nil {
		t.Fatalf("WriteSession(rid-latest-no-device-info) error = %v", err)
	}

	merged := mergeRefreshSessionFromSeed(models.SessionFile{}, models.LoginManifest{
		RequestType: "refresh",
		TargetURL:   "https://example.com",
	}, paths)

	if len(merged.Cookies) != 1 || merged.Cookies[0].Value != "latest" {
		t.Fatalf("merged cookies = %+v, want latest session cookies", merged.Cookies)
	}
	if merged.DeviceInfo == nil || merged.DeviceInfo.APNToken != "token-old" {
		t.Fatalf("merged device info = %+v, want fallback device info", merged.DeviceInfo)
	}
}
