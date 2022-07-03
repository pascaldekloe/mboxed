The following command distrubites a GMail “takeout” dump into separate mailboxes.

    mboxmux -d mailboxes -header X-Gmail-Labels -default Misc \
    	-tokentrim ',Opened' -tokentrim ',Unread' \
    	-tokentrim ',IMAP_Forwarded' -tokentrim ',IMAP_NotJunk' \
    	-tokentrim ',IMAP_$Forwarded' -tokentrim ',IMAP_$NotJunk' \
    	-tokentrim ',IMAP_$MailFlagBit0' -tokentrim ',IMAP_$MailFlagBit1' \
	bu1.mbox bu2.mbox bu3.mbox
