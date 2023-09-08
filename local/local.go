package local

import (
	"fmt"
	"os"
	"path/filepath"
)

func Backup(input string, dest string) error {
	finalFile := filepath.Join(dest, filepath.Base(input))
	fmt.Println("Writing to", finalFile)
	if err := os.Rename(input, finalFile); err != nil {
		return fmt.Errorf("moving %s to %s: %w", input, finalFile, err)
	}
	fmt.Println("Done")
	return nil
}
