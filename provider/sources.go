package provider

import (
	"github.com/denis-tingajkin/go-header/models"
)

//Sources means sources provider
type Sources interface {
	Get() []*models.Source
}
