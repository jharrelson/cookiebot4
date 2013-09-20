package main

import (
//	"encoding/json"
	"fmt"
	"io/ioutil"
//	"strings"
)

func LoadConfig() []byte {
	buf, err := ioutil.ReadFile("bot.cfg")
	if err != nil {
		fmt.Errorf(err.Error())
	}

//	decoder := json.NewDecoder(strings.NewReader(string(buf)))

	return buf
}

