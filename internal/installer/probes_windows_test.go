//go:build windows

package installer

import "testing"

func TestSafeLocationRejectsNetworkAndSynchronizedRoots(t *testing.T) {
	for _, path := range []string{
		`\\server\share\Majucau`,
		`C:\Users\Gisele\OneDrive - Company\Majucau`,
		`C:\Users\Gisele\Dropbox\Majucau`,
		`C:\Users\Gisele\SharePoint\Majucau`,
	} {
		if safeLocation(path) {
			t.Fatalf("safeLocation(%q) = true; want false", path)
		}
	}
}

func TestSafeLocationAcceptsLocalAbsolutePath(t *testing.T) {
	if !safeLocation(`C:\ProgramData\Majucau`) {
		t.Fatal("local absolute path should be accepted")
	}
}
