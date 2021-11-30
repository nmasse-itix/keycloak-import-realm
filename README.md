# Keycloak Realm Generator & Importer

This tool is used to load data into Keycloak before running a performance test.

## Usage

### Generation

Generate 5 realms, each one containing 10 clients and 100 users.
Generated files are saved in the current directory.

```sh
kci generate --realms 5 --clients 10 --users 100 
```

Same as above but save files in the specified directory.

```sh
kci generate --realms 5 --clients 10 --users 100 --target realms/
```

Generate realms using the provided template.

```sh
kci generate --realms 5 --clients 10 --users 100 --realm my.template
```

### Import

Configure your target Keycloak instance.

```sh
kci config set realm --value master
kci config set login --value admin
kci config set password --value S3cr3t
kci config set keycloak_url --value http://localhost:8080/auth
```

Import the previously generated realms.

```sh
kci import *.json
```

By default, 5 workers are used to speed up the loading process.
You can change this with:

```sh
kci config set workers --value 10
```

## Container image

An up-to-date container image is built by a Tekton pipeline and pushed to [quay.io/itix/kci](https://quay.io/repository/itix/kci?tab=tags).
Images are built for x86_64, armv5 and arm64.
Each tag in the git repository ends up being a tag in the Quay registry, `latest` being the continuous build of the `main` branch.

To use this tool from the command line with podman, you would need to define the following function in your .bashrc / .zshrc.

```sh
function kci () {
    touch ~/.kci.yaml
    podman run --rm -v $HOME/.kci.yaml:/.kci.yaml:z -v $PWD:/home:z -w /home quay.io/itix/kci:latest "$@"
}
```
