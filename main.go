// Package main contains a simple backup program.
// TODO: split main into structs: config reader, local exporter, gcs exporter, add tests
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/attwad/backupi/config"
	"github.com/attwad/backupi/gcs"
	"github.com/attwad/backupi/local"
)

var (
	configPath = flag.String("config_path", "./config.yaml", "Path to the config file")
	credsFile  = flag.String("creds_file", "", "Path to the credentials file used to authenticate to Google APIs, if necessary")
)

func main() {
	flag.Parse()

	fmt.Println("Reading config from", *configPath)

	c, err := config.Read(*configPath)
	if err != nil {
		log.Fatal(err)
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
		if err := gcs.Backup(context.Background(), tempOutputFile, c.Dest.GCS.Bucket, *credsFile); err != nil {
			log.Fatal(err)
		}
	}

	// Do this last as it moves the file.
	if c.Dest.LocalDir != nil && c.Dest.LocalDir.Path != "" {
		if err := local.Backup(tempOutputFile, c.Dest.LocalDir.Path); err != nil {
			log.Fatal(err)
		}
	}
}
