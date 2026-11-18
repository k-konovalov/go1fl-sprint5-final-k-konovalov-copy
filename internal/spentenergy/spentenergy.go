package spentenergy

import (
	"errors"
	"time"
)

// Основные константы, необходимые для расчетов.
const (
	mInKm                      = 1000 // количество метров в километре.
	minInH                     = 60   // количество минут в часе.
	stepLengthCoefficient      = 0.45 // коэффициент для расчета длины шага на основе роста.
	walkingCaloriesCoefficient = 0.5  // коэффициент для расчета калорий при ходьбе.
)

// Функция принимает количество шагов steps, рост пользователя height и продолжительность активности duration и возвращает количество калорий, которые потребляет пользователь шагом за прошедшее время.
func WalkingSpentCalories(steps int, weight, height float64, duration time.Duration) (float64, error) {
	runningSpentCalories, err := RunningSpentCalories(steps, weight, height, duration)
	if (err != nil) {
		return 0.0, err
	}
	spentCalories := runningSpentCalories * walkingCaloriesCoefficient
	return spentCalories, nil
}

// Функция принимает количество шагов steps, рост пользователя height и продолжительность активности duration и возвращает количество калорий, которые потребляет пользователь бегом за прошедшее время.
func RunningSpentCalories(steps int, weight, height float64, duration time.Duration) (float64, error) {
	isParametersNotValid := weight <= 0.0 || height <= 0.0 || steps <= 0 || duration.Milliseconds() <= 0
	if isParametersNotValid {
		return 0.0, errors.New("Параметры не могут быть меньше нуля")
	}
	spentCalories := (weight * MeanSpeed(steps, height, duration) * duration.Minutes()) / minInH
	return spentCalories, nil
}

// MeanSpeed рассчитывает и возвращает среднюю скорость в км/ч
// по количеству шагов, росту пользователя и длительности активности.
// Если длительность равна нулю или отрицательна, возвращает 0.
func MeanSpeed(steps int, height float64, duration time.Duration) float64 {
	if (duration.Milliseconds() <= 0) {
		return 0
	}
	avgSpeed := Distance(steps, height) / duration.Hours()
	return avgSpeed
}

// Distance рассчитывает и возвращает дистанцию в километрах
// по количеству шагов и росту пользователя в метрах.
func Distance(steps int, height float64) float64 {
	return (float64(steps) * stepLengthCoefficient * height ) / mInKm 
}
