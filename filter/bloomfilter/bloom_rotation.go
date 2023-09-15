package bloomfilter

import (
	"fmt"
	"time"
)

type BloomRotationInterface interface {
	BitSetClear(name string)
	BitSetOnline(name string, online bool)
}
type BloomRotation struct {
	bloom     BloomRotationInterface
	metaInfo  BloomMeta
	metaStore BloomMetaStore
}

func (f *BloomRotation) SetBloomMetaStore(store BloomMetaStore) {
	f.metaStore = store
}
func (f *BloomRotation) StartRotation(activeProviderName string, rotationList []string, rotationInterval int64, overwirte bool) {
	storeMeta, err := f.metaStore.Get()
	if err != nil {
		panic(fmt.Sprintf("bloom filter start rotation error ,err=%v", err))
	}

	if storeMeta == nil {
		// first init
		f.metaInfo.init(activeProviderName, rotationList, rotationInterval)
		f.metaStore.Save(&f.metaInfo)
	} else if overwirte == true {
		f.metaStore.Save(&f.metaInfo)
	} else {
		// load from metaStore
		f.metaInfo = *storeMeta
	}
	go f.loopRotaion()
}
func (f *BloomRotation) loopRotaion() {
	for {

		if name, ret := f.metaInfo.rotaionDb(); ret == true {
			rotationList := f.metaInfo.rotationList
			var newRotationList []string
			for _, dbname := range rotationList {
				if name != dbname {
					newRotationList = append(newRotationList, dbname)
				}
			}
			f.metaInfo.rotationList = newRotationList
			f.bloom.BitSetOnline(name, false)
			// distribute lock
			err := f.metaStore.Lock()
			if err == nil {
				// lock success
				time.Sleep(65 * time.Second)
				err = f.metaStore.Save(&f.metaInfo)
				fmt.Println(err, "save")
				f.bloom.BitSetClear(name)

				f.metaInfo.rotationList = rotationList
				for {
					err := f.metaStore.Save(&f.metaInfo)
					if err == nil {
						break
					}

					fmt.Println("metaInfo save error", err)
					time.Sleep(time.Second)
				}

			} else {
				// lock fail
				time.Sleep(70 * time.Second)
			}

			f.bloom.BitSetOnline(name, true)
		}
		storeMeta, err := f.metaStore.Get()
		if err == nil && storeMeta != nil {
			f.metaInfo = *storeMeta
		}
		time.Sleep(time.Minute)
	}
}
