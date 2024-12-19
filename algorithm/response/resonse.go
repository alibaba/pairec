package response

type AlgoResponse interface {
	GetScore() float64
	GetScoreMap() map[string]float64
	GetModuleType() bool
}

type AlgoMultiClassifyResponse interface {
	GetClassifyMap() map[string][]float64
}

type ResponseFunc func(interface{}) ([]AlgoResponse, error)
