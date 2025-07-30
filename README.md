# Route53 DDNS

- Sick of using Dynamic DNS services that are costly or annoying?
- Already use Route53 for your domain DNS?
- Do you have a dynamically assigned IP address?

This is the command for you.

```bash
Update route53 DNS record with your current IP address
Usage: route53-ddns [--setup] [--configfile CONFIGFILE]

Options:
  --setup, -s            Setup wizard
  --configfile CONFIGFILE, -c CONFIGFILE
                         Config File to use (default: ~/.aws/route53-ddns.json)
  --help, -h             display this help and exit
  --version              display version and exit
```

## Install

#### RHEL based distributions

RPM hosted on [yum.jc21.com](https://yum.jc21.com)

#### Go install

```bash
go install github.com/jc21/route53-ddns@latest
```

#### Building

```bash
git clone https://github.com/jc21/route53-ddns && cd route53-ddns
go build -o bin/route53-ddns cmd/route53-ddns/main.go
./bin/route53-ddns -h
```
