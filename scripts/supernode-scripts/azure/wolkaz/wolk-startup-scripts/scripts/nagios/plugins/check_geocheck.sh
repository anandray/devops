#!/bin/bash
#lynx --dump http://`hostname`/ads/systems/check.php?check=geo
curl http://`hostname`/ads/systems/check.php?check=geo
