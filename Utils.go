package main

//CheckError takes an error, checks if it is nil and if not panics
//This is a deliberate verbose handling for now whilst the solution is prototyped.
func CheckError(err error) {
	//Stay verbose for the time being
	if err != nil {
		panic(err)
	}
}

//ArrayContainsString checks if a string array contains a string value
func ArrayContainsString(input []string, value string) bool {
	for _, current := range input {
		if current == value {
			return true
		}
	}
	return false
}
