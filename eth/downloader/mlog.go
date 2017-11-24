package downloader

import "github.com/ethereumproject/go-ethereum/logger"

var mlogDownloader = logger.MLogRegisterAvailable("downloader", mLogLines)

var mLogLines = []logger.MLogT{
	mlogDownloaderRegisterPeer,
	mlogDownloaderUnregisterPeer,
}

var mlogDownloaderRegisterPeer= logger.MLogT{
	Description: "Called when a block announcement is discarded.",
	Receiver:    "DOWNLOADER",
	Verb:        "REGISTER",
	Subject:     "PEER",
	Details: []logger.MLogDetailT{
		{"PEER", "ID", "STRING"},
		{"PEER", "VERSION", "INT"},
		{"REGISTER", "ERROR", "STRING_OR_NULL"},
	},
}

var mlogDownloaderUnregisterPeer= logger.MLogT{
	Description: "Called when a block announcement is discarded.",
	Receiver:    "DOWNLOADER",
	Verb:        "UNREGISTER",
	Subject:     "PEER",
	Details: []logger.MLogDetailT{
		{"PEER", "ID", "STRING"},
		{"UNREGISTER", "ERROR", "STRING_OR_NULL"},
	},
}