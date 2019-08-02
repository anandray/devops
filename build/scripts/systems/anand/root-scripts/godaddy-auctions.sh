#!/bin/bash
  
for i in {1..58}
do
        #crosschannel
        watch -n1 curl -s "https://auctions.godaddy.com/trpItemListing.aspx?miid=245825961" &> /dev/null;

        #mdotm.com
        watch -n1 curl -s "https://auctions.godaddy.com/trpItemListing.aspx?miid=245891077" &> /dev/null;

        #eetuh.com
        watch -n1 curl -s "https://auctions.godaddy.com/trpItemListing.aspx?miid=245825982" &> /dev/null;

        #freegreatapps.com
        watch -n1 curl -s "https://auctions.godaddy.com/trpItemListing.aspx?miid=245825992" &> /dev/null;

        #rodta.com
        watch -n1 curl -s "https://auctions.godaddy.com/trpItemListing.aspx?miid=245825967" &> /dev/null;

        #trackinstall.com
        watch -n1 curl -s "https://auctions.godaddy.com/trpItemListing.aspx?miid=245825978" &> /dev/null;

        sleep 1;
done
