package translator

import (
	"fmt"
	"time"
)

func Translate(onTranslate func(string) string) {
	go func() {
		time.Sleep(2 * time.Second) // Simulate a delay for translation
		text := "Translated text"
		if onTranslate != nil {
			text = onTranslate(text)
		}
		fmt.Println(text) // Output the translated text
	}()
}
