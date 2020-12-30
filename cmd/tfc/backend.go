package main

import (
	"github.com/vs-ude/tinyfyr/internal/backends/backend"
	"github.com/vs-ude/tinyfyr/internal/backends/dummy"
)

/*
	"github.com/vs-ude/fyrlang/internal/backends/c99"
*/

func setupBackend() backend.Backend {
	/*
		if config.BuildTarget().C99 != nil {
			return c99.NewBackend()
		}
	*/
	return dummy.Backend{}
}
