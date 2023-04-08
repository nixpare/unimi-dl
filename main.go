package main

import (
	"fmt"
	"log"
	"os"
)

const TEST_PAGE = "https://nbasilicoae2.ariel.ctu.unimi.it/v5/frm3/ThreadList.aspx?fc=BxPlwgRH%2b296WJzdnwYOWDcStR%2fDkD%2fqUMKoIPKGaa2N408CbbDWmWxRudLIAoTV&roomid=227362"

func main() {
	defer func() {
		fmt.Println("\nPress Enter to exit")
		var b []byte
		fmt.Scanf("%s\n", &b)
	}()

	logF, _ := os.OpenFile("unimi-dl.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0777)
	log.SetOutput(logF)

	unimiDL, err := NewUnimiDL(TEST_PAGE)
	if err != nil {
		log.Fatalln(err)
	}

	err = unimiDL.GetAllLectures()
	if err != nil {
		log.Fatalln(err)
	}

	for _, l := range unimiDL.Lectures {
		if len(l.Videos) != 0 {
			fmt.Println(l.Videos[0].manifestURL)
			break
		}
	}
}
