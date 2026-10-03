//go:build !windows

package backup

import (
	"errors"
	"io"
)

func decryptV1Stream(_ io.Reader, _ int64, _, _ []byte, _ io.Writer) error {
	return errors.New("V1 backup restore requires Windows CNG")
}
