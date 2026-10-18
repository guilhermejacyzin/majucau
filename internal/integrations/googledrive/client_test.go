package googledrive

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/securetemp"
)

var testFolderIDs = []string{"folder-a", "folder-b", "folder-c"}

func TestListFilesRestrictsFolderAndRetries429(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls++
		if request.Method != http.MethodGet || request.URL.Path != "/drive/v3/files" {
			t.Errorf("request = %s %s, want GET /drive/v3/files", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-access-token" {
			t.Errorf("authorization header missing or wrong")
		}
		query := request.URL.Query()
		if query.Get("q") != "'folder-a' in parents and trashed = false" || query.Get("pageSize") != "100" {
			t.Errorf("query = %#v, want approved-folder filter and page size", query)
		}
		if query.Get("includeItemsFromAllDrives") != "true" || query.Get("supportsAllDrives") != "true" {
			t.Errorf("shared-drive flags missing")
		}
		if calls == 1 {
			writer.Header().Set("Retry-After", "1")
			writer.WriteHeader(http.StatusTooManyRequests)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"files":[{"id":"file-1","name":"ledger.csv","mimeType":"text/csv","size":"11","modifiedTime":"2026-10-18T09:00:00Z","md5Checksum":"abc","version":7,"parents":["folder-a"]}],"nextPageToken":"next"}`)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	client.wait = func(context.Context, time.Duration) error { return nil }

	page, err := client.ListFiles(context.Background(), "folder-a", "")
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if calls != 2 || len(page.Files) != 1 || page.NextPageToken != "next" || page.Files[0].ID != "file-1" || page.Files[0].Size != 11 {
		t.Fatalf("ListFiles() = %#v after %d calls", page, calls)
	}
}

func TestListFilesRejectsUnapprovedFolderWithoutNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected request") }))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "token", nil }))
	if _, err := client.ListFiles(context.Background(), "root", ""); !errors.Is(err, ErrFileNotAllowed) {
		t.Fatalf("ListFiles() error = %v, want folder policy error", err)
	}
}

func TestListFilesDoesNotRetryAuthorizationFailure(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		calls++
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	if _, err := client.ListFiles(context.Background(), "folder-a", ""); !errors.Is(err, ErrAuthorization) {
		t.Fatalf("ListFiles() error = %v, want authorization failure", err)
	}
	if calls != 1 {
		t.Fatalf("authorization request attempts = %d, want 1", calls)
	}
}

func TestWalkFilesStreamsPagesAndReturnsResumeTokenAtLimit(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls++
		writer.Header().Set("Content-Type", "application/json")
		if request.URL.Query().Get("pageToken") == "" {
			_, _ = io.WriteString(writer, `{"files":[{"id":"file-1","name":"one.csv","mimeType":"text/csv","parents":["folder-a"]},{"id":"file-2","name":"two.csv","mimeType":"text/csv","parents":["folder-a"]}],"nextPageToken":"page-two"}`)
			return
		}
		if got := request.URL.Query().Get("pageToken"); got != "page-two" {
			t.Errorf("pageToken = %q, want page-two", got)
		}
		_, _ = io.WriteString(writer, `{"files":[{"id":"file-3","name":"three.csv","mimeType":"text/csv","parents":["folder-a"]}]}`)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	var seen []string
	visit := func(file File) error {
		seen = append(seen, file.ID)
		return nil
	}

	first, err := client.WalkFiles(context.Background(), "folder-a", "", 1, visit)
	if !errors.Is(err, ErrFileWalkLimit) {
		t.Fatalf("first WalkFiles() error = %v, want page-limit signal", err)
	}
	if first.Complete || first.PagesRead != 1 || first.FilesVisited != 2 || first.NextPageToken != "page-two" {
		t.Fatalf("first WalkFiles() result = %#v", first)
	}
	second, err := client.WalkFiles(context.Background(), "folder-a", first.NextPageToken, 1, visit)
	if err != nil {
		t.Fatalf("resumed WalkFiles() error = %v", err)
	}
	if !second.Complete || second.PagesRead != 1 || second.FilesVisited != 1 || second.NextPageToken != "" {
		t.Fatalf("resumed WalkFiles() result = %#v", second)
	}
	if calls != 2 || strings.Join(seen, ",") != "file-1,file-2,file-3" {
		t.Fatalf("walk calls=%d files=%v", calls, seen)
	}
}

func TestWalkFilesStopsOnRepeatedContinuationToken(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls++
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"files":[],"nextPageToken":"repeat"}`)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "token", nil }))
	result, err := client.WalkFiles(context.Background(), "folder-a", "", 5, func(File) error { return nil })
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("WalkFiles() error = %v, want repeated-token response failure", err)
	}
	if calls != 2 || result.PagesRead != 1 || result.Complete {
		t.Fatalf("walk calls=%d result=%#v", calls, result)
	}
}

