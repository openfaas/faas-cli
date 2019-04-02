package platform

//Platform system platform
var Platform string

const (
	//defaultPlatform platform x86_64
	defaultPlatform = "x86_64"
)

//GetPlatform get platform
func GetPlatform() string {
	if len(Platform) == 0 {
		return defaultPlatform
	}

	return Platform
}
