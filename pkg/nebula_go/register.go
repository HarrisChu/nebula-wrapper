package nebula_go

import (
	"github.com/harrischu/nebula-wrapper/pkg/nebula_go/driver"
	nebula2_6 "github.com/harrischu/nebula-wrapper/pkg/nebula_go/driver/nebula2_6"
	nebula3_0 "github.com/harrischu/nebula-wrapper/pkg/nebula_go/driver/nebula3_0"
)

var drivers map[string]driver.Driver

func Register(version string, d driver.Driver) {
	if drivers == nil {
		drivers = make(map[string]driver.Driver)
	}
	drivers[version] = d
}

func init() {
	Register("2.6", &nebula2_6.Driver{})
	Register("3.0", &nebula3_0.Driver{})
}
