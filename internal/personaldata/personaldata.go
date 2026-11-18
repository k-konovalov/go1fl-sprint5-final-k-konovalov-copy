// Package personaldata хранит персональные данные пользователя: имя, вес и рост.
package personaldata

import "fmt"

// Personal хранит персональные данные пользователя: имя, вес и рост.
type Personal struct {
	Name string
	Weight float64
	Height float64
}

// Print выводит персональные данные пользователя (имя, вес и рост) в stdout.
func (p Personal) Print() {
	str := fmt.Sprintf(
		"Имя: %s\nВес: %.2f\nРост: %.2f", 
		p.Name, p.Weight, p.Height,
	)
	fmt.Println(str)
}
