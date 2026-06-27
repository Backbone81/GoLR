package ielr1go_test

import (
	"fmt"
	"math/rand"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/core/ielr1go"
)

var _ = Describe("Digraph Algorithm", func() {
	// The contract of the algorithm is that every node ends up with the union of the initial follow sets of all nodes
	// reachable from it through the relation (including itself). Each node n is seeded with the single token n, so the
	// expected follow set of a node is exactly the set of nodes reachable from it. We assert presence and absence of
	// every token, so over-propagation (a node receiving a follow it should not see) is caught as well.
	DescribeTable("should propagate follow sets to every reachable node and to no others",
		func(nodeCount int, edges []ielr1go.Edge, expected [][]int) {
			gotoRecords := make([]ielr1go.GotoRecord, nodeCount)
			for nodeIdx := range gotoRecords {
				gotoRecords[nodeIdx].GotoFollows.Add(nodeIdx)
			}

			digraph := ielr1go.NewDigraphAlgorithm(gotoRecords, edges, func(fromGotoIdx int, toGotoIdx int) {
				gotoRecords[fromGotoIdx].GotoFollows.Merge(&gotoRecords[toGotoIdx].GotoFollows)
			})
			digraph.Execute()

			for nodeIdx := range gotoRecords {
				for token := range nodeCount {
					want := slices.Contains(expected[nodeIdx], token)
					Expect(gotoRecords[nodeIdx].GotoFollows.Contains(token)).To(
						Equal(want),
						"node %d should have token %d == %v", nodeIdx, token, want,
					)
				}
			}
		},

		// No edges: nothing propagates, every node keeps only its own token.
		Entry("isolated nodes without edges",
			3,
			[]ielr1go.Edge{},
			[][]int{{0}, {1}, {2}},
		),

		// A single node looping onto itself must not break and must keep just its own token.
		Entry("a self loop",
			1,
			[]ielr1go.Edge{{FromIdx: 0, ToIdx: 0}},
			[][]int{{0}},
		),

		// A linear chain 0 -> 1 -> 2: follows accumulate towards the head of the chain.
		Entry("a linear chain",
			3,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 2},
			},
			[][]int{{0, 1, 2}, {1, 2}, {2}},
		),

		// A diamond with a shared child: 0 -> 1, 0 -> 2, 1 -> 3, 2 -> 3. Node 3 is reached along two paths but its
		// follows must reach node 0 exactly once and completely.
		Entry("a diamond with a shared child",
			4,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 0, ToIdx: 2},
				{FromIdx: 1, ToIdx: 3},
				{FromIdx: 2, ToIdx: 3},
			},
			[][]int{{0, 1, 2, 3}, {1, 3}, {2, 3}, {3}},
		),

		// Two disconnected components must not leak follows into each other and Execute must visit both roots.
		Entry("two disconnected components",
			4,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 2, ToIdx: 3},
			},
			[][]int{{0, 1}, {1}, {2, 3}, {3}},
		),

		// A 2-node loop 0 <-> 1: both nodes share the union of their follows.
		Entry("a loop of size two",
			2,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 0},
			},
			[][]int{{0, 1}, {0, 1}},
		),

		// A 3-node loop 0 -> 1 -> 2 -> 0. Every member of a strongly connected component must end up with the union of
		// all members' initial sets (DeRemer and Pennello, "F(Top of S) <- F x"). A 2-node loop gets this right by
		// accident during stack unwinding, so at least three nodes are needed to exercise the copy back across an SCC.
		Entry("a loop of size three",
			3,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 2},
				{FromIdx: 2, ToIdx: 0},
			},
			[][]int{{0, 1, 2}, {0, 1, 2}, {0, 1, 2}},
		),

		// A 4-node loop 0 -> 1 -> 2 -> 3 -> 0: every member of the cycle gets every token.
		Entry("a loop of size four",
			4,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 2},
				{FromIdx: 2, ToIdx: 3},
				{FromIdx: 3, ToIdx: 0},
			},
			[][]int{{0, 1, 2, 3}, {0, 1, 2, 3}, {0, 1, 2, 3}, {0, 1, 2, 3}},
		),

		// A cycle with a chord 0 -> 1 -> 2 -> 0 plus a shortcut 0 -> 2. The shortcut visits an already active node and
		// must not confuse the strongly connected component detection: all three nodes still share every token.
		Entry("a loop with a shortcut chord",
			3,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 2},
				{FromIdx: 2, ToIdx: 0},
				{FromIdx: 0, ToIdx: 2},
			},
			[][]int{{0, 1, 2}, {0, 1, 2}, {0, 1, 2}},
		),

		// Two strongly connected components linked by a single edge: {0,1} -> {2,3}. The downstream component's follows
		// must reach every member of the upstream component, but not the other way around.
		Entry("two loops linked by an edge",
			4,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 0},
				{FromIdx: 1, ToIdx: 2},
				{FromIdx: 2, ToIdx: 3},
				{FromIdx: 3, ToIdx: 2},
			},
			[][]int{{0, 1, 2, 3}, {0, 1, 2, 3}, {2, 3}, {2, 3}},
		),

		// An inner loop {1,2} reached from an entry node 0 and leading to a sink node 3. This exercises a nested cycle
		// that is not the whole graph, so the four nodes end up with four different follow sets.
		Entry("an inner loop with an entry and a sink",
			4,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 2},
				{FromIdx: 2, ToIdx: 1},
				{FromIdx: 2, ToIdx: 3},
			},
			[][]int{{0, 1, 2, 3}, {1, 2, 3}, {1, 2, 3}, {3}},
		),

		// Two independent loops {1,2} and {3,4} reachable from a common root 0. This is the situation of DeRemer and
		// Pennello's diagram (4.3): a single graph holds several distinct strongly connected components that must not be
		// merged. The two loops share no follows; only the root sees all of them.
		Entry("two independent loops reachable from a common root",
			5,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 2},
				{FromIdx: 2, ToIdx: 1},
				{FromIdx: 0, ToIdx: 3},
				{FromIdx: 3, ToIdx: 4},
				{FromIdx: 4, ToIdx: 3},
			},
			[][]int{{0, 1, 2, 3, 4}, {1, 2}, {1, 2}, {3, 4}, {3, 4}},
		),

		// Loop {3,4} has an edge into the already resolved loop {1,2} via 4 -> 1. The follows of the first loop must flow
		// into the second loop, but the two loops must stay distinct strongly connected components. This is the case
		// that relies on popped component members being marked as resolved so they are not pulled back into a later one.
		Entry("an edge into an already resolved loop",
			5,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 2},
				{FromIdx: 2, ToIdx: 1},
				{FromIdx: 0, ToIdx: 3},
				{FromIdx: 3, ToIdx: 4},
				{FromIdx: 4, ToIdx: 3},
				{FromIdx: 4, ToIdx: 1},
			},
			[][]int{{0, 1, 2, 3, 4}, {1, 2}, {1, 2}, {1, 2, 3, 4}, {1, 2, 3, 4}},
		),

		// An adversarial interleaving for the strongly connected component detection. The loop {1,2} is resolved first
		// at a shallow stack depth. Afterwards the entry node 3 leads into the loop {4,5}, which reaches back into the
		// already resolved loop via 5 -> 1. Node 3 is only an entry into {4,5} and must remain its own component: in
		// particular nodes 4 and 5 must not receive token 3. This only holds if resolved component members are excluded
		// from later traversals, so this case fails if that exclusion is dropped.
		Entry("an entry node into a loop that reaches an earlier resolved loop",
			6,
			[]ielr1go.Edge{
				{FromIdx: 0, ToIdx: 1},
				{FromIdx: 1, ToIdx: 2},
				{FromIdx: 2, ToIdx: 1},
				{FromIdx: 0, ToIdx: 3},
				{FromIdx: 3, ToIdx: 4},
				{FromIdx: 4, ToIdx: 5},
				{FromIdx: 5, ToIdx: 4},
				{FromIdx: 5, ToIdx: 1},
			},
			[][]int{{0, 1, 2, 3, 4, 5}, {1, 2}, {1, 2}, {1, 2, 3, 4, 5}, {1, 2, 4, 5}, {1, 2, 4, 5}},
		),
	)

	It("should match the naive reachability oracle on random small graphs", func() {
		// We use Ginkgo's own random seed so a failing run is reproducible by re-running with the same --seed.
		random := rand.New(rand.NewSource(GinkgoRandomSeed()))
		// Do not lower this iteration count carelessly. It needs to stay well above the number of iterations it takes to
		// surface the subtle bugs in the digraph algorithm. Mutation testing on small random graphs showed that a missing
		// copy-back merge is detected within about 10 iterations, and a missing MaxInt marking (the rarer of the two,
		// because it needs a specific interleaving of strongly connected components) within about 40 iterations. The 2000
		// here keeps a large safety margin for even rarer interleavings we have not enumerated.
		for iteration := range 2000 {
			nodeCount, edges := randomGraph(random)
			mismatch, ok := compareDigraphToOracle(nodeCount, edges)
			Expect(ok).To(BeTrue(), "iteration %d: %s", iteration, mismatch)
		}
	})
})

