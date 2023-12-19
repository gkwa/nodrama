package main

import (
	"os"

	"github.com/taylormonacelli/nodrama"
)

func main() {
	code := nodrama.Execute()
	os.Exit(code)
}
