# naptr_loadbalance

A CoreDNS plugin that shuffles NAPTR records in DNS responses for round-robin load balancing.

## Why

CoreDNS's built-in `loadbalance` plugin shuffles A, AAAA, and MX records but ignores NAPTR records. In 3GPP/telecom environments, NAPTR records are used for service discovery (e.g., PGW selection), and returning them in a fixed order means clients always hit the same node.

This plugin fills that gap by shuffling NAPTR records per query, distributing traffic across multiple targets.

## Syntax

```
naptr_loadbalance [single]
```

- **No argument**: Shuffles all NAPTR records in the response. All records are returned in randomized order.
- **`single`**: Shuffles and returns only **one** NAPTR record per response. This is useful when DNS caching prevents effective load balancing — each query gets exactly one target, so cached responses still distribute traffic.

## Examples

### Shuffle all NAPTRs (default)

```
example.org {
    file db.example.org
    naptr_loadbalance
}
```

Query returns all NAPTR records in randomized order:

```
;; ANSWER SECTION:
*.telnyx.apn.epc...  NAPTR  20 100 "a" "x-3gpp-pgw:..." node03.epc...
*.telnyx.apn.epc...  NAPTR  20 100 "a" "x-3gpp-pgw:..." node01.epc...
```

### Single mode — one NAPTR per response

```
example.org {
    file db.example.org
    naptr_loadbalance single
}
```

Each query returns only one NAPTR, alternating between targets:

```
Query 1 → node01.epc...
Query 2 → node03.epc...
Query 3 → node01.epc...
```

## Building CoreDNS with this plugin

Add the following to CoreDNS `plugin.cfg`:

```
naptr_loadbalance:github.com/sjagannath05/coredns-naptr-loadbalance
```

Then build:

```bash
go generate && go build
```

## How it works

The plugin wraps the DNS ResponseWriter and intercepts outgoing responses. For each response:

1. Separates NAPTR records from non-NAPTR records
2. Shuffles NAPTR records using the same Fisher-Yates algorithm as CoreDNS's built-in `loadbalance` plugin
3. In `single` mode, keeps only the first record after shuffling
4. Reassembles the response and sends it

Non-NAPTR records (A, AAAA, SOA, NS, etc.) pass through untouched. If the response contains fewer than 2 NAPTR records, the plugin is a no-op.

## Use case: 3GPP DNS

In 3GPP networks, NAPTR records map APNs to PGW/GGSN nodes:

```
*.telnyx  IN  NAPTR  20 100 "a" "x-3gpp-pgw:x-s5-gtp:x-s8-gtp:x-gn:x-gp" "" node01.epc.mnc210.mcc311.3gppnetwork.org.
*.telnyx  IN  NAPTR  20 100 "a" "x-3gpp-pgw:x-s5-gtp:x-s8-gtp:x-gn:x-gp" "" node03.epc.mnc210.mcc311.3gppnetwork.org.
```

Without this plugin, CoreDNS always returns `node01` first. With `naptr_loadbalance single`, each DNS query gets one node, alternating — effectively load balancing PGW selection at the DNS layer.
