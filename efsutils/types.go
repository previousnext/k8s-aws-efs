package efsutils

// LifeCycleState contains all the states which an EFS can be.
type LifeCycleState string

const (
	// LifeCycleStateReady shows when the Elastic Filesystem is ready.
	LifeCycleStateReady LifeCycleState = "Ready"

	// LifeCycleStateNotReady shows when the Elastic Filesystem is not ready.
	LifeCycleStateNotReady LifeCycleState = "Not Ready"

	// LifeCycleStateUnknown shows when we are not sure what the state is of the Elastic Filesystem.
	LifeCycleStateUnknown LifeCycleState = "Unknown"
)
