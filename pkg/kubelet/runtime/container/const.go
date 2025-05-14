package container

// String representation of the container state.
// Can be one of "created", "running", "paused", "restarting", "removing", "exited", or "dead"

const ContainerStateCreated = "created"
const ContainerStateRunning = "running"
const ContainerStatePaused = "paused"
const ContainerStateRestarting = "restarting"
const ContainerStateRemoving = "removing"
const ContainerStateExited = "exited"
const ContainerStateDead = "dead"
