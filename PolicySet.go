package main

func (instance PolicySet) validate(request HTTPRequest) bool {
	policies, err := instance.Policies()

	if err != nil {
		panic(err)
	}

	if policies.Len() == 0 {
		return true
	}

	for i := 0; i < policies.Len(); i++ {

		if policies.At(i).validate(request) {
			return true
		}
	}

	return false
}
