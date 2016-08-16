package mocker

// Config for mock server
type Config struct {
	RAMLFile                       string
	CheckRAMLVersion               bool
	CacheDir                       string
	Port                           int64
	Proxy                          string
	Resources                      map[string]bool
	AllowRequiredPropertyToBeEmpty bool
}

// BuildResourcesMap return resource map by resources string slice
func BuildResourcesMap(resources []string) map[string]bool {
	resmap := map[string]bool{}
	for _, respath := range resources {
		resmap[respath] = true
		resmap[toRAMLResource(respath)] = true
		resmap[toGinResource(respath)] = true
	}
	return resmap
}

var config = &Config{}
