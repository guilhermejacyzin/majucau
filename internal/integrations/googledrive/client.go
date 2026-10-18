// Package googledrive provides a read-only, bounded Google Drive API adapter.
package googledrive

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/httpretry"
)

const (
	APIBaseURL                    = "https://www.googleapis.com/drive/v3"
	approvedFolderCount           = 3
	maxPageSize                   = 100
	maxPageTokenBytes             = 4096
	maxFileWalkPagesPerCall       = 10000
	maxTokenBytes                 = 16 << 10
	maxJSONBytes            int64 = 2 << 20
	requestTimeout                = 2 * time.Minute
)

var (
	ErrConfiguration   = errors.New("Google Drive integration is not configured")
	ErrAuthorization   = errors.New("Google Drive authorization is unavailable")
	ErrRejected        = errors.New("Google Drive rejected the request")
	ErrRateLimited     = errors.New("Google Drive rate limit reached")
	ErrUnavailable     = errors.New("Google Drive is temporarily unavailable")
	ErrInvalidResponse = errors.New("Google Drive returned an invalid response")
	ErrFileNotAllowed  = errors.New("Google Drive file is outside the approved folders")
	ErrFileChanged     = errors.New("Google Drive file changed while it was being read")
	ErrFileTooLarge    = errors.New("Google Drive file exceeds the configured size limit")
	ErrFileWalkLimit   = errors.New("Google Drive file listing reached the per-call page limit")
	ErrFileVisitor     = errors.New("Google Drive file processing failed")
	ErrUnsafeFileType  = errors.New("Google Drive file type is not supported for direct download")
	ErrSink            = errors.New("Google Drive staging operation failed")
)

// AccessTokenProvider supplies a short-lived Google access token without
// exposing how the credential is stored or refreshed. Production wiring must
// use the approved protected secret store; tests may inject a fixed token.
type AccessTokenProvider interface {
	AccessToken(context.Context) (string, error)
}

// AccessTokenFunc adapts a function to AccessTokenProvider.
type AccessTokenFunc func(context.Context) (string, error)

func (f AccessTokenFunc) AccessToken(ctx context.Context) (string, error) {
	return f(ctx)
}

// File contains only metadata required to identify and safely stream a Drive
// file. Names are treated as untrusted display data and are never used as paths.
type File struct {
	ID           string
	Name         string
	MIMEType     string
	Size         int64
	SizeKnown    bool
	ModifiedTime string
	MD5Checksum  string
	Version      string
	Trashed      bool
	Parents      []string
}

type FilePage struct {
	Files         []File
	NextPageToken string
}

// FileWalkResult reports progress at a page boundary. NextPageToken is safe to
// persist as a resume point only after a whole page was visited successfully.
// If a visitor fails mid-page, it points back to that page, so previously
// visited files may be delivered again and must be handled idempotently.
type FileWalkResult struct {
	PagesRead     int
	FilesVisited  int64
	NextPageToken string
	Complete      bool
}

// StagingSink must discard uncommitted bytes on Reset and Rollback. Commit is
// called only after size, checksum, version, and final metadata are verified.
// Implementations should write to a private temporary file or database staging
// transaction; they must not publish bytes from Write before Commit.
type StagingSink interface {
	io.Writer
	Reset() error
	Commit() error
	Rollback() error
}

type StreamResult struct {
	File   File
	Bytes  int64
	SHA256 string
}

type Client struct {
	apiBase          *url.URL
	httpClient       *http.Client
	tokens           AccessTokenProvider
	allowedFolderIDs map[string]struct{}
	wait             func(context.Context, time.Duration) error
	random           func() float64
}

// NewClient creates a production client restricted to exactly three approved
// folder IDs. It intentionally does not discover folders or access Drive root.
func NewClient(tokens AccessTokenProvider, folderIDs []string) (*Client, error) {
	return newClient(APIBaseURL, tokens, folderIDs, nil)
}

