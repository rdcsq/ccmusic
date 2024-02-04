package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

var (
	listeningAddress string
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func main() {
	loadEnv()

	http.HandleFunc("/get", handleCcRequest)

	log.Println("Starting server on", listeningAddress)
	http.ListenAndServe(listeningAddress, nil)
}

func loadEnv() {
	// load .env if its available
	godotenv.Load()

	listeningAddress = os.Getenv("LISTENING_ADDRESS")
	if listeningAddress == "" {
		listeningAddress = "127.0.0.1:3000"
	}
}

// generates a random 32 character string
func generateRandomString() string {
	s := make([]rune, 32)
    for i := range s {
        s[i] = letters[rand.Intn(len(letters))]
    }
    return string(s)
}

func handleCcRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// get url
	urlParam := r.URL.Query().Get("url")
	if urlParam == "" {
		log.Println("ERR - Invalid url:", urlParam)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Downloading", urlParam)

	fileName := generateRandomString()

	// get resulting file name
	ytdlpOutput := filepath.Join(os.TempDir(), fileName + ".%(ext)s")
	ytdlpResult, err := exec.Command("yt-dlp", "-f", "ba", urlParam, "-o", ytdlpOutput, "--get-filename").Output()
	if err != nil {
		log.Println("ERR - Couldn't run yt-dlp:", err)
	}

	// download
	if err := exec.Command("yt-dlp", "-f", "ba", urlParam, "-o", ytdlpOutput).Run(); err != nil {
		log.Println("ERR - Couldn't download url:", urlParam)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	audioFilePath := strings.Replace(string(ytdlpResult), "\n", "", -1)
	defer os.Remove(audioFilePath)

	// convert
	outFile := filepath.Join(os.TempDir(), fileName + ".dfpwm")
	cmd := exec.Command("ffmpeg", "-i", audioFilePath, "-ac", "1", "-c:a", "dfpwm", "-ar", "48k", outFile)
	if out, err := cmd.Output(); err != nil {
		log.Println("ERR - Couldn't convert:", urlParam)
		log.Println("ffmpeg log:", out)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, err := os.Open(outFile)
	if err != nil {
		log.Println("ERR - Couldn't open file:", urlParam)
		w.WriteHeader(http.StatusInternalServerError)
		os.Remove(outFile)
		return
	}
	defer file.Close()
	defer os.Remove(outFile)
	
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", "attachment; filename=output.dfpwm")
	io.Copy(w, file)
}