package module

import (
	"sort"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserCustomRecallDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
}

// customRecallItemIdScore extracts the score from an item_ids entry whose
// format is "itemid:retrieveid:score". Entries without a score segment are
// treated as 0.
//
// NOTE: this rule (split on ":", take the 3rd segment via utils.ToFloat(...,0))
// is duplicated by the *Item emit loop in every UserCustomRecall*Dao
// (hologres/tablestore/redis/featurestore/clickhouse); the two sides are not
// linked by code. If the entry format ever changes (extra segment, score moved
// to another index, different separator), update this function AND those emit
// loops together. Drifting here fails silently: a mismatch makes every score 0,
// so sortAndTruncateItemIds degrades from top-score truncation to
// first-N-in-original-order with no error or log.
func customRecallItemIdScore(id string) float64 {
	strs := strings.Split(id, ":")
	if len(strs) == 3 {
		return utils.ToFloat(strs[2], float64(0))
	}
	return float64(0)
}

// sortAndTruncateItemIds replaces the previous random-shuffle-then-truncate
// behavior: it sorts item ids by their parsed score in descending order and
// keeps the top recallCount entries, so that high score items are no longer
// randomly discarded. The sort is stable, therefore entries with equal score
// (including entries without score) keep their original relative order. When
// the candidates do not exceed recallCount, the original slice is returned
// unchanged.
func sortAndTruncateItemIds(itemIds []string, recallCount int) []string {
	// Preserve the original boundary semantics: the previous code truncated with
	// itemIds[:recallCount], so a non-positive recallCount produced an empty
	// result (and a negative one panicked on the slice bound). Returning an empty
	// slice keeps recallCount<=0 meaning "recall nothing" exactly as before, while
	// avoiding the original negative-index panic.
	if recallCount <= 0 {
		return itemIds[:0]
	}
	if len(itemIds) <= recallCount {
		return itemIds
	}
	type scoredId struct {
		id    string
		score float64
	}
	scored := make([]scoredId, len(itemIds))
	for i, id := range itemIds {
		scored[i] = scoredId{id: id, score: customRecallItemIdScore(id)}
	}
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	result := make([]string, recallCount)
	for i := 0; i < recallCount; i++ {
		result[i] = scored[i].id
	}
	return result
}

func NewUserCustomRecallDao(config recconf.RecallConfig) UserCustomRecallDao {
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Mysql {
		return NewUserCustomRecallMysqlDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewUserCustomRecallTableStoreDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewUserCustomRecallHologresDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Redis {
		return NewUserCustomRecallRedisDao(config)
	} else if config.DaoConf.AdapterType == recconf.DataSource_Type_ClickHouse {
		return NewUserCustomRecallClickHouseDao(config)
	} else if config.DaoConf.AdapterType == recconf.DataSource_Type_FeatureStore {
		return NewUserCustomRecallFeatureStoreDao(config)
	}

	panic("not found UserCustomRecallDao implement")
}
