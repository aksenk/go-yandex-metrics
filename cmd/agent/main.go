package main

import (
	"log"
	"runtime"
)

func main() {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	log.Printf("%+v", m)
}
