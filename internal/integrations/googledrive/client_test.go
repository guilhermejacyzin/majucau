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
