package persistent

import "go.uber.org/fx"

var Module = fx.Module("order.persistence", fx.Provide(NewRepository))
