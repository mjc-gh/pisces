package engine

import "errors"

var ErrNoCrawlerVisit = errors.New("no visit from crawler")
var ErrUnknownAction = errors.New("unknown action")
