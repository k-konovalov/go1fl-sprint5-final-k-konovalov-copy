package personaldata

import "fmt"

// Данные пользователя
type Personal struct {
	Name string
	Weight float64
	Height float64
}


func (p Personal) Print() {
	str := fmt.Sprintf(
		"Имя: %s\nВес: %.2f\nРост: %.2f", 
		p.Name, p.Weight, p.Height,
	)
	fmt.Println(str)
}
