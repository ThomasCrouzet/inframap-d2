package render

import (
	"github.com/ThomasCrouzet/inframap-d2/internal/config"
	"github.com/ThomasCrouzet/inframap-d2/internal/model"
)

// Renderer defines the interface for diagram generators.
type Renderer interface {
	Render(infra *model.Infrastructure, cfg *config.Config) string
}

// RenderD2 generates a D2 diagram from infrastructure data.
func RenderD2(infra *model.Infrastructure, cfg *config.Config) string {
	r := &D2Renderer{
		DetailLevel: cfg.Render.DetailLevel,
	}
	return r.Render(infra, cfg)
}
