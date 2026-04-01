package config

import (
	"os"
	"testing"
	"time"

	"cookey/internal/models"
)

func TestListLocalRIDsReturnsUnionNewestFirst(t *testing.T) {
	root := t.TempDir()
	paths := NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := ensureDirectory(dir, 0o700); err != nil {
			t.Fatalf("ensureDirectory(%q) error = %v", dir, err)
		}
	}

	if err := WriteSession(models.SessionFile{Cookies: []models.BrowserCookie{}, Origins: []models.OriginState{}}, "rid-session", paths); err != nil {
		t.Fatalf("WriteSession(rid-session) error = %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := WriteDaemon(models.DaemonDescriptor{RID: "rid-daemon", Status: models.DaemonStateWaiting}, paths); err != nil {
		t.Fatalf("WriteDaemon(rid-daemon) error = %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := WriteSession(models.SessionFile{Cookies: []models.BrowserCookie{}, Origins: []models.OriginState{}}, "rid-both", paths); err != nil {
		t.Fatalf("WriteSession(rid-both) error = %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := WriteDaemon(models.DaemonDescriptor{RID: "rid-both", Status: models.DaemonStateReady}, paths); err != nil {
		t.Fatalf("WriteDaemon(rid-both) error = %v", err)
	}

	rids, err := ListLocalRIDs(paths)
	if err != nil {
		t.Fatalf("ListLocalRIDs() error = %v", err)
	}

	want := []string{"rid-both", "rid-daemon", "rid-session"}
	if len(rids) != len(want) {
		t.Fatalf("ListLocalRIDs() len = %d, want %d (%v)", len(rids), len(want), rids)
	}
	for index := range want {
		if rids[index] != want[index] {
			t.Fatalf("ListLocalRIDs()[%d] = %q, want %q (all=%v)", index, rids[index], want[index], rids)
		}
	}
}

func TestDeleteLocalRIDRemovesSessionAndDaemon(t *testing.T) {
	root := t.TempDir()
	paths := NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := ensureDirectory(dir, 0o700); err != nil {
			t.Fatalf("ensureDirectory(%q) error = %v", dir, err)
		}
	}

	if err := WriteSession(models.SessionFile{Cookies: []models.BrowserCookie{}, Origins: []models.OriginState{}}, "rid-1", paths); err != nil {
		t.Fatalf("WriteSession() error = %v", err)
	}
	if err := WriteDaemon(models.DaemonDescriptor{RID: "rid-1", Status: models.DaemonStateReady}, paths); err != nil {
		t.Fatalf("WriteDaemon() error = %v", err)
	}

	sessionDeleted, daemonDeleted, err := DeleteLocalRID(paths, "rid-1")
	if err != nil {
		t.Fatalf("DeleteLocalRID() error = %v", err)
	}
	if !sessionDeleted || !daemonDeleted {
		t.Fatalf("DeleteLocalRID() = (%t, %t), want both true", sessionDeleted, daemonDeleted)
	}
	if fileExists(paths.SessionPath("rid-1")) || fileExists(paths.DaemonPath("rid-1")) {
		t.Fatal("expected local files to be deleted")
	}
}

func TestDeleteLocalRIDRejectsActiveDaemon(t *testing.T) {
	root := t.TempDir()
	paths := NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := ensureDirectory(dir, 0o700); err != nil {
			t.Fatalf("ensureDirectory(%q) error = %v", dir, err)
		}
	}

	if err := WriteDaemon(models.DaemonDescriptor{
		RID:    "rid-active",
		PID:    int32(os.Getpid()),
		Status: models.DaemonStateWaiting,
	}, paths); err != nil {
		t.Fatalf("WriteDaemon() error = %v", err)
	}

	sessionDeleted, daemonDeleted, err := DeleteLocalRID(paths, "rid-active")
	if err != ErrActiveRequest {
		t.Fatalf("DeleteLocalRID() error = %v, want ErrActiveRequest", err)
	}
	if sessionDeleted || daemonDeleted {
		t.Fatalf("DeleteLocalRID() = (%t, %t), want both false", sessionDeleted, daemonDeleted)
	}
	if !fileExists(paths.DaemonPath("rid-active")) {
		t.Fatal("expected active daemon descriptor to remain present")
	}
}

