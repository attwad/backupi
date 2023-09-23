package local

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackup(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()
	input := filepath.Join(tmpDir1, "some.input")
	if err := os.WriteFile(input, []byte("foo"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := Backup(input, tmpDir2); err != nil {
		t.Errorf("Backup() error = %v", err)
	}
}
