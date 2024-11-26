package main

import (
	"crds/controllers/bar"
	clientTools "crds/tools/client"
)

func main() {

	officalClient := clientTools.GetOfficialClientSet()

	ctx := clientTools.SetupSignalHandler()

	barClient := clientTools.GetExampleClientSet()

	controller := bar.NewBarController(ctx, officalClient, barClient)

	controller.Run(ctx)

}
