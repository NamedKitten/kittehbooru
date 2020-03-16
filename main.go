package main

import (
	"github.com/NamedKitten/kittehimageboard/start"
	"os"
	"os/signal"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	start.Start("settings.yaml", c)
}
