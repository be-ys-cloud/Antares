# Antares

Antares is a tool who can synchronize an HashiCorp Vault server, and a KeePass file.

In a nutshell, you can import your KeePass secrets to a Hashicorp Vault instance, and export Vault secrets to an offline Keepass.

## Configuration

### Arguments 

The most of the configuration is made through CLI. When exporting, you must also provide a JSON file who will include the Vault paths you want to export.

Configuration arguments:
* Mandatory:
  * `-import` or `-export` : Define the action you want to do. Import will import secrets from KeePass to Vault, export will export Vault secret to a new KeePass file.
  * `-password=<your_password>` : The password of your KeePass database. If you are exporting secrets, the KeePass file will be created with this password.
    * This argument can also be provided through `ANTARES_KEEPASS_PASSWORD` environment variable.
  * `-vaultToken=<you_vault_token>` : The Vault token corresponding to your account. You can get one by simply log-in into the Vault UI.
    * This argument can also be provided through `ANTARES_VAULT_TOKEN` environment variable.
  * `-vaultServer=<your_vault_server>` : The address of your Vault instance.
* Optional :
  * `-keepassFile=<path_to_your_keepass.kdbx>` : The input/output KeePass location. 
    * By default, this value is set to `./vault-keepass.kdbx`
  * `-configFile=<path_to_your_configuration_file.kdbx>` : The configuration file, as described below.
    * By default, this value is set to `./configuration.json`

### Configuration File (export only)

The configuration file allows you to specify which paths of your Vault you want to export. If a secret ends with a `/`, the sub-secrets will also be retrieved.

For example:
  
```json 
[
  {
    "engine": "gitlab",
    "paths": [
      "/project1/",
      "/project2/secret"
    ]
  },
  {
    "engine": "home",
    "paths": [
      "/"
    ]
  }
]
```

With this configuration file, the tool will save in your KeePass:
* All secrets in `project1` folder in `gitlab` engine;
* The specific secret named `project2/secret` in `gitlab` engine;
* All secrets that will be find in `home` engine.


## Build

* `go mod download`
* `go build -o antares ./cmd`
* `go run ./antares <your_arguments>`

## License

This project is released under MIT license.