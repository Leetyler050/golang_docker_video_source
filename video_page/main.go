package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileInfo holds information about a file for templating.
type FileInfo struct {
	Name    string
	Path    string
	IsDir   bool
	ModTime string
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
    <title>File and Folder Listing</title>
    <style>
        body { font-family: sans-serif; }
        table { border-collapse: collapse; width: 100%; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f2f2f2; }
        a { text-decoration: none; color: #007bff; }
        a:hover { text-decoration: underline; }
        .folder { color: #ffa500; font-weight: bold; }
        .file { color: #007bff; }
    </style>
</head>
<body>
    <h1>Files and Folders</h1>
    <table>
        <tr>
            <th>Name</th>
            <th>Last Modified</th>
        </tr>
        {{range .Files}}
            <tr>
                <td>
                    {{if .IsDir}}
                        <a class="folder" href="{{.Path}}">📁 {{.Name}}/</a>
                    {{else}}
                        <a class="file" href="{{.Path}}">📄 {{.Name}}</a>
                    {{end}}
                </td>
                <td>{{.ModTime}}</td>
            </tr>
        {{end}}
    </table>
</body>
</html>
`

func main() {
	// Create the directory to serve, for demonstration purposes.
	// err := os.MkdirAll(dirToServe, 0755)
	// if err != nil {
	// 	log.Fatal("Could not create directory:", err)
	// }

	// Create some dummy files to be listed.
	// os.WriteFile(filepath.Join(dirToServe, "file1.txt"), []byte("Hello, file 1!"), 0644)
	// os.WriteFile(filepath.Join(dirToServe, "test.md"), []byte("# Markdown Test"), 0644)

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

	ips, err := GetLocalIPs()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ips)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))

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

	// Get the requested path and construct the full directory path
	requestedPath := strings.TrimPrefix(r.URL.Path, "/")
	fullPath := filepath.Join(dirToServe, requestedPath)

	// Security check: ensure the path doesn't escape dirToServe
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	absDirToServe, _ := filepath.Abs(dirToServe)
	if !strings.HasPrefix(absPath, absDirToServe) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Read all directory entries from the requested directory
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var files []FileInfo
	// Include both files and folders
	for _, entry := range entries {
		// Skip hidden files/folders and macOS metadata files
		if entry.Name()[0] != '.' && !strings.HasPrefix(entry.Name(), "._") {
			info, _ := entry.Info()
			relativePath := filepath.Join(requestedPath, entry.Name())
			files = append(files, FileInfo{
				Name:    entry.Name(),
				Path:    "/videos/" + relativePath,
				IsDir:   entry.IsDir(),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
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
