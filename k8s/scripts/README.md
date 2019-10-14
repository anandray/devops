# Wolk builds

## Production:

This will pull the latest `wolk` binary from `github.com` and restart the wolk.service all nodes:
```
sudo build/scripts/pushwolkservice
```

This uses the following configuration
```
binary: wolk        
httpport: 80       [DefaultConfig]
p2p: 30300         
datadir: /usr/local/wolk   [DefaultConfig]
toml: wolk.toml    
```

## Pushing Staging: (1..9)

To make parallel development easy, this will push a similar binary (`wolk1` ... `wolk9`) based on an integer 1..9
```
sudo build/scripts/pushstaging 1..9
```
which has a different binary with flags overriding the default config
```
binary: wolk1..wolk9
httpport: 81..89  [passed via command line flag]
p2p: 30301..3309      [passed via command line flag]
tomlfile: wolk1.toml ... wolk9.toml [passed via command line flag]
datadir: /usr/local/wolk1 ... /usr/local/wolk9   [passed via command line flag]
```

The `pushstaging` script will push to the same instances as production, but by virtue of having different ports/binary/toml/datadir we can run independent blockchains.
In principle, 9 different "staging" blockchains can exist for development:
 * Sourabh 1
 * Rodney 2
 * Mayumi 3
 * Alina 4
 * Michael 5
 * Bruce 6
 * Yaron 7
 * Anand 8

The process is:
 1. `make wolk`: compile a new `wolk` binary locally
 2. (if necessary) change `wolk.toml` (which will be in .gitignore)
 3. `pushstaging [stage]`