func newClient(baseURL string, tokens AccessTokenProvider, folderIDs []string, httpClient *http.Client) (*Client, error) {
	base, err := url.Parse(baseURL)
	if err != nil || base == nil || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" || !safeBaseURLScheme(base) {
		return nil, ErrConfiguration
	}
	if tokens == nil || len(folderIDs) != approvedFolderCount {
		return nil, ErrConfiguration
	}
	allowed := make(map[string]struct{}, len(folderIDs))
	for _, folderID := range folderIDs {
		if !validDriveID(folderID) {
			return nil, ErrConfiguration
		}
		if _, exists := allowed[folderID]; exists {
			return nil, ErrConfiguration
		}
		allowed[folderID] = struct{}{}
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: requestTimeout}
	}
	// Never forward a bearer token to a redirect target, even if a caller
	// supplies a client with a permissive redirect policy.
	clientCopy := *httpClient
	clientCopy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	base.Path = strings.TrimRight(base.Path, "/")
	return &Client{
		apiBase:          base,
		httpClient:       &clientCopy,
		tokens:           tokens,
		allowedFolderIDs: allowed,
		wait:             httpretry.Wait,
		random:           nil,
	}, nil
}

// ListFiles reads one bounded page from one of the three approved folders.
// The caller owns pagination and must impose a maximum number of pages.
func (c *Client) ListFiles(ctx context.Context, folderID, pageToken string) (FilePage, error) {
	if err := c.validateFolder(folderID); err != nil {
		return FilePage{}, err
	}
	if len(pageToken) > maxPageTokenBytes {
		return FilePage{}, ErrConfiguration
	}
	endpoint := c.endpoint("files")
	query := endpoint.Query()
	query.Set("q", "'"+folderID+"' in parents and trashed = false")
	query.Set("pageSize", strconv.Itoa(maxPageSize))
	query.Set("supportsAllDrives", "true")
	query.Set("includeItemsFromAllDrives", "true")
	query.Set("fields", "nextPageToken,files(id,name,mimeType,size,modifiedTime,md5Checksum,version,trashed,parents)")
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	endpoint.RawQuery = query.Encode()

	var wire listFilesResponse
	err := c.withGETRetry(ctx, func() error {
		wire = listFilesResponse{}
		return c.getJSONAttempt(ctx, endpoint, &wire)
	})
	if err != nil {
		return FilePage{}, err
	}
	if len(wire.Files) > maxPageSize || len(wire.NextPageToken) > maxPageTokenBytes {
		return FilePage{}, ErrInvalidResponse
	}
	page := FilePage{NextPageToken: wire.NextPageToken, Files: make([]File, 0, len(wire.Files))}
	for _, item := range wire.Files {
		file, parseErr := parseFile(item)
		if parseErr != nil || file.Trashed || !hasParent(file.Parents, folderID) {
			return FilePage{}, ErrInvalidResponse
		}
		page.Files = append(page.Files, file)
	}
	return page, nil
}

// WalkFiles delivers files one at a time and retains at most one bounded API
// page in memory. The caller supplies a per-call page limit and may resume a
// later call with result.NextPageToken. The total number of files is not capped
// across calls. The visitor's writes must be transactional/idempotent because
// a retry after a mid-page error can deliver earlier files from that page
// again.
func (c *Client) WalkFiles(ctx context.Context, folderID, startPageToken string, maxPages int, visit func(File) error) (result FileWalkResult, returnErr error) {
	if ctx == nil || visit == nil || maxPages <= 0 || maxPages > maxFileWalkPagesPerCall || len(startPageToken) > maxPageTokenBytes {
		return FileWalkResult{}, ErrConfiguration
	}
	if err := c.validateFolder(folderID); err != nil {
		return FileWalkResult{}, err
	}
	pageToken := startPageToken
	result.NextPageToken = pageToken
	seenTokens := make(map[[sha256.Size]byte]struct{}, maxPages)
	if pageToken != "" {
		seenTokens[sha256.Sum256([]byte(pageToken))] = struct{}{}
	}
	for result.PagesRead < maxPages {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		page, err := c.ListFiles(ctx, folderID, pageToken)
		if err != nil {
			return result, err
		}
		for _, file := range page.Files {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			if err := visit(file); err != nil {
				if ctx.Err() != nil {
					return result, ctx.Err()
				}
				return result, &fileVisitFailure{cause: err}
			}
			result.FilesVisited++
		}
		result.PagesRead++
		if page.NextPageToken == "" {
			result.NextPageToken = ""
			result.Complete = true
			return result, nil
		}
		nextTokenHash := sha256.Sum256([]byte(page.NextPageToken))
		if _, exists := seenTokens[nextTokenHash]; exists {
			result.NextPageToken = pageToken
			return result, ErrInvalidResponse
		}
		seenTokens[nextTokenHash] = struct{}{}
		pageToken = page.NextPageToken
		result.NextPageToken = pageToken
	}
	return result, ErrFileWalkLimit
}

