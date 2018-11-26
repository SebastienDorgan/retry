package retry

//FalseCondition utility function that always return false
func FalseCondition(v interface{}, e error) bool {
	return false
}
