# cfbench

Benchmark of cf {push,scale}

```
export CF_AUTH_TOKEN=$(cf oauth-token);
DOPPLER_ADDR=$(cf curl /v2/info| jq -r .doppler_logging_endpoint)
go build;

./cfbench  -app-dir=/Users/pivotal/go/src/github.com/cloudfoundry/cf-acceptance-tests/assets/dora -doppler-address=$DOPPLER_ADDR
./cfbench --action scale -instances=10 -app-dir=/Users/pivotal/go/src/github.com/cloudfoundry/cf-acceptance-tests/assets/dora  -doppler-address=$DOPPLER_ADDR
```

Logs are going to stderr
`-json` output is going to stdout
