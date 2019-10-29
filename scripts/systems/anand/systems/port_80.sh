#!/bin/bash
telnet localhost 80 << EOF
HEAD / HTTP/1.1

^]
q
EOF
