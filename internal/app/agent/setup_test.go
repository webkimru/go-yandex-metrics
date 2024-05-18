package agent

import (
	"github.com/webkimru/go-yandex-metrics/internal/app/server/grpc"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories/store"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	db := store.NewFakeStorage()

	repoGRPC := grpc.NewRepo(db)
	grpc.NewMetricHandlers(repoGRPC)

	os.Exit(m.Run())
}
