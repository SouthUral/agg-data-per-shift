package utils

import (
	"fmt"
)

// функция оборачиватель ошибок
func Wrapper(errFirst, errSecond error) error {
	return fmt.Errorf("%w: %w", errFirst, errSecond)
}
