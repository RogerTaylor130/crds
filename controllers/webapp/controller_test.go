package webapp

import (
	"context"
	"crds/pkg/generated/clientset/versioned/fake"
	"testing"
)

func TestFakeClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := fake.NewSimpleClientset()
	client
}
