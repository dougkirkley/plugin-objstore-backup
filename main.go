package main

import (
	"fmt"
	"os"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/backup"
	"github.com/cloudnative-pg/cnpg-i/pkg/operator"
	"github.com/cloudnative-pg/cnpg-i/pkg/wal"
	"google.golang.org/grpc"

	backupImpl "github.com/dougkirkley/plugin-objstore-backup/internal/backup"
	"github.com/dougkirkley/plugin-objstore-backup/internal/identity"
	operatorImpl "github.com/dougkirkley/plugin-objstore-backup/internal/operator"
	walImpl "github.com/dougkirkley/plugin-objstore-backup/internal/wal"
)

func main() {
	cmd := pluginhelper.CreateMainCmd(identity.Identity{}, func(server *grpc.Server) {
		operator.RegisterOperatorServer(server, operatorImpl.Operator{})
		wal.RegisterWALServer(server, walImpl.WAL{})
		backup.RegisterBackupServer(server, backupImpl.BackupServer{})
	})
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
