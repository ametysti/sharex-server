package main

import (
	"io"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	if _, err := os.Stat("files"); os.IsNotExist(err) {
		slog.Info("Files directory doesnt exists. Creating")
		os.Mkdir("files", 0755)
	}

	godotenv.Load(".env")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Post("/sharex", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(0)

		if !strings.HasPrefix(r.Header.Get("User-Agent"), "ShareX/") {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed parsing multipart form", 500)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Println(err)
			http.Error(w, "did not receive any file. note that url shorteners are unsupported", 406)
			return
		}
		defer file.Close()

		randomNumber := rand.Intn(4-1+1) + 1

		fileLocation := fmt.Sprintf("%s.%s", generateRandomString(randomNumber), header.Filename)

		slog.Debug(fmt.Sprintf("writing file to files/%s", fileLocation))
		dst, err := os.Create("files/" + fileLocation)
		io.Copy(dst, file)

		if err != nil {
			slog.Error(err.Error())
			w.WriteHeader(500)
			w.Write([]byte("failed to write"))

			return
		}

		slog.Debug(fmt.Sprintf("sending file output link to client: https://%s/%s", os.Getenv("URL"), fileLocation))
		w.Write([]byte(fmt.Sprintf("https://%s/%s", os.Getenv("URL"), fileLocation)))
	})

	port := "127.0.0.1:3001"

	if _, err := os.Stat("/.dockerenv"); err == nil {
		port = ":3000"
	}

	slog.Info("running")
	http.ListenAndServe(port, r)

}
