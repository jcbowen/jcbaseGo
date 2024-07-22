package security

import (
	"github.com/jcbowen/jcbaseGo/component/helper"
	"html"
	"regexp"
	"strings"
)

type SanitizeInput struct {
	Value        interface{}
	DefaultValue interface{}
}

var (
	htmlEntityRegex = regexp.MustCompile(`&((#(\d{3,5}|x[a-fA-F0-9]{4}));)`)
	sqlRegex        = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|truncate|alter|exec|;|--)`)
	badStr          = []string{"\000", "%00", "%3C", "%3E", "<?", "<%", "<?php", "{php", "{if", "{foreach", "{for", "../"}
	replacementStr  = []string{"", "", "<", ">", "", "", "", "", "", "", "", ".."}
)

// Belong checks if the value belongs to the allowed list
func (s SanitizeInput) Belong(allow []interface{}, strict bool) interface{} {
	for _, v := range allow {
		if strict && v == s.Value {
			return s.Value
		} else if !strict && (helper.Convert{Value: v}.ToString()) == (helper.Convert{Value: s.Value}.ToString()) {
			return s.Value
		}
	}
	return s.DefaultValue
}

// String sanitizes and returns the string representation of the value
func (s SanitizeInput) String() string {
	val, ok := s.Value.(string)
	if !ok {
		if defVal, ok := s.DefaultValue.(string); ok {
			return defVal
		}
		return ""
	}

	val = s.badStrReplace(val)
	val = htmlEntityRegex.ReplaceAllString(val, "&$1")
	val = sqlRegex.ReplaceAllString(val, "")
	val = html.EscapeString(val)

	if val == "" && s.DefaultValue != "" {
		return s.DefaultValue.(string)
	}
	return val
}

// badStrReplace replaces potentially harmful substrings
func (s SanitizeInput) badStrReplace(str string) string {
	if str == "" {
		return ""
	}
	for i := range badStr {
		str = strings.ReplaceAll(str, badStr[i], replacementStr[i])
	}
	return str
}
