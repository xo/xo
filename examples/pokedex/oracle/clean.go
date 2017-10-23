// +build ignore

package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	flagPath = flag.String(
		"path",
		os.Getenv("GOPATH")+"/src/github.com/knq/xo/examples/pokedex/pokedex/pokedex/data",
		"path",
	)
)

func walk(p string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !fi.IsDir() && filepath.Ext(fi.Name()) == ".csv" {
		buf, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		r := csv.NewReader(bytes.NewReader(buf))
		rows, err := r.ReadAll()
		if err != nil {
			return err
		}

		for i, row := range rows {
			for j, s := range row {
				var r []rune
				var skipped bool
				for _, z := range s {
					if z < 128 {
						r = append(r, z)
					} else {
						r = append(r, ' ')
						skipped = true
					}
				}

				s := string(r)
				if y := strings.TrimSpace(string(s)); len(y) == 0 && skipped {
					s = fmt.Sprintf("a-%d-%d", i, j)
				}
				row[j] = s
			}
		}

		f, err := os.OpenFile(p, os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		w := csv.NewWriter(f)
		err = w.WriteAll(rows)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()

	err := filepath.Walk(*flagPath, walk)
	if err != nil {
		log.Fatal(err)
	}
}
