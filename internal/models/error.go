package models

// ErrorResponse стандартная структура ошибки
// @Description Ошибка выполнения запроса
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`                // Код ошибки (HTTP status)
	Message string `json:"message" example:"Некорректный ID"` // Сообщение об ошибке
}