func TestSyncDeviceInfoUpdatesMatchingSessions(t *testing.T) {
	root := t.TempDir()
	paths := NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := ensureDirectory(dir, 0o700); err != nil {
			t.Fatalf("ensureDirectory(%q) error = %v", dir, err)
		}
	}

	currentRID := "rid-current"
	otherRID := "rid-other"
	unrelatedRID := "rid-unrelated"

	matchingDeviceInfo := &models.DeviceInfo{
		DeviceID:       "device-shared",
		APNEnvironment: "sandbox",
		APNToken:       "token-old",
		PublicKey:      "public-old",
	}

	if err := WriteSession(models.SessionFile{
		Cookies:    []models.BrowserCookie{},
		Origins:    []models.OriginState{},
		DeviceInfo: matchingDeviceInfo,
	}, currentRID, paths); err != nil {
		t.Fatalf("WriteSession(current) error = %v", err)
	}

	if err := WriteSession(models.SessionFile{
		Cookies: []models.BrowserCookie{},
		Origins: []models.OriginState{},
		DeviceInfo: &models.DeviceInfo{
			DeviceID:       "device-shared",
			APNEnvironment: "production",
			APNToken:       "token-stale",
			PublicKey:      "public-stale",
		},
	}, otherRID, paths); err != nil {
		t.Fatalf("WriteSession(other) error = %v", err)
	}

	if err := WriteSession(models.SessionFile{
		Cookies: []models.BrowserCookie{},
		Origins: []models.OriginState{},
		DeviceInfo: &models.DeviceInfo{
			DeviceID:       "device-other",
			APNEnvironment: "production",
			APNToken:       "token-keep",
			PublicKey:      "public-keep",
		},
	}, unrelatedRID, paths); err != nil {
		t.Fatalf("WriteSession(unrelated) error = %v", err)
	}

	updatedInfo := &models.DeviceInfo{
		DeviceID:       "device-shared",
		APNEnvironment: "sandbox",
		APNToken:       "token-new",
		PublicKey:      "public-new",
	}

	if err := SyncDeviceInfo(paths, currentRID, updatedInfo); err != nil {
		t.Fatalf("SyncDeviceInfo() error = %v", err)
	}

	otherSession, err := ReadJSON[models.SessionFile](paths.SessionPath(otherRID))
	if err != nil {
		t.Fatalf("ReadJSON(other) error = %v", err)
	}
	if otherSession.DeviceInfo == nil {
		t.Fatal("expected other session device info to be updated")
	}
	if *otherSession.DeviceInfo != *updatedInfo {
		t.Fatalf("updated device info = %+v, want %+v", *otherSession.DeviceInfo, *updatedInfo)
	}

	currentSession, err := ReadJSON[models.SessionFile](paths.SessionPath(currentRID))
	if err != nil {
		t.Fatalf("ReadJSON(current) error = %v", err)
	}
	if currentSession.DeviceInfo == nil {
		t.Fatal("expected current session device info to remain present")
	}
	if *currentSession.DeviceInfo != *matchingDeviceInfo {
		t.Fatalf("current session device info = %+v, want %+v", *currentSession.DeviceInfo, *matchingDeviceInfo)
	}

	unrelatedSession, err := ReadJSON[models.SessionFile](paths.SessionPath(unrelatedRID))
	if err != nil {
		t.Fatalf("ReadJSON(unrelated) error = %v", err)
	}
	if unrelatedSession.DeviceInfo == nil {
		t.Fatal("expected unrelated session device info to remain present")
	}
	if unrelatedSession.DeviceInfo.DeviceID != "device-other" || unrelatedSession.DeviceInfo.APNToken != "token-keep" {
		t.Fatalf("unrelated session device info changed unexpectedly: %+v", *unrelatedSession.DeviceInfo)
	}
}

