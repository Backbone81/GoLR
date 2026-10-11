package server

type Lifecycle int

const (
	LifecycleUninitialized Lifecycle = iota
	LifecycleRunning
	LifecycleShutDown
)
