package bcspcp

import "os"

var ProtocolVersionHeader = []byte("BCSPCP/1.0\n")

const SockFilePerm = os.FileMode(0700)
const DefaultSockDir = ".bcspcp"
const DefaultSockName = "bcspcp.sock"
const NetworkName = "unix"
