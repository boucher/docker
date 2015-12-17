// +build experimental

package daemon

import (
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/container"
)

func addExperimentalFields(container *container.Container, data *types.ContainerJSON) {
	data.State.Checkpointed = container.State.Checkpointed
	data.State.CheckpointedAt = container.State.CheckpointedAt.Format(time.RFC3339Nano)
}
