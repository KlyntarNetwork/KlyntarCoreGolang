package life

var SPAM_FLAG = false

var FINALIZATION_PROOFS_CACHE map[string]string

type ProofsGrabber struct{}

var PROOFS_GRABBER ProofsGrabber

func BlocksSharingAndProofsGrabingThread() {}
