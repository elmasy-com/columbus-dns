# columbus-dns

A DNS server to collect domains to Columbus Server.

The goal of `columbus-dns` is to make it easy/possible for users to easily  contribute to the Columbus Database.

By setting the system's DNS servers to `columbus-dns` servers while enumerationg subdomains, hunting bugs, etc  you can contribute to the Database.

## Design

```
                               _________________
  |---> dig exmaple.com A ---->| COLUMBUS-DNS  | -----> /insert/example.com -----> columbus.elmasy.com
  |                            -----------------
  ^                                   |
  |                                   V
Alice <------- 93.184.216.34 <---------

```

- Only domains with valid answer will be sent to the server (eg.: not `NXDOMAIN`)

# IMPORTANT!

**THIS SERVER IS NOT MEANT TO USED AS A DAILY DNS RESOLVER!** 

The IP will be not logged, but the domain will!
