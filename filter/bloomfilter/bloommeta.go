package bloomfilter

import "time"

type BloomMetaStore interface {
	Get() (*BloomMeta, error)
	Save(*BloomMeta) error
	Lock() error
}
type BloomMeta struct {
	currActiveDbName string
	nextRotationTime int64 // next rotation time, uint second
	rotationList     []string
	createTime       int64
	updateTime       int64
	rotationInterval int64
}

func (m *BloomMeta) init(activeProviderName string, rotationList []string, rotationInterval int64) {
	m.currActiveDbName = activeProviderName
	m.rotationList = rotationList

	m.createTime = time.Now().Unix()
	m.updateTime = time.Now().Unix()
	m.rotationInterval = rotationInterval
	m.nextRotationTime = time.Now().Unix() + rotationInterval
}
func (m *BloomMeta) changeNextActiveDb() {
	index := 0
	for i, v := range m.rotationList {
		if v == m.currActiveDbName {
			index = i + 1
			break
		}
	}

	if index == len(m.rotationList) {
		index = 0
	}

	m.currActiveDbName = m.rotationList[index]
}
func (m *BloomMeta) rotaionDb() (string, bool) {
	if time.Now().Unix() < m.nextRotationTime {
		return m.currActiveDbName, false
	}

	m.nextRotationTime = time.Now().Unix() + m.rotationInterval
	m.updateTime = time.Now().Unix()
	currActiveDbName := m.currActiveDbName
	m.changeNextActiveDb()

	return currActiveDbName, true
}
