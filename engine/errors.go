package engine

import "errors"

var ErrNoCrawlerVisit = errors.New("no visit from crawler")
var ErrInvalidCrawlerState = errors.New("crawler has no current visit")
var ErrUnknownAction = errors.New("unknown action")
