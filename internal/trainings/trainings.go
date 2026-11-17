package trainings

import (
    "errors"
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
	if (len([]rune(datastring)) == 0) {
		return errors.New("")
	}
	// to slice
	slice := strings.SplitN(datastring, ",", 3)
	if (len(slice) != 3) {
		return errors.New("")
	}
	// Steps
	steps, parseStepErr := strconv.Atoi(slice[0])
	if parseStepErr != nil {
	 	return errors.Join(
			errors.New(""),
			parseStepErr,
		)
	}
	t.Steps = steps
	// TrainingType
	t.TrainingType = slice[1]
	// Duration
	duration, parseDurationErr := time.ParseDuration(slice[3])
	if parseDurationErr != nil {
		return errors.New("")
	}
	t.Duration = duration

	return nil
}

func (t Training) ActionInfo() (string, error) {
	// TODO: реализовать функцию
}
