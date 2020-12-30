package main

import (
	"fmt"

	"github.com/vs-ude/tinyfyr/internal/config"
)

func main() {
	fmt.Println(config.EncodedPlatformName())
}
