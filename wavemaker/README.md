## Wavemaker

Wavemaker is a tool that generates dummy pods in waves to validate the handling of pod churn on a cluster

### Usage

```console
./wavemaker --interval 1m --duration 1m --resources cpu=100m,memory=100Mi --count 100
```