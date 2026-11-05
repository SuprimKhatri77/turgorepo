package validator

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var once sync.Once

// personNamePattern allows letters and combining marks in any script, separated
// by spaces or the punctuation found in real names (O'Brien, Mary-Jane, J. Smith).
var personNamePattern = regexp.MustCompile(`^[\p{L}\p{M}]+(?:[ '’.\-]+[\p{L}\p{M}]+)*\.?$`)

func Init() {
	once.Do(func() {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}

		_ = v.RegisterValidation("person_name", func(fl validator.FieldLevel) bool {
			return personNamePattern.MatchString(fl.Field().String())
		})

		// max_bytes limits encoded length. bcrypt rejects passwords over 72 bytes,
		// which max= can't catch because it counts characters.
		_ = v.RegisterValidation("max_bytes", func(fl validator.FieldLevel) bool {
			limit, err := strconv.Atoi(fl.Param())
			if err != nil {
				return false
			}
			return len(fl.Field().String()) <= limit
		})

		_ = v.RegisterValidation("date_format", func(fl validator.FieldLevel) bool {
			_, err := time.Parse("2006-01-02", fl.Field().String())
			return err == nil
		})

		_ = v.RegisterValidation("not_blank", func(fl validator.FieldLevel) bool {
			return strings.TrimSpace(fl.Field().String()) != ""
		})
	})
}
