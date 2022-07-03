The following command distrubites a GMail “takeout” into separate mailboxes.

    mboxmux -d mailboxes -header X-Gmail-Labels -default Misc \
    	-tokentrim ',Opened' -tokentrim ',Unread' \
    	-tokentrim ',Archived' \
    	-tokentrim ',Category Personal' \
    	-tokentrim ',Category Promotions' \
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
