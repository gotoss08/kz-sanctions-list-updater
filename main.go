package main

import (
	"bytes"
	"fmt"
	"github.com/jlaffaye/ftp"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func setup_logger() {

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	// Create a MultiWriter that writes to both os.Stdout and the logFile
	multi_writer := io.MultiWriter(os.Stdout, file)

	log.SetOutput(multi_writer)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

}

func fetch_sanction_terrorist_list(token string, status string) []byte {

	base_url := "https://api.websfm.kz/v1/sanctions/sanction-terrorist"

	// Prepare query parameters
	params := url.Values{}
	params.Add("category", "sanction_category_1")
	params.Add("type", "person")
	params.Add("status", status)
	params.Add("offset", "0")
	params.Add("limit", "999999999")

	full_url := base_url + "?" + params.Encode()

	log.Printf("TOKEN: %s", token)
	log.Printf("GET %s", full_url)

	// Create a new HTTP GET request
	req, err := http.NewRequest(http.MethodGet, full_url, nil)
	if err != nil {
		log.Panicf("Error creating request: %v", err)
	}

	req.Header.Set("X-Api-Key", token)

	client := &http.Client{}

	// Make GET request
	resp, err := client.Do(req)
	if err != nil {
		log.Panicf("Error executing request: %v", err)
	}
	defer resp.Body.Close()

	// Read full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("Error reading response body: %v", err)
	}

	return body

}

func connect_to_ftp(host string, user string, pass string) *ftp.ServerConn {

	conn, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = conn.Login(user, pass)
	if err != nil {
		log.Fatal(err)
	}

	return conn

}

func disconnect_from_ftp(conn *ftp.ServerConn) {

	if err := conn.Quit(); err != nil {
		log.Fatal(err)
	}

}

func main() {

	setup_logger()

	token := os.Getenv("API_WEBSFM_KZ_TOKEN")

	ftp_host := os.Getenv("FTP_HOST")
	ftp_user := os.Getenv("FTP_USER")
	ftp_pass := os.Getenv("FTP_PASS")

	statuses := []string{"acting"}

	ftp_conn := connect_to_ftp(ftp_host, ftp_user, ftp_pass)
	defer disconnect_from_ftp(ftp_conn)

	for _, status := range statuses {

		body := fetch_sanction_terrorist_list(token, status)
		ftp_file_name := fmt.Sprintf("test-%s.json", status)

		err := ftp_conn.Stor(ftp_file_name, bytes.NewReader(body))
		if err != nil {
			log.Printf("Error storing file \"%s\" to ftp: %s", ftp_file_name, err)
		}

	}
}
