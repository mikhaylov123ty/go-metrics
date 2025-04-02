package collector

import (
	"testing"
)

func BenchmarkCollectMetrics(b *testing.B) {
	buf := CollectMetrics
	stats := &Stats{}

	b.Run("collect metrics", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			buf(stats)
		}
	})
}

func BenchmarkBuildMetrics(b *testing.B) {
	buf := CollectMetrics
	stats := &Stats{}
	buf(stats)

	b.Run("build metrics", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			stats.BuildMetrics()
		}
	})
}
