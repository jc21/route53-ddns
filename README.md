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

### Installing

```bash
go get github.com/jc21/route53-ddns
```

or build with:

```bash
git clone https://github.com/jc21/route53-ddns && cd route53-ddns
make
./bin/route53-ddns -h
```

