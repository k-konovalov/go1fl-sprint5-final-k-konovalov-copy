package actioninfo

import (
	"fmt"
	"log"
)

type DataParser interface {
	Parse(datastring string) (err error)
	ActionInfo() (string, error)
}

func Info(dataset []string, dp DataParser) {
	for _, data := range dataset {
		parseErr := dp.Parse(data)
		if parseErr != nil {
			log.Printf("failed to parse data: %v", parseErr)
			continue
		}
		str, actionErr := dp.ActionInfo()
		if actionErr != nil {
			log.Printf("failed to get action info: %v", actionErr)
			continue
		}
		fmt.Println(str)
	}

}
