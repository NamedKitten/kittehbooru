package main

import (
	"github.com/NamedKitten/kittehbooru/start"
	"os"
	"os/signal"
	"flag"
)

var conf = flag.String("conf", "settings.yaml", "config file")

func main() {
	flag.Parse()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	start.Start(*conf, c)
}
