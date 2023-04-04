# Github Issue Label Syncer

## Usage

```console
export ACCESS_TOKEN=<GH_ACCESS_TOKEN>
go run main.go --from-owner=aws --to-owner=aws --from-repo=karpenter --to-repo=karpenter-core --access-token=$ACCESS_TOKEN
```