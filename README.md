# Mboxed

… library plus tools for mbox files, written in the Go programming
language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


## mboxmux

The mboxmux(1) tool can split mailboxes based on header values. Run
`go install github.com/pascaldekloe/mboxed/cmd/mboxmux@latest` for
a local build.

```
Usage of mboxmux:
  -d directory
    	Set the directory for output files. (default ".")
  -default file-name
    	Sets a default output file-name for messages that would have been omitted otherwise, which are no name, . and .. specifically.
  -escape replacement
    	Sets the replacement for '/' occurences in output files. (default "_")
  -header name
    	Define the header (name) used for file distribution.
  -tokentrim pattern
    	Add a pattern for token omission on the output files. The first character in the pattern defines the token separator, and the remainder sets the token to be excluded. E.g., -tokentrim ,Opened omits any Opened occurences in a comma-separated list, i.e., Inbox,Opened,Important would become Inbox,Important. Multiple tokentrim arguments are applied in conjuntion.
```

The following command splits a “takeout” from GMail into separate
mailboxes.

    mboxmux -d mailboxes -header X-Gmail-Labels -default Misc \
    	-tokentrim ',Opened' -tokentrim ',Unread' \
    	-tokentrim ',Archived' \
    	-tokentrim ',Category Personal' \
    	-tokentrim ',Category Promotions' \
    	-tokentrim ',Category Purchases' \
    	-tokentrim ',Category Social' \
    	-tokentrim ',Category Travel' \
    	-tokentrim ',Category Updates' \
    	-tokentrim ',IMAP_Forwarded' -tokentrim ',IMAP_$Forwarded' \
    	-tokentrim ',IMAP_Redirected' \
    	-tokentrim ',IMAP_NotJunk' -tokentrim ',IMAP_$NotJunk' \
    	-tokentrim ',IMAP_Junk' -tokentrim ',IMAP_$Junk' \
    	-tokentrim ',IMAP_JunkRecorded' \
    	-tokentrim ',IMAP_$MailFlagBit0' \
    	-tokentrim ',IMAP_$MailFlagBit1' \
    	-tokentrim ',IMAP_$MailFlagBit2' \
    	bu1.mbox bu2.mbox bu3.mbox
