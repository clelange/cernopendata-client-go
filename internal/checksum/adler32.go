package checksum

import (
	"fmt"
	"hash/adler32"
	"os"
)

func CalculateChecksum(filePath string) (string, error) {
	data, err := os.ReadFile(filePath) // #nosec G304
	if err != nil {
		return "", err
	}

	adler := adler32.Checksum(data)
	return fmt.Sprintf("adler32:%08x", adler), nil
}

func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
