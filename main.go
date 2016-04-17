package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/mcarrowd/oneclogbeat/beater"
)

var Name = "oneclogbeat"

func main() {
	if err := beat.Run(Name, "", beater.New()); err != nil {
		os.Exit(1)
	}
}
