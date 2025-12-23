package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Модель отзыва с текстом и оценкой
type Review struct {
	Text   string `json:"text"`
	Rating int    `json:"rating"`
}

// Модель места с названием, описанием, категорией и отзывами
type Place struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Reviews     []Review `json:"reviews"`
}

// Основная модель пользователя с именем и списком мест
type User struct {
	Name   string  `json:"name"`
	Places []Place `json:"places"`
}

var currentUser *User

// Инициализация глобального объекта текущего пользователя
func init() {
	currentUser = &User{}
	loadData(currentUser)
}

// Загрузка данных из файла
func loadData(user *User) {
	file, err := os.Open("places.json")
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Ошибка открытия файла: %v\n", err)
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&user.Places)
	if err != nil && err.Error() != "EOF" {
		log.Printf("Ошибка парсинга JSON: %v\n", err)
	}
}

// Сохранение данных в файл
func saveData(user *User) {
	file, err := os.Create("places.json")
	if err != nil {
		log.Fatalf("Ошибка создания файла: %v\n", err)
		return
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(user.Places)
	if err != nil {
		log.Fatalf("Ошибка сохранения данных: %v\n", err)
	}
}

// Создание нового места
func createPlace(c *fiber.Ctx) error {
	var place Place
	err := c.BodyParser(&place)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ошибка парсинга тела"})
	}

	for _, existingPlace := range currentUser.Places {
		if existingPlace.Name == place.Name {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "место с таким названием уже существует"})
		}
	}

	currentUser.Places = append(currentUser.Places, place)
	saveData(currentUser)
	return c.JSON(fiber.Map{"message": "место успешно создано"})
}

// Просмотр всех мест
func listPlaces(c *fiber.Ctx) error {
	return c.JSON(currentUser.Places)
}

// Редактирование описания места
func updatePlaceDescription(c *fiber.Ctx) error {
	idStr := c.Params("id")
	index, err := parseID(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "невозможно преобразовать ID"})
	}

	if index >= len(currentUser.Places) || index < 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "место не найдено"})
	}

	description := c.Query("description")
	if description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "описание не задано"})
	}

	currentUser.Places[index].Description = description
	saveData(currentUser)
	return c.JSON(fiber.Map{"message": "описание обновлено"})
}

// Удаление места
func deletePlace(c *fiber.Ctx) error {
	idStr := c.Params("id")
	index, err := parseID(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "невозможно преобразовать ID"})
	}

	if index >= len(currentUser.Places) || index < 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "место не найдено"})
	}

	currentUser.Places = append(currentUser.Places[:index], currentUser.Places[index+1:]...)
	saveData(currentUser)
	return c.JSON(fiber.Map{"message": "место удалено"})
}

// Парсер индекса
func parseID(s string) (int, error) {
	i, err := fmt.Sscanf(s, "%d", new(int))
	if err != nil || i != 1 {
		return 0, fmt.Errorf("некорректный формат id")
	}

	num := s[:len(s)-1]                 // Здесь получаем строку без последнего символа
	parsedNum, err := strconv.Atoi(num) // Конвертируем строку в число
	if err != nil {
		return 0, fmt.Errorf("не удалось конвертировать ID в число")
	}

	return parsedNum, nil
}

// Настройка маршрутов
func setupRoutes(app *fiber.App) {
	app.Post("/places", createPlace)
	app.Get("/places", listPlaces)
	app.Put("/places/:id", updatePlaceDescription)
	app.Delete("/places/:id", deletePlace)

	// Добавляем обработчик для корня "/"
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Добро пожаловать на мой сервис мест!")
	})
}

// Запуск серверного приложения
func main() {
	app := fiber.New()
	setupRoutes(app)
	fmt.Println("Запущен сервер на порте :8082")
	log.Fatal(app.Listen(":8082"))
}
