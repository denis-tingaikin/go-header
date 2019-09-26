package provider

import (
	"github.com/go-header/models"
)

//Sources means sources provider
type Sources interface {
	Get() []*models.Source
}