// compareDigraphToOracle runs the digraph algorithm and the naive oracle on the same graph and reports the first
// disagreement. Each node is seeded with the single token equal to its own index, so a node's expected follow set is
// exactly the set of nodes reachable from it. Seeding singletons is sufficient to exercise propagation completely:
// follow sets distribute over union, so correctness for per-node singletons implies correctness for any initial sets.
func compareDigraphToOracle(nodeCount int, edges []ielr1go.Edge) (string, bool) {
	gotoRecords := make([]ielr1go.GotoRecord, nodeCount)
	for nodeIdx := range gotoRecords {
		gotoRecords[nodeIdx].GotoFollows.Add(nodeIdx)
	}
	digraph := ielr1go.NewDigraphAlgorithm(gotoRecords, edges, func(fromGotoIdx int, toGotoIdx int) {
		gotoRecords[fromGotoIdx].GotoFollows.Merge(&gotoRecords[toGotoIdx].GotoFollows)
	})
	digraph.Execute()

	want := naiveReachabilityFollows(nodeCount, edges)

	for nodeIdx := range nodeCount {
		for token := range nodeCount {
			got := gotoRecords[nodeIdx].GotoFollows.Contains(token)
			if got != want[nodeIdx][token] {
				return fmt.Sprintf(
					"node %d token %d: got %v, want %v (nodeCount=%d, edges=%v)",
					nodeIdx, token, got, want[nodeIdx][token], nodeCount, edges,
				), false
			}
		}
	}
	return "", true
}

