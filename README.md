# ctrreschk - tooling to inspect resources assigned to containers/processes

`ctrreschk` is the improved successor of [numalign](https://github.com/ffromani/numalign).
It incorporate lessons learned over the years, builds on [better infrastructure](https://github.com/jaypipes/ghw),
provides more information.

## license
(C) 2024 Red Hat Inc and licensed under the Apache License v2

## build
just run
```bash
make build
```

## container images

pre-built container images are hosted on [quay.io](https://quay.io/repository/fromani/ctrreschk)

## usage

`ctrreschk` is meant to be used as container entry point and performs alignment check.

In general a resource is deemed "aligned" if all its items are taken from the same pool.
So, if more than a single resource pool is needed, then by definition it cannot be aligned.
Example of resource pools:

- virtual cpus (Hyperthreading, SMT) are taken from physical CPUs. Usually the pool size is 2.
- CPUs (virtual or physical) are taken from a Uncore cache pool, because multiple cpus use the same
  uncore cache block. Usually the pool size is in the range 8-16.
- CPUs (virtual or physical) are taken from a NUMA node, because multiple cpus are physically part
  of a NUMA node. Usually the pool size is in the range 8-256.

### online help

```bash
$ ./_out/ctrreschk 
Usage:
  ctrreschk [flags]
  ctrreschk [command]

Available Commands:
  align       show resource alignment properties
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  info        show machine properties
  k8s         show properties in kubernetes format

Flags:
  -h, --help           help for ctrreschk
  -S, --sleepforever   run and sleep forever after executing the command
  -v, --verbose int    log verbosity

Use "ctrreschk [command] --help" for more information about a command.
```

here `align` is the most important subcommand which performs the alignment checks

```bash
$ ./_out/ctrreschk align -h
show resource alignment properties

Usage:
  ctrreschk align [flags]

Flags:
  -h, --help                 help for align
  -M, --machinedata string   read fake machine data from path, don't read real data from the system

Global Flags:
  -S, --sleepforever   run and sleep forever after executing the command
  -v, --verbose int    log verbosity
```

example output on developer's laptop

```bash
$ ./_out/ctrreschk align | jq .
{
  "alignment": {         <- terse summary
    "smt": true,         <- all CPUs the container is using are SMT aligned
    "llc": true,         <- ditto for LLC (uncore)
    "numa": true         <- ditto for NUMA
  },
  "aligned": {          <- breakdown of aligned resources
    "llc": {            <- LLC cpus
      "0": {            <- CPUs pertaining to LLC/Uncore block "0"
        "cpus": [
          0,
          1,
          2,
          3,
          4,
          5,
          6,
          7
        ]
      }
    },
    "numa": {           <- NUMA cpus
      "0": {            <- CPUs pertaining to NUMA node "0"
        "cpus": [
          0,
          1,
          2,
          3,
          4,
          5,
          6,
          7
        ]
      }
    }
  }
}
```

## APIs

the "API" definition represent the tool output in such a way which is standardized and easily
consumable by third party tools. It's a formalization of the output contract of the tool.
The API and therefore the output contract is not final yet and still subject to change (hence v0).
There is no provision for an API contract for the input, but may be added in the future.
