package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Initialize S3 client
var svc *s3.S3

func init() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Change to your desired region
	})
	if err != nil {
		log.Fatalf("failed to create session: %v", err)
	}
	svc = s3.New(sess)
}

// Upload a file to S3
func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Could not get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Specify the bucket name
	bucket := "your-s3-bucket-name" // Change this to your bucket name

	// Upload the file to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("uploads/" + r.FormValue("filename")),
		Body:        file,
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully!")
}

// List files in S3
func listFiles(w http.ResponseWriter, r *http.Request) {
	bucket := "your-s3-bucket-name" // Change this to your bucket name

	// List objects in the bucket
	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	var files []string
	for _, item := range result.Contents {
		files = append(files, *item.Key)
	}

	fmt.Fprintf(w, "Files in bucket:\n%s", strings.Join(files, "\n"))
}

// Delete a file from S3
func deleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	bucket := "your-s3-bucket-name" // Change this to your bucket name

	// Delete the file from S3
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File deleted successfully!")
}

func main() {
	http.HandleFunc("/upload", uploadFile) // POST to upload a file
	http.HandleFunc("/list", listFiles)    // GET to list files
	http.HandleFunc("/delete", deleteFile) // DELETE to remove a file

	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
