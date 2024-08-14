package domain

import (
	"fmt"
	"sync"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/utils"
)

type Model struct {
	*api.Model
	project                 *Project
	featureViewMap          map[string]FeatureView
	featureEntityMap        map[string]*FeatureEntity
	featureNamesMap         map[string][]string               // featureview : feature names
	aliasNamesMap           map[string]map[string]string      // featureview : alias names
	featureEntityJoinIdMap  map[string]map[string]FeatureView // feature entity joinid : featureviews
	featureEntityJoinIdList []string
}

func NewModel(model *api.Model, p *Project) *Model {
	m := &Model{
		Model:                  model,
		project:                p,
		featureViewMap:         make(map[string]FeatureView),
		featureEntityMap:       make(map[string]*FeatureEntity),
		featureNamesMap:        make(map[string][]string),
		aliasNamesMap:          make(map[string]map[string]string),
		featureEntityJoinIdMap: make(map[string]map[string]FeatureView),
	}

	for _, feature := range m.Features {
		featureView := m.project.GetFeatureView(feature.FeatureViewName)

		featureEntity := m.project.GetFeatureEntity(featureView.GetFeatureEntityName())
		m.featureViewMap[feature.FeatureViewName] = featureView
		m.featureEntityMap[featureView.GetFeatureEntityName()] = featureEntity
		m.featureNamesMap[feature.FeatureViewName] = append(m.featureNamesMap[feature.FeatureViewName], featureView.Offline2Online(feature.Name))

		if feature.AliasName != "" {
			aliasMap, ok := m.aliasNamesMap[feature.FeatureViewName]
			if !ok {
				aliasMap = make(map[string]string)
			}
			aliasMap[feature.Name] = feature.AliasName
			m.aliasNamesMap[feature.FeatureViewName] = aliasMap
		}
		featureViewMap, ok := m.featureEntityJoinIdMap[featureEntity.FeatureEntityJoinid]
		if !ok {
			featureViewMap = make(map[string]FeatureView)
		}
		featureViewMap[feature.FeatureViewName] = featureView
		m.featureEntityJoinIdMap[featureEntity.FeatureEntityJoinid] = featureViewMap

	}
	for joinid := range m.featureEntityJoinIdMap {
		m.featureEntityJoinIdList = append(m.featureEntityJoinIdList, joinid)
	}

	//fmt.Println(m)
	return m
}

func (m *Model) GetOnlineFeatures(joinIds map[string][]interface{}) ([]map[string]interface{}, error) {

	size := -1
	for _, joinid := range m.featureEntityJoinIdList {
		if _, ok := joinIds[joinid]; !ok {
			return nil, fmt.Errorf("join id:%s not found", joinid)
		}
		if size == -1 {
			size = len(joinIds[joinid])
		} else {
			if size != len(joinIds[joinid]) {
				return nil, fmt.Errorf("join id:%s length not equal", joinid)
			}
		}
	}

	var mu sync.Mutex

	var wg sync.WaitGroup
	joinIdFeaturesMap := make(map[string][]map[string]interface{})
	for joinId, keys := range joinIds {
		featureViewMap := m.featureEntityJoinIdMap[joinId]

		for _, featureView := range featureViewMap {
			wg.Add(1)
			go func(featureView FeatureView, joinId string, keys []interface{}) {
				defer wg.Done()
				features, err := featureView.GetOnlineFeatures(keys, m.featureNamesMap[featureView.GetName()], m.aliasNamesMap[featureView.GetName()])
				if err != nil {
					fmt.Println(err)
				}

				mu.Lock()
				joinIdFeaturesMap[joinId] = append(joinIdFeaturesMap[joinId], features...)
				mu.Unlock()

			}(featureView, joinId, keys)
		}
	}
	wg.Wait()

	featuresResult := make([]map[string]interface{}, 0, size)
	for i := 0; i < size; i++ {
		features := make(map[string]interface{}, len(m.Features)+len(m.featureEntityJoinIdMap))
		for _, joinid := range m.featureEntityJoinIdList {
			joinIdValue := joinIds[joinid][i]
			for _, joinIdFeatures := range joinIdFeaturesMap[joinid] {
				if utils.ToString(joinIdFeatures[joinid], "") == utils.ToString(joinIdValue, " ") {
					for k, v := range joinIdFeatures {
						features[k] = v
					}
				}
			}
		}

		featuresResult = append(featuresResult, features)

	}
	return featuresResult, nil
}

func (m *Model) GetOnlineFeaturesWithEntity(joinIds map[string][]interface{}, featureEntityName string) ([]map[string]interface{}, error) {
	featureEntity, ok := m.featureEntityMap[featureEntityName]
	if !ok {
		return nil, fmt.Errorf("feature entity name:%s not found", featureEntityName)
	}
	size := -1
	if _, ok := joinIds[featureEntity.FeatureEntityJoinid]; !ok {
		return nil, fmt.Errorf("join id:%s not found", featureEntity.FeatureEntityJoinid)
	}

	size = len(joinIds[featureEntity.FeatureEntityJoinid])

	var wg sync.WaitGroup
	joinIdFeaturesMap := make(map[string][]map[string]interface{})
	featureViewMap := m.featureEntityJoinIdMap[featureEntity.FeatureEntityJoinid]

	var mu sync.Mutex

	for _, featureView := range featureViewMap {
		wg.Add(1)
		go func(featureView FeatureView, joinId string, keys []interface{}) {
			defer wg.Done()
			features, err := featureView.GetOnlineFeatures(keys, m.featureNamesMap[featureView.GetName()], m.aliasNamesMap[featureView.GetName()])
			if err != nil {
				fmt.Println(err)
			}
			mu.Lock()
			joinIdFeaturesMap[joinId] = append(joinIdFeaturesMap[joinId], features...)
			mu.Unlock()

		}(featureView, featureEntity.FeatureEntityJoinid, joinIds[featureEntity.FeatureEntityJoinid])
	}
	wg.Wait()

	featuresResult := make([]map[string]interface{}, 0, size)
	for i := 0; i < size; i++ {
		features := make(map[string]interface{}, len(m.Features))
		joinIdValue := joinIds[featureEntity.FeatureEntityJoinid][i]
		for _, joinIdFeatures := range joinIdFeaturesMap[featureEntity.FeatureEntityJoinid] {
			if utils.ToString(joinIdFeatures[featureEntity.FeatureEntityJoinid], "") == utils.ToString(joinIdValue, " ") {
				for k, v := range joinIdFeatures {
					features[k] = v
				}
			}
		}

		featuresResult = append(featuresResult, features)

	}
	return featuresResult, nil
}
