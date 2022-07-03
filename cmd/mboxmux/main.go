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
		if *defaultFlag != "" {
			key = *defaultFlag
		} else {
			log.Printf("%s: %q skipped on output-file name %q; see -default option", name, fromLine[:len(fromLine)-2], key)
			return
		}
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
	headerFlag  = flag.String("header", "", "Define the header (`name`) used for file distribution.")
	outDirFlag  = flag.String("d", ".", "Set the `directory` for output files.")
	escapeFlag  = flag.String("escape", "_", fmt.Sprintf("Sets the `replacement` for %q occurences in output files.", filepath.Separator))
	defaultFlag = flag.String("default", "", "Sets a default output `file-name` for messages that would have been omitted otherwise, which are no name, . and .. specifically.")
)

var tokenTrims []string

func main() {
	log.SetFlags(0)
	flag.Func("tokentrim", "Add a `pattern` for token omission on the output files. The first character in the pattern defines the token separator, and the remainder sets the token to be excluded. E.g., -tokentrim ,Opened omits any Opened occurences in a comma-separated list, i.e., Inbox,Opened,Important would become Inbox,Important. Multiple tokentrim arguments are applied in conjuntion.", func(s string) error {
		tokenTrims = append(tokenTrims, s)
		return nil
	})
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
			s := msg.Header.Get(*headerFlag)
			for _, pattern := range tokenTrims {
				s = trimToken(s, pattern)
			}
			return s
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

func trimToken(s, exclude string) string {
	if exclude == "" {
		return s
	}
	separator := exclude[:1]
	token := exclude[1:]

	// fast-path for token absense
	if strings.Index(s, token) < 0 {
		return s
	}

	tokens := strings.Split(s, separator)
	for i := len(tokens) - 1; i > 0; i-- {
		if strings.TrimSpace(tokens[i]) == token {
			// delete
			tokens = append(tokens[:i], tokens[i+1:]...)
		}
	}
	return strings.Join(tokens, separator)
}
