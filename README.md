# ins-bsc-tx-tool

A testing tool for sending tx to Inscription or BSC with one or multiple bls key signing payload 

### Setup config

1. Edit `config/config.json` and input your Inscription or BSC private key and BLS private key(s)
2. Fill in either `ins_config` or `bsc_config` within `config/config.json` depends on which chain want to send tx to


### Build
```shell script
make build
````


### Run

Send tx to Inscription
```shell script
./build/ins-bsc-tx-tool --tx-chain ins
````
Send tx to BSC
```shell script
./build/ins-bsc-tx-tool --tx-chain bsc
````
