package combine

import (
	"context"
)

func Run() {
	ctx := context.Background()
  combineWallets(ctx)

	//CombineStores(ctx)
	//SaveWallets(ctx)
}
