package main

import (
	"strings"
	"regexp"
)

func WildcardCompare(s1 string, s2 string) bool {
        s1 = strings.ToLower(s1)
        s2 = strings.ToLower(s2)

        s1 = createRegex(s1)
        regex := regexp.MustCompile(s1)

        return regex.MatchString(s2)
}

func createRegex(s string) string {
        var wc string
        for i := 0; i < len(s); i++ {
                if s[i] == '*' {
                        wc += ".*"
                } else if s[i] == '?' {
                        wc += "."
                } else if s[i] >= '0' && s[i] <= '9' {
                        wc += string(s[i])
                } else if s[i] >= 'a' && s[i] <= 'z' {
                        wc += string(s[i])
                } else {
                        wc += "\\" + string(s[i])
                }
        }

        return "^" + wc + "$"
}
