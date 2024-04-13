package repository

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/logging"

	"github.com/dougkirkley/plugin-objstore-backup/internal/backup/provider"
)

const (
	PGDataLocation    = "/var/lib/postgresql/data/pgdata"
	TablespacesFolder = "pg_tblspc"
	WALFolder         = "pg_wal"
)

// Repository represents a backup repository where
// base directories are stored
type Repository struct {
	provider       string
	path           string
	cacheDirectory string
	configFile     string
}

// NewRepository creates a new repository in a certain
// path, ensuring that the repository is initialized and
// ready to accept backups
func NewRepository(ctx context.Context, p string, path string, configFile string, cacheDirectory string) (*Repository, error) {
	result := &Repository{
		provider:       p,
		path:           path,
		configFile:     configFile,
		cacheDirectory: cacheDirectory,
	}

	if !provider.Validate(p) {
		return nil, fmt.Errorf("provider not valid: %s", p)
	}

	return result, nil
}

func (repo *Repository) initializeRepository(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	args := []string{
		"kopia",
		"Repository",
		"create",
		repo.provider,
		fmt.Sprintf("--path=%s", repo.path),
		fmt.Sprintf("--config-file=%s", repo.configFile),
		fmt.Sprintf("--log-dir=%s/log", repo.cacheDirectory),
		fmt.Sprintf("--cache-directory=%s", repo.cacheDirectory),
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...) // nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(
			err,
			fmt.Sprintf("Error invoking kopia create %s command", repo.provider),
			"args", args,
			"output", string(output))
		return err
	}

	return repo.configureIgnoreFolders(ctx)
}

func (repo *Repository) configureIgnoreFolders(ctx context.Context) error {
	if err := repo.addIgnoreFolder(ctx, path.Join(PGDataLocation, WALFolder)); err != nil {
		return err
	}

	if err := repo.addIgnoreFolder(ctx, path.Join(PGDataLocation, TablespacesFolder)); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) addIgnoreFolder(ctx context.Context, folder string) error {
	logger := logging.FromContext(ctx)

	args := []string{
		"kopia",
		"policy",
		"set",
		folder,
		fmt.Sprintf("--log-dir=%s/log", repo.cacheDirectory),
		"--add-ignore=.",
		fmt.Sprintf("--config-file=%s", repo.configFile),
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...) // nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(
			err,
			"Error invoking kopia policy set command",
			"args", args,
			"output", string(output))
		return err
	}

	return nil
}

// Snapshot takes a Kopia snapshot of a certain path, adding a set of tags
func (repo *Repository) Snapshot(ctx context.Context, path string, tags map[string]string) error {
	logger := logging.FromContext(ctx)

	args := []string{
		"kopia",
		"snapshot",
		"create",
		fmt.Sprintf("--log-dir=%s/log", repo.cacheDirectory),
		fmt.Sprintf("--config-file=%s", repo.configFile),
		path,
	}

	tagsOption := ""
	for k, v := range tags {
		if len(tagsOption) > 0 {
			tagsOption += ","
		}
		tagsOption += fmt.Sprintf("%s:%v", k, v)
	}

	if len(tagsOption) > 0 {
		args = append(args, "--tags="+tagsOption)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...) // nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(
			err,
			"Error invoking kopia snapshot create command",
			"args", args,
			"output", string(output))
		return err
	}

	return nil
}
