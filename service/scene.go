package service

import (
	"fmt"

	"github.com/alibaba/pairec/v2/recconf"
)

var scenes = make(map[string]*Scene)

func GetSence(sceneId string) (*Scene, error) {

	scene, ok := scenes[sceneId]
	if !ok {
		return nil, fmt.Errorf("Scene:not found, SceneId:%s", sceneId)
	}
	return scene, nil
}

func Load(conf *recconf.RecommendConfig) {

	for sceneId, categoryConfs := range conf.SceneConfs {
		var scene *Scene
		if _, ok := scenes[sceneId]; ok {
			scene = scenes[sceneId]
		} else {
			scene = NewScene(sceneId)
		}

		for categoryName, categoryConf := range categoryConfs {
			var category *Category
			if _, ok := scene.Categories[categoryName]; ok {
				category = scene.Categories[categoryName]
				category.Init(categoryConf)
			} else {
				category = NewCategory(categoryName)
				category.Init(categoryConf)
				scene.AddCategory(categoryName, category)
			}
		}

		if _, ok := scenes[sceneId]; !ok {
			scenes[sceneId] = scene
		}
	}
}

type Scene struct {
	SceneId    string
	Categories map[string]*Category
}

func NewScene(senceId string) *Scene {
	scene := Scene{SceneId: senceId}
	scene.Categories = make(map[string]*Category)
	return &scene
}
func (s *Scene) Init(config recconf.SceneConfig) {
}
func (s *Scene) AddCategory(name string, category *Category) {
	s.Categories[name] = category
}
func (s *Scene) GetCategory(name string) (*Category, error) {
	category, ok := s.Categories[name]
	if !ok {
		return s.Categories["default_category"], nil
	}
	return category, nil
}
