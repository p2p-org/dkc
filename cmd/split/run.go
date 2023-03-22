package split

import (
	"context"
)

func Run() {
	ctx := context.Background()
	CreateWallets(ctx)
	//CombineStores(ctx)
	//SaveWallets(ctx)
}
