package counterexample

import (
	"errors"
	"fmt"

	"github.com/backbone81/golr/internal/utils"
)

// ProductConfigurationQueue is a priority queue of configurations in order of increasing cost, held as
// PendingConfiguration. It holds a bucket of configurations per cost, starting at minCost. As every action of the
// search has a positive cost, no configuration is added which is cheaper than the one removed last, so buckets are only
// removed at the start and added at the end. An emptied bucket moves from the start to the end with its storage, so the
// buckets reuse it.
//
// The buckets hold the configurations by value in a utils.ChunkedStack, which keeps the garbage collector from tracking
// every configuration as an object of its own.
type ProductConfigurationQueue struct {
	// buckets holds the configurations of cost minCost+i at index i.
	buckets utils.DynamicRingBuffer[*utils.ChunkedStack[PendingConfiguration]]
	minCost int
	length  int
}

// initialBucketCount covers the costs from the cheapest configuration to its most expensive successor.
const initialBucketCount = productionStepCost + repeatedProductionStepCost + 1

// NewProductConfigurationQueue returns an empty queue.
func NewProductConfigurationQueue() ProductConfigurationQueue {
	buckets := utils.NewDynamicRingBufferWithCapacity[*utils.ChunkedStack[PendingConfiguration]](initialBucketCount)
	for range initialBucketCount {
		buckets.Add(&utils.ChunkedStack[PendingConfiguration]{})
	}
	return ProductConfigurationQueue{
		buckets: buckets,
	}
}

// Len returns the number of configurations.
func (q *ProductConfigurationQueue) Len() int {
	return q.length
}

// IsEmpty reports if the queue holds no configuration.
func (q *ProductConfigurationQueue) IsEmpty() bool {
	return q.Len() == 0
}

// Add adds the configuration to the queue. It must not be cheaper than the configuration removed last.
func (q *ProductConfigurationQueue) Add(c PendingConfiguration) {
	utils.DebugAssert(func() error {
		if c.Cost < q.minCost {
			return fmt.Errorf("configuration of cost %d is cheaper than the queue allows (%d)", c.Cost, q.minCost)
		}
		return nil
	})
	bucketIdx := c.Cost - q.minCost
	for q.buckets.Length() <= bucketIdx {
		q.buckets.Add(&utils.ChunkedStack[PendingConfiguration]{})
	}
	q.buckets.Get(bucketIdx).Push(c)
	q.length++
}

// Remove removes the cheapest configuration from the queue and returns it. The queue must not be empty.
func (q *ProductConfigurationQueue) Remove() PendingConfiguration {
	utils.DebugAssert(func() error {
		if q.IsEmpty() {
			return errors.New("configuration queue is empty")
		}
		return nil
	})
	for q.buckets.Get(0).IsEmpty() {
		q.buckets.Add(q.buckets.Remove())
		q.minCost++
	}
	bucket := q.buckets.Get(0)
	result := bucket.Top()
	bucket.Pop()
	q.length--
	return result
}
