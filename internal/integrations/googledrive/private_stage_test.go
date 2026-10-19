package googledrive

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"majucau.local/financial-intelligence/internal/securetemp"
)

func TestWithPrivateStageStreamsVerifiedFileAndRemovesWorkspace(t *testing.T) {
	content := []byte("external financial input\n")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("alt") == "media" {
			_, _ = writer.Write(content)
			return
		}
		_, _ = fmt.Fprintf(writer, `{"id":"file-1","name":"report.csv","mimeType":"text/csv","size":"%d","modifiedTime":"2026-10-18T09:00:00Z","version":7,"parents":["folder-a"]}`, len(content))
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	parent := t.TempDir()
	wantHash := sha256.Sum256(content)
	called := false
	result, err := client.WithPrivateStage(context.Background(), parent, "file-1", 1<<20,
		func(_ context.Context, file File, reader io.Reader, size int64, hash string) error {
			called = true
			got, err := io.ReadAll(reader)
			if err != nil {
				return err
			}
			if string(got) != string(content) {
				t.Fatalf("staged content = %q, want original streamed bytes", got)
			}
			if file.ID != "file-1" || size != int64(len(content)) || hash != hex.EncodeToString(wantHash[:]) {
				t.Fatalf("callback metadata = (%q, %d, %q)", file.ID, size, hash)
			}
			return nil
		})
	if err != nil {
		t.Fatalf("WithPrivateStage() error = %v", err)
	}
	if !called || result.Bytes != int64(len(content)) || result.SHA256 != hex.EncodeToString(wantHash[:]) {
		t.Fatalf("result = %#v, callback called = %v", result, called)
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("private staging workspace remains after callback: %d entries", len(entries))
	}
}

func TestWithPrivateStageRemovesWorkspaceWhenConsumerFails(t *testing.T) {
	content := []byte("external financial input")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("alt") == "media" {
			_, _ = writer.Write(content)
			return
		}
		_, _ = fmt.Fprintf(writer, `{"id":"file-1","name":"report.csv","mimeType":"text/csv","size":"%d","version":7,"parents":["folder-a"]}`, len(content))
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	parent := t.TempDir()
	_, err := client.WithPrivateStage(context.Background(), parent, "file-1", 1<<20,
		func(_ context.Context, _ File, _ io.Reader, _ int64, _ string) error {
			return errors.New("sensitive parser detail")
		})
	if !errors.Is(err, ErrFileVisitor) || strings.Contains(err.Error(), "sensitive parser detail") {
		t.Fatalf("consumer error = %v, want sanitized file-processing failure", err)
	}
	entries, readErr := os.ReadDir(parent)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("workspace cleanup: entries=%d error=%v", len(entries), readErr)
	}
}

func TestWithPrivateStageRetryReplacesPartialDownload(t *testing.T) {
	content := []byte("complete verified source bytes")
	mediaAttempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("alt") == "media" {
			mediaAttempts++
			writer.Header().Set("Content-Length", fmt.Sprint(len(content)))
			if mediaAttempts == 1 {
				_, _ = writer.Write(content[:len(content)/2])
				writer.(http.Flusher).Flush()
				return
			}
			_, _ = writer.Write(content)
			return
		}
		_, _ = fmt.Fprintf(writer, `{"id":"file-1","name":"report.csv","mimeType":"text/csv","size":"%d","version":7,"parents":["folder-a"]}`, len(content))
	}))
	defer server.Close()
	client := testClient(t, server, AccessTokenFunc(func(context.Context) (string, error) { return "test-access-token", nil }))
	parent := t.TempDir()
	callbackCalls := 0
	result, err := client.WithPrivateStage(context.Background(), parent, "file-1", 1<<20,
		func(_ context.Context, _ File, reader io.Reader, size int64, _ string) error {
			callbackCalls++
			got, err := io.ReadAll(reader)
			if err != nil {
				return err
			}
			if string(got) != string(content) || size != int64(len(content)) {
				t.Fatalf("callback saw partial or duplicate bytes: %q (%d bytes)", got, size)
			}
			return nil
		})
	if err != nil {
		t.Fatalf("WithPrivateStage() error = %v", err)
	}
	if mediaAttempts != 2 || callbackCalls != 1 || result.Bytes != int64(len(content)) {
		t.Fatalf("attempts=%d callbacks=%d result=%#v", mediaAttempts, callbackCalls, result)
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 0 {
		t.Fatalf("workspace cleanup after retry: entries=%d error=%v", len(entries), err)
	}
}

func TestPrivateFileSinkDoesNotExceedConfiguredLimitAndRollsBack(t *testing.T) {
	workspace, err := securetemp.NewWorkspace(t.TempDir(), "majucau-drive-stage-")
	if err != nil {
		t.Fatal(err)
	}
	defer workspace.RemoveAll()
	sink, err := newPrivateFileSink(workspace, 3)
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.Reset(); err != nil {
		t.Fatal(err)
	}
	if n, err := sink.Write([]byte("abc")); err != nil || n != 3 {
		t.Fatalf("Write() = (%d, %v), want (3, nil)", n, err)
	}
	if n, err := sink.Write([]byte("d")); n != 0 || !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("oversize Write() = (%d, %v), want limit error without partial write", n, err)
	}
	if err := sink.Rollback(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(workspace.Path())
	if err != nil || len(entries) != 0 {
		t.Fatalf("rollback left staged files: entries=%d error=%v", len(entries), err)
	}
}
