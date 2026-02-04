package checksum

import (
	"fmt"
	"hash/adler32"
	"io"
	"os"
)

func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath) // #nosec G304
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hasher := adler32.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("adler32:%08x", hasher.Sum32()), nil
}

func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
