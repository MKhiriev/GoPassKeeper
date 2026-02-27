package tui

import "sync/atomic"

var sessionUserID int64

func setSessionUserID(userID int64) {
	atomic.StoreInt64(&sessionUserID, userID)
}

func getSessionUserID() int64 {
	return atomic.LoadInt64(&sessionUserID)
}

func clearSessionUserID() {
	atomic.StoreInt64(&sessionUserID, 0)
}
