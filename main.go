package main

import "github.com/NamedKitten/kittehimageboard/start"
import "os"
import "runtime/pprof"

func main() {
	f, _ := os.Create("cpu.prof")

	defer f.Close()
	pprof.StartCPUProfile(f)
	start.Start()
	pprof.StopCPUProfile()
}
