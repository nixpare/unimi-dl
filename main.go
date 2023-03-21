package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	defer func() {
		fmt.Println("\nPress Enter to exit")
		var b []byte
		fmt.Scanf("%s\n", &b)
	}()

	pageFile, err := os.Open("page2.html")
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(pageFile)
	if err != nil {
		panic(err)
	}

	lectures, err := FindAllLectures(string(data))
	if err != nil {
		panic(err)
	}

	for _, l := range lectures {
		fmt.Println(l)
		fmt.Println()
	}
}
