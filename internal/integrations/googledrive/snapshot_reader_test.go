package googledrive

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestDecodeSnapshotFileValidatesAndMapsMetadata(t *testing.T) {
	metadata := driveFileMetadata{
		ID: "file_123", Name: "pagamentos.csv", MIMEType: "text/csv",
		Size: 42, SizeKnown: true, ModifiedTime: "2026-10-18T12:00:00Z",
		MD5Checksum: "0123456789abcdef0123456789abcdef", Version: "7",
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}

	file, err := decodeSnapshotFile(metadata.ID, encoded)
	if err != nil {
		t.Fatalf("decodeSnapshotFile returned an error: %v", err)
	}
	if file.ID != metadata.ID || file.Name != metadata.Name || file.MIMEType != metadata.MIMEType || file.Size != metadata.Size || !file.SizeKnown || file.ModifiedTime != metadata.ModifiedTime || file.MD5Checksum != metadata.MD5Checksum || file.Version != metadata.Version {
		t.Fatalf("decoded metadata = %#v", file)
	}
}

func TestDecodeSnapshotFileRejectsInvalidMetadata(t *testing.T) {
	base := driveFileMetadata{
		ID: "file_123", Name: "pagamentos.csv", MIMEType: "text/csv",
		Size: 42, SizeKnown: true, ModifiedTime: "2026-10-18T12:00:00Z", Version: "7",
	}
	tests := []struct {
		name     string
		fileID   string
		metadata driveFileMetadata
		encoded  []byte
	}{
		{name: "database id does not match metadata", fileID: "other_file", metadata: base},
		{name: "missing version", fileID: base.ID, metadata: driveFileMetadata{ID: base.ID, Name: base.Name, MIMEType: base.MIMEType, SizeKnown: true, ModifiedTime: base.ModifiedTime}},
		{name: "negative size", fileID: base.ID, metadata: driveFileMetadata{ID: base.ID, Name: base.Name, MIMEType: base.MIMEType, Size: -1, SizeKnown: true, ModifiedTime: base.ModifiedTime, Version: base.Version}},
		{name: "unknown size cannot carry a value", fileID: base.ID, metadata: driveFileMetadata{ID: base.ID, Name: base.Name, MIMEType: base.MIMEType, Size: 1, ModifiedTime: base.ModifiedTime, Version: base.Version}},
		{name: "malformed json", fileID: base.ID, encoded: []byte(`{"id":`)},
		{name: "metadata too large", fileID: base.ID, encoded: []byte(strings.Repeat("x", 16<<10+1))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded := test.encoded
			if encoded == nil {
				var err error
				encoded, err = json.Marshal(test.metadata)
				if err != nil {
					t.Fatal(err)
				}
			}
			if _, err := decodeSnapshotFile(test.fileID, encoded); !errors.Is(err, ErrSnapshotInvalid) {
				t.Fatalf("decodeSnapshotFile error = %v, want %v", err, ErrSnapshotInvalid)
			}
		})
	}
}