func TestWalkFilePagesDeliversPageAndCheckpointTogether(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls++
		writer.Header().Set("Content-Type", "application/json")
		if request.URL.Query().Get("pageToken") == "" {
			_, _ = io.WriteString(writer, `{"files":[{"id":"file-1","name":"one.csv","mimeType":"text/csv","parents":["folder-a"]},{"id":"file-2","name":"two.csv","mimeType":"text/csv","parents":["folder-a"]}],"nextPageToken":"page-two"}`)
			return
		}
		_, _ = io.WriteString(writer, `{"files":[{"id":"file-3","name":"three.csv","mimeType":"text/csv","parents":["folder-a"]}]}`)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "token", nil }))
	var storedToken string
	var visited []string
	first, err := client.WalkFilePages(context.Background(), "folder-a", "", 1, func(page FilePage, nextPageToken string) error {
		if len(page.Files) != 2 || nextPageToken != "page-two" {
			t.Fatalf("first page callback got page=%#v token=%q", page, nextPageToken)
		}
		for _, file := range page.Files {
			visited = append(visited, file.ID)
		}
		storedToken = nextPageToken // A caller commits these effects and token in one transaction.
		return nil
	})
	if !errors.Is(err, ErrFileWalkLimit) || first.PagesRead != 1 || first.FilesVisited != 2 || first.NextPageToken != "page-two" || storedToken != first.NextPageToken {
		t.Fatalf("first page result=%#v stored token=%q error=%v", first, storedToken, err)
	}
	second, err := client.WalkFilePages(context.Background(), "folder-a", storedToken, 1, func(page FilePage, nextPageToken string) error {
		if len(page.Files) != 1 || page.Files[0].ID != "file-3" || nextPageToken != "" {
			t.Fatalf("second page callback got page=%#v token=%q", page, nextPageToken)
		}
		visited = append(visited, page.Files[0].ID)
		storedToken = nextPageToken
		return nil
	})
	if err != nil || !second.Complete || second.PagesRead != 1 || second.FilesVisited != 1 || storedToken != "" {
		t.Fatalf("second page result=%#v stored token=%q error=%v", second, storedToken, err)
	}
	if calls != 2 || strings.Join(visited, ",") != "file-1,file-2,file-3" {
		t.Fatalf("page calls=%d files=%v", calls, visited)
	}
}

func TestWalkFilePagesFailureKeepsResumeTokenAtCurrentPage(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		calls++
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"files":[{"id":"file-1","name":"one.csv","mimeType":"text/csv","parents":["folder-a"]}],"nextPageToken":"page-two"}`)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "token", nil }))
	privateCause := errors.New("database path must remain private")
	first, err := client.WalkFilePages(context.Background(), "folder-a", "", 1, func(FilePage, string) error { return privateCause })
	if !errors.Is(err, ErrFileVisitor) || !errors.Is(err, privateCause) || strings.Contains(err.Error(), privateCause.Error()) {
		t.Fatalf("first WalkFilePages() error=%v, want sanitized callback error", err)
	}
	if first.PagesRead != 0 || first.FilesVisited != 0 || first.NextPageToken != "" || first.Complete {
		t.Fatalf("failed page advanced progress: %#v", first)
	}
	retry, retryErr := client.WalkFilePages(context.Background(), "folder-a", first.NextPageToken, 1, func(FilePage, string) error { return nil })
	if !errors.Is(retryErr, ErrFileWalkLimit) || retry.PagesRead != 1 || retry.FilesVisited != 1 || retry.NextPageToken != "page-two" || calls != 2 {
		t.Fatalf("retry result=%#v calls=%d error=%v", retry, calls, retryErr)
	}
}

func TestWalkFilesSanitizesVisitorFailureAndReplaysCurrentPage(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		calls++
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"files":[{"id":"file-1","name":"one.csv","mimeType":"text/csv","parents":["folder-a"]},{"id":"file-2","name":"two.csv","mimeType":"text/csv","parents":["folder-a"]}],"nextPageToken":"page-two"}`)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "token", nil }))
	privateCause := errors.New("private file path must not appear")
	var visited []string
	failedOnce := false
	result, err := client.WalkFiles(context.Background(), "folder-a", "", 1, func(file File) error {
		if file.ID == "file-2" && !failedOnce {
			failedOnce = true
			return privateCause
		}
		visited = append(visited, file.ID)
		return nil
	})
	if !errors.Is(err, ErrFileVisitor) || !errors.Is(err, privateCause) {
		t.Fatalf("WalkFiles() error = %v, want sanitized visitor failure retaining cause", err)
	}
	if strings.Contains(err.Error(), privateCause.Error()) || result.Complete || result.PagesRead != 0 || result.FilesVisited != 1 || result.NextPageToken != "" {
		t.Fatalf("visitor failure leaked or returned unsafe resume point: result=%#v error=%v", result, err)
	}

	retryResult, retryErr := client.WalkFiles(context.Background(), "folder-a", result.NextPageToken, 1, func(file File) error {
		visited = append(visited, file.ID)
		return nil
	})
	if !errors.Is(retryErr, ErrFileWalkLimit) || retryResult.NextPageToken != "page-two" || retryResult.FilesVisited != 2 {
		t.Fatalf("retry result=%#v error=%v, want current page replayed and next page token", retryResult, retryErr)
	}
	if calls != 2 || strings.Join(visited, ",") != "file-1,file-1,file-2" {
		t.Fatalf("retry calls=%d delivered=%v, want first file replayed safely", calls, visited)
	}
}

