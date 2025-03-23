package containers

import "time"

type Container struct {
	Owner         string            `json:"owner"`    // Namespace
	Name          string            `json:"name"`     // Unique identifier
	ImageTag      string            `json:"imageTag"` // Container image tag
	Replicas      int32             `json:"replicas"`
	ContainerTags map[string]string `json:"containerTags"`
	HasPorts      bool              `json:"hasPorts"`
	Port          string            `json:"ports,omitempty"`
	MappedPort    int32             `json:"mappedPort,omitempty"`
	CreationDate  time.Time         `json:",omitempty"`
}
