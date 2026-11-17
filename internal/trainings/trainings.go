package trainings

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Yandex-Practicum/tracker/internal/personaldata"
	"github.com/Yandex-Practicum/tracker/internal/spentenergy"
)

type Training struct {
	Steps int
	TrainingType string
	Duration time.Duration
	personaldata.Personal
}

func (t *Training) Parse(datastring string) (err error) {
	if len(datastring) == 0 {
		return errors.New("empty training data")
	}

	slice := strings.SplitN(datastring, ",", 3)
	if len(slice) != 3 {
		return fmt.Errorf("invalid training data format: expected 3 fields (steps, type, duration), got %d", len(slice))
	}

	// Steps
	steps, parseStepErr := strconv.Atoi(slice[0])
	if parseStepErr != nil {
		return fmt.Errorf("invalid steps value %q: %w", slice[0], parseStepErr)
	}
	if steps <= 0 {
		return fmt.Errorf("steps must be positive, got %d", steps)
	}
	t.Steps = steps

	// TrainingType
	t.TrainingType = slice[1]

	// Duration
	duration, parseDurationErr := time.ParseDuration(slice[2])
	if parseDurationErr != nil {
		return fmt.Errorf("invalid duration %q: %w", slice[2], parseDurationErr)
	}
	if duration.Milliseconds() <= 0 {
		return errors.New("duration must be positive")
	}
	t.Duration = duration

	return nil
}

func (t Training) ActionInfo() (string, error) {
	distance := strconv.FormatFloat(spentenergy.Distance(t.Steps, t.Height), 'f', 2, 64)
	duration := strconv.FormatFloat(t.Duration.Hours(), 'f', 2, 64)
	speed := strconv.FormatFloat(spentenergy.MeanSpeed(t.Steps, t.Height, t.Duration), 'f', 2, 64)

	var calories string
	var err error

	switch t.TrainingType {
	case "Бег":
		var spentCalories float64
		spentCalories, err = spentenergy.RunningSpentCalories(t.Steps, t.Weight, t.Height, t.Duration)
		if err != nil {
			return "", fmt.Errorf("failed to calculate running calories: %w", err)
		}
		calories = strconv.FormatFloat(spentCalories, 'f', 2, 64)
	case "Ходьба":
		var spentCalories float64
		spentCalories, err = spentenergy.WalkingSpentCalories(t.Steps, t.Weight, t.Height, t.Duration)
		if err != nil {
			return "", fmt.Errorf("failed to calculate walking calories: %w", err)
		}
		calories = strconv.FormatFloat(spentCalories, 'f', 2, 64)
	default:
		return "", fmt.Errorf("unknown training type %q", t.TrainingType)
	}

	str := fmt.Sprintf(
		strings.Join(
			[]string{
				"Тип тренировки: %s\n",
				"Длительность: %s ч.\n",
				"Дистанция: %s км.\n",
				"Скорость: %s км/ч\n",
				"Сожгли калорий: %s\n",
			},
			"",
		),
		t.TrainingType, duration, distance, speed, calories,
	)

	return str, nil
}
