package main

import capnp "zombiezen.com/go/capnproto2"

//EXTENSION METHODS FOR THE GENERATED CODE!

//KeyValuePolicyListFromMap creates a new KeyValuePolicy_List from a map
//TODO: Instead of checking the errors let them propagate up the chain
func KeyValuePolicyListFromMap(input map[string][]string) (KeyValuePolicy_List, error) {
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	CheckError(err)

	keyValueList, err := NewKeyValuePolicy_List(seg, int32(len(input)))
	CheckError(err)

	count := 0
	for key, values := range input {
		keyValuePolicy, err := NewKeyValuePolicy(seg)
		CheckError(err)

		keyValuePolicy.SetKey(key)

		valueList, err := capnp.NewTextList(seg, int32(len(values)))
		CheckError(err)

		for index, Value := range values {
			valueList.Set(index, Value)
		}

		keyValuePolicy.SetValues(valueList)
		keyValueList.Set(count, keyValuePolicy)
		count++
	}

	return keyValueList, nil
}
