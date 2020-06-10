package postflight

import (
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
)

type PostflightOptions struct {
	Verbose bool
}

type QliksensePostflight struct {
	Q  *qliksense.Qliksense
	P  *PostflightOptions
	CG *api.ClientGoUtils
}
