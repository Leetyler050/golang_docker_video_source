package main

import (
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

// FileInfo holds information about a file for templating.
type FileInfo struct {
	Name string
	Path string
}

// Data for the HTML template.
type TemplateData struct {
	Files []FileInfo
}

// The path to the directory you want to serve.
const dirToServe = "./videos"

// Define the HTML template.
const listTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>File Listing</title>
    <style>
        body { font-family: sans-serif; }
        ul { list-style-type: none; padding: 0; }
        li { margin: 10px 0; }
        a { text-decoration: none; color: #007bff; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>Files in Directory</h1>
    <ul>
        {{range .Files}}
            <li><a href="{{.Path}}">{{.Name}}</a></li>
        {{end}}
    </ul>
</body>
</html>
`

func main() {
	// Create the directory to serve, for demonstration purposes.
	err := os.MkdirAll(dirToServe, 0755)
	if err != nil {
		log.Fatal("Could not create directory:", err)
	}

	// Create some dummy files to be listed.
	os.WriteFile(filepath.Join(dirToServe, "file1.txt"), []byte("Hello, file 1!"), 0644)
	os.WriteFile(filepath.Join(dirToServe, "test.md"), []byte("# Markdown Test"), 0644)

	// Create a new ServeMux for routing.
	mux := http.NewServeMux()

	// Handle requests to the root path ("/") with our custom handler.
	mux.HandleFunc("/", listFilesHandler)

	// Serve the actual files from the directory.
	// `http.StripPrefix` is used so that the `http.FileServer` serves files
	// from the `dirToServe` directory without exposing the directory name in the URL.
	// For example, `http://localhost:8080/file1.txt` will serve `./files/file1.txt`.
	fs := http.FileServer(http.Dir(dirToServe))
	mux.Handle("/videos/", http.StripPrefix("/videos/", fs))
	//mux.Handle("/files/", http.StripPrefix("/files/", fs))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))

	ips, err := GetLocalIPs()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(ips)
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	//test ip restriction
	allowedIP := "192.168.65.1" // Change to allowed IP
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	log.Printf("Connection from IP: %s", clientIP)
	if clientIP != allowedIP {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Only handle requests to the root path.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Read all directory entries from the specified directory.
	entries, err := os.ReadDir(dirToServe)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var files []FileInfo
	// Populate the FileInfo struct for each file.
	for _, entry := range entries {
		// Skip directories and hidden files.
		if !entry.IsDir() && entry.Name()[0] != '.' {
			files = append(files, FileInfo{
				Name: entry.Name(),
				Path: "/videos/" + entry.Name(),
				//Path: "/files/" + entry.Name(),
			})
		}
	}

	// Parse the HTML template.
	t, err := template.New("listing").Parse(listTemplate)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	// Execute the template, writing the output to the HTTP response.
	err = t.Execute(w, TemplateData{Files: files})
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}

// Get preferred outbound ip of this machine
func GetLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips, nil
}
