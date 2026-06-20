package validator

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var once sync.Once

func Init() {
	once.Do(func() {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}

		v.RegisterValidation("alphaspace", func(fl validator.FieldLevel) bool {
			// only letters (upper+lower) and spaces allowed
			return regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(fl.Field().String())
		})

		v.RegisterValidation("date_format", func(fl validator.FieldLevel) bool {
			_, err := time.Parse("2006-01-02", fl.Field().String())
			return err == nil
		})

		v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
			return strings.TrimSpace(fl.Field().String()) != ""
		})

	})
}
