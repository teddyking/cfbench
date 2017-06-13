## CFTrace

  Benchmark cf push 

### How to use

```
export CF_AUTH_TOKEN=$(cf oauth-token); export DOPPLER_ADDR=$(cf curl /v2/info| jq -r .doppler_logging_endpoint)
go build main.go -o ~/bin/cftrace

# either
cftrace --app-dir ~/workspace/cf-acceptance-tests/assets/dora
# or
cd ~/workspace/cf-acceptance-tests/assets/dora
cftrace 
```
