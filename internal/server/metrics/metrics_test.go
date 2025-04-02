package metrics

import (
	"testing"

	"metrics/internal/storage/memory"
)

var fileStorage *MetricsFileStorage

func init() {
	newMemStorage := memory.NewMemoryStorage()
	fileStorage = &MetricsFileStorage{
		storageCommands: newMemStorage,
	}

}

func BenchmarkInitMetricsFromFile(b *testing.B) {
	fileStorage.fileStorage = "tempFile.txt"

	b.Run("init metrics from file", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if err := fileStorage.InitMetricsFromFile(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkStoreMetrics(b *testing.B) {
	fileStorage.fileStorage = "testFileStorage.txt"

	b.Run("store metrics to file", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if err := fileStorage.StoreMetrics(); err != nil {
				b.Fatal(err)
			}
		}
	})
}
