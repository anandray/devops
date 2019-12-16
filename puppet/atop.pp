include 'agents/*.pp'

node default { }

package { "atop":
  ensure => "present"
}
