package backup

import (
	"context"
	"time"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/logging"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/backup"

	"github.com/dougkirkley/plugin-objstore-backup/internal/backup/storage"

	"github.com/dougkirkley/plugin-objstore-backup/internal/backup/executor"
	"github.com/dougkirkley/plugin-objstore-backup/internal/backup/repository"
	"github.com/dougkirkley/plugin-objstore-backup/pkg/metadata"
)

// BackupServer is the implementation of the identity service
type BackupServer struct {
	backup.BackupServer
}

// GetCapabilities gets the capabilities of the Backup service
func (BackupServer) GetCapabilities(
	context.Context,
	*backup.BackupCapabilitiesRequest,
) (*backup.BackupCapabilitiesResult, error) {
	return &backup.BackupCapabilitiesResult{
		Capabilities: []*backup.BackupCapability{
			{
				Type: &backup.BackupCapability_Rpc{
					Rpc: &backup.BackupCapability_RPC{
						Type: backup.BackupCapability_RPC_TYPE_BACKUP,
					},
				},
			},
		},
	}, nil
}

// Backup take a physical backup using Kopia
func (BackupServer) Backup(
	ctx context.Context,
	request *backup.BackupRequest,
) (*backup.BackupResult, error) {
	contextLogger := logging.FromContext(ctx)

	helper, err := pluginhelper.NewDataBuilder(metadata.Data.Name, request.ClusterDefinition).Build()
	if err != nil {
		contextLogger.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	backupObject, err := helper.DecodeBackup(request.BackupDefinition)
	if err != nil {
		contextLogger.Error(err, "Error while decoding backup definition from CNPG")
		return nil, err
	}

	cluster := helper.GetCluster()
	rep, err := repository.NewRepository(
		ctx,
		"s3",
		storage.GetBasePath(cluster.Name),
		storage.GetKopiaConfigFilePath(cluster.Name),
		storage.GetKopiaCacheDirectory(cluster.Name),
	)
	if err != nil {
		return nil, err
	}

	exec := executor.NewLocalExecutor(
		cluster,
		backupObject,
		rep,
	)

	startedAt := time.Now()
	backupInfo, err := exec.Backup(ctx)
	if err != nil {
		return nil, err
	}

	return &backup.BackupResult{
		BackupId:          backupInfo.BackupName,
		BackupName:        backupInfo.BackupName,
		StartedAt:         startedAt.Unix(),
		StoppedAt:         time.Now().Unix(),
		BeginWal:          exec.GetBeginWal(),
		EndWal:            exec.GetEndWal(),
		BeginLsn:          string(backupInfo.BeginLSN),
		EndLsn:            string(backupInfo.EndLSN),
		BackupLabelFile:   backupInfo.LabelFile,
		TablespaceMapFile: backupInfo.SpcmapFile,
		Online:            true,
	}, nil
}
