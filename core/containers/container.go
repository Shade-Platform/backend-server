package containers

import "time"

type Container struct {
	UserName     string    `json:"username"`
	ContainerTag string    `json:"containerTag"`
	OpenedPort  int32    `json:"openedPorts,omitempty"`
	CreationDate time.Time `json:",omitempty"`
}
