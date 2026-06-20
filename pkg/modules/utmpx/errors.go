package utmpx

import "errors"

var (
	errParseUtmpx    = errors.New("error parsing utmpx")
	errReadUtmpxFile = errors.New("error reading utmpx file")
	errWriteRecord   = errors.New("simulated write failure")
)
