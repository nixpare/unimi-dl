package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eiannone/keyboard"
)

const TEST_PAGE = "https://nbasilicoae2.ariel.ctu.unimi.it/v5/frm3/ThreadList.aspx?fc=BxPlwgRH%2b296WJzdnwYOWDcStR%2fDkD%2fqUMKoIPKGaa2N408CbbDWmWxRudLIAoTV&roomid=227362"

func main() {
	defer func() {
		keyEvents, err := keyboard.GetKeys(50)
		if err != nil {
			return
		}

		fmt.Println("\nPress Enter to exit")
		for e := range keyEvents {
			if e.Err != nil {
				continue
			}

			if e.Key == keyboard.KeyEnter {
				break
			}
		}
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
			l.Videos[0].Download(unimiDL.Client, "test")
			break
		}
	}
}
