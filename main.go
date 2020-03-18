package main

import (
	"flag"

	"github.com/NamedKitten/kittehbooru/start"
)

var conf = flag.String("conf", "settings.yaml", "config file")

func main() {
	flag.Parse()
	start.Start(*conf)
}
