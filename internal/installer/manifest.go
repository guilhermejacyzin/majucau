package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	cleanManifestPath, ok := cleanManifestRelativePath(manifestPath)
	if !ok {
		result.add(ManifestIssuePackageInvalid, "", "manifest path is unsafe")
		return result
	}
	manifestBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(cleanManifestPath)))
	if err != nil {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest is unavailable")
		return result
	}
	decoder := json.NewDecoder(strings.NewReader(string(manifestBytes)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result.Manifest); err != nil {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest JSON is invalid")
		return result
	}
	if result.Manifest.ManifestVersion != ReleaseManifestVersion ||
		strings.TrimSpace(result.Manifest.AppVersion) == "" ||
		strings.TrimSpace(result.Manifest.SchemaVersion) == "" {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "manifest identity is incomplete or unsupported")
	}
	if !validSchemaRange(result.Manifest.MinSchema, result.Manifest.MaxSchema) {
		result.add(ManifestIssuePackageInvalid, cleanManifestPath, "schema compatibility range is invalid")
	} else if !schemaInRange(currentSchemaVersion, result.Manifest.MinSchema, result.Manifest.MaxSchema) {
		result.add(ManifestIssueSchemaIncompatible, cleanManifestPath, "current database schema is outside the package compatibility range")
	}

	seen := map[string]struct{}{strings.ToLower(cleanManifestPath): {}}
	validateFiles := func(files []ManifestFile, kind string) {
		if kind == "artifacts" && len(files) == 0 {
			result.add(ManifestIssuePackageInvalid, "", "manifest has no artifacts")
		}
		for _, file := range files {
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
			if err := validateManifestFile(root, cleanPath, file); err != nil {
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

func validSHA256(value string) bool {
	if len(strings.TrimSpace(value)) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimSpace(value))
	return err == nil
}

func validateManifestFile(root, relativePath string, expected ManifestFile) error {
	fullPath := filepath.Join(root, filepath.FromSlash(relativePath))
	info, err := os.Lstat(fullPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return os.ErrNotExist
	}
	payload, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(payload)
	if !strings.EqualFold(hex.EncodeToString(digest[:]), strings.TrimSpace(expected.SHA256)) {
		return os.ErrInvalid
	}
	if expected.SizeBytes != 0 && int64(len(payload)) != expected.SizeBytes {
		return os.ErrInvalid
	}
	return nil
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
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) == 0 {
		return nil, false
	}
	result := make([]int, len(parts))
	for index, part := range parts {
		if part == "" {
			return nil, false
		}
		number := 0
		for _, char := range part {
			if char < '0' || char > '9' {
				return nil, false
			}
			number = number*10 + int(char-'0')
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
