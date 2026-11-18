// Package actioninfo предоставляет единый интерфейс для разбора и вывода
// информации о тренировках и прогулках.
package actioninfo

import (
	"fmt"
	"log"
)

// DataParser — интерфейс для разбора данных и формирования информации
// о тренировочных и прогулочных активностях. Типы, реализующие этот интерфейс,
// должны предоставлять методы Parse() и ActionInfo().
type DataParser interface {
	Parse(datastring string) (err error)
	ActionInfo() (string, error)
}

// Info перебирает срез строк с данными, разбирает каждую с помощью переданного
// DataParser и выводит отформатированную информацию об активности.
// Ошибки при разборе или формировании информации логируются и пропускаются.
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
