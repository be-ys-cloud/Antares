package main

import (
	"github.com/be-ys-cloud/antares/internal/helpers"
	"github.com/be-ys-cloud/antares/internal/services"
)

func main() {
	if helpers.Configuration.Export {
		services.Export()
	} else {
		services.Import()
	}
}
