package response

type AlgoResponse interface {
	GetScore() float64
	GetScoreMap() map[string]float64
	GetModuleType() bool
}

type ResponseFunc func(interface{}) ([]AlgoResponse, error)
