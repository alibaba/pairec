package domain

type FeatureSelector struct {
	FeatureEntity string
	FeatureView   string
	Features      []string
	Alias         map[string]string
}
