package main

import (
	"html/template"
	"log"
)

var temps *template.Template

// parsing views
// if any error exit the application
func parseTemplate(pattern string) *template.Template {
	temp, err := template.ParseGlob(pattern)
	if err != nil {
		log.Fatal(err)
	}
	// for later assign to global variable
	return temp
}
