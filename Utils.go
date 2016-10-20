package main

func CheckError(err error) {

	//Stay verbose for the time being
	if err != nil {
		panic(err)
	}
}
