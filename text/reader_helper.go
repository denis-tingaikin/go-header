package text

import "log"

//LengthNotEqual return func for reading specific string by length
func LengthNotEqual(n int) func(rune) bool {
	if n < 0 {
		log.Fatal("value should not be negative for func LengthNotEqual")
	}
	i := 0
	return func(_ rune) bool {
		result := i == n
		i++
		return !result
	}
}
