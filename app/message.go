package main

import (
  "regexp"
  "strings"
)

var rxEmail = regexp.MustCompile(".+@.+\\..+")

type Message struct {
  Email string
  FirstName string
  OtherFirstName string
  OtherLastName string
  OtherEmail string
  Uniqueid string
  Content string
  Errors  map[string]string
}
type MessagePost struct {
  Email []string
  FirstName string
  OtherFirstName string
  OtherLastName string
  OtherEmail []string
  Uniqueid string
  Content string
  Errors  map[string]string
  Url string
  hidden string
}
func (msg *MessagePost) Validate() bool {
  msg.Errors = make(map[string]string)

  match := rxEmail.Match([]byte(strings.Join(msg.Email,"")))
  if match == false {
    msg.Errors["Email"] = "Please enter a valid email address"
  }

  if strings.TrimSpace(msg.Content) == "" {
    msg.Errors["Content"] = "Please enter a message"
  }

  return len(msg.Errors) == 0
}
