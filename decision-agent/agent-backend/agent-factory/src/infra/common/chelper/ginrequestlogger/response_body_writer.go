package ginrequestlogger

// Write 实现io.Writer接口，同时写入buffer和原始ResponseWriter
func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}
