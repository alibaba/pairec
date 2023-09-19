package sort

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestDiversitryRuleSortByIntervalSize(t *testing.T) {
	config := recconf.SortConfig{

		DiversityRules: []recconf.DiversityRuleConfig{
			{
				Dimensions:   []string{"tag"},
				IntervalSize: 1,
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"
		if i < 5 {
			item.AddProperty("tag", "t1")
		} else {
			item.AddProperty("tag", "t2")
		}

		items = append(items, item)
	}
	context := context.NewRecommendContext()
	context.Size = 10
	sortData := SortData{Data: items, Context: context}

	NewDiversityRuleSort(config).Sort(&sortData)

	result := sortData.Data.([]*module.Item)

	for i, item := range result {
		if i%2 == 0 && item.StringProperty("tag") != "t1" {
			t.Error("item error")
		}
	}
}

func TestDiversitryRuleSortByIntervalSize2(t *testing.T) {
	config := recconf.SortConfig{

		DiversityRules: []recconf.DiversityRuleConfig{
			{
				Dimensions:   []string{"tag"},
				IntervalSize: 1,
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"
		item.AddProperty("tag", "t1")

		items = append(items, item)
	}
	context := context.NewRecommendContext()
	context.Size = 10
	sortData := SortData{Data: items, Context: context}

	NewDiversityRuleSort(config).Sort(&sortData)

	result := sortData.Data.([]*module.Item)

	for i, item := range result {
		if string(item.Id) != strconv.Itoa(i) {
			t.Error("item error")
		}
	}
}

func TestDiversitryRuleSort(t *testing.T) {
	config := recconf.SortConfig{

		DiversityRules: []recconf.DiversityRuleConfig{
			{
				Dimensions:    []string{"tag"},
				WindowSize:    5,
				FrequencySize: 2,
				IntervalSize:  1,
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"
		if i < 3 {
			item.AddProperty("tag", "t1")
		} else if i < 6 {
			item.AddProperty("tag", "t2")
		} else {
			item.AddProperty("tag", "t3")
		}

		items = append(items, item)
	}
	context := context.NewRecommendContext()
	context.Size = 10
	sortData := SortData{Data: items, Context: context}

	NewDiversityRuleSort(config).Sort(&sortData)

	result := sortData.Data.([]*module.Item)

	for i, item := range result {
		fmt.Println(i, item)
	}
}

//func readCsvFile(filePath string) [][]string {
//	f, err := os.Open(filePath)
//	if err != nil {
//		log.Fatal("Unable to read input file "+filePath, err)
//	}
//	//defer f.Close()
//
//	csvReader := csv.NewReader(f)
//	records, err := csvReader.ReadAll()
//	if err != nil {
//		log.Fatal("Unable to parse file as CSV for "+filePath, err)
//	}
//
//	return records
//}

//func TestDiversityRuleSort(t *testing.T) {
//	config := recconf.SortConfig{
//		DiversityRules: []recconf.DiversityRuleConfig{
//			{
//				Dimensions:    []string{"goods_id"},
//				WindowSize:    10,
//				FrequencySize: 1,
//			},
//			{
//				Dimensions:   []string{"root_brand_id"},
//				IntervalSize: 1,
//			},
//			{
//				Dimensions:    []string{"root_category_id"},
//				WindowSize:    3,
//				FrequencySize: 2,
//			},
//			{
//				Dimensions:    []string{"child_category_id"},
//				WindowSize:    3,
//				FrequencySize: 1,
//			},
//		},
//	}
//
//	records := readCsvFile("/Users/weisu.yxd/Code/recommand/shihuo/config/t.txt")
//
//	var items []*module.Item
//	for i, record := range records {
//		if i == 0 {
//			continue // skip header
//		}
//
//		item := module.NewItem(record[0])
//		item.AddProperty("goods_id", record[1])
//		item.AddProperty("root_category_id", record[2])
//		item.AddProperty("child_category_id", record[3])
//		item.AddProperty("root_brand_id", record[4])
//		item.AddProperty("vertical_name", record[5])
//		if s, err := strconv.ParseFloat(record[6], 64); err == nil {
//			item.Score = s
//		} else {
//			fmt.Println("parse score failed:" + record[6])
//		}
//		item.RetrieveId = "r1"
//		items = append(items, item)
//	}
//	sort.Sort(sort.Reverse(ItemScoreSlice(items)))
//
//	context := context.NewRecommendContext()
//	context.Size = 10
//	sortData := SortData{Data: items, Context: context}
//
//	NewDiversityRuleSort(config).Sort(&sortData)
//
//	result := sortData.Data.([]*module.Item)
//
//	for i, item := range result {
//		fmt.Println(i, item)
//	}
//}
