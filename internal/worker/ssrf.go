package worker

import (
	"github.com/SirZeck/ping-gopher/internal/validator"
)

var (
	ErrInvalidScheme   = validator.ErrInvalidScheme
	ErrEmptyHost       = validator.ErrEmptyHost
	ErrSSRFForbiddenIP = validator.ErrSSRFForbiddenIP
	ValidateSafeURL    = validator.ValidateSafeURL
)
