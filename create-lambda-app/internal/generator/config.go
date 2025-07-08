package generator

// Config holds the configuration for project generation
type Config struct {
	Name             string
	Description      string
	DeploymentTool   string
	Architecture     string
	TestingFramework string
	Features         map[string]bool
	SkipGit          bool
	SkipInstall      bool
	Module           string // Go module name
}

// HasFeature checks if a feature is enabled
func (c *Config) HasFeature(feature string) bool {
	return c.Features[feature]
}

// GetEnabledFeatures returns a list of enabled features
func (c *Config) GetEnabledFeatures() []string {
	var features []string
	for feature, enabled := range c.Features {
		if enabled {
			features = append(features, feature)
		}
	}
	return features
}