package proxmox

import (
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes/storage"
	log "github.com/sirupsen/logrus"
)

func (client *Client) StorageCreate(cluster string, storageConfig storage.StorageConfig) error {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		return err
	}

	log.Infof("Creating storage, cluster: %s config: %+v", cluster, storageConfig)
	err = storage.Create(apiConfig, storageConfig)
	if err != nil {
		return err
	}

	return nil
}
