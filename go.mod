// +heroku goVersion go1.13
// +heroku install ./cmd/...
module github.com/meddion/web-blog

go 1.13

require (
	github.com/gabriel-vasile/mimetype v1.0.4
	github.com/gorilla/mux v1.7.4
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.10.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.mongodb.org/mongo-driver v1.3.0
	golang.org/x/crypto v0.0.0-20200219234226-1ad67e1f0ef4
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20200219091948-cb0a6d8edb6c // indirect
)
