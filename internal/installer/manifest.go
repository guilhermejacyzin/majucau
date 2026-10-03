package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ReleaseManifestVersion identifies the JSON contract for a release/backup
// package manifest. It is intentionally independent from the preflight and
// install-journal schema versions.
const ReleaseManifestVersion = "1"

const defaultReleaseManifestPath = "release-manifest.json"

// Release manifests are control data and must stay small enough to parse with
// a bounded amount of memory, even when a package is untrusted.
const (
	maxReleaseManifestBytes        int64 = 4 << 20
	maxReleaseManifestEntries            = 4096
	maxReleaseManifestPathBytes          = 1024
	maxManifestVersionFieldBytes         = 128
)

const ManifestValidationSchemaVersion = "1.0"

const (
	ManifestIssuePackageInvalid     = "PACKAGE_INVALID"
	ManifestIssueChecksumMismatch   = "CHECKSUM_MISMATCH"
	ManifestIssueSchemaIncompatible = "SCHEMA_INCOMPATIBLE"
)

// ReleaseManifest is the side-effect-free contract used before a package can
// be accepted by a future update/restore flow. The manifest itself is not
// included in Artifacts because including its own digest would be recursive.
type ReleaseManifest struct {
	ManifestVersion string           `json:"manifest_version"`
	AppVersion      string           `json:"app_version"`
	SchemaVersion   string           `json:"schema_version"`
	MinSchema      string           `json:"min_schema_version"`
	MaxSchema      string           `json:"max_schema_version"`
	Migrations     []ManifestFile   `json:"migrations"`
	Artifacts      []ManifestFile   `json:"artifacts"`
}

type ManifestFile struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

type ManifestIssue struct {
	Code   string `json:"code"`
	Path   string `json:"path,omitempty"`
	Detail string `json:"detail"`
}

type ManifestValidation struct {
	Valid  bool             `json:"valid"`
	Issues []ManifestIssue  `json:"issues,omitempty"`
	Manifest ReleaseManifest `json:"manifest"`
}

