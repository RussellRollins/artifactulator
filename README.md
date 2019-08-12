# artifactulator

An Artifactory stressing tool.

Extremely Work in Progress

## Install

Install using Go (go 1.12 required).

```
cd cmd/artifactulator
go install
```

## Run

Export required env vars using envchain

```
envchain --set --noecho default ARTIFACTORY_HOST
envchain --set --noecho default ARTIFACTORY_USER
envchain --set --noecho default ARTIFACTORY_TOKEN
```

Put load on the Artifactory instance.

```
envchain default artifactulator stress --repo="test"
```

## More Info

```
artifactulator --help
```
