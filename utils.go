package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ExtractTextFromHTML(selection *goquery.Selection) string {
	textPlit := strings.Split(
		strings.ReplaceAll(selection.Text(), "\t", " ",),
		" ",
	)
	
	result := make([]string, 0)
	for _, s := range textPlit {
		if s == "" {
			continue
		}
		result = append(result, strings.TrimSpace(s))
	}

	return strings.Join(result, " ")
}