package runconfig

import (
    "github.com/docker/libcontainer"
)

type RestoreConfig struct {
    CriuOpts libcontainer.CriuOpts
    ForceRestore bool
}
