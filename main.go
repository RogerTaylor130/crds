package main

import (
	"crds/controllers/bar"
	clientTools "crds/tools/client"
)

func main() {

	ctx := clientTools.SetupSignalHandler()
	bar.RunBarController(ctx)
}
