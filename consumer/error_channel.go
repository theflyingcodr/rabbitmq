package consumer

import(
	log "github.com/sirupsen/logrus"
	"fmt"
	"time"
)

type Node struct{
	TimeStamp time.Time
	Error string
}

// Queue is a basic FIFO queue based on a circular list that resizes as needed.
type Queue struct {
	nodes []*Node
	size  int
	head  int
	tail  int
	count int
}

// NewQueue returns a new queue with the given initial size.
func NewQueue(size int) *Queue {
	return &Queue{
		nodes: make([]*Node, size),
		size:  size,
	}
}

// Push adds a node to the queue.
func (q *Queue) Push(n *Node) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*Node, len(q.nodes)+q.size)
		copy(nodes, q.nodes[q.head:])
		copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
		q.head = 0
		q.tail = len(q.nodes)
		q.nodes = nodes
	}
	q.nodes[q.tail] = n
	q.tail = (q.tail + 1) % len(q.nodes)
	q.count++
}

// Pop removes and returns a node from the queue in first to last order.
func (q *Queue) Pop() *Node {
	if q.count == 0 {
		return nil
	}
	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return node
}

func (q *Queue) ReadAll() []*Node{
	for q.count > 10{
		q.Pop()
	}
	return q.nodes
}

type RabbitErrorChannel struct{
	q *Queue
}

type ErrorChannel interface{
	LogErrorf(format string, a ...interface{})
	LogError(err error)
	GetErrors() []*Node
}

func NewErrorChannel() ErrorChannel{
	return &RabbitErrorChannel{
		q: NewQueue(10),
	}
}

func (e *RabbitErrorChannel) LogErrorf(format string, a ...interface{}) {
	log.Errorf(format, a)
	e.q.Push(&Node{
		Error:fmt.Sprintf(format, a),
		TimeStamp:time.Now().UTC(),
	})
}


func (e *RabbitErrorChannel) LogError(err error) {
	log.Error(err)
	e.q.Push(&Node{
		Error:err.Error(),
		TimeStamp:time.Now().UTC(),
	})
}

func (e *RabbitErrorChannel) GetErrors() []*Node{
	return e.q.ReadAll()
}