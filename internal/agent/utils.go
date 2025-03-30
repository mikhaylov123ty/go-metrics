package agent

// Структура для канала заданий метрик
type metricJob struct {
	data *[]byte
	//urlPath string
}

// Структура для канала ответов заданий метрик
type jobResponse struct {
	worker int
	err    error
}

//type ClientResponse interface {
//	Error() string
//}
//
//type GRPCResponse struct {
//	err error
//}
//
//type HTTPResponse struct {
//	err error
//}
//
////
////func (g GRPCResponse) Error() string {
////	return g.err.Error()
////}
////
////func (h HTTPResponse) Error() string {
////	return h.err.Error()
////}
//
//func (j jobResponse) Error() string {
//	return j.ClientResponse.Error()
//}
