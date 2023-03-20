package combine

import (
	"context"
	"fmt"
)

func Run() {
	ctx := context.Background()
	t, _ := combineWallets(ctx)
	fmt.Println(t)
	//CombineStores(ctx)
	//SaveWallets(ctx)
}
