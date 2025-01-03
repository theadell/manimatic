package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type Register struct {
	stringVars []*string
	intVars    []*int
	boolVars   []*bool
}

func (r *Register) String(ptr *string, name, usage string, defValue string) {
	*ptr = defValue
	if envVal := os.Getenv(name); envVal != "" {
		*ptr = envVal
	}
	flag.StringVar(ptr, strings.ToLower(strings.ReplaceAll(name, "_", "-")), *ptr, usage)
	r.stringVars = append(r.stringVars, ptr)
}

func (r *Register) Int(ptr *int, name, usage string, defValue int) {
	*ptr = defValue
	if envVal := os.Getenv(name); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			*ptr = val
		}
	}
	flag.IntVar(ptr, strings.ToLower(strings.ReplaceAll(name, "_", "-")), *ptr, usage)
	r.intVars = append(r.intVars, ptr)
}

func (r *Register) Bool(ptr *bool, name, usage string, defValue bool) {
	*ptr = defValue
	if envVal := os.Getenv(name); envVal != "" {
		if val, err := strconv.ParseBool(envVal); err == nil {
			*ptr = val
		}
	}
	flag.BoolVar(ptr, strings.ToLower(strings.ReplaceAll(name, "_", "-")), *ptr, usage)
	r.boolVars = append(r.boolVars, ptr)
}