// ValidateReleaseManifest validates only package contents and compatibility.
// It never writes files, starts services, invokes PostgreSQL or mutates the
// machine. A valid result is a prerequisite for, but is not evidence of, a
// successful backup, restore, update or rollback.
func ValidateReleaseManifest(root, manifestPath, currentSchemaVersion string) ManifestValidation {
	result := ManifestValidation{Valid: false}
	root = strings.TrimSpace(root)
	if root == "" || !filepath.IsAbs(root) {
		result.add(ManifestIssuePackageInvalid, "", "package root must be an absolute directory")
		return result
	}
	rootInfo, err := os.Stat(root)
	if err != nil || !rootInfo.IsDir() {
		result.add(ManifestIssuePackageInvalid, "", "package root is unavailable")
		return result
	}

	manifestPath = strings.TrimSpace(manifestPath)
	if manifestPath == "" {
		manifestPath = defaultReleaseManifestPath
	}
	if len(manifestPath) > maxReleaseManifestPathBytes {
		result.add(ManifestIssuePackageInvalid, "", "manifest path exceeds the supported size")
		return result
	}
	cleanManifestPath, ok := cleanManifestRelativePath(manifestPath)
	if !ok {
		result.add(ManifestIssuePackageInvalid, "", "manifest path is unsafe")
		return result
	}
	packageRoot, err := os.OpenRoot(root)
	if err != nil {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "package root is unavailable")
		return result
	}
	defer packageRoot.Close()
	manifestFile, err := openRegularPackageFile(packageRoot, cleanManifestPath)
	if err != nil {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest is unavailable")
		return result
	}
	defer manifestFile.Close()
	manifestReader := &io.LimitedReader{R: manifestFile, N: maxReleaseManifestBytes + 1}
	decoder := json.NewDecoder(manifestReader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result.Manifest); err != nil {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest JSON is invalid")
		return result
	}
	if manifestReader.N == 0 {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest exceeds the supported size")
		return result
	}
	var trailingData struct{}
	if err := decoder.Decode(&trailingData); err != io.EOF {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest JSON is invalid")
		return result
	}
	if manifestReader.N == 0 {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest exceeds the supported size")
		return result
	}
	if result.Manifest.ManifestVersion != ReleaseManifestVersion ||
		!validManifestVersionText(result.Manifest.AppVersion) ||
		!validManifestVersionText(result.Manifest.SchemaVersion) ||
		!validManifestVersionText(result.Manifest.MinSchema) ||
		!validManifestVersionText(result.Manifest.MaxSchema) {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest identity is incomplete or unsupported")
	}
	if !validSchemaRange(result.Manifest.MinSchema, result.Manifest.MaxSchema) {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "schema compatibility range is invalid")
	} else if !schemaInRange(currentSchemaVersion, result.Manifest.MinSchema, result.Manifest.MaxSchema) {
		result.add(ManifestIssueSchemaIncompatible, cleanManifestPath, "current database schema is outside the package compatibility range")
	}
	if len(result.Manifest.Migrations) > maxReleaseManifestEntries ||
		len(result.Manifest.Artifacts) > maxReleaseManifestEntries-len(result.Manifest.Migrations) {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest contains too many files")
		return result
	}

	seen := map[string]struct{}{strings.ToLower(cleanManifestPath): {}}
	validateFiles := func(files []ManifestFile, kind string) {
		if kind == "artifacts" && len(files) == 0 {
			result.add(ManifestIssuePackageInvalid, "", "manifest has no artifacts")
		}
		for _, file := range files {
			if len(file.Path) > maxReleaseManifestPathBytes {
				result.add(ManifestIssuePackageInvalid, "", "package entry path exceeds the supported size")
				continue
			}
			cleanPath, pathOK := cleanManifestRelativePath(file.Path)
			if !pathOK {
				result.add(ManifestIssuePackageInvalid, "", "package entry path is unsafe or contains protected material")
				continue
			}
			if prohibitedManifestPath(cleanPath) {
				result.add(ManifestIssuePackageInvalid, cleanPath, "package entry path is unsafe or contains protected material")
				continue
			}
			key := strings.ToLower(cleanPath)
			if _, exists := seen[key]; exists {
				result.add(ManifestIssuePackageInvalid, cleanPath, "package entry is duplicated")
				continue
			}
			seen[key] = struct{}{}
			if !validSHA256(file.SHA256) || file.SizeBytes < 0 {
				result.add(ManifestIssuePackageInvalid, cleanPath, "package entry digest or size is invalid")
				continue
			}
			if err := validateManifestFile(packageRoot, cleanPath, file); err != nil {
				result.add(ManifestIssueChecksumMismatch, cleanPath, "package entry does not match its manifest")
			}
		}
	}
	validateFiles(result.Manifest.Migrations, "migrations")
	validateFiles(result.Manifest.Artifacts, "artifacts")

	if len(result.Issues) == 0 {
		result.Valid = true
	}
	return result
}

func (r *ManifestValidation) add(code, entryPath, detail string) {
	r.Issues = append(r.Issues, ManifestIssue{Code: code, Path: entryPath, Detail: detail})
}

func validManifestVersionText(value string) bool {
	return len(value) > 0 && len(value) <= maxManifestVersionFieldBytes && strings.TrimSpace(value) != ""
}

func validSHA256(value string) bool {
	if len(strings.TrimSpace(value)) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimSpace(value))
	return err == nil
}

func validateManifestFile(root *os.Root, relativePath string, expected ManifestFile) error {
	file, err := openRegularPackageFile(root, relativePath)
	if err != nil {
		return err
	}
	defer file.Close()

	digest := sha256.New()
	size, err := io.Copy(digest, file)
	if err != nil {
		return err
	}
	if !strings.EqualFold(hex.EncodeToString(digest.Sum(nil)), strings.TrimSpace(expected.SHA256)) {
		return os.ErrInvalid
	}
	if expected.SizeBytes != 0 && size != expected.SizeBytes {
		return os.ErrInvalid
	}
	return nil
}

