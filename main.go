package main

import (
	"crds/controllers/webapp"
	clientTools "crds/tools/client"
	"flag"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	ctx := clientTools.SetupSignalHandler()
	//bar.RunBarController(ctx)
	webapp.RunWebappController(ctx)
}
