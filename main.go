// Package main contains a simple backup program.
// TODO: split main into structs: config reader, local exporter, gcs exporter, add tests
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"gopkg.in/yaml.v3"
)

type config struct {
	Path []string    `yaml:"path"`
	Dest destination `yaml:"dest"`
}

type destination struct {
	LocalDir *localDest `yaml:"localDir,omitempty"`
	GCS      *gcs       `yaml:"gcs,omitempty"`
}

type localDest struct {
	// Path is the absolute path to a directory in which to store the backup.tar file.
	Path string `yaml:"path"`
}

type gcs struct {
	// Bucket is the bucket in which to put the backup.tar file.
	Bucket string `yaml:"bucket"`
}

var (
	configPath = flag.String("config_path", "./config.yaml", "Path to the config file")
	credsFile  = flag.String("creds_file", "", "Path to the credentials file used to authenticate to Google APIs, if necessary")
)

func main() {
	flag.Parse()

	fmt.Println("Reading config from", *configPath)

	var c config
	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("reading %s: %v", *configPath, err)
	}
	fmt.Printf("config:\n%s\n", string(data))
	if err := yaml.Unmarshal(data, &c); err != nil {
		log.Fatalf("unmarshaling config yaml: %v", err)
	}

	dir, err := os.MkdirTemp("", "backupi")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fmt.Println("Will write to", dir)
	tempOutputFile := filepath.Join(dir, "backup.tar")

	for i, path := range c.Path {
		var arg string
		if i == 0 {
			// The first time we create the archive.
			arg = "c" // Create
		} else {
			// After the first time, we append files to it.
			arg = "r" // Append
		}
		cmd := exec.Command("tar", fmt.Sprintf("-%sf", arg), tempOutputFile, path)
		fmt.Println("Executing", cmd.String())
		b, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("executing %s: %v, output: %s", cmd.String(), err, string(b))
		}
		fmt.Println("Done adding", path, "to the archive")
	}
	// Now gzip the tar file.
	cmd := exec.Command("gzip", tempOutputFile)
	fmt.Println("Executing", cmd.String())
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("executing %s: %v, output: %s", cmd.String(), err, string(b))
	}
	fmt.Println("Done gzipping", tempOutputFile)
	tempOutputFile = tempOutputFile + ".gz"

	if c.Dest.GCS != nil && c.Dest.GCS.Bucket != "" {
		fmt.Println("Uploading to GCS at", c.Dest.GCS.Bucket)
		ctx := context.Background()

		file, err := os.Open(tempOutputFile)
		if err != nil {
			log.Fatalf("opening %s: %v", tempOutputFile, err)
		}

		srv, err := storage.NewClient(ctx, option.WithCredentialsFile(*credsFile))
		if err != nil {
			log.Fatalf("creating GCS client: %v", err)
		}

		wc := srv.Bucket(c.Dest.GCS.Bucket).Object("backup.tar.gz").NewWriter(ctx)
		if _, err := io.Copy(wc, file); err != nil {
			log.Fatalf("io.Copy: %v", err)
		}
		if err := wc.Close(); err != nil {
			log.Fatalf("Writer.Close: %v", err)
		}
		fmt.Println("Copied to GCS under name:", wc.Name)
	}

	// Do this last as it moves the file.
	if c.Dest.LocalDir != nil && c.Dest.LocalDir.Path != "" {
		finalFile := filepath.Join(c.Dest.LocalDir.Path, tempOutputFile)
		fmt.Println("Writing to", finalFile)
		if err := os.Rename(tempOutputFile, finalFile); err != nil {
			log.Fatalf("Moving %s to %s: %v", tempOutputFile, finalFile, err)
		}
		fmt.Println("Done")
	}
}
