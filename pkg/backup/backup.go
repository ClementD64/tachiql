package backup

//go:generate deno run --allow-net=api.github.com,raw.githubusercontent.com --allow-write ../../script/schema.ts tachiyomi.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go_opt=Mtachiyomi.proto=github.com/clementd64/tachiql/pkg/backup tachiyomi.proto

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"google.golang.org/protobuf/proto"
)

func LoadBackup(filename string) (*Backup, error) {
	in, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	file, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, file); err != nil {
		return nil, err
	}

	backup := &Backup{}
	if err := proto.Unmarshal(buffer.Bytes(), backup); err != nil {
		return nil, err
	}

	return backup, nil
}

func LoadFromDirectory(dirname string) (*Backup, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	filename := ""

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".proto.gz") {
			continue
		}

		if file.Name() > filename {
			filename = file.Name()
		}
	}

	if filename != "" {
		return LoadBackup(path.Join(dirname, filename))
	}

	return nil, errors.New("no backup found")
}