func TestWalkFilesRejectsInvalidLimitsAndCanceledContextWithoutNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected request") }))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "token", nil }))
	visit := func(File) error { return nil }
	for _, limit := range []int{0, maxFileWalkPagesPerCall + 1} {
		if _, err := client.WalkFiles(context.Background(), "folder-a", "", limit, visit); !errors.Is(err, ErrConfiguration) {
			t.Fatalf("WalkFiles(maxPages=%d) error = %v, want configuration error", limit, err)
		}
	}
	if _, err := client.WalkFiles(context.Background(), "folder-a", "", 1, nil); !errors.Is(err, ErrConfiguration) {
		t.Fatalf("WalkFiles(nil visitor) error = %v, want configuration error", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.WalkFiles(ctx, "folder-a", "", 1, visit); !errors.Is(err, context.Canceled) {
		t.Fatalf("WalkFiles(canceled context) error = %v, want context cancellation", err)
	}
}

func TestStreamFileWritesVerifiedContentAndCommits(t *testing.T) {
	content := []byte("hello drive")
	md5sum := md5.Sum(content)
	sha := sha256.Sum256(content)
	metadataCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.Header.Get("Authorization") != "Bearer test-access-token" {
			t.Errorf("unexpected request method or missing bearer token")
		}
		switch {
		case request.URL.Path == "/drive/v3/files/file-1" && request.URL.Query().Get("alt") == "media":
			writer.Header().Set("Content-Type", "application/octet-stream")
			_, _ = writer.Write(content)
		case request.URL.Path == "/drive/v3/files/file-1":
			metadataCalls++
			_, _ = io.WriteString(writer, fmt.Sprintf(`{"id":"file-1","name":"ledger.csv","mimeType":"text/csv","size":"%d","modifiedTime":"2026-10-18T09:00:00Z","md5Checksum":"%s","version":7,"parents":["folder-a"]}`, len(content), hex.EncodeToString(md5sum[:])))
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	sink := &memorySink{}
	result, err := client.StreamFile(context.Background(), "file-1", 1<<20, sink)
	if err != nil {
		t.Fatalf("StreamFile() error = %v", err)
	}
	if !sink.committed || sink.rolledBack || sink.String() != string(content) || result.Bytes != int64(len(content)) || result.SHA256 != hex.EncodeToString(sha[:]) || metadataCalls != 2 {
		t.Fatalf("result=%#v sink=%#v metadata calls=%d", result, sink, metadataCalls)
	}
}

func TestStreamFileRollsBackWhenSourceChanges(t *testing.T) {
	content := []byte("hello")
	metadataCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("alt") == "media" {
			_, _ = writer.Write(content)
			return
		}
		metadataCalls++
		version := 7
		if metadataCalls > 1 {
			version = 8
		}
		_, _ = io.WriteString(writer, fmt.Sprintf(`{"id":"file-1","name":"ledger.csv","mimeType":"text/csv","size":"%d","modifiedTime":"2026-10-18T09:00:00Z","version":%d,"parents":["folder-a"]}`, len(content), version))
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	sink := &memorySink{}
	if _, err := client.StreamFile(context.Background(), "file-1", 1<<20, sink); !errors.Is(err, ErrFileChanged) {
		t.Fatalf("StreamFile() error = %v, want source-version failure", err)
	}
	if sink.committed || !sink.rolledBack || sink.Len() != 0 {
		t.Fatalf("sink published or retained bytes after change: %#v", sink)
	}
}

func TestStreamFilePreservesManualCleanupPathForAuthorizedOperator(t *testing.T) {
	content := []byte("hello")
	metadataCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("alt") == "media" {
			_, _ = writer.Write(content)
			return
		}
		metadataCalls++
		version := 7
		if metadataCalls > 1 {
			version = 8
		}
		_, _ = io.WriteString(writer, fmt.Sprintf(`{"id":"file-1","name":"ledger.csv","mimeType":"text/csv","size":"%d","modifiedTime":"2026-10-18T09:00:00Z","version":%d,"parents":["folder-a"]}`, len(content), version))
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	cleanup := &securetemp.CleanupError{Path: `C:\private\temporary-data`, Cause: errors.New("denied")}
	sink := &memorySink{rollbackErr: cleanup}
	_, err := client.StreamFile(context.Background(), "file-1", 1<<20, sink)
	var got *securetemp.CleanupError
	if !errors.As(err, &got) || got.Path != cleanup.Path {
		t.Fatalf("cleanup error = %v, want the typed path for authorized cleanup", err)
	}
	if strings.Contains(err.Error(), cleanup.Path) {
		t.Fatalf("cleanup path leaked into ordinary error: %v", err)
	}
}

func TestStreamFileRejectsOversizeBeforeDownload(t *testing.T) {
	downloads := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("alt") == "media" {
			downloads++
			return
		}
		_, _ = io.WriteString(writer, `{"id":"file-1","name":"ledger.csv","mimeType":"text/csv","size":"20","version":7,"parents":["folder-a"]}`)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	sink := &memorySink{}
	if _, err := client.StreamFile(context.Background(), "file-1", 10, sink); !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("StreamFile() error = %v, want file-size failure", err)
	}
	if downloads != 0 || !sink.rolledBack || sink.Len() != 0 {
		t.Fatalf("downloads=%d sink=%#v", downloads, sink)
	}
}

func TestRedirectNeverForwardsBearerToken(t *testing.T) {
	redirectHits := 0
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { redirectHits++ }))
	defer redirectTarget.Close()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, redirectTarget.URL, http.StatusFound)
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	if _, err := client.ListFiles(context.Background(), "folder-a", ""); !errors.Is(err, ErrRejected) {
		t.Fatalf("ListFiles() error = %v, want rejected redirect", err)
	}
	if redirectHits != 0 {
		t.Fatalf("redirect target received %d requests", redirectHits)
	}
}

