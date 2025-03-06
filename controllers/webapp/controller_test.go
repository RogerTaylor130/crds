package webapp

import (
	"context"
	clientTools "crds/tools/client"
	"fmt"
	kubeinformers "k8s.io/client-go/informers"
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

func TestSharedInformerFactory(t *testing.T) {
	officialClient := clientTools.GetOfficialClientSet()

	officialInformerFactory := kubeinformers.NewSharedInformerFactory(officialClient, 0)

	//ctx := context.TODO()

	// NOTE
	// Informer() func will call factory's InformerFor() func that will:
	// returns the SharedIndexInformer for obj using an internal and
	// append the SharedIndexInformer to factory's informers(map[reflect.Type]cache.SharedIndexInformer)
	officialInformerFactory.Core().V1().Pods().Informer()
	officialInformerFactory.Apps().V1().Deployments().Informer()

	officialInformerFactory.Start(make(<-chan struct{}))

	fmt.Println("Waiting for cache to sync")

	res := officialInformerFactory.WaitForCacheSync(make(<-chan struct{}))

	//if ok := cache.WaitForCacheSync(ctx.Done(), officialInformerFactory.Core().V1().Pods().Informer().HasSynced); !ok {
	//	fmt.Errorf("failed to wait for caches to sync")
	//}
	fmt.Println("Cache synced")
	fmt.Println(res)

	//informers := make(map[reflect.Type]cache.SharedIndexInformer)
	//for informerType, informer := range informers {
	//	fmt.Println(informerType)
	//	fmt.Println(informer)
	//}

}
