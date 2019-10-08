package code

import (
	"fmt"
	"github.com/pkg/errors"
)

func Exec() error {
	fmt.Println("Hello Code")
	return errors.WithMessage(fmt.Errorf("this is error"), "Test")
}
