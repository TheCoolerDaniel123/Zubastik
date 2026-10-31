package e

import "fmt"

// WrapIfErr - обертка для обработки ошибок
func WrapIfErr(err error, msg string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}
