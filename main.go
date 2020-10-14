package main

import (
	"fmt"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/miquella/ask"
	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/storage"
)

var args struct {
	RsfHost   string `long:"rsf"`
	RsHost    string `long:"rs"`
	AccessKey string `long:"ak"`
	SecretKey string `long:"sk"`
	Bucket    string `long:"bucket"`
	Prefix    string `long:"prefix"`
}

func main() {
	_, err := flags.ParseArgs(&args, os.Args[1:])
	if err != nil {
		return
	}

	mac := auth.New(args.AccessKey, args.SecretKey)
	bucketManager := storage.NewBucketManager(mac, &storage.Config{
		RsHost:  args.RsHost,
		RsfHost: args.RsfHost,
	})
	var (
		allEntries []storage.ListItem
		marker     = ""
		hasNext    = true
	)
	for hasNext {
		var entries []storage.ListItem

		entries, _, marker, hasNext, err = bucketManager.ListFiles(args.Bucket, args.Prefix, "", marker, 1000)
		if err != nil {
			panic(err)
		}
		allEntries = append(allEntries, entries...)
	}

	if len(allEntries) == 0 {
		fmt.Printf("No file with prefix `%s` in bucket `%s` is found\n", args.Prefix, args.Bucket)
		return
	}

	for _, entry := range allEntries {
		fmt.Printf("Rename: %s => %s\n", entry.Key, renameKey(entry.Key))
	}
	answer, err := ask.Ask("Rename all? [y/N]: ")
	if err != nil {
		panic(err)
	}
	if answer != "y" {
		return
	}

	moved := 0
	for _, entry := range allEntries {
		if err = bucketManager.Move(args.Bucket, entry.Key, args.Bucket, renameKey(entry.Key), false); err != nil {
			fmt.Fprintf(os.Stderr, "Rename %s => %s Error: %s\n", entry.Key, renameKey(entry.Key), err)
		} else {
			moved += 1
		}
	}

	fmt.Printf("Done: %d files renamed\n", moved)
}

func renameKey(key string) string {
	return strings.TrimPrefix(key, args.Prefix)
}
