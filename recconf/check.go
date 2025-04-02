package recconf

import (
	"encoding/json"
)

type CheckResult struct {
	Issues   []string
	RefCount int
}

func CheckRecommendConfig(configData string) (map[ModuleIndex]*CheckResult, error) {
	var conf RecommendConfig
	if err := json.Unmarshal([]byte(configData), &conf); err != nil {
		return nil, err
	}

	modules := conf.GetModules()
	results := make(map[ModuleIndex]*CheckResult, len(modules))
	refMap := make(map[ModuleIndex]int)

	for index, module := range modules {
		results[index] = new(CheckResult)

		if dependent, ok := module.(Dependent); ok {
			if err := dependent.Requirements().Check(modules); err != nil {
				results[index].Issues = append(results[index].Issues, err.Error())
			}

			for refIndex := range dependent.Requirements() {
				refMap[refIndex]++
			}
		}

		if validator, ok := module.(Validator); ok {
			if err := validator.Validate(); err != nil {
				results[index].Issues = append(results[index].Issues, err.Error())
			}
		}
	}

	for index, result := range results {
		result.RefCount = refMap[index]
	}

	return results, nil
}
