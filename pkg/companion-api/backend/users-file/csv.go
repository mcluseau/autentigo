package usersfile

import (
	"encoding/csv"
	"os"
	"sync"

	uuid "github.com/nu7hatch/gouuid"
)

type usersFileWriter struct {
	mutex       *sync.Mutex
	writer      *csv.Writer
	tmpFileName string
	tmpFile     *os.File
}

func newUsersFileWriter() (ufw *usersFileWriter, err error) {
	var f *os.File
	tmpFileName := tmpFileName()

	f, err = os.Create(tmpFileName)
	if err != nil {
		return
	}

	writer := csv.NewWriter(f)
	writer.Comma = ':'

	ufw = &usersFileWriter{
		writer:      writer,
		mutex:       &sync.Mutex{},
		tmpFileName: tmpFileName,
		tmpFile:     f,
	}
	return
}

func (ufw *usersFileWriter) write(record []string) {
	ufw.mutex.Lock()
	ufw.writer.Write(record)
	ufw.mutex.Unlock()
}

func (ufw *usersFileWriter) flush() {
	ufw.mutex.Lock()
	ufw.writer.Flush()
	ufw.mutex.Unlock()
}

func (ufw *usersFileWriter) save(filePath string) (err error) {
	ufw.flush()
	ufw.tmpFile.Close()

	if err = os.Remove(filePath); err != nil {
		return
	}

	err = os.Rename(ufw.tmpFileName, filePath)
	return
}

type usersFileReader struct {
	reader *csv.Reader
	file   *os.File
}

func newUsersFileReader(fileName string) (ufr *usersFileReader, err error) {
	var f *os.File

	f, err = os.Open(fileName)
	if err != nil {
		return
	}

	reader := csv.NewReader(f)
	reader.Comma = ':'

	ufr = &usersFileReader{
		reader: reader,
		file:   f,
	}
	return
}

func (ufr *usersFileReader) read() (record []string, err error) {
	return ufr.reader.Read()
}

func (ufr *usersFileReader) close() (err error) {
	return ufr.file.Close()
}

func tmpFileName() (fileName string) {
	uuid, err := uuid.NewV4()
	if err == nil {
		fileName = uuid.String()
	} else {
		fileName = "autorizo-companion-api-tmp.csv"
	}
	return
}
