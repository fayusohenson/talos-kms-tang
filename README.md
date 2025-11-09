# talos-kms-tang

[Talos](https://github.com/siderolabs/talos) KMS proxy server for [Tang](https://github.com/latchset/tang)

## Usage

```
./talos-kms-tang -tang-endpoint host[:port]
```

## Additional Options

```
$ ./talos-kms-tang --help
Usage of ./talos-kms-tang:
  -kms-api-endpoint string
    	gRPC API endpoint for the KMS (default ":4050")
  -tang-endpoint string
    	tang server endpoint
  -tang-thumbprint string
    	thumbprint of a trusted signing key
  -tls-cert-path string
    	path to TLS certificate file
  -tls-enable
    	whether to enable tls or not
  -tls-key-path string
    	path to TLS private key file
```
