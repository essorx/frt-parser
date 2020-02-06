package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
	pdf "github.com/unidoc/unipdf/v3/model"
)

type Summary struct {
	Id      string `json:"id"`
	Make    string `json:"make"`
	Model	string `json:"model"`
	Manufacturer    string `json:"manufacturer"`
	Type            string `json:"type"`
	Action                  string `json:"action"`
	ManufacturerCountry     string `json:"manufacturer_country"`
	LegalClassification     string `json:"legal_classification"`
}

var (
	summaries []Summary

	manufacturers   map[string]struct{}
	types           map[string]struct{}
	actions         map[string]struct{}
	models          map[string][]string
	classifications map[string]struct{}
)

func walkFunc(path string, info os.FileInfo, inErr error) (err error) {
	if !strings.HasSuffix(strings.ToLower(path), ".pdf") {
		//fmt.Println("ignoring", path)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return err
	}

	method := pdfReader.GetEncryptionMethod()
	if len(method) != 0 {
		_, err = pdfReader.Decrypt([]byte{})
		if err != nil {
			return err
		}
	}

	page, err := pdfReader.GetPage(1)
	if err != nil {
		return err
	}

	ex, err := extractor.New(page)
	if err != nil {
		return err
	}

	text, err := ex.ExtractText()
	if err != nil {
		return err
	}

	lines := strings.Split(text, "\n")

	summary := Summary{}

	for _, line := range lines {
		tokens := strings.Split(line, ": ")
		if len(tokens) < 2 {
			continue
		}

		switch tokens[0] {
		case "Firearm Reference No.":
			summary.Id = tokens[1]
		case "Make":
			summary.Make = tokens[1]
		case "Model":
			summary.Model = tokens[1]
		case "Manufacturer":
			summary.Manufacturer = tokens[1]
		case "Type":
			summary.Type = tokens[1]
		case "Action":
			summary.Action = tokens[1]
		case "Country of Manufacturer":
			summary.ManufacturerCountry = tokens[1]
		case "Legal Classification":
			summary.LegalClassification = tokens[1]
		}
	}


	summaries = append(summaries, summary)

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(os.Args[0], "<directory>")
		os.Exit(1)
	}

	// For debugging.
	// common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))

	root := os.Args[1]

	err := filepath.Walk(root, walkFunc)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	data, err := json.Marshal(summaries)

	fmt.Println(string(data))
}

