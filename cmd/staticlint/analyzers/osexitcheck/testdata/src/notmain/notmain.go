package notmain

import "os"

func Run() {
	os.Exit(0) // no error expected - not in main package
}
