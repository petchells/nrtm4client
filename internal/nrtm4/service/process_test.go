package service

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/petchells/nrtm4client/internal/nrtm4/persist"
	"github.com/petchells/nrtm4client/internal/nrtm4/pg"
	"github.com/petchells/nrtm4client/internal/nrtm4/testresources"
)

var stubNotificationURL = "https://example.com/source1/notification.json"
var stubSnapshot2URL = "https://example.com/ca128382-78d9-41d1-8927-1ecef15275be/nrtm-snapshot.2.047595d0fae972fbed0c51b4a41c7a349e0c47bb.json.gz"

func TestFileRefSorter(t *testing.T) {
	refs := []persist.FileRefJSON{
		{
			Version: 4,
			URL:     "https://xxx.xxx.xxx/4",
			Hash:    "4444",
		},
		{
			Version: 6,
			URL:     "https://xxx.xxx.xxx/6",
			Hash:    "6666",
		},
		{
			Version: 3,
			URL:     "https://xxx.xxx.xxx/3",
			Hash:    "3333",
		},
		{
			Version: 5,
			URL:     "https://xxx.xxx.xxx/5",
			Hash:    "5",
		},
	}
	sort.Sort(fileRefsByVersion(refs))
	expect := [...]uint32{3, 4, 5, 6}
	for idx, v := range expect {
		if refs[idx].Version != v {
			t.Error("Expected", v, "but got", refs[idx].Version)
		}
	}
}

func TestE2EConnect(t *testing.T) {
	testresources.SetEnvVarsFromFile(t, "../testresources/env.test.conf")
	pgTestRepo := pgRepo()
	testresources.TruncateDatabase(t)
	stubClient := NewStubClient(t)
	tmpDir, err := os.MkdirTemp("", "nrtmtest*")
	if err != nil {
		t.Fatal("Could not create temp test directory")
	}
	defer os.RemoveAll(tmpDir)

	conf := AppConfig{
		NRTMFilePath: tmpDir,
	}
	processor := NewNRTMProcessor(conf, pgTestRepo, stubClient)
	if err = processor.Connect(stubNotificationURL, ""); err != nil {
		t.Fatal("Failed to Connect", err)
	}
	sources, err := processor.ListSources()
	if len(sources) != 1 {
		t.Error("Should only be a single source")
	}
	src := sources[0]
	if src.Source != "EXAMPLE" {
		t.Error("Source should be EXAMPLE")
	}
	if src.Version != 3 {
		t.Error("Version should be 3")
	}
	if src.NotificationURL != stubNotificationURL {
		t.Error("NotificationURL should be", stubNotificationURL)
	}
	if src.SessionID != "ca128382-78d9-41d1-8927-1ecef15275be" {
		t.Error("SessionID should be", "ca128382-78d9-41d1-8927-1ecef15275be")
	}
}

func TestFindUpdatesSuccess(t *testing.T) {

	var notification persist.NotificationJSON
	readJSON("../testresources/ripe-notification-file.json", &notification)
	source := stubsource()

	fileRefs, err := findUpdates(notification, source)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	expectedLen := 9
	if len(fileRefs) != expectedLen {
		t.Fatalf("Unexpected slice length. Expected %d but was %d", expectedLen, len(fileRefs))
	}
}

func TestFindUpdatesErrors(t *testing.T) {

	var notification persist.NotificationJSON
	{
		readJSON("../testresources/ripe-notification-file.json", &notification)
		source := stubsource()
		source.Version = 350194 - 2

		expect := ErrNRTMNextConsecutiveDeltaUnavaliable

		_, err := findUpdates(notification, source)
		if err != expect {
			t.Errorf("Expected error %v but was %v", expect, err)
		}
	}
	{
		readJSON("../testresources/ripe-notification-file.json", &notification)
		source := stubsource()
		refs := *notification.DeltaRefs
		dr := append(refs[:10], refs[11:]...)
		notification.DeltaRefs = &dr

		expect := ErrNRTMNotificationDeltaSequenceBroken

		_, err := findUpdates(notification, source)
		if err != expect {
			t.Errorf("Expected error %v but was %v", expect, err)
		}
	}
	{
		readJSON("../testresources/ripe-notification-file.json", &notification)
		source := stubsource()
		refs := *notification.DeltaRefs
		dr := refs[:len(refs)-2]
		notification.DeltaRefs = &dr

		expect := ErrNRTMNotificationVersionDoesNotMatchDelta

		_, err := findUpdates(notification, source)
		if err != expect {
			t.Errorf("Expected error %v but was %v", expect, err)
		}
	}
	{
		readJSON("../testresources/ripe-notification-file.json", &notification)
		source := stubsource()
		refs := *notification.DeltaRefs
		dr := append(refs[:10], refs[9:]...)
		notification.DeltaRefs = &dr

		expect := ErrNRTMDuplicateDeltaVersion

		_, err := findUpdates(notification, source)
		if err != expect {
			t.Errorf("Expected error %v but was %v", expect, err)
		}
	}
	{
		readJSON("../testresources/ripe-notification-file.json", &notification)
		source := stubsource()
		dr := []persist.FileRefJSON{}
		notification.DeltaRefs = &dr

		expect := ErrNRTMNoDeltasInNotification

		_, err := findUpdates(notification, source)
		if err != expect {
			t.Errorf("Expected error %v but was %v", expect, err)
		}
	}
}

func pgRepo() persist.Repository {
	dbURL := os.Getenv("PG_DATABASE_URL")
	if len(dbURL) == 0 {
		log.Fatal("ERROR no url for database", dbURL)
		return nil
	}
	repo := pg.PostgresRepository{}
	if err := repo.Initialize(dbURL); err != nil {
		log.Fatal("Failed to initialize repository")
	}
	return &repo
}

func stubsource() persist.NRTMSource {
	t, err := time.Parse(time.RFC3339, "2025-01-04T23:01:00Z")
	if err != nil {
		log.Fatalln("bad timestamp")
	}
	src := persist.NRTMSource{
		ID:              576576257634,
		Source:          "TEST_SRC",
		SessionID:       "db44e038-1f07-4d54-a307-1b32339f141a",
		Version:         350684,
		NotificationURL: stubNotificationURL,
		Label:           "",
		Created:         t,
	}
	return src
}

func readJSON(fileName string, ptr any) {
	var err error

	var file *os.File
	if file, err = os.Open(fileName); err != nil {
		log.Println(err)
		return
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(bytes, ptr)
	if err != nil {
		log.Println(err)
	}
}
