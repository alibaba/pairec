package bloomfilter

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

const (
	Redis_BloomMeta_Key      = "bloommeta"
	Redis_BloomMeta_Lock_Key = "bloommeta_lock"
)

type ReidsBloomMetaStore struct {
	pool *redis.Pool
}

func NewReidsBloomMetaStore(pool *redis.Pool) *ReidsBloomMetaStore {
	return &ReidsBloomMetaStore{
		pool: pool,
	}
}
func (r *ReidsBloomMetaStore) Get() (*BloomMeta, error) {

	conn := r.pool.Get()
	defer conn.Close()

	replys, err := redis.Strings(conn.Do("HGETALL", Redis_BloomMeta_Key))
	if err != nil {
		return nil, err
	}
	meta := &BloomMeta{}
	for i := 0; i < len(replys); i += 2 {
		key := replys[i]
		if key == "currActiveDbName" {
			meta.currActiveDbName = replys[i+1]
		} else if key == "nextRotationTime" {
			meta.nextRotationTime, _ = strconv.ParseInt(replys[i+1], 10, 64)
		} else if key == "rotationList" {
			list := replys[i+1]
			meta.rotationList = strings.Split(list, ",")
		} else if key == "rotationInterval" {
			meta.rotationInterval, _ = strconv.ParseInt(replys[i+1], 10, 64)
		} else if key == "createTime" {
			meta.createTime, _ = strconv.ParseInt(replys[i+1], 10, 64)
		} else if key == "updateTime" {
			meta.updateTime, _ = strconv.ParseInt(replys[i+1], 10, 64)
		}
	}
	if meta.currActiveDbName == "" {
		return nil, nil
	}

	return meta, nil
}
func (r *ReidsBloomMetaStore) Save(meta *BloomMeta) error {
	conn := r.pool.Get()
	defer conn.Close()

	args := make([]interface{}, 0, 13)
	args = append(args, Redis_BloomMeta_Key)
	args = append(args, "currActiveDbName", meta.currActiveDbName)
	args = append(args, "nextRotationTime", meta.nextRotationTime)
	args = append(args, "rotationInterval", meta.rotationInterval)
	args = append(args, "createTime", meta.createTime)
	args = append(args, "updateTime", meta.updateTime)
	args = append(args, "rotationList", strings.Join(meta.rotationList, ","))

	_, err := conn.Do("HMSET", args...)
	return err
}
func (r *ReidsBloomMetaStore) Lock() error {
	conn := r.pool.Get()
	defer conn.Close()

	reply, err := redis.String(conn.Do("SET", Redis_BloomMeta_Lock_Key, "lock", "EX", 100, "NX"))
	if err != nil {
		return err
	}
	if reply != "OK" {
		return errors.New("get lock fail")
	}

	return nil
}
