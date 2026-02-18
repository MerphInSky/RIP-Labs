package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Device struct {
	ID          int
	Title       string
	Power_dis   float64
	Photo       string
	Description string
}

func (r *Repository) GetDevices() ([]Device, error) {
	devices := []Device{
		{
			ID:          1,
			Title:       "Микросхема BD9897FS",
			Power_dis:   0.95,
			Photo:       "Num1.jpg",
			Description: "Контроллер управления инвертором подсветки для ЖК-дисплеев в корпусе для поверхностного монтажа SOP-24",
		},
		{
			ID:          2,
			Title:       "Транзистор FDD8447L",
			Power_dis:   44,
			Photo:       "Num2.jpg",
			Description: "N-канальный полевой транзистор для силовой коммутации в корпусе D-PAK для поверхностного монтажа",
		},
		{
			ID:          3,
			Title:       "Транзистор B1261",
			Power_dis:   52,
			Photo:       "GIFKA.gif",
			Description: "Высокотоковый PNP-транзистор для усилителей мощности и схем управления в металлическом корпусе TO-3P",
		},
		{
			ID:          4,
			Title:       "Транзисторная сборка AO4606C",
			Power_dis:   2,
			Photo:       "Num4.jpg",
			Description: "Комплементарная пара полевых транзисторов N- и P-канального типов в компактном корпусе SO-8 для поверхностного монтажа",
		},
		{
			ID:          5,
			Title:       "Стабилитрон BZX55-33V",
			Power_dis:   0.5,
			Photo:       "Num5.jpg",
			Description: "Малосигнальный стабилитрон, напряжение стабилизации 33 В, мощность 500 мВт, допуск ±5%, корпус DO-35 (выводной монтаж) ",
		},
		{
			ID:          6,
			Title:       "Микросхема NE555H",
			Power_dis:   1,
			Photo:       "Num6.jpg",
			Description: "Контроллер импульсного источника питания с ШИМ, токовый режим управления, защита от перегрузки, корпус DIP-8",
		},
		{
			ID:          7,
			Title:       "Транзистор FQPF20N60C",
			Power_dis:   115,
			Photo:       "Numero7.jpg",
			Description: "Мощный биполярный NPN-транзистор общего назначения в металлическом корпусе ТО-3",
		},
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("Массив пустой")
	}

	return devices, nil
}

func (r *Repository) GetDevice(id int) (Device, error) {
	devices, err := r.GetDevices()
	if err != nil {
		return Device{}, err
	}

	for _, device := range devices {
		if device.ID == id {
			return device, nil
		}
	}
	return Device{}, fmt.Errorf("Заказ не найден")
}

func (r *Repository) GetDeviceByTitle(title string) ([]Device, error) {
	devices, err := r.GetDevices()
	if err != nil {
		return []Device{}, err
	}

	var result []Device
	for _, device := range devices {
		if strings.Contains(strings.ToLower(device.Title), strings.ToLower(title)) {
			result = append(result, device)
		}
	}
	return result, nil
}

func (r *Repository) GetCart() ([]Device, error) {
	devices, err := r.GetDevices()
	if err != nil {
		return []Device{}, err
	}

	var result []Device
	for _, device := range devices {
		if device.ID == 3 || device.ID == 7 {
			result = append(result, device)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("Массив пустой")
	}

	return result, nil
}
