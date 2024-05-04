package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

var tpl *template.Template
var str string

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/ascii-art", processor)
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))
	fmt.Println("HTTP SERVER RUNNING AT: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
// 500 Is a status code for internal error
func ascii_art(argument string, fonts string) (string, int) {
	// Read the banner file based on the specified font
	banner, err := ioutil.ReadFile("fonts/" + fonts + ".txt")
	if err != nil {
		return "Error 500\nInternal Error", 500
	}

	split := strings.Split(string(banner), "\n")
	if fonts == "thinkertoy" {
		split = strings.Split(string(banner), "\r\n")
	}

	myting := strings.Split(strings.ReplaceAll(argument, "\r", ""), "\n") // Splitting the input text by "\\n"
	
	for word := 0; word < len(myting); word++ {
		if word == 0 && len(myting) >= 3 {
			// Skip empty lines at the beginning of the input
			if len(myting[0]) == 0 && len(myting[1]) == 0 && len(myting[2]) == 0 {
				word += 1
			}
		}

		for k := 0; k < 8; k++ {
			// Skip processing if the line is empty and there are more lines
			if len(myting[word]) == 0 && len(myting) >= 2 {
				k = 7
			}

			for i := 0; i < len(myting[word]); i++ {
				str += split[(int(myting[word][i])-32)*9+1+k] // Generating ASCII art by mapping characters to the appropriate lines in the font file
			}

			if len(myting[word]) != 0 {
				str += "\n" // Add a newline after processing each line of the input text
			}

			if len(myting[word]) == 0 && len(myting) >= 2 {
				str += "\n"

				// Check for a new line which in this case is a backslash n (" \n")
				if len(myting) == 2 && word != len(myting)-1 {
					if len(myting[word+1]) == 0 {
						word++
					}
				}
			}
		}
	}

	return str, 200
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

// Render function processes the given string and returns the result along with a status code
func render(s string) (string, int) {
	noerr, _ := errorcheck(s)
	
	if len(s) >= 128 {
		return "Too long", 400
	}

	if s == "" {
		return "Enter a text!", 200
	}

	if noerr {
		result := strings.ReplaceAll(s,"\r\n", "\n")
		result = strings.ReplaceAll(result, "\r", "\n")
		return result, 200
	} else {
		return "Bad request", 400
	}
}

// Errorcheck function checks if the given string contains only ASCII characters
// and returns a boolean indicating the result and a status code
func errorcheck(s string) (bool, int) {
	a := []rune(s)
	for i, _ := range s {
		if a[i] <= 127 {
			continue
		} else {
			return false, 400
		}
	}
	return true, 200
}

// Processor function handles the HTTP requests to the "/ascii-art" endpoint
func processor(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		text := r.FormValue("ascii-data")
		fonts := r.FormValue("fonts")
		final, renderstatus := render(text)
		str = ""

		if fonts == "standard" || fonts == "shadow" || fonts == "thinkertoy" {
			if renderstatus == 400 {
				fmt.Printf("%s did a bad request (400)\nWith the text: %s\n", r.RemoteAddr, text)
				data, _ := ascii_art(final, fonts)
				d := struct {
					First string
				}{
					First: data,
				}
				w.WriteHeader(http.StatusBadRequest)
				tpl.ExecuteTemplate(w, "index.html", d)
				return
			}

			data, statuscode := ascii_art(final, fonts)
			if statuscode == 500 {
				fmt.Printf("%s got internal error (500) from %s\n", r.RemoteAddr, text)
				data, _ := ascii_art(final, fonts)
				d := struct {
					First string
				}{
					First: data,
				}
				w.WriteHeader(http.StatusInternalServerError)
				tpl.ExecuteTemplate(w, "index.html", d)
				return
			}
			if statuscode == 200 {
				fmt.Printf("%s sent the text: %s\nWith the font: %s\n", r.RemoteAddr, text, fonts)
				d := struct {
					First string
				}{
					First: data,
				}
				tpl.ExecuteTemplate(w, "index.html", d)
				return
			}

		} else {
			fmt.Printf("New connection from %s\n", r.RemoteAddr)
			tpl.ExecuteTemplate(w, "index.html", nil)
			return
		}
	}
	w.WriteHeader(http.StatusBadRequest)
	tpl.ExecuteTemplate(w, "500.html", nil)
	return
}


func index(w http.ResponseWriter, r *http.Request) {
	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(userAgent, "curl") {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	if r.URL.Path != "/ascii-art" && r.URL.Path != "/" {
		fmt.Printf("%s got a 404 with the path: %s\n", r.RemoteAddr, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		tpl.ExecuteTemplate(w, "404.html", nil)
		return
	}

	if r.Method == "POST" {
		// Handle POST request to the root path ("/")
		fmt.Printf("%s made a POST request to the root path\n", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		tpl.ExecuteTemplate(w, "400.html", nil)
		return
	}

	// Handle GET request to the root path ("/")
	fmt.Printf("New connection from %s\n", r.RemoteAddr)
	tpl.ExecuteTemplate(w, "index.html", nil)
}
