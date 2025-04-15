package models

// People представляет структуру данных пользователя
// @Description Обогащённые данные о человеке
type People struct {
	ID          int    `json:"id" example:"1"`                            // Уникальный идентификатор
	Name        string `json:"name" example:"Dmitriy"`                    // Имя
	Surname     string `json:"surname" example:"Ushakov"`                 // Фамилия
	Patronymic  string `json:"patronymic,omitempty" example:"Vasilevich"` // Отчество (опционально)
	Age         int    `json:"age" example:"30"`                          // Возраст (из внешнего API)
	Gender      string `json:"gender" example:"male"`                     // Пол (из внешнего API)
	Nationality string `json:"nationality" example:"RU"`                  // Национальность (из внешнего API)
}
