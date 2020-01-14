package qliksense

// Qliksense is the logic behind the qliksense client
type Qliksense struct {
	PorterExe string
}

// New qliksense client, initialized with useful defaults.
func New(porterExe string) *Qliksense {
	return &Qliksense{
		porterExe,
	}
}
