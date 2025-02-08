package webapp

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"testing"
)

//func TestFakeClient(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	client := fake.NewSimpleClientset()
//	client
//}

func TestIsEmpty(t *testing.T) {
	var args []string
	fmt.Println(args == nil)

}

func TestKLog(t *testing.T) {
	logger := klog.FromContext(context.TODO())
	logger.Info("Hello World")
}