func openRegularPackageFile(root *os.Root, relativePath string) (*os.File, error) {
	parts := strings.Split(relativePath, "/")
	if root == nil || len(parts) == 0 {
		return nil, os.ErrNotExist
	}
	current := root
	currentOwned := false
	defer func() {
		if currentOwned {
			_ = current.Close()
		}
	}()
	for _, segment := range parts[:len(parts)-1] {
		info, err := current.Lstat(segment)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, os.ErrNotExist
		}
		next, err := current.OpenRoot(segment)
		if err != nil {
			return nil, err
		}
		if currentOwned {
			if err := current.Close(); err != nil {
				_ = next.Close()
				return nil, err
			}
		}
		current = next
		currentOwned = true
	}

	name := parts[len(parts)-1]
	info, err := current.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, os.ErrNotExist
	}
	file, err := current.Open(name)
	if err != nil {
		return nil, err
	}
	openedInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if !os.SameFile(info, openedInfo) {
		_ = file.Close()
		return nil, os.ErrNotExist
	}
	return file, nil
}

func cleanManifestRelativePath(value string) (string, bool) {
	normalized := strings.TrimSpace(strings.ReplaceAll(value, `\`, "/"))
	if normalized == "" || strings.ContainsRune(normalized, 0) || strings.HasPrefix(normalized, "/") ||
		(len(normalized) >= 2 && normalized[1] == ':') {
		return "", false
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", false
		}
	}
	clean := path.Clean(normalized)
	if clean == "." || clean != normalized {
		return "", false
	}
	return clean, true
}

func prohibitedManifestPath(value string) bool {
	protected := map[string]struct{}{
		".env": {}, "token": {}, "tokens": {}, "secret": {}, "secrets": {},
		"credential": {}, "credentials": {}, "password": {}, "passwords": {},
		"private-key": {}, "private_key": {}, "privatekey": {},
	}
	for _, segment := range strings.Split(strings.ToLower(value), "/") {
		if _, found := protected[segment]; found {
			return true
		}
	}
	return false
}

func validSchemaRange(minimum, maximum string) bool {
	min, minOK := parseSchemaVersion(minimum)
	max, maxOK := parseSchemaVersion(maximum)
	return minOK && maxOK && compareSchemaVersions(min, max) <= 0
}

func schemaInRange(current, minimum, maximum string) bool {
	currentVersion, currentOK := parseSchemaVersion(current)
	min, minOK := parseSchemaVersion(minimum)
	max, maxOK := parseSchemaVersion(maximum)
	return currentOK && minOK && maxOK && compareSchemaVersions(currentVersion, min) >= 0 && compareSchemaVersions(currentVersion, max) <= 0
}

func parseSchemaVersion(value string) ([]int, bool) {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > maxManifestVersionFieldBytes {
		return nil, false
	}
	parts := strings.Split(value, ".")
	if len(parts) == 0 {
		return nil, false
	}
	result := make([]int, len(parts))
	maxInt := int(^uint(0) >> 1)
	for index, part := range parts {
		if part == "" {
			return nil, false
		}
		number := 0
		for _, char := range part {
			if char < '0' || char > '9' {
				return nil, false
			}
			digit := int(char - '0')
			if number > (maxInt-digit)/10 {
				return nil, false
			}
			number = number*10 + digit
		}
		result[index] = number
	}
	return result, true
}

func compareSchemaVersions(left, right []int) int {
	length := len(left)
	if len(right) > length {
		length = len(right)
	}
	for index := 0; index < length; index++ {
		leftValue, rightValue := 0, 0
		if index < len(left) {
			leftValue = left[index]
		}
		if index < len(right) {
			rightValue = right[index]
		}
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}
