<?php
#
# Copyright (c) 2010-2012 Curu Wong (http://www.pineapple.com.hk/blog/curu)
# PNP template for check_mailstat.pl plugin
#
$_LINE     = '#000000';
#
# Initial Logic ...
#

foreach ($this->DS as $KEY=>$VAL) {
        $vlabel   = "";
        $lower    = "";
        $upper    = "";

        $VAL['UNIT'] = $VAL['UNIT'] ? $VAL['UNIT'] : 'msgs/min';

        if ($VAL['UNIT'] == "%%") {
                $vlabel = "%";
                $upper = " --upper=101 ";
                $lower = " --lower=0 ";
        }
        else {
                $vlabel = $VAL['UNIT'];
        }

        $opt[$KEY] = '--vertical-label "' . $vlabel . '" --title "' . $this->MACRO['DISP_HOSTNAME'] . ' / ' . $this->MACRO['DISP_SERVICEDESC'] . '"' . $upper . $lower;
        $ds_name[$KEY] = $VAL['LABEL'];
        $def[$KEY]  = rrd::def     ("var1", $VAL['RRDFILE'], $VAL['DS'], "AVERAGE");
        $def[$KEY] .= rrd::gradient("var1", "BDC6DE", "3152A5", rrd::cut($VAL['NAME'],16), 20);
        $def[$KEY] .= rrd::line1   ("var1", $_LINE );
        $def[$KEY] .= rrd::gprint  ("var1", array("LAST","MAX","AVERAGE"), "%.2lf ".$VAL['UNIT']);
}
?>
