# AGENTS.md
go1fl-sprint5-final-k-konovalov-copy - /home/kirillkonovalov/Documents/git/itmo/ITMO/itmo_GO/go1fl-sprint5-final-k-konovalov-copy
Тип проекта: код (Go 1.24, фитнес-трекер — Yandex Practicum / ITMO, sprint 5)
Роль ИИ: Аудит/разбор кода — TASK_DESCRIPTION.md, структура проекта, ответы на вопросы по коду

## Документация проекта
- Описание проекта и техническое задание: @docs/TASK_DESCRIPTION.md (модуль фитнес-трекера, 5 пакетов: spentenergy, personaldata, trainings, daysteps, actioninfo).
- План работ: @docs/PLAN.md.
- Критерии оценивания ревью: @docs/REVIEW_POINTS.md.
- Паттерны и подходы к коду: @docs/CODE_PATTERN.md

## Правила
- Не используй субагентов.
- Для кода и навигации по нему не используй grep / find напрямую -> используй codegraph и ast-index.
- Не уверен - спроси пользователя.
- Рассуждай (thinking) на английском, независимо от языка запроса.

## Проект
- Проверка: `go mod tidy && go test -v ./...` (CI: golang:1.24, .github/workflows/unit-tests.yaml).
- Изменять содержимое тестов (_test.go) запрещено.
- Паттерны и подходы к коду: @docs/CODE_PATTERN.md (индекс подходов, чек-лист соответствия).