type fileVisitFailure struct{ cause error }

func (*fileVisitFailure) Error() string        { return ErrFileVisitor.Error() }
func (e *fileVisitFailure) Unwrap() error      { return e.cause }
func (*fileVisitFailure) Is(target error) bool { return target == ErrFileVisitor }

// StreamFile downloads a Drive binary file into an uncommitted staging sink.
// The file is read in bounded chunks, never retained in memory, and the sink is
// committed only after the source version and checksum are verified. On any
// failure the sink is rolled back; a cleanup error is returned to the caller.
func (c *Client) StreamFile(ctx context.Context, fileID string, maxBytes int64, sink StagingSink) (result StreamResult, returnErr error) {
	if c == nil || c.apiBase == nil || c.httpClient == nil || c.tokens == nil || !validDriveID(fileID) || maxBytes <= 0 || sink == nil {
		return StreamResult{}, ErrConfiguration
	}
	var committed bool
	defer func() {
		if committed {
			return
		}
		if cleanupErr := sink.Rollback(); cleanupErr != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("%w: rollback staging data", ErrSink), cleanupErr)
		}
	}()

	fileURL := c.endpoint("files", fileID)
	fileURL.RawQuery = "fields=id,name,mimeType,size,modifiedTime,md5Checksum,version,trashed,parents"
	var downloaded StreamResult
	err := c.withGETRetry(ctx, func() error {
		if resetErr := sink.Reset(); resetErr != nil {
			return &sinkFailure{cause: resetErr}
		}
		before, metadataErr := c.getMetadata(ctx, fileURL)
		if metadataErr != nil {
			return metadataErr
		}
		if !c.fileAllowed(before) {
			return ErrFileNotAllowed
		}
		if strings.HasPrefix(before.MIMEType, "application/vnd.google-apps.") {
			return ErrUnsafeFileType
		}
		if !before.SizeKnown || before.Size < 0 {
			return ErrInvalidResponse
		}
		if before.Size > maxBytes {
			return ErrFileTooLarge
		}

		downloadURL := c.endpoint("files", fileID)
		downloadURL.RawQuery = "alt=media"
		response, requestErr := c.doGETAttempt(ctx, downloadURL)
		if requestErr != nil {
			return requestErr
		}
		defer response.Body.Close()
		sha := sha256.New()
		md5Hash := md5.New()
		writer := io.MultiWriter(stagingWriter{sink: sink}, sha, md5Hash)
		written, copyErr := io.CopyN(writer, response.Body, before.Size)
		if copyErr != nil {
			if isRetryableTransport(copyErr) {
				return &transportFailure{cause: copyErr}
			}
			return ErrInvalidResponse
		}
		if written != before.Size {
			return ErrInvalidResponse
		}
		var extra [1]byte
		count, readErr := response.Body.Read(extra[:])
		if count != 0 {
			return ErrInvalidResponse
		}
		if readErr != io.EOF {
			if isRetryableTransport(readErr) {
				return &transportFailure{cause: readErr}
			}
			return ErrInvalidResponse
		}

		after, metadataErr := c.getMetadata(ctx, fileURL)
		if metadataErr != nil {
			return metadataErr
		}
		if !sameVersion(before, after) {
			return ErrFileChanged
		}
		if before.MD5Checksum != "" && !strings.EqualFold(before.MD5Checksum, hex.EncodeToString(md5Hash.Sum(nil))) {
			return ErrInvalidResponse
		}
		downloaded = StreamResult{File: before, Bytes: written, SHA256: hex.EncodeToString(sha.Sum(nil))}
		return nil
	})
	if err != nil {
		return StreamResult{}, err
	}
	if err := sink.Commit(); err != nil {
		return StreamResult{}, fmt.Errorf("%w: commit temporary file", ErrSink)
	}
	committed = true
	return downloaded, nil
}

func (c *Client) getMetadata(ctx context.Context, endpoint *url.URL) (File, error) {
	var wire fileWire
	if err := c.withGETRetry(ctx, func() error {
		wire = fileWire{}
		return c.getJSONAttempt(ctx, endpoint, &wire)
	}); err != nil {
		return File{}, err
	}
	file, err := parseFile(wire)
	if err != nil || file.Trashed {
		return File{}, ErrInvalidResponse
	}
	return file, nil
}

