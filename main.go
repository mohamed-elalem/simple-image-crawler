package main

import (
	"os"

	"./crawler"
)

func main() {
	crawler.Run(os.Args[1:])
}
