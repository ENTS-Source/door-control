package assets

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func SetupWeb() string {
	webPath, err := os.MkdirTemp(os.TempDir(), "dc-web")
	if err != nil {
		log.Fatal(err)
	}
	extractPrefixTo("web", webPath)
	return webPath
}

func extractPrefixTo(pathName string, destination string) {
	for f, b64 := range f2CompressedFiles {
		if !strings.HasPrefix(f, pathName) {
			continue
		}

		b, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			log.Fatal(err)
		}

		gr, err := gzip.NewReader(bytes.NewBuffer(b))
		if err != nil {
			log.Fatal(err)
		}

		dest := path.Join(destination, filepath.Base(f))
		log.Printf("Writing %s to %s", f, dest)
		file, err := os.Create(dest)
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(file, gr)
		if err != nil {
			log.Fatal(err)
		}

		_ = gr.Close()
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