func (c *Client) getJSONAttempt(ctx context.Context, endpoint *url.URL, target any) error {
	response, err := c.doGETAttempt(ctx, endpoint)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := decodeBoundedJSON(response.Body, maxJSONBytes, target); err != nil {
		if isRetryableTransport(err) {
			return &transportFailure{cause: err}
		}
		return ErrInvalidResponse
	}
	return nil
}

func (c *Client) doGETAttempt(ctx context.Context, endpoint *url.URL) (*http.Response, error) {
	if ctx == nil || ctx.Err() != nil {
		if ctx == nil {
			return nil, ErrConfiguration
		}
		return nil, ctx.Err()
	}
	token, err := c.tokens.AccessToken(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrAuthorization
	}
	if !validToken(token) {
		return nil, ErrAuthorization
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, ErrConfiguration
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &transportFailure{cause: err}
	}
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return response, nil
	}
	statusErr := &apiFailure{status: response.StatusCode, retryAfter: httpretry.ParseRetryAfter(response.Header.Get("Retry-After"), time.Now())}
	_ = response.Body.Close()
	return nil, statusErr
}

func (c *Client) withGETRetry(ctx context.Context, request func() error) error {
	_, err := httpretry.Do(ctx, httpretry.Policy{
		Method:    http.MethodGet,
		Retryable: isRetryable,
		RetryAfter: func(err error) time.Duration {
			var apiErr *apiFailure
			if errors.As(err, &apiErr) {
				return apiErr.retryAfter
			}
			return 0
		},
		Wait:   c.wait,
		Random: c.random,
	}, func() (struct{}, error) {
		if err := request(); err != nil {
			return struct{}{}, err
		}
		return struct{}{}, nil
	})
	return err
}

func (c *Client) endpoint(parts ...string) *url.URL {
	endpoint := *c.apiBase
	segments := []string{strings.TrimRight(endpoint.Path, "/")}
	for _, part := range parts {
		segments = append(segments, url.PathEscape(part))
	}
	endpoint.Path = strings.Join(segments, "/")
	endpoint.RawPath = ""
	return &endpoint
}

func (c *Client) validateFolder(folderID string) error {
	if c == nil || c.allowedFolderIDs == nil {
		return ErrConfiguration
	}
	if _, ok := c.allowedFolderIDs[folderID]; !ok {
		return ErrFileNotAllowed
	}
	return nil
}

func (c *Client) fileAllowed(file File) bool {
	if c == nil || file.Trashed || !validDriveID(file.ID) {
		return false
	}
	for _, parent := range file.Parents {
		if _, ok := c.allowedFolderIDs[parent]; ok {
			return true
		}
	}
	return false
}

type listFilesResponse struct {
	NextPageToken string     `json:"nextPageToken"`
	Files         []fileWire `json:"files"`
}

type fileWire struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	MIMEType     string          `json:"mimeType"`
	Size         string          `json:"size"`
	ModifiedTime string          `json:"modifiedTime"`
	MD5Checksum  string          `json:"md5Checksum"`
	Version      json.RawMessage `json:"version"`
	Trashed      bool            `json:"trashed"`
	Parents      []string        `json:"parents"`
}

func parseFile(wire fileWire) (File, error) {
	if !validDriveID(wire.ID) || wire.MIMEType == "" || len(wire.Name) > 4096 || len(wire.ModifiedTime) > 128 || len(wire.MD5Checksum) > 64 || len(wire.Parents) > approvedFolderCount {
		return File{}, ErrInvalidResponse
	}
	file := File{ID: wire.ID, Name: wire.Name, MIMEType: wire.MIMEType, ModifiedTime: wire.ModifiedTime, MD5Checksum: wire.MD5Checksum, Trashed: wire.Trashed, Parents: append([]string(nil), wire.Parents...)}
	if wire.Size != "" {
		size, err := strconv.ParseInt(wire.Size, 10, 64)
		if err != nil || size < 0 {
			return File{}, ErrInvalidResponse
		}
		file.Size, file.SizeKnown = size, true
	}
	if len(wire.Version) > 0 && string(wire.Version) != "null" {
		version := strings.Trim(string(wire.Version), "\"")
		if len(version) > 32 || version == "" {
			return File{}, ErrInvalidResponse
		}
		if _, err := strconv.ParseUint(version, 10, 64); err != nil {
			return File{}, ErrInvalidResponse
		}
		file.Version = version
	}
	for _, parent := range file.Parents {
		if !validDriveID(parent) {
			return File{}, ErrInvalidResponse
		}
	}
	return file, nil
}

