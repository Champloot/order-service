package ports

type HTTPServer interface {
	Start(addr string) error
}