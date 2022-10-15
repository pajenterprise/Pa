// +build linux_bpf

package bytecode

import (
	"fmt"
)

// ReadBPFModule from the asset file
func ReadBPFModule(bpfDir string, debug bool, useCore bool) (AssetReader, error) {
	var file string
	if useCore {
		bpfDir = "./c/co-re"
		file = "pkg/ebpf/c/co-re/tracer.o"
		if debug {
			file = "pkg/ebpf/c/co-re/tracer-debug.o"
		}
	} else {
		bpfDir = "./c"
		file = "pkg/ebpf/c/tracer-ebpf.o"
		if debug {
			file = "pkg/ebpf/c/tracer-ebpf-debug.o"
		}
	}

	ebpfReader, err := GetReader(bpfDir, file)
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}

	return ebpfReader, nil
}

// ReadOffsetBPFModule from the asset file
func ReadOffsetBPFModule(bpfDir string, debug bool) (AssetReader, error) {
	bpfDir = "./c"
	file := "pkg/ebpf/c/offset-guess.o"
	if debug {
		file = "pkg/ebpf/c/offset-guess-debug.o"
	}

	ebpfReader, err := GetReader(bpfDir, file)
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}

	return ebpfReader, nil
}
