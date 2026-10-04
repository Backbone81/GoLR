package counterexample

// Statistics counts the work of the searches for unifying counterexamples.
type Statistics struct {
	// Generated counts the configurations the searches built.
	Generated int

	// Queued counts the configurations the searches queued, which are the ones generated which were not queued before.
	Queued int

	// Processed counts the configurations the searches removed from their queues.
	Processed int

	// PeakQueueLength is the most configurations a single search held in its queue at once.
	PeakQueueLength int
}

// Add adds the counts of the other statistics.
func (s *Statistics) Add(other Statistics) {
	s.Generated += other.Generated
	s.Queued += other.Queued
	s.Processed += other.Processed
	s.PeakQueueLength = max(s.PeakQueueLength, other.PeakQueueLength)
}
