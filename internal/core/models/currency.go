package models

type Currency struct {
	Code            string `json:"code" db:"code"`                         // ISO 4217, например "RUB"
	Name            string `json:"name" db:"name"`                         // Полное название валюты
	MinorUnitName   string `json:"minor_unit_name" db:"minor_unit_name"`   // "копейка", "цент", "филс"
	MinorUnits      int64    `json:"minor_units" db:"minor_units"`           // Сколько минимальных единиц в основной валюте: 100 или 1000
	IsFractional    bool   `json:"is_fractional" db:"is_fractional"`       // Есть ли дробная часть у валюты (например, у японской иены — нет)
}