// Package daysteps предоставляет функционал для разбора строк с данными о прогулках
// и формирования отформатированной информации о дневной ходьбе.
package daysteps

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Yandex-Practicum/tracker/internal/personaldata"
	"github.com/Yandex-Practicum/tracker/internal/spentenergy"
)

// DaySteps хранит данные о дневной прогулке: количество шагов,
// длительность, а также персональные данные пользователя.
type DaySteps struct {
	Steps int
	Duration time.Duration
	personaldata.Personal
}

// Parse разбирает строку формата "678,0h50m" и записывает извлечённые
// значения (шаги и длительность) в структуру DaySteps.
// Возвращает ошибку, если формат данных некорректен или значения неположительны.
func (ds *DaySteps) Parse(datastring string) (err error) {
	if len(datastring) == 0 {
		return errors.New("empty daysteps data")
	}

	slice := strings.SplitN(datastring, ",", 2)
	if len(slice) != 2 {
		return fmt.Errorf("invalid daysteps format: expected 2 fields (steps, type, duration), got %d", len(slice))
	}

	// Steps
	steps, parseStepErr := strconv.Atoi(slice[0])
	if parseStepErr != nil {
		return fmt.Errorf("invalid steps value %q: %w", slice[0], parseStepErr)
	}
	if steps <= 0 {
		return fmt.Errorf("steps must be positive, got %d", steps)
	}
	ds.Steps = steps

	// Duration
	duration, parseDurationErr := time.ParseDuration(slice[1])
	if parseDurationErr != nil {
		return fmt.Errorf("invalid duration %q: %w", slice[1], parseDurationErr)
	}
	if duration.Milliseconds() <= 0 {
		return errors.New("duration must be positive")
	}
	ds.Duration = duration

	return nil
}

// ActionInfo формирует и возвращает отформатированную строку с информацией о прогулке,
// включая количество шагов, дистанцию и количество сожжённых калорий.
// Возвращает ошибку при сбое расчётов или некорректных входных значениях.
func (ds DaySteps) ActionInfo() (string, error) {
	if ds.Steps <= 0 || ds.Duration.Milliseconds() <= 0 {
		return "", errors.New("steps and duration must be positive")
	}
	if ds.Weight <= 0 || ds.Height <= 0 {
		return "", errors.New("weight and height must be positive")
	}

	steps := strconv.FormatInt(int64(ds.Steps), 10)
	distance := strconv.FormatFloat(spentenergy.Distance(ds.Steps, ds.Height), 'f', 2, 64)
	caloriesFloat, calErr := spentenergy.WalkingSpentCalories(ds.Steps, ds.Weight, ds.Height, ds.Duration)
	if calErr != nil {
		return "", fmt.Errorf("failed to calculate calories: %w", calErr)
	}
	calories := strconv.FormatFloat(caloriesFloat, 'f', 2, 64)

	str := fmt.Sprintf(
		strings.Join(
			[]string{
				"Количество шагов: %s.\n",
				"Дистанция составила %s км.\n",
				"Вы сожгли %s ккал.\n",
			},
			"",
		),
		steps, distance, calories,
	)

	return str, nil
}
