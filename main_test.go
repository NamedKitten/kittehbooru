package main

import (
	"github.com/NamedKitten/kittehimageboard/start"
	"os"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func Init(configFile string) (chan os.Signal) {
	c := make(chan os.Signal, 1)
	go func() {
		wg.Add(1)
		defer wg.Done()
		start.Start(configFile, c)
	}()
	return c
}

func TestMain(t *testing.T) {
	wg = sync.WaitGroup{}
	c := Init("settings_test.yaml")
	time.Sleep(time.Second)
	c <- os.Kill
	wg.Wait()
}