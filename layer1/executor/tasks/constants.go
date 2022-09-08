package tasks

import "time"

const (
	// How much time we are going to poll to check if the task is completed
	ExecutorPoolingTime time.Duration = 7 * time.Second
)

const (
	// The size of the buffered channels used by the TaskManager
	ManagerBufferSize uint64 = 1024
	// Timeout in seconds for the network interactions
	ManagerNetworkTimeout time.Duration = 1 * time.Second
	// Time in which the scheduler it's going to enter in the main loop to spawn tasks
	ManagerProcessingTime time.Duration = 3 * time.Second
	// how many blocks after we sent a kill sign to a task to consider it
	// unresponsive and removing from the scheduler mapping
	ManagerHeightToleranceBeforeRemoving uint64 = 50
	// how many block we wait before removing a response from the TaskManager
	ManagerResponseToleranceBeforeRemoving uint64 = 50
)
