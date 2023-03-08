package service

func CheckNumberByLuhn(number string) bool {
	var d int
	if len(number)%2 == 1 {
		d = 1
	}
	var sum int
	for i, c := range number {
		if c < '0' || c > '9' {
			return false
		}
		n := int(c - '0')
		if i%2 == d {
			if n*2 > 9 {
				sum = sum + n*2 - 9
			} else {
				sum = sum + n*2
			}
		} else {
			sum = sum + n
		}
	}
	return sum%10 == 0
}
