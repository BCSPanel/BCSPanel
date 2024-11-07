package bcspcp

import "os"

const SockFilePerm = os.FileMode(0700)
const DefaultSockDir = ".bcspcp"
const DefaultSockName = "bcspcp.sock"
const NetworkName = "unix"
