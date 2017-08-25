package efsutils

type LifeCycleState string

const (
	LifeCycleStateReady    LifeCycleState = "Ready"
	LifeCycleStateNotReady LifeCycleState = "Not Ready"
	LifeCycleStateUnknown  LifeCycleState = "Unknown"
)
