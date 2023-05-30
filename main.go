package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/seb-sep/sha3-go/sha3go"
)

const (
	cpuProfFile = "./sha3_cpu.prof"
	memProfFile = "./sha3_mem.prof"
)

func main() {
	var cpu, mem, print bool
	var input string
	flag.BoolVar(&cpu, "cpu", false, "Enable CPU profiling")
	flag.BoolVar(&mem, "mem", false, "Enable memory profiling")
	flag.BoolVar(&print, "print", false, "Print hashing outputs to stdout")
	flag.StringVar(&input, "input", "one-way hashing function", "Input for the SHA-3 algorithm to hash")

	flag.Parse()

	if cpu {
		f, err := os.Create(cpuProfFile)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if mem {
		f, err := os.Create(memProfFile)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		defer f.Close()
		defer pprof.WriteHeapProfile(f)
	}

	{
		if print {
			for i := 0; i < 1000; i++ {
				fmt.Println(string(sha3go.SHA3256([]byte(input))))
			}
		} else {
			for i := 0; i < 1000; i++ {
				sha3go.SHA3256([]byte(input))
			}
		}
	}
}
