package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/pascaldekloe/mboxed"
)

// SplitWriter writes messages to file <key>.mbox, whereby <key> gets defined by
// the KeyFunc. KeyFunc may return an empty string to skip a message. Files are
// not closed.
type SplitWriter struct {
	OutDir     string // base of output files
	KeyFunc    func(msg *mail.Message, raw []byte) string
	KeyEscapes *strings.Replacer
	PerKey     map[string]*bufio.Writer
}

// OnMessage implements mbox.MessageListener.
func (split *SplitWriter) OnMessage(fromLine string, raw []byte, msg *mail.Message) {
	key := split.KeyEscapes.Replace(split.KeyFunc(msg, raw))
	switch key {
	case "", ".", "..":
		log.Printf("%s: %q skipped on unusable key %q", name, fromLine, key)
		return
	}

	w, ok := split.PerKey[key]
	if !ok {
		f, err := os.Create(filepath.Join(split.OutDir, key))
		if err != nil {
			log.Fatalf("%s: %s", name, err)
		}
		// ⚠️ file Close relies on command exit

		w = bufio.NewWriter(f)
		split.PerKey[key] = w
	}

	// ⚠️ error check on Flush
	w.WriteString(fromLine)
	w.Write(raw)
}

// Name the command.
var name = os.Args[0]

// Command Invocation Options
var (
	headerFlag = flag.String("header", "", "Defines the header (`name`) used for file distribution.")
	outDirFlag = flag.String("d", ".", "Sets the `directory` for output files.")
	escapeFlag = flag.String("escape", "_", fmt.Sprintf("Sets the `replacement` for %q occurences in output files.", filepath.Separator))
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	err := os.MkdirAll(*outDirFlag, 0o755)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("%s: %s", name, err)
	}

	split := SplitWriter{
		OutDir:     *outDirFlag,
		KeyEscapes: strings.NewReplacer(string([]rune{filepath.Separator}), *escapeFlag),
		PerKey:     make(map[string]*bufio.Writer),
	}
	if split.OutDir == "" {
		split.OutDir = "."
	}

	switch {
	case *headerFlag != "":
		split.KeyFunc = func(msg *mail.Message, raw []byte) string {
			return msg.Header.Get(*headerFlag)
		}
	default:
		log.Fatalf("%s: no split key defined: use the header flag", name)
	}

	var failN int
	for i := 0; flag.Arg(i) != ""; i++ {
		err := mbox.ReadFile(flag.Arg(i), split.OnMessage)
		if err != nil {
			log.Print(err)
			failN++
		}
	}
	for _, w := range split.PerKey {
		err := w.Flush()
		if err != nil {
			log.Printf("%s: %s", name, err)
			failN++
		}
	}

	os.Exit(failN)
}
