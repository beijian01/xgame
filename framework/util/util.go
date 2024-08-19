package util

import (
	"fmt"
)

func Try(tryFn func(), catchFn func(errString string)) bool {
	var hasException = true
	func() {
		defer catchError(catchFn)
		tryFn()
		hasException = false
	}()
	return hasException
}

func catchError(catch func(errString string)) {
	if r := recover(); r != nil {
		catch(fmt.Sprint(r))
	}
}

// StringIn checks given string in string slice or not.
func StringIn(v string, sl []string) (int, bool) {
	for i, vv := range sl {
		if vv == v {
			return i, true
		}
	}
	return 0, false
}
