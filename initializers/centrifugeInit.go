package initializers

import (
	"log"

	"github.com/centrifugal/centrifuge"
)

var node *centrifuge.Node

func handleLog(e centrifuge.LogEntry) {
	log.Printf("%s: %v", e.Message, e.Fields)
}

func main() {
	node, _ = centrifuge.New(centrifuge.Config{
		LogLevel: centrifuge.LogLevelDebug,
		LogHandler: handleLog,
	})

	
}
