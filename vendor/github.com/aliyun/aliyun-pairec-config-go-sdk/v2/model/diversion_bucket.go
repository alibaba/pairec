package model

import (
	"crypto/md5"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/common"
)

// DiversionBucket is a interface by the bucket type to match diversion
type DiversionBucket interface {
	Match(*ExperimentContext) bool
}

func NewDiversionBucket(bucketType uint32) DiversionBucket {
	if bucketType == common.Bucket_Type_UID {
		return &UidDiversionBucket{}
	} else if bucketType == common.Bucket_Type_UID_HASH {
		return &UidHashDiversionBucket{}
	} else if bucketType == common.Bucket_Type_Custom {
		return &CustomDiversionBucket{}
	} else if bucketType == common.Bucket_Type_Filter {
		return &FilterDiversionBucket{}
	}

	return nil
}

type UidDiversionBucket struct {
	buckets     map[int]bool
	bucketCount int
}

func NewUidDiversionBucket(bucketCount int, bucketStr string) *UidDiversionBucket {
	diversionBucket := &UidDiversionBucket{
		bucketCount: bucketCount,
		buckets:     make(map[int]bool),
	}

	expBuckets := strings.Split(bucketStr, ",")
	for _, bucket := range expBuckets {
		if strings.Contains(bucket, "-") {
			bucketStrings := strings.Split(bucket, "-")
			if len(bucketStrings) == 2 {
				start, err1 := strconv.Atoi(bucketStrings[0])
				end, err2 := strconv.Atoi(bucketStrings[1])
				if err1 == nil && err2 == nil {
					for i := start; i < end; i++ {
						if i > int(diversionBucket.bucketCount) {
							break
						}
						diversionBucket.buckets[i] = true
					}
				}
			}
		} else {
			if val, err := strconv.Atoi(bucket); err == nil {
				diversionBucket.buckets[val] = true
			}
		}
	}
	return diversionBucket
}
func (b *UidDiversionBucket) Match(experimentContext *ExperimentContext) bool {

	uid, err := strconv.ParseUint(experimentContext.Uid, 10, 64)
	if err != nil {
		return false
	}

	mod := uid % uint64(b.bucketCount)
	if _, found := b.buckets[int(mod)]; found {
		return true
	}

	return false
}

type UidHashDiversionBucket struct {
	*UidDiversionBucket
}

func NewUidHashDiversionBucket(bucketCount int, bucketStr string) *UidHashDiversionBucket {
	diversionBucket := &UidHashDiversionBucket{
		UidDiversionBucket: NewUidDiversionBucket(bucketCount, bucketStr),
	}

	return diversionBucket
}

func (b *UidHashDiversionBucket) Match(experimentContext *ExperimentContext) bool {

	md5 := md5.Sum([]byte(experimentContext.Uid))
	hash := fnv.New64()
	hash.Write(md5[:])

	mod := hash.Sum64() % uint64(b.bucketCount)
	if _, found := b.buckets[int(mod)]; found {
		return true
	}

	return false
}

type CustomDiversionBucket struct {
}

func NewCustomDiversionBucket() *CustomDiversionBucket {
	return &CustomDiversionBucket{}
}
func (b *CustomDiversionBucket) Match(experimentContext *ExperimentContext) bool {

	return false
}

type FilterDiversionBucket struct {
	Filter              string
	EvaluableExpression *govaluate.EvaluableExpression
}

// NewFilterDiversionBucket return instance of FilterDiversionBucket
func NewFilterDiversionBucket(filter string) (*FilterDiversionBucket, error) {
	diversionBucket := &FilterDiversionBucket{
		Filter: filter,
	}
	evaluableExpression, err := govaluate.NewEvaluableExpression(filter)
	if err != nil {
		return nil, err
	}

	diversionBucket.EvaluableExpression = evaluableExpression
	return diversionBucket, nil
}

// Match is a function of FilterDiversionBucket implements the DiversionBucket interface
func (b *FilterDiversionBucket) Match(experimentContext *ExperimentContext) bool {
	if b.EvaluableExpression != nil && experimentContext.FilterParams != nil {
		if result, err := b.EvaluableExpression.Evaluate(experimentContext.FilterParams); err == nil {
			return common.ToBool(result, false)
		} else {
			fmt.Println(err)
		}
	}
	return false
}
