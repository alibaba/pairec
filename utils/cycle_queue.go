package utils

import "sync"

type CycleQueue struct {
	data  []interface{} //存储空间
	front int           //前指针,前指针负责弹出数据移动
	rear  int           //尾指针,后指针负责添加数据移动
	cap   int           //设置切片最大容量
	lock  sync.Mutex
}

// NewCycleQueue init Queue
func NewCycleQueue(cap int) *CycleQueue {
	cap++ // 因为有一个元素用不了
	return &CycleQueue{
		data:  make([]interface{}, cap),
		cap:   cap,
		front: 0,
		rear:  0,
	}
}

func (q *CycleQueue) Freeze() {
	q.lock.Lock()
}

func (q *CycleQueue) Unfreeze() {
	q.lock.Unlock()
}

// Length 因为是循环队列, 后指针减去前指针 加上最大值, 然后与最大值 取余
func (q *CycleQueue) Length() int {
	return (q.rear - q.front + q.cap) % q.cap
}

// Push 入队操作
// 判断队列是否队满,队满则不允许添加数据
func (q *CycleQueue) Push(data interface{}) bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	// check queue is full
	if (q.rear+1)%q.cap == q.front { //队列已满时，不执行入队操作
		return false
	}
	q.data[q.rear] = data         //将元素放入队列尾部
	q.rear = (q.rear + 1) % q.cap //尾部元素指向下一个空间位置,取模运算保证了索引不越界（余数一定小于除数）
	return true
}

// Pop 出队操作 需要考虑: 队空没有数据返回了
func (q *CycleQueue) Pop() interface{} {
	if q.rear == q.front {
		return nil
	}
	data := q.data[q.front]
	q.data[q.front] = nil
	q.front = (q.front + 1) % q.cap
	return data
}

func (q *CycleQueue) Peek() interface{} {
	if q.rear == q.front {
		return nil
	}
	return q.data[q.front]
}

func (q *CycleQueue) Tail() interface{} {
	if q.rear == q.front {
		return nil
	}
	return q.data[q.rear]
}

func (q *CycleQueue) IsFull() bool {
	if (q.rear+1)%q.cap == q.front {
		return true
	}
	return false
}

func (q *CycleQueue) IsEmpty() bool {
	if q.front == q.rear {
		return true
	}
	return false
}
