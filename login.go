package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"syscall"

	"golang.org/x/term"
)

const (
	LOGIN_URL = "https://elearning.unimi.it/authentication/skin/portaleariel/login.aspx?url=https://ariel.unimi.it/"
	LOGIN_CONTENT_TYPE = "application/x-www-form-urlencoded"
	LOGIN_PAYLOAD_FORMAT = "hdnSilent=true&tbLogin=%s&tbPassword=%s&ddlType="
)

func performLogin(c *http.Client) error {
	fmt.Print("Enter Email: ")
    var email string
	fmt.Scanln(&email)

    fmt.Print("Enter Password: ")
    bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
    if err != nil {
		log.Printf("Password error: %v\n", err)
		return errors.New("An error has occurred while reading the password")
    }
    password := string(bytePassword)
	
	payload := fmt.Sprintf(
		LOGIN_PAYLOAD_FORMAT,
		url.QueryEscape(strings.TrimSpace(email)),
		url.QueryEscape(strings.TrimSpace(password)),
	)

	resp, err := c.Post(LOGIN_URL, LOGIN_CONTENT_TYPE, strings.NewReader(payload))
	if err != nil || resp.StatusCode != 200 {
		log.Printf("Login error: %v\n\tHTTP response: %v\n", err, resp)
		return errors.New("An error has occurred while trying to login")
	}

	return nil
}

func getPage(c *http.Client, url string) (string, error) {
	resp, err := c.Get(url)
	if err != nil || resp.StatusCode >= 400 {
		log.Printf("Login error: %v\n\tHTTP response: %v\n", err, resp)
		return "", errors.New("An error has occurred while getting Ariel home page")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Login error: %v\n\tHTTP response: %v\n", err, resp)
		return "", errors.New("An error has occurred while reading response body")
	}

	return string(body), nil
}
