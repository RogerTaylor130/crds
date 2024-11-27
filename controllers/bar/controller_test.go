package bar

import (
	clientTools "crds/tools/client"
	"testing"
)

func TestMakeDefaultController(t *testing.T) {
	officalClient := clientTools.GetOfficialClientSet()

	ctx := clientTools.SetupSignalHandler()

	barClient := clientTools.GetExampleClientSet()

	officalInformer := barClient

	controller := NewBarController(ctx, officalClient, barClient)
}