func TestSyncDeviceInfoSkipsMissingDeviceIdentifier(t *testing.T) {
	root := t.TempDir()
	paths := NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := ensureDirectory(dir, 0o700); err != nil {
			t.Fatalf("ensureDirectory(%q) error = %v", dir, err)
		}
	}

	rid := "rid-1"
	original := models.SessionFile{
		Cookies: []models.BrowserCookie{},
		Origins: []models.OriginState{},
		DeviceInfo: &models.DeviceInfo{
			DeviceID:       "device-1",
			APNEnvironment: "sandbox",
			APNToken:       "token-1",
			PublicKey:      "public-1",
		},
	}
	if err := WriteSession(original, rid, paths); err != nil {
		t.Fatalf("WriteSession() error = %v", err)
	}

	if err := SyncDeviceInfo(paths, "rid-current", &models.DeviceInfo{APNToken: "token-2"}); err != nil {
		t.Fatalf("SyncDeviceInfo() error = %v", err)
	}

	session, err := ReadJSON[models.SessionFile](paths.SessionPath(rid))
	if err != nil {
		t.Fatalf("ReadJSON() error = %v", err)
	}
	if session.DeviceInfo == nil {
		t.Fatal("expected device info to remain present")
	}
	if *session.DeviceInfo != *original.DeviceInfo {
		t.Fatalf("device info changed unexpectedly: %+v", *session.DeviceInfo)
	}
}

func TestLatestDeviceInfoForTargetReturnsNewestMatchingDeviceInfo(t *testing.T) {
	root := t.TempDir()
	paths := NewAppPaths(root)
	for _, dir := range []string{paths.Root, paths.Sessions, paths.Daemons} {
		if err := ensureDirectory(dir, 0o700); err != nil {
			t.Fatalf("ensureDirectory(%q) error = %v", dir, err)
		}
	}

	if err := WriteSession(models.SessionFile{
		Cookies: []models.BrowserCookie{},
		Origins: []models.OriginState{},
		Metadata: &models.SessionMetadata{TargetURL: "https://example.com"},
	}, "rid-no-device-info", paths); err != nil {
		t.Fatalf("WriteSession(no-device-info) error = %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	olderDeviceInfo := &models.DeviceInfo{
		DeviceID:       "device-1",
		APNEnvironment: "sandbox",
		APNToken:       "token-old",
		PublicKey:      "public-old",
	}
	if err := WriteSession(models.SessionFile{
		Cookies:    []models.BrowserCookie{},
		Origins:    []models.OriginState{},
		DeviceInfo: olderDeviceInfo,
		Metadata:   &models.SessionMetadata{TargetURL: "https://example.com"},
	}, "rid-device-old", paths); err != nil {
		t.Fatalf("WriteSession(device-old) error = %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	newerDeviceInfo := &models.DeviceInfo{
		DeviceID:       "device-1",
		APNEnvironment: "production",
		APNToken:       "token-new",
		PublicKey:      "public-new",
	}
	if err := WriteSession(models.SessionFile{
		Cookies:    []models.BrowserCookie{},
		Origins:    []models.OriginState{},
		DeviceInfo: newerDeviceInfo,
		Metadata:   &models.SessionMetadata{TargetURL: "https://example.com"},
	}, "rid-device-new", paths); err != nil {
		t.Fatalf("WriteSession(device-new) error = %v", err)
	}

	deviceInfo, rid, err := LatestDeviceInfoForTarget(paths, "https://example.com")
	if err != nil {
		t.Fatalf("LatestDeviceInfoForTarget() error = %v", err)
	}
	if rid != "rid-device-new" {
		t.Fatalf("LatestDeviceInfoForTarget() rid = %q, want %q", rid, "rid-device-new")
	}
	if deviceInfo == nil {
		t.Fatal("expected device info to be returned")
	}
	if *deviceInfo != *newerDeviceInfo {
		t.Fatalf("LatestDeviceInfoForTarget() device info = %+v, want %+v", *deviceInfo, *newerDeviceInfo)
	}

	missingDeviceInfo, missingRID, err := LatestDeviceInfoForTarget(paths, "https://missing.example.com")
	if err != nil {
		t.Fatalf("LatestDeviceInfoForTarget(missing) error = %v", err)
	}
	if missingDeviceInfo != nil || missingRID != "" {
		t.Fatalf("LatestDeviceInfoForTarget(missing) = (%+v, %q), want (nil, \"\")", missingDeviceInfo, missingRID)
	}
}