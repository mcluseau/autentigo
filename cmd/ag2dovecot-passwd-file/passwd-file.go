package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

func loadFile() {
	in, err := os.Open(*passwdFile)
	if err != nil {
		log.Fatal("failed to read file: ", err)
	}

	defer in.Close()

	scan := bufio.NewScanner(in)
	for scan.Scan() {
		line := strings.SplitN(scan.Text(), ":", 2)
		if len(line) != 2 {
			log.Fatal("bad file format, it should be <user>:<password>")
		}

		values[line[0]] = line[1]
	}
}

func save(rev int64) {
	// get keys in order
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	buf := &bytes.Buffer{}
	for _, key := range keys {
		fmt.Fprintf(buf, "%s:%s\n", key, values[key])
	}

	// write new file
	newPath := *passwdFile + ".new"

	if err := ioutil.WriteFile(newPath, buf.Bytes(), 0600); err != nil {
		log.Fatal("failed to write new file: ", err)
	}

	// replace current file
	if err := os.Rename(newPath, *passwdFile); err != nil {
		log.Fatal("failed to replace current file: ", err)
	}

	// save state
	saveState(rev)
}
