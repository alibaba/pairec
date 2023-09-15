package eas

type wrrscheduler struct {
	currentWeight int
	current       int
	dataSet       map[string]int
	ipaddr        []string
	weight        []int
	maxS          int
	gcdS          int
	lenS          int
	inited        bool
}

func gcd(a, b int) int {
	if a%b > 0 {
		return gcd(b, a%b)
	}
	return b
}

func newwrr() *wrrscheduler {
	return &wrrscheduler{
		currentWeight: 0,
		current:       -1,
		maxS:          -1,
		gcdS:          1,
		lenS:          0,
	}
}

// func (w *wrrscheduler) reset() {
// 	w.currentWeight = 0
// 	w.current = -1
// 	w.dataSet = make(map[string]int)
// 	w.maxS = -1
// 	w.gcdS = 1
// 	w.lenS = 0
// }

func (w *wrrscheduler) init(endpoints map[string]int) {
	w.currentWeight = 0
	w.current = -1
	w.maxS = -1
	w.gcdS = 1
	w.inited = true
	w.lenS = len(endpoints)
	w.ipaddr = []string{}
	w.weight = []int{}
	w.dataSet = endpoints
	for k, v := range w.dataSet {
		w.ipaddr = append(w.ipaddr, k)
		w.weight = append(w.weight, v)
		w.gcdS = gcd(w.gcdS, v)
		if v > w.maxS {
			w.maxS = v
		}
	}
}

func wrrScheduler(endpoints map[string]int) wrrscheduler {
	w := &wrrscheduler{}
	w.init(endpoints)
	return *w
}

func (w *wrrscheduler) schedule() string {
	for true {
		if w.lenS == 0 {
			return ""
		}
		w.current = (w.current + 1) % w.lenS
		if w.current == 0 {
			w.currentWeight = w.currentWeight - w.gcdS
			if w.currentWeight <= 0 {
				w.currentWeight = w.maxS
				if w.currentWeight == 0 {
					return ""
				}
			}
		}
		if w.weight[w.current] >= w.currentWeight {
			return w.ipaddr[w.current]
		}
	}
	return ""
}

func (w *wrrscheduler) getNext() string {
	return w.schedule()
}
