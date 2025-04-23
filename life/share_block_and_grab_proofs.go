package life

var SPAM_FLAG = false

var FINALIZATION_PROOFS_CACHE = make(map[string]map[string]string)

type ProofsGrabber struct{}

var PROOFS_GRABBER ProofsGrabber

func BlocksSharingAndProofsGrabingThread() {}
