package ports

//go:generate mockgen -source=server.go -destination=../mocks/mock_server.go -package=mocks

type HTTPServer interface {
	Start(addr string) error
}