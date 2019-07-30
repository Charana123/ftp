package utils

import (
	"log"
)

func HandleFatalError(preFunction func(), errs ...error) {
	for _, err := range errs {
		if err != nil {
			if preFunction != nil {
				preFunction()
			}
			log.Panic("error: " + err.Error())
			return
		}
	}
}

func HandleWarning(preFunction func(), errs ...error) bool {
	for _, err := range errs {
		if err != nil {
			if preFunction != nil {
				preFunction()
			}
			log.Println("error: " + err.Error())
			return true
		}
	}
	return false
}
