package proxy

import "errors"

var (
	ErrorQueryFlag           = errors.New("the --query flags must take the form of key=value (= not found)")
	ErrorEmptyQueryFlag      = errors.New("the --query flag must take the form of: key=value (empty value given, or value ends in =)")
	ErrorUnauthorizedGateway = errors.New("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
)
