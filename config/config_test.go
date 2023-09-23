package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRead(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(testFile, []byte(`
path:
- /Users/who/Desktop/dev/backupi/test_config.yaml
dest:
  localDir:
    path: /tmp
  gcs:
    bucket: who-backups
  `), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	c, err := Read(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(c, &Config{
		Path: []string{"/Users/who/Desktop/dev/backupi/test_config.yaml"},
		Dest: Destination{
			LocalDir: &LocalDest{
				Path: "/tmp",
			},
			GCS: &GCS{
				Bucket: "who-backups",
			},
		},
	}); diff != "" {
		t.Errorf("diff=\n%s", diff)
	}
}
