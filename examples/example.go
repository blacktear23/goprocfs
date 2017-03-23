package main

import (
	"fmt"
	"log"
	"os"
	"time"

	gpfs "github.com/blacktear23/goprocfs"
)

var Data string = "Default"

func main() {
	fp, _ := os.Open("/bin/ls")
	defer fp.Close()

	pfs := gpfs.NewProcFS()

	pfs.RegisterReadOnlyFile("datetime", 0444,
		func() []byte {
			fmt.Println("datetime file OnRead")
			now := time.Now()
			return []byte(fmt.Sprintf("%v\n", now))
		},
	)

	pfs.RegisterFile("readwrite", 0666,
		func() []byte {
			fmt.Println("readwrite file OnRead")
			return []byte(Data)
		},
		func(buf []byte) {
			fmt.Println("readwrite file OnWrite")
			fmt.Println(buf)
			Data = string(buf)
		},
	)

	err := pfs.Mount("/tmp/dumps", nil)
	if err != nil {
		log.Fatal(err)
	}
	pfs.Serve()
}
