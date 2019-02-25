package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	stateFile = flag.String("state-file", "passwd.rev", "Sync state file")
)

func loadState() (rev int64) {
	data, err := ioutil.ReadFile(*stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// no state file, start from scratch
			return
		}
		log.Fatal("failed to read state: ", err)
	}

	fmt.Sscanf(string(data), "%20d", &rev)
	return
}

func saveState(rev int64) {
	data := &bytes.Buffer{}
	fmt.Fprintf(data, "%20d", rev)

	if err := ioutil.WriteFile(*stateFile, data.Bytes(), 0600); err != nil {
		log.Fatal("failed to write state: ", err)
	}
}
