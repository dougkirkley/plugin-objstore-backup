package wal

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/logging"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/wal"

	"github.com/dougkirkley/plugin-objstore-backup/internal/backup/storage"
	"github.com/dougkirkley/plugin-objstore-backup/pkg/metadata"
)

type walStatMode string

const (
	walStatModeFirst = "first"
	walStatModeLast  = "last"
)

// Status gets the statistics of the WAL file archive
func (WAL) Status(
	ctx context.Context,
	request *wal.WALStatusRequest,
) (*wal.WALStatusResult, error) {
	contextLogger := logging.FromContext(ctx)

	helper, err := pluginhelper.NewDataBuilder(metadata.Data.Name, request.ClusterDefinition).Build()
	if err != nil {
		contextLogger.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	walPath := storage.GetWALPath(helper.GetCluster().Name)
	contextLogger = contextLogger.WithValues(
		"walPath", walPath,
		"clusterName", helper.GetCluster().Name,
	)

	walDirEntries, err := os.ReadDir(walPath)
	if err != nil {
		contextLogger.Error(err, "Error while reading WALs directory")
		return nil, err
	}

	firstWal, err := getWALStat(helper.GetCluster().Name, walDirEntries, walStatModeFirst)
	if err != nil {
		contextLogger.Error(err, "Error while reading WALs directory (getting first WAL)")
		return nil, err
	}

	lastWal, err := getWALStat(helper.GetCluster().Name, walDirEntries, walStatModeLast)
	if err != nil {
		contextLogger.Error(err, "Error while reading WALs directory (getting first WAL)")
		return nil, err
	}

	return &wal.WALStatusResult{
		FirstWal: firstWal,
		LastWal:  lastWal,
	}, nil
}

func getWALStat(clusterName string, entries []fs.DirEntry, mode walStatMode) (string, error) {
	entry, ok := getEntry(entries, mode)
	if !ok {
		return "", nil
	}

	if !entry.IsDir() {
		return "", fmt.Errorf("%s is not a directory", entry)
	}

	entryAbsolutePath := path.Join(storage.GetWALPath(clusterName), entry.Name())
	subFolderEntries, err := os.ReadDir(entryAbsolutePath)
	if err != nil {
		return "", fmt.Errorf("while reading %s entries: %w", entry, err)
	}

	selectSubFolderEntry, ok := getEntry(subFolderEntries, mode)
	if !ok {
		return "", nil
	}

	return selectSubFolderEntry.Name(), nil
}

func getEntry(entries []fs.DirEntry, mode walStatMode) (fs.DirEntry, bool) {
	if len(entries) == 0 {
		return nil, false
	}

	switch mode {
	case walStatModeFirst:
		return entries[0], true

	case walStatModeLast:
		return entries[len(entries)-1], true

	default:
		return nil, false
	}
}