func decodeBoundedJSON(reader io.Reader, limit int64, target any) error {
	if reader == nil || target == nil || limit <= 0 {
		return ErrInvalidResponse
	}
	counted := &countingReader{reader: io.LimitReader(reader, limit+1)}
	decoder := json.NewDecoder(counted)
	if err := decoder.Decode(target); err != nil {
		if counted.count > limit {
			return ErrInvalidResponse
		}
		return err
	}
	var trailing json.RawMessage
	err := decoder.Decode(&trailing)
	if counted.count > limit {
		return ErrInvalidResponse
	}
	if err != io.EOF {
		return ErrInvalidResponse
	}
	return nil
}

type countingReader struct {
	reader io.Reader
	count  int64
}

func (r *countingReader) Read(buffer []byte) (int, error) {
	n, err := r.reader.Read(buffer)
	r.count += int64(n)
	return n, err
}

type apiFailure struct {
	status     int
	retryAfter time.Duration
}

func (e *apiFailure) Error() string {
	switch e.status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrAuthorization.Error()
	case http.StatusTooManyRequests:
		return ErrRateLimited.Error()
	default:
		if e.status >= http.StatusInternalServerError {
			return ErrUnavailable.Error()
		}
		return ErrRejected.Error()
	}
}

func (e *apiFailure) Is(target error) bool {
	switch {
	case target == ErrRateLimited:
		return e.status == http.StatusTooManyRequests
	case target == ErrUnavailable:
		return e.status >= http.StatusInternalServerError && e.status <= 599
	case target == ErrAuthorization:
		return e.status == http.StatusUnauthorized || e.status == http.StatusForbidden
	case target == ErrRejected:
		return e.status != http.StatusUnauthorized && e.status != http.StatusForbidden && e.status != http.StatusTooManyRequests && (e.status < http.StatusInternalServerError || e.status > 599)
	default:
		return false
	}
}

type transportFailure struct{ cause error }

func (e *transportFailure) Error() string { return "Google Drive transport failed" }
func (e *transportFailure) Unwrap() error { return e.cause }

type sinkFailure struct{ cause error }

func (e *sinkFailure) Error() string { return ErrSink.Error() }
func (e *sinkFailure) Unwrap() error { return e.cause }

type stagingWriter struct{ sink io.Writer }

func (w stagingWriter) Write(buffer []byte) (int, error) {
	n, err := w.sink.Write(buffer)
	if err != nil {
		return n, &sinkFailure{cause: err}
	}
	return n, nil
}

func isRetryable(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var apiErr *apiFailure
	if errors.As(err, &apiErr) {
		return apiErr.status == http.StatusRequestTimeout || apiErr.status == http.StatusTooManyRequests || apiErr.status >= 500 && apiErr.status <= 599
	}
	var sinkErr *sinkFailure
	if errors.As(err, &sinkErr) {
		return false
	}
	var transportErr *transportFailure
	return errors.As(err, &transportErr) && isRetryableTransport(transportErr.cause)
}

func isRetryableTransport(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.EPIPE) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsTimeout || dnsErr.IsTemporary
	}
	var netErr net.Error
	return errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary())
}

func validDriveID(value string) bool {
	if value == "" || len(value) > 256 {
		return false
	}
	for _, char := range value {
		if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_' || char == '-') {
			return false
		}
	}
	return true
}

func safeBaseURLScheme(value *url.URL) bool {
	if value.Scheme == "https" {
		return true
	}
	if value.Scheme != "http" {
		return false
	}
	host := value.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func validToken(value string) bool {
	return value != "" && len(value) <= maxTokenBytes && !strings.ContainsAny(value, "\r\n \t")
}

func hasParent(parents []string, target string) bool {
	for _, parent := range parents {
		if parent == target {
			return true
		}
	}
	return false
}

func sameVersion(left, right File) bool {
	return left.ID == right.ID && left.Version != "" && left.Version == right.Version && left.SizeKnown && right.SizeKnown && left.Size == right.Size && left.ModifiedTime == right.ModifiedTime && strings.EqualFold(left.MD5Checksum, right.MD5Checksum) && left.MIMEType == right.MIMEType && hasCommonParent(left.Parents, right.Parents)
}

func hasCommonParent(left, right []string) bool {
	for _, parent := range left {
		if hasParent(right, parent) {
			return true
		}
	}
	return false
}
