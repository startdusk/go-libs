package orm

import (
	"github.com/startdusk/go-libs/orm/model"

	"github.com/startdusk/go-libs/orm/internal/valuer"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       model.Registry

	mdls []Middleware
}
