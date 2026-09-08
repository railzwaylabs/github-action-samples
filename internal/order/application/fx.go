package application

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"order.application",
	fx.Provide(NewService),
)
