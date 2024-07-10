package isservermode

import "os"

var IsServerMode = len(os.Args) == 2 && os.Args[1] == "__main__"
