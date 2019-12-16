user { 'antara':
  ensure             => 'present',
#  gid                => 1002,
  home               => '/home/testuser',
  password           => '$6$bSclpF2g$p6HaylOYb35..1wj39vUB7vE7wUPmOEXaASbwicM40.s1JtYPeEzRVi949WTwfvsKUc4.AbJ68soDPkodlgFH/',
  password_max_age   => 99999,
  password_min_age   => 0,
  password_warn_days => 7,
  shell              => '/bin/bash',
#  uid                => 1001,
}
