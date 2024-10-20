package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"net/http"
)

const bucket = "go-bucket123" // Change this to your bucket name

// Root handler for the "/" route
func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the AWS S3 Go API! Use the following endpoints:\n")
	fmt.Fprintf(w, "GET /list - List files\n")
	fmt.Fprintf(w, "POST /upload - Upload a file\n")
	fmt.Fprintf(w, "DELETE /delete?key=filename - Delete a file\n")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Create a new S3 session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Change to your region
	})
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		log.Println("Error creating session:", err)
		return
	}

	svc := s3.New(sess)

	// Upload the file
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})

	if err != nil {
		http.Error(w, "Unable to upload file", http.StatusInternalServerError)
		log.Println("Error uploading file:", err) // Log the actual error
		return
	}

	fmt.Fprintf(w, "File uploaded successfully: %s", filename)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Change to your region
	})

	svc := s3.New(sess)

	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		http.Error(w, "Unable to list files", http.StatusInternalServerError)
		return
	}

	for _, item := range result.Contents {
		fmt.Fprintf(w, "File: %s\n", *item.Key)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	// Allow both GET and DELETE methods for testing purposes
	if r.Method != http.MethodDelete && r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the 'key' parameter from the URL query string
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	log.Println("Received key:", key)

	// Create a new S3 session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Change to your region
	})
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		log.Println("Error creating session:", err)
		return
	}

	svc := s3.New(sess)

	// Check if the object exists before trying to delete
	_, err = svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if the error is a "Not Found" error (404)
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			http.Error(w, "File not found", http.StatusNotFound)
			log.Println("File not found:", key)
		} else {
			http.Error(w, "No such file exists", http.StatusInternalServerError)
			log.Println("No such file exists:", err)
		}
		return
	}

	// Proceed to delete the file
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		http.Error(w, "Unable to delete file", http.StatusInternalServerError)
		log.Println("Error deleting file:", err)
		return
	}

	// Wait until the object is deleted
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		http.Error(w, "Error waiting for object deletion", http.StatusInternalServerError)
		log.Println("Error waiting for object deletion:", err)
		return
	}

	fmt.Fprintf(w, "File deleted successfully: %s", key)
}

func main() {
	// Handlers for the routes
	http.HandleFunc("/", rootHandler) // Root route handler
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/delete", deleteHandler)

	// Start the server
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
