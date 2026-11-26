// Command metadata asks which source line supplied a field and whether that
// occurrence survived the merge.
package main

import (
	"fmt"

	"github.com/adambarber/goconf"
)

type (
	region string
	port   int
)

type config struct {
	Region region `json:"region"`
	Port   port   `json:"port"`
}

const appName goconf.Name = "app"

func main() {
	_, _, meta, err := goconf.LoadBytesWithMeta[config](
		[]byte("{\n  \"region\": \"eu\",\n  \"port\": 8080\n}"),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
	)
	if err != nil {
		panic(err)
	}

	occurrences := meta.ByPath(goconf.OccurrencePath("port"))
	fmt.Println(occurrences[0].Path, occurrences[0].Line, occurrences[0].Effective)
}
