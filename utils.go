package main

import (
	"fmt"
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

func IndentMultilineString(a any, nSpaces int) string {
	var spacePrefix string
	for i := 0; i < nSpaces; i++ {
		spacePrefix += " "
	}

	split := strings.Split(fmt.Sprint(a), "\n")
	for i := 1; i < len(split); i++ {
		split[i] = spacePrefix + split[i]
	}

	return strings.Join(split, "\n")
}