// naiveReachabilityFollows is the defining specification of the digraph algorithm, implemented in the slowest and most
// obviously correct way: every node starts with its own token, and follow sets are pushed backward across every edge,
// repeatedly, until nothing changes. The result for each node is the union of the initial tokens of all nodes reachable
// from it (including itself). It shares no mechanism with DigraphAlgorithm (no strongly connected component detection,
// no traversal order, no stack), which is what makes it a trustworthy oracle rather than a second copy of the same
// algorithm. It is defined in the test file so it can never be reached from production code, and it is far too slow to
// be used there anyway.
func naiveReachabilityFollows(nodeCount int, edges []ielr1go.Edge) []map[int]bool {
	follows := make([]map[int]bool, nodeCount)
	for nodeIdx := range follows {
		follows[nodeIdx] = map[int]bool{nodeIdx: true}
	}
	for changed := true; changed; {
		changed = false
		for _, edge := range edges {
			for token := range follows[edge.ToIdx] {
				if !follows[edge.FromIdx][token] {
					follows[edge.FromIdx][token] = true
					changed = true
				}
			}
		}
	}
	return follows
}

// randomGraph builds a small random graph for the property test. Keeping the node count tiny ensures the naive oracle
// stays cheap while still producing cycles, shared children and nested components often enough to be interesting.
func randomGraph(random *rand.Rand) (int, []ielr1go.Edge) {
	const maxNodes = 8
	nodeCount := random.Intn(maxNodes) + 1
	edgeCount := random.Intn(nodeCount*nodeCount + 1)
	edges := make([]ielr1go.Edge, edgeCount)
	for edgeIdx := range edges {
		edges[edgeIdx] = ielr1go.Edge{
			FromIdx: random.Intn(nodeCount),
			ToIdx:   random.Intn(nodeCount),
		}
	}
	return nodeCount, edges
}
