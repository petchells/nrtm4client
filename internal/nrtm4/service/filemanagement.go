package service

import (
	"bufio"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"

	"gitlab.com/etchells/nrtm4client/internal/nrtm4/jsonseq"
	"gitlab.com/etchells/nrtm4client/internal/nrtm4/persist"
)

var GZIP_SNAPSHOT_EXTENTION = ".gz"

type fileManager struct {
	repo   persist.Repository
	client Client
}

func (fm fileManager) readSnapshotRecords(
	snapshotFile *os.File,
	fn jsonseq.RecordReaderFunc,
) error {

	var err error

	log.Println("DEBUG opening snapshotFile for reading", snapshotFile.Name())
	var reader io.Reader
	if reader, err = os.Open(snapshotFile.Name()); err != nil {
		return err
	}
	bufioReader := new(bufio.Reader)
	if snapshotFile.Name()[len(snapshotFile.Name())-len(GZIP_SNAPSHOT_EXTENTION):] == GZIP_SNAPSHOT_EXTENTION {
		var gzreader *gzip.Reader
		if gzreader, err = gzip.NewReader(reader); err != nil {
			return err
		}
		bufioReader = bufio.NewReader(gzreader)
	} else {
		bufioReader = bufio.NewReader(reader)
	}
	err = jsonseq.ReadRecords(bufioReader, func(bytes []byte, err error) error {
		return fn(bytes, err)
	})
	return err
}

// func (fm fileManager) fileToDatabase(url string, nrtmFile nrtm4model.NrtmFile, filetype persist.NTRMFileType, path string) (*os.File, error) {
// 	var file *os.File
// 	var err error
// 	if file, err = fm.writeResourceToPath(url, path); err != nil {
// 		return file, err
// 	}
// 	// defer func() {
// 	// 	if err := file.Close(); err != nil {
// 	// 		panic(err)
// 	// 	}
// 	// }()
// 	ds := NrtmDataService{fm.repo}
// 	return file, ds.saveState(url, nrtmFile, filetype, file)
// }

func (fm fileManager) writeResourceToPath(url string, path string) (*os.File, error) {
	fileName := filepath.Base(url)
	if f, err := os.Open(filepath.Join(path, fileName)); err == nil {
		return f, err
	}
	var reader io.Reader
	var err error
	if reader, err = fm.client.getResponseBody(url); err != nil {
		log.Println("ERROR Failed to fetch file", url, err)
		return nil, err
	}
	return readerToFile(reader, path, fileName)
}

func readerToFile(reader io.Reader, path string, fileName string) (*os.File, error) {
	var outFile *os.File
	var err error
	if outFile, err = os.Create(filepath.Join(path, fileName)); err != nil {
		log.Println("ERROR Failed to open file on disk", err)
		return nil, err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			panic(err)
		}
	}()
	if err = transferReaderToFile(reader, outFile); err != nil {
		log.Println("ERROR writing file")
		return nil, err
	}
	return outFile, err
}

func transferReaderToFile(from io.Reader, to *os.File) error {
	buf := make([]byte, FILE_BUFFER_LENGTH)
	for {
		n, err := from.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := to.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}
