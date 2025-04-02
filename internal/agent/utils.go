package agent

// Структура для канала заданий метрик
type metricJob struct {
	data *[]byte
}

// Структура для канала ответов заданий метрик
type jobResponse struct {
	worker int
	err    error
}