func TestAccessTokenFailureDoesNotExposeCredentialDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected request") }))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "", errors.New("secret-material-must-not-appear") }))
	_, err := client.ListFiles(context.Background(), "folder-a", "")
	if !errors.Is(err, ErrAuthorization) || strings.Contains(fmt.Sprint(err), "secret-material") {
		t.Fatalf("token error = %v, want sanitized authorization error", err)
	}
}

func TestClientRejectsDuplicateOrUnsafeFoldersAndIDs(t *testing.T) {
	token := AccessTokenFunc(func(context.Context) (string, error) { return "token", nil })
	if _, err := newClient("https://example.com/drive/v3", token, []string{"folder-a", "folder-a", "folder-c"}, nil); !errors.Is(err, ErrConfiguration) {
		t.Fatalf("duplicate folders error = %v", err)
	}
	if _, err := newClient("https://example.com/drive/v3", token, []string{"folder-a", "../folder-b", "folder-c"}, nil); !errors.Is(err, ErrConfiguration) {
		t.Fatalf("unsafe folder error = %v", err)
	}
	if _, err := newClient("http://example.com/drive/v3", token, testFolderIDs, nil); !errors.Is(err, ErrConfiguration) {
		t.Fatalf("non-loopback HTTP error = %v", err)
	}
}

type memorySink struct {
	strings.Builder
	committed   bool
	rolledBack  bool
	rollbackErr error
}

func (s *memorySink) Reset() error {
	s.Builder.Reset()
	s.committed = false
	s.rolledBack = false
	return nil
}

func (s *memorySink) Commit() error {
	s.committed = true
	return nil
}

func (s *memorySink) Rollback() error {
	s.Builder.Reset()
	s.rolledBack = true
	return s.rollbackErr
}

func testClient(t *testing.T, server *httptest.Server, tokens AccessTokenProvider) *Client {
	t.Helper()
	client, err := newClient(server.URL+"/drive/v3", tokens, testFolderIDs, server.Client())
	if err != nil {
		t.Fatalf("newClient() error = %v", err)
	}
	client.wait = func(context.Context, time.Duration) error { return nil }
	client.random = func() float64 { return 0 }
	return client
}
