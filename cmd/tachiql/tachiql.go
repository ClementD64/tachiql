package main

import (
	"flag"

	"github.com/clementd64/tachiql/pkg/tachiql/server"
)

func main() {
	backupDir := flag.String("backup-dir", "", "Backup directory")
	backupFile := flag.String("backup", "", "Backup file")
	bind := flag.String("bind", ":8080", "Listen Port")
	queryPath := flag.String("query-path", "/", "Resolve query at this path")
	thumbnailDir := flag.String("thumbnail", "./thumbnail", "Thumbnail download path")
	thumbnailPath := flag.String("thumbnail-path", "", "Serve thumbnail at this path")

	flag.Parse()

	server := server.New(server.Config{
		BackupDir:     *backupDir,
		BackupFile:    *backupFile,
		Bind:          *bind,
		QueryPath:     *queryPath,
		ThumbnailDir:  *thumbnailDir,
		ThumbnailPath: *thumbnailPath,
	})

	server.Run()
}
